package libravdbstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchadmission"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type scheduledWaitV24 struct {
	Kind                      researchadmission.Kind
	Begin, Acquired, Released time.Time
	Error                     string
}
type scheduledWitnessV24 struct {
	*witnessStoreV23
	scheduler     *researchadmission.Scheduler
	mode          bool
	traceMu       sync.Mutex
	trace         []scheduledWaitV24
	beforeJournal func() // test-only handoff barrier, nil in loaded trials
}

// Finish admitted journal durability before releasing its read lease, even
// if the caller cancels. Return cancellation afterward, not a delivered packet.
func (s *scheduledWitnessV24) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.beforeJournal != nil {
		s.beforeJournal()
	}
	if err := s.witnessStoreV23.PutBayesianJournal(context.WithoutCancel(ctx), entry); err != nil {
		return err
	}
	return ctx.Err()
}

func (s *scheduledWitnessV24) acquire(ctx context.Context, kind researchadmission.Kind) (func(), error) {
	row := scheduledWaitV24{Kind: kind, Begin: time.Now().UTC()}
	var lease *researchadmission.Lease
	var err error
	if s.mode {
		lease, err = s.scheduler.Acquire(ctx, kind)
	}
	row.Acquired = time.Now().UTC()
	if err != nil {
		row.Error = err.Error()
		row.Released = row.Acquired
		s.traceMu.Lock()
		s.trace = append(s.trace, row)
		s.traceMu.Unlock()
		return nil, err
	}
	return func() {
		// Record the exclusive/shared interval before releasing admission. The
		// next grant can never overlap the recorded protected work interval.
		row.Released = time.Now().UTC()
		if lease != nil {
			lease.Release()
		}
		s.traceMu.Lock()
		s.trace = append(s.trace, row)
		s.traceMu.Unlock()
	}, nil
}
func (s *scheduledWitnessV24) recall(ctx context.Context, svc *service.Service, request model.RecallRequest) (model.ContextPacket, *publishedRecallViewV8, error) {
	release, err := s.acquire(ctx, researchadmission.Recall)
	if err != nil {
		return model.ContextPacket{}, nil, err
	}
	defer release()
	return s.recallV23(ctx, svc, request)
}
func (s *scheduledWitnessV24) append(ctx context.Context, writes []ResearchEventWrite, interrupt bool) (model.Snapshot, error) {
	release, err := s.acquire(ctx, researchadmission.Ingest)
	if err != nil {
		return model.Snapshot{}, err
	}
	defer release()
	return s.appendV23(ctx, writes, interrupt)
}
func (s *scheduledWitnessV24) ApplyBayesianOutcome(ctx context.Context, request model.BayesianOutcomeRequest, key, parent, digest string, weight float64, change bayes.ChangePolicy, group bayes.GroupPolicy, observation model.ResidualObservation, policy residual.Policy) (store.BayesianOutcomeResult, error) {
	release, err := s.acquire(ctx, researchadmission.Outcome)
	if err != nil {
		return store.BayesianOutcomeResult{}, err
	}
	defer release()
	return s.witnessStoreV23.ApplyBayesianOutcome(ctx, request, key, parent, digest, weight, change, group, observation, policy)
}
func (s *scheduledWitnessV24) Close() error {
	s.scheduler.Close()
	return s.publishedRecallLoadStoreV16.Close()
}

func attachScheduledV24(t *testing.T, f *witnessFixtureV23, mode bool) *scheduledWitnessV24 {
	t.Helper()
	scheduler, err := researchadmission.NewScheduler(8, 512)
	if err != nil {
		t.Fatal(err)
	}
	s := &scheduledWitnessV24{witnessStoreV23: f.adapter, scheduler: scheduler, mode: mode}
	before := f.adapter.gate.store.Snapshot(context.Background())
	// Identical policy binding is idempotent; no source/version retagging.
	svc, err := service.New(s, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		t.Fatal(err)
	}
	if before != f.adapter.gate.store.Snapshot(context.Background()) {
		t.Fatal("attachment changed source state")
	}
	f.svc = svc
	return s
}

