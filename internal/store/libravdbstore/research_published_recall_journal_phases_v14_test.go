package libravdbstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type publishedRecallLoadStoreV14 struct {
	store.EventStore
	gate     *incrementalSortGate
	owner    sync.Mutex
	current  atomic.Pointer[publishedRecallViewV8]
	searches atomic.Int64
	stales   atomic.Int64
}

type recallPhaseKeyV14 struct{}

type recallPhaseV14 struct {
	searchNS          atomic.Int64
	graphNS           atomic.Int64
	certificateNS     atomic.Int64
	journalWaitNS     atomic.Int64
	journalAppendNS   atomic.Int64
	journalPublishNS  atomic.Int64
	gateMutexWaitNS   atomic.Int64
	markerReadNS      atomic.Int64
	beforeLSNNS       atomic.Int64
	beforeLSNValue    atomic.Uint64
	journalWriteNS    atomic.Int64
	afterLSNNS        atomic.Int64
	afterLSNValue     atomic.Uint64
	journalReadbackNS atomic.Int64
	markerCommitNS    atomic.Int64
	residualCalls     atomic.Int64
	residualMu        sync.Mutex
	residualStart     time.Time
	residualEnd       time.Time
}

func phaseV14(ctx context.Context) *recallPhaseV14 {
	phase, _ := ctx.Value(recallPhaseKeyV14{}).(*recallPhaseV14)
	return phase
}

func (p *recallPhaseV14) residualSpan() time.Duration {
	p.residualMu.Lock()
	defer p.residualMu.Unlock()
	if p.residualStart.IsZero() || p.residualEnd.IsZero() {
		return 0
	}
	return p.residualEnd.Sub(p.residualStart)
}

// This is a test-only timed copy of appendJournal; its transition checks
// stay identical so the phase probe measures the durable path itself.
func (g *incrementalSortGate) appendJournalPhasesV14(ctx context.Context, entry model.BayesianJournalEntry, stopAt string, phase *recallPhaseV14) error {
	started := time.Now()
	g.mu.Lock()
	defer g.mu.Unlock()
	if phase != nil {
		phase.gateMutexWaitNS.Store(time.Since(started).Nanoseconds())
	}
	started = time.Now()
	marker, err := g.readMarker(ctx)
	if phase != nil {
		phase.markerReadNS.Store(time.Since(started).Nanoseconds())
	}
	if err != nil {
		return err
	}
	if g.verified == "" || marker.phase != "READY" || marker.digest != g.verified ||
		marker.count != g.wantRows || marker.lsn != g.verifiedLSN {
		return errors.New("frontier journal gate is not current")
	}
	started = time.Now()
	g.store.writeMu.RLock()
	beforeSnapshot := g.store.snapshot
	beforeLSN, err := g.store.db.LatestCommitLSN(ctx)
	g.store.writeMu.RUnlock()
	if phase != nil {
		phase.beforeLSNNS.Store(time.Since(started).Nanoseconds())
		phase.beforeLSNValue.Store(beforeLSN)
	}
	if err != nil {
		g.verified = ""
		return err
	}
	if beforeLSN != marker.lsn {
		g.verified = ""
		return fmt.Errorf("LibraVDB moved outside frontier journal gate: %d != %d", beforeLSN, marker.lsn)
	}
	started = time.Now()
	writeErr := g.store.PutBayesianJournal(ctx, entry)
	if phase != nil {
		phase.journalWriteNS.Store(time.Since(started).Nanoseconds())
	}
	if writeErr != nil {
		g.verified = ""
		return writeErr
	}
	started = time.Now()
	g.store.writeMu.RLock()
	afterSnapshot := g.store.snapshot
	afterLSN, err := g.store.db.LatestCommitLSN(ctx)
	g.store.writeMu.RUnlock()
	if phase != nil {
		phase.afterLSNNS.Store(time.Since(started).Nanoseconds())
		phase.afterLSNValue.Store(afterLSN)
	}
	if err != nil || afterLSN < beforeLSN || afterSnapshot != beforeSnapshot {
		g.verified = ""
		return fmt.Errorf("frontier journal moved EventFrame state: LSN %d -> %d: %v", beforeLSN, afterLSN, err)
	}
	started = time.Now()
	saved, err := g.store.GetBayesianJournal(ctx, entry.TenantID, entry.ID)
	if err != nil {
		g.verified = ""
		return err
	}
	wantJSON, wantErr := json.Marshal(entry)
	gotJSON, gotErr := json.Marshal(saved)
	if phase != nil {
		phase.journalReadbackNS.Store(time.Since(started).Nanoseconds())
	}
	if wantErr != nil || gotErr != nil || !bytes.Equal(wantJSON, gotJSON) {
		g.verified = ""
		return fmt.Errorf("frontier journal readback mismatch: want=%v got=%v", wantErr, gotErr)
	}
	if afterLSN == beforeLSN {
		return nil
	}
	if stopAt == "after_db" {
		return errors.New("injected interruption after frontier journal DB commit")
	}
	started = time.Now()
	tx, err := g.sidecar.BeginTx(ctx, nil)
	if err != nil {
		g.verified = ""
		return err
	}
	defer tx.Rollback()
	updated, err := tx.ExecContext(ctx,
		"UPDATE marker SET lsn=? WHERE id=1 AND phase='READY' AND lsn=? AND row_count=? AND digest=?",
		afterLSN, marker.lsn, marker.count, marker.digest)
	if err != nil {
		g.verified = ""
		return err
	}
	if affected, err := updated.RowsAffected(); err != nil || affected != 1 {
		g.verified = ""
		return fmt.Errorf("frontier journal marker update affected %d rows: %v", affected, err)
	}
	if err := tx.Commit(); err != nil {
		g.verified = ""
		return err
	}
	if phase != nil {
		phase.markerCommitNS.Store(time.Since(started).Nanoseconds())
	}
	if stopAt == "after_sqlite" {
		return errors.New("injected interruption after frontier journal SQLite commit")
	}
	g.verifiedLSN = afterLSN
	return nil
}