type scheduledJournalV24 struct {
	Index    int
	Snapshot model.Snapshot
	SHA256   string
}

func TestResearchScheduledWitnessCancellationV24(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SCHEDULED_WITNESS_V24") != "1" {
		t.Skip("isolated scheduled durable handoff")
	}
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachScheduledV24(t, f, true)
	entered := make(chan struct{})
	done := make(chan error, 1)
	s.beforeJournal = func() { s.owner.Lock(); close(entered) }
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "cancel")); done <- err }()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("handoff did not start")
	}
	cancel()
	select {
	case err := <-done:
		s.owner.Unlock()
		t.Fatal("caller returned before durability terminal", err)
	case <-time.After(10 * time.Millisecond):
	}
	if stats := s.scheduler.Snapshot(); stats.ActiveReaders != 1 || stats.ActiveWriter {
		s.owner.Unlock()
		t.Fatal("premature lease release", stats)
	}
	s.owner.Unlock()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("handoff did not finish")
	}
	if len(s.state.Journals) != 1 || s.scheduler.Snapshot().ActiveReaders != 0 {
		t.Fatal("durable handoff missing or lease leaked")
	}
	if _, ready := s.gate.capture(context.Background()); !ready {
		t.Fatal("marker not ready after canceled handoff")
	}
}

type scheduledTrialV24 struct {
	witnessTrialV23
	Scheduled     bool
	Admission     researchadmission.Stats
	Trace         []scheduledWaitV24
	JournalProofs []scheduledJournalV24
}

func TestResearchScheduledWitnessLifecycleV24(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SCHEDULED_WITNESS_V24") != "1" {
		t.Skip("isolated scheduled native lifecycle")
	}
	ctx := context.Background()
	for _, visible := range []bool{false, true} {
		t.Run(fmt.Sprintf("visible-%v", visible), func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachScheduledV24(t, f, true)
			prime, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if err != nil {
				t.Fatal(err)
			}
			response, err := f.feedback(ctx, prime.BayesianShadow.JournalID, "past100", "first", false)
			if err != nil {
				t.Fatal(err)
			}
			at := f.origin.Add(time.Hour)
			if visible {
				at = f.origin
			}
			w := pinnedWrite("new-scheduled", at)
			w.Vector = denseRowV6(f.query, 7, .0003)
			if _, err := s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
			after, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "after"))
			if err != nil || (witnessBeliefsV23(after) > 0) == visible {
				t.Fatal("wrong scheduled transport", err)
			}
			p, err := s.gate.store.GetBayesianPosterior(ctx, "tenant-a", "past100")
			if err != nil || p.EvidenceEpoch != response.Posterior.EvidenceEpoch {
				t.Fatal("retagged source", err)
			}
			for _, alter := range []string{"query", "vector", "selection", "old"} {
				request := f.request(time.Now().UTC(), "alter")
				switch alter {
				case "query":
					request.Query = "different"
				case "vector":
					request.Embedding = denseRowV6(f.query, 11, .1)
				case "selection":
					request.PackK = 9
				case "old":
					request.AsOf = p.UpdatedAt.Add(-time.Nanosecond)
				}
				other, _, err := s.recall(ctx, f.svc, request)
				if err != nil || witnessBeliefsV23(other) != 0 {
					t.Fatal("altered request accepted", alter, err)
				}
			}
			f.reopen(t)
			s = attachScheduledV24(t, f, true)
			after, _, err = s.recall(ctx, f.svc, f.request(time.Now().UTC(), "reopen"))
			if err != nil || (witnessBeliefsV23(after) > 0) == visible {
				t.Fatal("reopened scheduled transport", err)
			}
		})
	}
	for _, interrupt := range []bool{false, true} {
		t.Run(fmt.Sprintf("gap-%v", interrupt), func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachScheduledV24(t, f, true)
			w := pinnedWrite("unaccounted", f.origin.Add(time.Hour))
			w.Vector = denseRowV6(f.query, 2, .003)
			if interrupt {
				if _, err := s.append(ctx, []ResearchEventWrite{w}, true); err == nil {
					t.Fatal("missing interruption")
				}
			} else {
				if err := appendDenseV6(ctx, s.gate, []ResearchEventWrite{w}); err != nil {
					t.Fatal(err)
				}
			}
			f.reopen(t)
			s = attachScheduledV24(t, f, true)
			if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "gap")); err == nil {
				t.Fatal("unwitnessed backend accepted")
			}
		})
	}
}