func (s *publishedRecallLoadStoreV14) publishLocked(ctx context.Context) error {
	lsn, ready := s.gate.capture(ctx)
	if !ready {
		s.current.Store(nil)
		return fmt.Errorf("journaled EventFrame view is not READY")
	}
	s.current.Store(&publishedRecallViewV8{lsn: lsn, snapshot: s.gate.store.Snapshot(ctx), at: time.Now()})
	return nil
}

func (s *publishedRecallLoadStoreV14) appendBatch(ctx context.Context, writes []ResearchEventWrite) (model.Snapshot, error) {
	s.owner.Lock()
	defer s.owner.Unlock()
	if err := appendDenseV6(ctx, s.gate, writes); err != nil {
		s.current.Store(nil)
		return model.Snapshot{}, err
	}
	if err := s.publishLocked(ctx); err != nil {
		return model.Snapshot{}, err
	}
	return s.current.Load().snapshot, nil
}

func (s *publishedRecallLoadStoreV14) Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
	started := time.Now()
	if phase := phaseV14(ctx); phase != nil {
		defer func() { phase.searchNS.Store(time.Since(started).Nanoseconds()) }()
	}
	if tenantID != "tenant-a" {
		return nil, fmt.Errorf("unexpected tenant %q", tenantID)
	}
	pin, _ := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8)
	if pin == nil {
		return nil, fmt.Errorf("Recall has no per-call published view pin")
	}
	view := s.current.Load()
	if view == nil || time.Since(view.at) >= 250*time.Millisecond {
		return nil, fmt.Errorf("published view unavailable or expired")
	}
	results, err := queryDensePublishedV6(ctx, s.gate, view.lsn, availableBy, vector, limit)
	if err != nil {
		return nil, err
	}
	if time.Since(view.at) >= 250*time.Millisecond || s.current.Load() == nil {
		return nil, fmt.Errorf("published view expired during Search")
	}
	pin.view = view
	s.searches.Add(1)
	return results, nil
}

func (s *publishedRecallLoadStoreV14) GetPredictiveGraph(ctx context.Context, tenantID string) (model.PredictiveGraph, error) {
	started := time.Now()
	graph, err := s.EventStore.GetPredictiveGraph(ctx, tenantID)
	if phase := phaseV14(ctx); phase != nil {
		phase.graphNS.Store(time.Since(started).Nanoseconds())
	}
	return graph, err
}

func (s *publishedRecallLoadStoreV14) GetSelectionCertificate(ctx context.Context, tenantID string) (model.SelectionSupportCertificate, error) {
	started := time.Now()
	certificate, err := s.EventStore.GetSelectionCertificate(ctx, tenantID)
	if phase := phaseV14(ctx); phase != nil {
		phase.certificateNS.Add(time.Since(started).Nanoseconds())
	}
	return certificate, err
}

func (s *publishedRecallLoadStoreV14) GetOmittedInfluenceCertificate(ctx context.Context, tenantID string) (model.OmittedInfluenceCertificate, error) {
	started := time.Now()
	certificate, err := s.EventStore.GetOmittedInfluenceCertificate(ctx, tenantID)
	if phase := phaseV14(ctx); phase != nil {
		phase.certificateNS.Add(time.Since(started).Nanoseconds())
	}
	return certificate, err
}

func (s *publishedRecallLoadStoreV14) GetResidualCandidates(ctx context.Context, tenantID, actionKey, generalKey string) (model.ResidualCandidates, error) {
	phase := phaseV14(ctx)
	if phase == nil {
		return s.EventStore.GetResidualCandidates(ctx, tenantID, actionKey, generalKey)
	}
	started := time.Now()
	phase.residualMu.Lock()
	if phase.residualStart.IsZero() || started.Before(phase.residualStart) {
		phase.residualStart = started
	}
	phase.residualMu.Unlock()
	candidates, err := s.EventStore.GetResidualCandidates(ctx, tenantID, actionKey, generalKey)
	ended := time.Now()
	phase.residualMu.Lock()
	if ended.After(phase.residualEnd) {
		phase.residualEnd = ended
	}
	phase.residualMu.Unlock()
	phase.residualCalls.Add(1)
	return candidates, err
}

func (s *publishedRecallLoadStoreV14) Snapshot(ctx context.Context) model.Snapshot {
	if pin, ok := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8); ok && pin.view != nil {
		return pin.view.snapshot
	}
	return s.gate.store.Snapshot(ctx)
}

func (s *publishedRecallLoadStoreV14) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	phase := phaseV14(ctx)
	waitStarted := time.Now()
	s.owner.Lock()
	defer s.owner.Unlock()
	if phase != nil {
		phase.journalWaitNS.Store(time.Since(waitStarted).Nanoseconds())
	}
	current := s.gate.store.Snapshot(ctx)
	if !store.JournalSnapshotCompatible(entry.Snapshot, current, entry.AsOf, s.gate.store.ingestMotion) {
		s.stales.Add(1)
		return store.ErrStaleSnapshot
	}
	appendStarted := time.Now()
	appendErr := s.gate.appendJournalPhasesV14(ctx, entry, "", phase)
	if phase != nil {
		phase.journalAppendNS.Store(time.Since(appendStarted).Nanoseconds())
	}
	if appendErr != nil {
		s.current.Store(nil)
		return appendErr
	}
	publishStarted := time.Now()
	err := s.publishLocked(ctx)
	if phase != nil {
		phase.journalPublishNS.Store(time.Since(publishStarted).Nanoseconds())
	}
	return err
}

func (s *publishedRecallLoadStoreV14) Close() error { return nil }

type recallLoadOfferV14 struct {
	index   int
	offered time.Time
}

type recallLoadResultV14 struct {
	index                   int
	offered, started, ended time.Time
	admissionWait           time.Duration
	view                    *publishedRecallViewV8
	packet                  model.ContextPacket
	phase                   *recallPhaseV14
	err                     error
}

type recallLoadAckV14 struct {
	id      string
	offered time.Time
	acked   time.Time
	version uint64
}

type recallLoadWriterV14 struct {
	acks           []recallLoadAckV14
	groups         []int
	leaseWaitNanos []int64
	err            error
}