func runScheduledWitnessLoadV24(t *testing.T, trial int, scheduled, visible bool) scheduledTrialV24 {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachScheduledV24(t, f, scheduled)
	a := f.adapter
	prime, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
	if err != nil {
		t.Fatal(err)
	}
	var events []string
	for _, d := range prime.BayesianShadow.Decisions {
		if d.Activated {
			events = append(events, d.EventID)
			if len(events) == 16 {
				break
			}
		}
	}
	if len(events) != 16 {
		t.Fatal("prime activation")
	}
	start := time.Now().Add(20 * time.Millisecond)
	jobs := make(chan recallLoadOfferV16, 128)
	writeJobs := make(chan sortLoadWriteJob, 128)
	outcomeJobs := make(chan outcomeOfferV18, 16)
	reads := make(chan witnessReadV23, 128)
	writes := make(chan []witnessWriteV23, 1)
	outcomes := make(chan []witnessOutcomeV23, 1)
	waitUntil := func(at time.Time) bool {
		timer := time.NewTimer(time.Until(at))
		select {
		case <-timer.C:
			return true
		case <-ctx.Done():
			timer.Stop()
			return false
		}
	}
	go func() {
		defer close(jobs)
		for i := 0; i < 128; i++ {
			if !waitUntil(start.Add(time.Duration(i) * 4 * time.Millisecond)) {
				return
			}
			jobs <- recallLoadOfferV16{index: i, offered: time.Now().UTC()}
		}
	}()
	go func() {
		defer close(writeJobs)
		for i := 0; i < 128; i++ {
			if !waitUntil(start.Add(time.Duration(i) * 4 * time.Millisecond)) {
				return
			}
			at := f.origin.Add(time.Hour)
			if visible {
				at = f.origin
			}
			w := pinnedWrite(fmt.Sprintf("write-%03d", i), at)
			w.Vector = denseRowV6(f.query, i, .00213*float64(i+1))
			writeJobs <- sortLoadWriteJob{write: w, offered: time.Now().UTC()}
		}
	}()
	go func() {
		var rows []witnessWriteV23
		for {
			group, closed := sortLoadTakeGroup(writeJobs, 16, 16*time.Millisecond)
			if len(group) == 0 {
				break
			}
			batch := make([]ResearchEventWrite, len(group))
			for j, job := range group {
				batch[j] = job.write
			}
			snap, err := s.append(ctx, batch, false)
			ack := time.Now().UTC()
			for _, job := range group {
				r := witnessWriteV23{Index: len(rows), Offer: job.offered, Ack: ack, Snapshot: snap, Available: job.write.Event.AvailableAt}
				if err != nil {
					r.Error = err.Error()
				}
				rows = append(rows, r)
			}
			if closed || err != nil {
				break
			}
		}
		writes <- rows
	}()
	go func() {
		defer close(outcomeJobs)
		for i, event := range events {
			if !waitUntil(start.Add(20*time.Millisecond + time.Duration(i)*16*time.Millisecond)) {
				return
			}
			offer := time.Now().UTC()
			outcomeJobs <- outcomeOfferV18{index: i, request: f.outcomeRequest(prime.BayesianShadow.JournalID, event, fmt.Sprintf("load-%d", i), i%2 == 0, offer)}
		}
	}()
	go func() {
		var rows []witnessOutcomeV23
		for job := range outcomeJobs {
			response, err := f.svc.ObserveBayesianOutcome(ctx, job.request)
			r := witnessOutcomeV23{Index: job.index, EventID: job.request.EventID, Useful: job.request.Useful, Offer: job.request.AvailableAt, Published: time.Now().UTC(), Response: response, Request: job.request}
			if err != nil {
				r.Error = err.Error()
			}
			rows = append(rows, r)
			if err != nil {
				break
			}
		}
		outcomes <- rows
	}()
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				began := time.Now().UTC()
				request := f.request(job.offered, "load")
				packet, view, err := s.recall(ctx, f.svc, request)
				r := witnessReadV23{Index: job.index, Offer: job.offered, Start: began, End: time.Now().UTC(), Snapshot: packet.Snapshot, Journal: packet.BayesianShadow.JournalID, Decisions: packet.BayesianShadow.Decisions, Packed: packet.Candidates, Request: request, Selection: packet.BayesianShadow.SelectionSupportCertified, Omitted: packet.BayesianShadow.OmittedInfluenceCertified}
				if view != nil {
					r.ViewAt = view.at.UTC()
					r.PinSnapshot = view.snapshot
				}
				if err != nil {
					r.Error = err.Error()
				}
				reads <- r
			}
		}()
	}
	go func() { wg.Wait(); close(reads) }()
	row := scheduledTrialV24{witnessTrialV23: witnessTrialV23{Trial: trial, Enabled: true, Visible: visible, Metrics: map[string]int64{}, Origin: f.origin, InitialSnapshot: prime.Snapshot}, Scheduled: scheduled}
	for r := range reads {
		row.Reads = append(row.Reads, r)
	}
	row.Writes = <-writes
	row.Outcomes = <-outcomes
	row.Admission = s.scheduler.Snapshot()
	s.traceMu.Lock()
	row.Trace = append([]scheduledWaitV24(nil), s.trace...)
	s.traceMu.Unlock()
	row.Transported = a.transports.Load()
	row.WitnessHash = a.state.Hash
	row.FinalSnapshot = a.gate.store.Snapshot(ctx)
	row.FinalWitness = *a.state
	commitRows, err := a.gate.sidecar.QueryContext(ctx, "SELECT seq,payload,prior,digest FROM witness_v23 ORDER BY seq")
	if err != nil {
		t.Fatal(err)
	}
	for commitRows.Next() {
		var c witnessCommitV23
		if err := commitRows.Scan(&c.Seq, &c.Payload, &c.Prior, &c.Digest); err != nil {
			commitRows.Close()
			t.Fatal(err)
		}
		row.Commits = append(row.Commits, c)
	}
	if err := commitRows.Err(); err != nil {
		t.Fatal(err)
	}
	commitRows.Close()
	for _, event := range events {
		p, err := a.gate.store.GetBayesianPosterior(ctx, "tenant-a", event)
		if err != nil {
			row.Errors = append(row.Errors, "missing durable source")
		}
		row.DurableSources = append(row.DurableSources, p)
	}
	if len(row.Reads) != 128 || len(row.Writes) != 128 || len(row.Outcomes) != 16 {
		row.Errors = append(row.Errors, "incomplete offered work")
	}
	var readNS, offerNS, viewNS, writeNS, outcomeNS []int64
	learned := 0
	for _, r := range row.Reads {
		readNS = append(readNS, r.End.Sub(r.Start).Nanoseconds())
		offerNS = append(offerNS, r.End.Sub(r.Offer).Nanoseconds())
		viewNS = append(viewNS, r.End.Sub(r.ViewAt).Nanoseconds())
		if r.Error != "" || r.Journal == "" || len(r.Decisions) != 150 {
			row.Errors = append(row.Errors, "invalid Recall/journal")
		}
		saved, err := a.gate.store.GetBayesianJournal(ctx, "tenant-a", r.Journal)
		got, marshalErr := json.Marshal(saved.Report.Decisions)
		want, _ := json.Marshal(r.Decisions)
		if err != nil || marshalErr != nil || saved.Snapshot != r.Snapshot || !bytes.Equal(got, want) {
			row.Errors = append(row.Errors, "wire journal mismatch")
		}
		h := sha256.Sum256(got)
		row.JournalProofs = append(row.JournalProofs, scheduledJournalV24{r.Index, saved.Snapshot, hex.EncodeToString(h[:])})
		for _, d := range r.Decisions {
			if d.Forecast.BeliefLaw != nil {
				learned++
			}
		}
	}
	for _, w := range row.Writes {
		writeNS = append(writeNS, w.Ack.Sub(w.Offer).Nanoseconds())
		if w.Error != "" {
			row.Errors = append(row.Errors, "write: "+w.Error)
		}
	}
	for _, o := range row.Outcomes {
		outcomeNS = append(outcomeNS, o.Published.Sub(o.Offer).Nanoseconds())
		if o.Error != "" || o.Response.Duplicate {
			row.Errors = append(row.Errors, "outcome: "+o.Error)
		}
	}
	row.Metrics["learned_decisions"] = int64(learned)
	for name, values := range map[string][]int64{"call": readNS, "offer": offerNS, "view": viewNS, "write": writeNS, "outcome": outcomeNS} {
		row.Metrics[name+"_p99_ns"] = pinnedPercentile(values, .99).Nanoseconds()
		row.Metrics[name+"_max_ns"] = pinnedPercentile(values, 1).Nanoseconds()
	}
	if a.poison.Load() || a.state.Head != row.FinalSnapshot {
		row.Errors = append(row.Errors, "invalid witness head")
	}
	if _, ready := a.gate.capture(ctx); !ready {
		row.Errors = append(row.Errors, "marker not READY")
	}
	if row.Admission.ActiveWriter || row.Admission.ActiveReaders != 0 || row.Admission.Queued != ([3]int{}) {
		row.Errors = append(row.Errors, "undrained admission")
	}
	row.Pass = len(row.Errors) == 0 && row.Metrics["call_p99_ns"] < int64(100*time.Millisecond) && row.Metrics["offer_p99_ns"] < int64(100*time.Millisecond) && row.Metrics["write_p99_ns"] < int64(250*time.Millisecond) && row.Metrics["outcome_p99_ns"] < int64(100*time.Millisecond) && row.Metrics["outcome_max_ns"] < int64(250*time.Millisecond) && row.Metrics["view_max_ns"] < int64(250*time.Millisecond)
	if !visible {
		row.Pass = row.Pass && learned > 0 && row.Transported > 0
	} else {
		row.Pass = row.Pass && row.Transported == 0
	}
	return row
}
func TestResearchScheduledWitnessLoadV24(t *testing.T) {
	output := os.Getenv("EVENTFRAME_SCHEDULED_WITNESS_V24_OUTPUT")
	if output == "" {
		t.Skip("exclusive frozen scheduled workload")
	}
	manifest, err := os.ReadFile("../../../research/scheduled-witness-v24/freeze.json")
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(manifest)
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Type": "header", "Trials": 8, "FreezeSHA256": hex.EncodeToString(h[:]), "Time": time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	for trial := 1; trial <= 2; trial++ {
		for _, visible := range []bool{false, true} {
			for _, scheduled := range []bool{false, true} {
				r := runScheduledWitnessLoadV24(t, trial, scheduled, visible)
				if err := enc.Encode(r); err != nil {
					t.Fatal(err)
				}
				if err := f.Sync(); err != nil {
					t.Fatal(err)
				}
				t.Logf("trial=%d scheduled=%v visible=%v pass=%v errors=%v metrics=%v transported=%d", trial, scheduled, visible, r.Pass, r.Errors, r.Metrics, r.Transported)
			}
		}
	}
	if err := enc.Encode(map[string]any{"Type": "footer", "Trials": 8}); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}