func runPublishedRecallLoadV14(t *testing.T, raceCorrectnessOnly bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	query := denseQueryV6()
	rawQuery := make([]float32, len(query))
	for i, value := range query {
		rawQuery[i] = 2 * value
	}
	gate := createUnpublishedDenseGateV8(t, t.TempDir(), query)
	defer gate.close()
	adapter := &publishedRecallLoadStoreV14{EventStore: gate.store, gate: gate}
	var admission sync.RWMutex
	em, err := embed.NewHashEmbedder(256)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := service.New(adapter, em, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer svc.Close()
	if err := gate.publish(ctx); err != nil {
		t.Fatal(err)
	}
	if err := gate.initJournal(ctx); err != nil {
		t.Fatal(err)
	}
	if err := adapter.publishLocked(ctx); err != nil {
		t.Fatal(err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	allRows := []denseOracleRowV7{{"past100", 0}, {"at120", 0}}
	for start := 0; start < 198; start += 16 {
		batch := make([]ResearchEventWrite, 0, min(16, 198-start))
		for i := start; i < min(198, start+16); i++ {
			angle := 0.005 * float64(i+1)
			id := fmt.Sprintf("eligible-%03d", i)
			write := pinnedWrite(id, second.Add(100*time.Millisecond))
			write.Vector = denseRowV6(query, i, angle)
			batch = append(batch, write)
			allRows = append(allRows, denseOracleRowV7{id, angle})
		}
		if _, err := adapter.appendBatch(ctx, batch); err != nil {
			t.Fatal("seed eligible", err)
		}
	}
	for start := 0; start < 16; start += 16 {
		batch := make([]ResearchEventWrite, 0, 16)
		for i := start; i < start+16; i++ {
			write := pinnedWrite(fmt.Sprintf("future-%03d", i), asOf.Add(5*time.Millisecond))
			write.Vector = denseRowV6(query, i, 0.001*float64(i+1))
			batch = append(batch, write)
		}
		if _, err := adapter.appendBatch(ctx, batch); err != nil {
			t.Fatal("seed future", err)
		}
	}
	seedVersion := adapter.current.Load().snapshot.RuntimeVersion
	oracleByVersion := map[uint64][]denseOracleRowV7{seedVersion: sortedDenseOracleV7(allRows)}
	writeJobs := make(chan sortLoadWriteJob, 128)
	readJobs := make(chan recallLoadOfferV14, 128)
	writeOffers, readOffers := make(chan []time.Time, 1), make(chan []time.Time, 1)
	writerDone := make(chan recallLoadWriterV14, 1)
	readReports := make(chan recallLoadResultV14, 128)
	start := make(chan struct{})
	go func() {
		<-start
		defer close(writeJobs)
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		times := make([]time.Time, 0, 128)
		defer func() { writeOffers <- times }()
		for i := 0; i < 128; i++ {
			if i > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
			offered := time.Now()
			times = append(times, offered)
			write := pinnedWrite(fmt.Sprintf("visible-%03d", i), asOf)
			write.Vector = denseRowV6(query, i, 0.00213*float64(i+1))
			select {
			case writeJobs <- sortLoadWriteJob{write: write, offered: offered}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		<-start
		defer close(readJobs)
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		times := make([]time.Time, 0, 128)
		defer func() { readOffers <- times }()
		for i := 0; i < 128; i++ {
			if i > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
			offered := time.Now()
			times = append(times, offered)
			select {
			case readJobs <- recallLoadOfferV14{index: i, offered: offered}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		<-start
		report := recallLoadWriterV14{}
		for {
			group, closed := sortLoadTakeGroup(writeJobs, 16, 16*time.Millisecond)
			if len(group) == 0 {
				writerDone <- report
				return
			}
			batch := make([]ResearchEventWrite, len(group))
			for i, job := range group {
				batch[i] = job.write
			}
			waiting := time.Now()
			admission.Lock()
			leaseWait := time.Since(waiting)
			snapshot, err := adapter.appendBatch(ctx, batch)
			admission.Unlock()
			report.leaseWaitNanos = append(report.leaseWaitNanos, leaseWait.Nanoseconds())
			if err != nil {
				report.err = err
				writerDone <- report
				return
			}
			for _, job := range group {
				var index int
				if _, err := fmt.Sscanf(strings.TrimPrefix(job.write.Event.ID, "visible-"), "%d", &index); err != nil {
					report.err = err
					writerDone <- report
					return
				}
				allRows = append(allRows, denseOracleRowV7{job.write.Event.ID, 0.00213 * float64(index+1)})
			}
			oracleByVersion[snapshot.RuntimeVersion] = sortedDenseOracleV7(allRows)
			acked := time.Now()
			report.groups = append(report.groups, len(group))
			for _, job := range group {
				report.acks = append(report.acks, recallLoadAckV14{job.write.Event.ID, job.offered, acked, snapshot.RuntimeVersion})
			}
			if closed {
				writerDone <- report
				return
			}
		}
	}()
	var readers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			<-start
			for offer := range readJobs {
				began := time.Now()
				pin := &publishedRecallPinV8{}
				phase := &recallPhaseV14{}
				callCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
				callCtx = context.WithValue(callCtx, recallPhaseKeyV14{}, phase)
				request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion,
					TenantID: "tenant-a", SessionID: fmt.Sprintf("lab-%03d", offer.index),
					Query: "public vector search fixture", Embedding: rawQuery, EmbeddingModel: em.ModelKey(),
					AsOf: asOf, RecallK: 50, PackK: 10, TokenBudget: 10000}
				waiting := time.Now()
				admission.RLock()
				leaseWait := time.Since(waiting)
				packet, err := svc.Recall(callCtx, request)
				admission.RUnlock()
				readReports <- recallLoadResultV14{index: offer.index, offered: offer.offered, started: began,
					ended: time.Now(), admissionWait: leaseWait, view: pin.view, packet: packet, phase: phase, err: err}
			}
		}()
	}
	go func() { readers.Wait(); close(readReports) }()
	close(start)
	reports := make([]recallLoadResultV14, 0, 128)
	callNS, offerNS, viewAgeNS := make([]int64, 0, 128), make([]int64, 0, 128), make([]int64, 0, 128)
	readLeaseNS := make([]int64, 0, 128)
	for report := range readReports {
		reports = append(reports, report)
		callNS = append(callNS, report.ended.Sub(report.started).Nanoseconds())
		offerNS = append(offerNS, report.ended.Sub(report.offered).Nanoseconds())
		readLeaseNS = append(readLeaseNS, report.admissionWait.Nanoseconds())
		if report.view != nil {
			viewAgeNS = append(viewAgeNS, report.ended.Sub(report.view.at).Nanoseconds())
		}
	}
	writer := <-writerDone
	writeTimes, readTimes := <-writeOffers, <-readOffers
	violations := make([]string, 0)
	if writer.err != nil {
		violations = append(violations, "writer: "+writer.err.Error())
	}
	if len(writer.acks) != 128 || len(reports) != 128 || len(writeTimes) != 128 || len(readTimes) != 128 {
		violations = append(violations, fmt.Sprintf("counts ack=%d recall=%d write_offer=%d read_offer=%d", len(writer.acks), len(reports), len(writeTimes), len(readTimes)))
	}
	for _, report := range reports {
		if report.err != nil {
			violations = append(violations, fmt.Sprintf("recall %d: %v", report.index, report.err))
			continue
		}
		if report.view == nil || report.packet.Snapshot != report.view.snapshot {
			violations = append(violations, fmt.Sprintf("recall %d lost per-call pinned snapshot", report.index))
			continue
		}
		journal, err := gate.store.GetBayesianJournal(ctx, "tenant-a", report.packet.BayesianShadow.JournalID)
		if err != nil || journal.Snapshot != report.packet.Snapshot || journal.Report.JournalID != report.packet.BayesianShadow.JournalID {
			violations = append(violations, fmt.Sprintf("recall %d missing/mismatched durable journal: %v", report.index, err))
			continue
		}
		expected, ok := oracleByVersion[report.packet.Snapshot.RuntimeVersion]
		if !ok || len(expected) < 150 || report.packet.Recalled != 150 || len(journal.Report.Decisions) != 150 {
			violations = append(violations, fmt.Sprintf("recall %d missing 150-row version oracle/frontier", report.index))
			continue
		}
		want := make(map[string]bool, 150)
		for _, row := range expected[:150] {
			want[row.id] = true
		}
		for _, decision := range journal.Report.Decisions {
			if !want[decision.EventID] || strings.HasPrefix(decision.EventID, "future") {
				violations = append(violations, fmt.Sprintf("recall %d bad nominated event %s", report.index, decision.EventID))
			}
			delete(want, decision.EventID)
		}
		if len(want) != 0 {
			violations = append(violations, fmt.Sprintf("recall %d omitted %d oracle nominees", report.index, len(want)))
		}
		for _, candidate := range report.packet.Candidates {
			if candidate.Event.AvailableAt.After(asOf) || strings.HasPrefix(candidate.Event.ID, "future") {
				violations = append(violations, fmt.Sprintf("recall %d packed future event %s", report.index, candidate.Event.ID))
			}
		}
		latestBeforeOffer := seedVersion
		for _, ack := range writer.acks {
			if !ack.acked.After(report.offered) && ack.version > latestBeforeOffer {
				latestBeforeOffer = ack.version
			}
		}
		if report.packet.Snapshot.RuntimeVersion < latestBeforeOffer {
			violations = append(violations, fmt.Sprintf("recall %d omitted an acknowledged-before-offer version", report.index))
		}
	}
	final := adapter.current.Load()
	if final == nil || final.snapshot.RuntimeVersion != seedVersion+128 {
		violations = append(violations, "final published snapshot did not include all visible writes")
	}
	if _, ready := gate.capture(ctx); !ready {
		violations = append(violations, "final journal publication gate is not READY")
	}
	if stale := adapter.stales.Load(); stale != 0 {
		violations = append(violations, fmt.Sprintf("read admission still allowed %d stale-snapshot rejections", stale))
	}
	writeAgeNS := make([]int64, 0, len(writer.acks))
	for _, ack := range writer.acks {
		writeAgeNS = append(writeAgeNS, ack.acked.Sub(ack.offered).Nanoseconds())
	}
	writeP99 := pinnedPercentile(writeAgeNS, .99)
	callP99 := pinnedPercentile(callNS, .99)
	offerP99 := pinnedPercentile(offerNS, .99)
	maxViewAge := pinnedPercentile(viewAgeNS, 1)
	phaseTimes := make(map[string][]int64)
	for _, report := range reports {
		if report.err != nil || report.phase == nil {
			continue
		}
		phase := report.phase
		residualWall := phase.residualSpan().Nanoseconds()
		if phase.residualCalls.Load() != 150 || residualWall <= 0 || phase.searchNS.Load() <= 0 ||
			phase.graphNS.Load() <= 0 || phase.certificateNS.Load() <= 0 ||
			phase.journalAppendNS.Load() <= 0 || phase.journalPublishNS.Load() <= 0 ||
			phase.markerReadNS.Load() <= 0 ||
			phase.beforeLSNNS.Load() <= 0 || phase.journalWriteNS.Load() <= 0 ||
			phase.afterLSNNS.Load() <= 0 || phase.journalReadbackNS.Load() <= 0 ||
			phase.markerCommitNS.Load() <= 0 {
			violations = append(violations, fmt.Sprintf("recall %d has incomplete phase coverage: residual_calls=%d before_lsn=%d after_lsn=%d marker_commit_ns=%d write_ns=%d readback_ns=%d", report.index,
				phase.residualCalls.Load(), phase.beforeLSNValue.Load(), phase.afterLSNValue.Load(), phase.markerCommitNS.Load(), phase.journalWriteNS.Load(), phase.journalReadbackNS.Load()))
			continue
		}
		measured := phase.searchNS.Load() + phase.graphNS.Load() + phase.certificateNS.Load() +
			residualWall + phase.journalWaitNS.Load() + phase.journalAppendNS.Load() + phase.journalPublishNS.Load()
		other := report.ended.Sub(report.started).Nanoseconds() - report.admissionWait.Nanoseconds() - measured
		if other < 0 {
			violations = append(violations, fmt.Sprintf("recall %d has overlapping phase spans", report.index))
			continue
		}
		gateMeasured := phase.gateMutexWaitNS.Load() + phase.markerReadNS.Load() + phase.beforeLSNNS.Load() +
			phase.journalWriteNS.Load() + phase.afterLSNNS.Load() + phase.journalReadbackNS.Load() + phase.markerCommitNS.Load()
		gateOther := phase.journalAppendNS.Load() - gateMeasured
		if gateOther < 0 {
			violations = append(violations, fmt.Sprintf("recall %d has overlapping journal gate spans", report.index))
			continue
		}
		phaseTimes["search"] = append(phaseTimes["search"], phase.searchNS.Load())
		phaseTimes["graph"] = append(phaseTimes["graph"], phase.graphNS.Load())
		phaseTimes["certificates"] = append(phaseTimes["certificates"], phase.certificateNS.Load())
		phaseTimes["residual_wall"] = append(phaseTimes["residual_wall"], residualWall)
		phaseTimes["journal_wait"] = append(phaseTimes["journal_wait"], phase.journalWaitNS.Load())
		phaseTimes["journal_append"] = append(phaseTimes["journal_append"], phase.journalAppendNS.Load())
		phaseTimes["gate_mutex_wait"] = append(phaseTimes["gate_mutex_wait"], phase.gateMutexWaitNS.Load())
		phaseTimes["marker_read"] = append(phaseTimes["marker_read"], phase.markerReadNS.Load())
		phaseTimes["before_lsn"] = append(phaseTimes["before_lsn"], phase.beforeLSNNS.Load())
		phaseTimes["journal_write"] = append(phaseTimes["journal_write"], phase.journalWriteNS.Load())
		phaseTimes["after_lsn"] = append(phaseTimes["after_lsn"], phase.afterLSNNS.Load())
		phaseTimes["journal_readback"] = append(phaseTimes["journal_readback"], phase.journalReadbackNS.Load())
		phaseTimes["marker_commit"] = append(phaseTimes["marker_commit"], phase.markerCommitNS.Load())
		phaseTimes["gate_other"] = append(phaseTimes["gate_other"], gateOther)
		phaseTimes["journal_publish"] = append(phaseTimes["journal_publish"], phase.journalPublishNS.Load())
		phaseTimes["other_service"] = append(phaseTimes["other_service"], other)
	}
	for _, name := range []string{"search", "graph", "certificates", "residual_wall", "journal_wait", "journal_append", "gate_mutex_wait", "marker_read", "before_lsn", "journal_write", "after_lsn", "journal_readback", "marker_commit", "gate_other", "journal_publish", "other_service"} {
		values := phaseTimes[name]
		if len(values) != 128 {
			violations = append(violations, fmt.Sprintf("phase %s coverage=%d/128", name, len(values)))
			continue
		}
		t.Logf("phase=%s p50=%s p99=%s", name, pinnedPercentile(values, .5), pinnedPercentile(values, .99))
	}
	if !raceCorrectnessOnly && (writeP99 >= 250*time.Millisecond || callP99 >= 100*time.Millisecond ||
		offerP99 >= 100*time.Millisecond || maxViewAge >= 250*time.Millisecond) {
		violations = append(violations, fmt.Sprintf("normal latency/freshness gate write_p99=%s call_p99=%s offer_p99=%s view_age_max=%s", writeP99, callP99, offerP99, maxViewAge))
	}
	t.Logf("mode_race_correctness_only=%v writes=%d groups=%d recalls=%d violations=%d searches=%d stale_rejects=%d write_gap_p50=%s write_gap_p99=%s read_gap_p50=%s read_gap_p99=%s writer_lease_p50=%s writer_lease_p99=%s reader_lease_p50=%s reader_lease_p99=%s write_age_p50=%s write_age_p99=%s call_p50=%s call_p99=%s offer_p50=%s offer_p99=%s view_age_p50=%s view_age_max=%s",
		raceCorrectnessOnly, len(writer.acks), len(writer.groups), len(reports), len(violations),
		adapter.searches.Load(), adapter.stales.Load(),
		pinnedPercentile(sortLoadOfferGaps(writeTimes), .5), pinnedPercentile(sortLoadOfferGaps(writeTimes), .99),
		pinnedPercentile(sortLoadOfferGaps(readTimes), .5), pinnedPercentile(sortLoadOfferGaps(readTimes), .99),
		pinnedPercentile(writer.leaseWaitNanos, .5), pinnedPercentile(writer.leaseWaitNanos, .99),
		pinnedPercentile(readLeaseNS, .5), pinnedPercentile(readLeaseNS, .99),
		pinnedPercentile(writeAgeNS, .5), writeP99,
		pinnedPercentile(callNS, .5), callP99, pinnedPercentile(offerNS, .5), offerP99,
		pinnedPercentile(viewAgeNS, .5), maxViewAge)
	if len(violations) > 0 {
		for i, violation := range violations {
			if i >= 12 {
				break
			}
			t.Error(violation)
		}
		t.Fatalf("loaded Service Recall failed with %d violations", len(violations))
	}
}

func TestResearchPublishedRecallLoadV14(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V14") != "1" {
		t.Skip("opt-in full Service Recall published-LSN load screen")
	}
	raceCorrectnessOnly := os.Getenv("EVENTFRAME_RACE_CORRECTNESS_ONLY") == "1"
	for trial := 1; trial <= 2; trial++ {
		t.Run(fmt.Sprintf("trial-%d", trial), func(t *testing.T) {
			runPublishedRecallLoadV14(t, raceCorrectnessOnly)
		})
	}
}

func TestResearchJournalPhaseParityV14(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_JOURNAL_PHASE_PARITY_V14") != "1" {
		t.Skip("opt-in timed journal gate parity and interruption controls")
	}
	ctx := context.Background()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC).Add(120 * time.Millisecond)
	type outcome struct {
		failed         bool
		immediateReady bool
		reopenedReady  bool
		journalPresent bool
		retryMovedLSN  bool
		retryFailed    bool
	}
	for _, stopAt := range []string{"", "after_db", "after_sqlite"} {
		var prior *outcome
		for _, timed := range []bool{false, true} {
			root := t.TempDir()
			gate := createIncrementalSortGate(t, root)
			entry := researchSortJournal(gate, "phase-parity-"+stopAt, asOf)
			appendOne := func() error {
				if timed {
					return gate.appendJournalPhasesV14(ctx, entry, stopAt, &recallPhaseV14{})
				}
				return gate.appendJournal(ctx, entry, stopAt)
			}
			result := outcome{failed: appendOne() != nil}
			_, result.immediateReady = gate.capture(ctx)
			if stopAt == "" {
				before, err := gate.readMarker(ctx)
				if err != nil {
					t.Fatal("read duplicate-control marker", err)
				}
				result.retryFailed = appendOne() != nil
				after, err := gate.readMarker(ctx)
				if err != nil {
					t.Fatal("read duplicate-control marker after retry", err)
				}
				result.retryMovedLSN = before.lsn != after.lsn
			}
			if err := gate.close(); err != nil {
				t.Fatal("close parity gate", err)
			}
			gate, err := openIncrementalSortGate(root)
			if err != nil {
				t.Fatal("reopen parity gate", err)
			}
			_, result.reopenedReady = gate.capture(ctx)
			_, err = gate.store.GetBayesianJournal(ctx, entry.TenantID, entry.ID)
			result.journalPresent = err == nil
			if err := gate.close(); err != nil {
				t.Fatal("close reopened parity gate", err)
			}
			if prior == nil {
				prior = &result
			} else if result != *prior {
				t.Fatalf("timed journal transition differs at %q: original=%+v timed=%+v", stopAt, *prior, result)
			}
		}
		if prior == nil || !prior.journalPresent || prior.failed != (stopAt != "") ||
			prior.reopenedReady != (stopAt != "after_db") || prior.immediateReady != (stopAt == "") ||
			prior.retryFailed || prior.retryMovedLSN {
			t.Fatalf("unexpected original journal transition at %q: %+v", stopAt, prior)
		}
	}
}
