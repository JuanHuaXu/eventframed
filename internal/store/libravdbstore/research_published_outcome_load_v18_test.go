package libravdbstore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	libra "github.com/xDarkicex/libravdb/libravdb"
	_ "modernc.org/sqlite"
)

func createLiveDenseGateV18(t *testing.T, root string, query []float32, origin time.Time) *incrementalSortGate {
	t.Helper()
	ctx := context.Background()
	s, err := Open(Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 256,
		EmbeddingModel: "research:d256", Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "sort-publication.sqlite")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		s.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		s.Close()
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		s.Close()
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{"PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", "PRAGMA busy_timeout=5000"} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			s.Close()
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`CREATE TABLE marker(
		id INTEGER PRIMARY KEY CHECK(id=1), phase TEXT NOT NULL,
		lsn INTEGER NOT NULL, row_count INTEGER NOT NULL, digest TEXT NOT NULL);
		INSERT INTO marker(id,phase,lsn,row_count,digest) VALUES(1,'PENDING',0,0,'')`); err != nil {
		db.Close()
		s.Close()
		t.Fatal(err)
	}
	g := &incrementalSortGate{&sortPublicationGate{store: s, sidecar: db,
		name: collectionName("tenant-a", "research:d256"), wantRows: 3}}
	_, err = s.db.CreateCollection(ctx, g.name, libra.WithDimension(256),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true), libra.WithMetadataSchema(libra.MetadataSchema{
			"available_at": libra.StringField, "available_at_sort": libra.StringField,
			"event_json": libra.StringField, "corpus_text": libra.StringField,
			"raw_content": libra.StringField,
		}))
	if err != nil {
		g.close()
		t.Fatal(err)
	}
	writes := []ResearchEventWrite{
		pinnedWrite("past100", origin),
		pinnedWrite("at120", origin.Add(time.Second)),
		pinnedWrite("future125", origin.Add(time.Hour)),
	}
	for i := range writes {
		writes[i].Vector, err = normalizedVectorV4(denseRowV6(query, i, 0), 256)
		if err != nil {
			g.close()
			t.Fatal(err)
		}
	}
	if _, lsn, err := s.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil || lsn == 0 {
		g.close()
		t.Fatal("live dense genesis", lsn, err)
	}
	return g
}

type outcomeOfferV18 struct {
	request model.BayesianOutcomeRequest
	index   int
}

type outcomeResultV18 struct {
	index              int
	eventID            string
	offered, published time.Time
	view               *publishedRecallViewV8
	response           model.BayesianOutcomeResponse
	err                error
}

type outcomeWriterReportV19 struct {
	recallLoadWriterV16
	refreshNS []int64
}

func runPublishedOutcomeLoadV18(t *testing.T, feedback, raceCorrectnessOnly, refresh, expectLearning bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	query := denseQueryV6()
	rawQuery := make([]float32, len(query))
	for i, value := range query {
		rawQuery[i] = 2 * value
	}
	origin := time.Now().UTC().Add(-10 * time.Second)
	gate := createLiveDenseGateV18(t, t.TempDir(), query, origin)
	defer gate.close()
	base := &publishedRecallLoadStoreV16{EventStore: gate.store, gate: gate,
		journalJobs: make(chan *batchJournalJobV16, 128), workerEnd: make(chan struct{})}
	adapter := &publishedOutcomeStoreV17{publishedRecallLoadStoreV16: base}
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
	go base.runJournalWorker()
	defer svc.Close()
	seedRows := []denseOracleRowV7{{"past100", 0}, {"at120", 0}}
	for start := 0; start < 198; start += 16 {
		batch := make([]ResearchEventWrite, 0, min(16, 198-start))
		for i := start; i < min(198, start+16); i++ {
			id := fmt.Sprintf("eligible-%03d", i)
			angle := 0.005 * float64(i+1)
			write := pinnedWrite(id, origin)
			write.Vector, err = normalizedVectorV4(denseRowV6(query, i, angle), 256)
			if err != nil {
				t.Fatal(err)
			}
			batch = append(batch, write)
			seedRows = append(seedRows, denseOracleRowV7{id, angle})
		}
		if _, lsn, err := gate.store.PutResearchEventBatchSortableReceipt(ctx, batch); err != nil || lsn == 0 {
			t.Fatal("seed eligible", lsn, err)
		}
	}
	for i := 0; i < 16; i++ {
		write := pinnedWrite(fmt.Sprintf("future-%03d", i), origin.Add(time.Hour))
		write.Vector, err = normalizedVectorV4(denseRowV6(query, i, 0.001*float64(i+1)), 256)
		if err != nil {
			t.Fatal(err)
		}
		if _, lsn, err := gate.store.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{write}); err != nil || lsn == 0 {
			t.Fatal("seed future", lsn, err)
		}
	}
	gate.wantRows = 217
	snapshot := gate.store.Snapshot(ctx)
	now := time.Now().UTC()
	selection := model.SelectionSupportCertificate{
		ID: "selection-v18", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion,
		EvidenceEpoch: snapshot.EvidenceEpoch, MinSelectionProbability: .2,
		SimultaneousCoverage: .95, Procedure: "synthetic fixture only", Issuer: "research-fixture",
		ExternalAudit: true, ValidFrom: origin, ValidUntil: now.Add(time.Hour),
	}
	if _, err := gate.store.PublishSelectionCertificate(ctx, selection); err != nil {
		t.Fatal(err)
	}
	omitted := model.OmittedInfluenceCertificate{
		ID: "omitted-v18", TenantID: "tenant-a", PolicyVersion: snapshot.PolicyVersion,
		EvidenceEpoch: snapshot.EvidenceEpoch, DivergenceUCB: .02, DivergenceLimit: .05,
		AuditProbability: 1, SimultaneousCoverage: .95,
		Procedure: "synthetic fixture only", Issuer: "research-fixture",
		ExternalAudit: true, ValidUntil: now.Add(time.Hour),
	}
	if _, err := gate.store.PublishOmittedInfluenceCertificate(ctx, omitted); err != nil {
		t.Fatal(err)
	}
	if err := gate.publish(ctx); err != nil {
		t.Fatal(err)
	}
	if err := gate.initJournal(ctx); err != nil {
		t.Fatal(err)
	}
	if err := base.publishLocked(ctx); err != nil {
		t.Fatal(err)
	}
	seedVersion := base.current.Load().snapshot.RuntimeVersion
	writeJobs := make(chan sortLoadWriteJob, 128)
	readJobs := make(chan recallLoadOfferV16, 128)
	readReports := make(chan recallLoadResultV16, 128)
	writerDone := make(chan outcomeWriterReportV19, 1)
	start := make(chan struct{})
	go func() {
		<-start
		defer close(writeJobs)
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		for i := 0; i < 128; i++ {
			if i > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
			write := pinnedWrite(fmt.Sprintf("visible-%03d", i), origin)
			write.Vector = denseRowV6(query, i, 0.00213*float64(i+1))
			select {
			case writeJobs <- sortLoadWriteJob{write: write, offered: time.Now()}:
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
		for i := 0; i < 128; i++ {
			if i > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
			select {
			case readJobs <- recallLoadOfferV16{index: i, offered: time.Now()}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		<-start
		report := outcomeWriterReportV19{}
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
			adapter.admission.Lock()
			snapshot, err := base.appendBatch(ctx, batch)
			if err == nil && refresh {
				started := time.Now()
				selection, omitted := syntheticCertificatesV19(snapshot, started.UTC())
				snapshot, err = adapter.publishSyntheticCertificatesV19(ctx, selection, omitted)
				report.refreshNS = append(report.refreshNS, time.Since(started).Nanoseconds())
			}
			adapter.admission.Unlock()
			if err != nil {
				report.err = err
				writerDone <- report
				return
			}
			acked := time.Now()
			report.groups = append(report.groups, len(group))
			for _, job := range group {
				report.acks = append(report.acks, recallLoadAckV16{
					id: job.write.Event.ID, offered: job.offered, acked: acked, version: snapshot.RuntimeVersion,
				})
			}
			if closed {
				writerDone <- report
				return
			}
		}
	}()
	var readers sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			<-start
			for offer := range readJobs {
				began := time.Now()
				pin := &publishedRecallPinV8{}
				callCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
				request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion,
					TenantID: "tenant-a", SessionID: fmt.Sprintf("lab-%03d", offer.index),
					Query: "public vector search fixture", Embedding: rawQuery, EmbeddingModel: em.ModelKey(),
					AsOf: offer.offered, RecallK: 50, PackK: 10, TokenBudget: 10000}
				adapter.admission.RLock()
				packet, err := svc.Recall(callCtx, request)
				adapter.admission.RUnlock()
				readReports <- recallLoadResultV16{index: offer.index, offered: offer.offered,
					started: began, ended: time.Now(), view: pin.view, packet: packet, err: err}
			}
		}()
	}
	go func() { readers.Wait(); close(readReports) }()
	feedbackJobs := make(chan outcomeOfferV18, 16)
	feedbackDone := make(chan []outcomeResultV18, 1)
	if feedback {
		go func() {
			results := make([]outcomeResultV18, 0, 16)
			ticker := time.NewTicker(16 * time.Millisecond)
			defer ticker.Stop()
			for job := range feedbackJobs {
				if job.index > 0 {
					select {
					case <-ticker.C:
					case <-ctx.Done():
						feedbackDone <- results
						return
					}
				}
				offered := time.Now().UTC()
				job.request.ObservedAt, job.request.AvailableAt = offered, offered
				response, err := svc.ObserveBayesianOutcome(ctx, job.request)
				results = append(results, outcomeResultV18{index: job.index, eventID: job.request.EventID, offered: offered,
					published: time.Now(), view: base.current.Load(), response: response, err: err})
			}
			feedbackDone <- results
		}()
	}
	close(start)
	reports := make([]recallLoadResultV16, 0, 128)
	feedbackScheduled := false
	for report := range readReports {
		reports = append(reports, report)
		if !feedback || feedbackScheduled || report.err != nil || report.packet.BayesianShadow.JournalID == "" {
			continue
		}
		seen := make(map[string]bool)
		jobs := make([]outcomeOfferV18, 0, 16)
		for _, decision := range report.packet.BayesianShadow.Decisions {
			if !decision.Activated || seen[decision.EventID] || strings.HasPrefix(decision.EventID, "future") {
				continue
			}
			seen[decision.EventID] = true
			jobs = append(jobs, outcomeOfferV18{index: len(jobs), request: model.BayesianOutcomeRequest{
				ProtocolVersion: model.ProtocolVersion, IdempotencyKey: fmt.Sprintf("v18-%03d", len(jobs)),
				TenantID: "tenant-a", JournalID: report.packet.BayesianShadow.JournalID,
				EventID: decision.EventID, Useful: true, Source: model.OutcomeFullStream,
				InclusionProbability: 1,
			}})
			if len(jobs) == 16 {
				break
			}
		}
		if len(jobs) == 16 {
			for _, job := range jobs {
				feedbackJobs <- job
			}
			feedbackScheduled = true
		}
	}
	close(feedbackJobs)
	writer := <-writerDone
	var outcomes []outcomeResultV18
	if feedback {
		outcomes = <-feedbackDone
	}
	violations := make([]string, 0)
	if writer.err != nil {
		violations = append(violations, "writer: "+writer.err.Error())
	}
	if len(writer.acks) != 128 || len(reports) != 128 || feedback && (len(outcomes) != 16 || !feedbackScheduled) {
		violations = append(violations, fmt.Sprintf("counts writes=%d recalls=%d outcomes=%d scheduled=%v",
			len(writer.acks), len(reports), len(outcomes), feedbackScheduled))
	}
	callNS, offerNS, viewAgeNS, writeAgeNS, outcomeAgeNS := make([]int64, 0, 128), make([]int64, 0, 128),
		make([]int64, 0, 128), make([]int64, 0, 128), make([]int64, 0, 16)
	for _, ack := range writer.acks {
		writeAgeNS = append(writeAgeNS, ack.acked.Sub(ack.offered).Nanoseconds())
	}
	learned := 0
	labelledCandidates, certifiedLabelledCandidates, epochMatchedLabelledCandidates, epochAlignedLabelledCandidates := 0, 0, 0, 0
	certifiedRecalls, afterOutcomeRecalls, afterOutcomeCertified := 0, 0, 0
	var firstOutcomePublished time.Time
	if len(outcomes) > 0 {
		firstOutcomePublished = outcomes[0].published
	}
	for _, report := range reports {
		callNS = append(callNS, report.ended.Sub(report.started).Nanoseconds())
		offerNS = append(offerNS, report.ended.Sub(report.offered).Nanoseconds())
		if report.view != nil {
			viewAgeNS = append(viewAgeNS, report.ended.Sub(report.view.at).Nanoseconds())
		}
		if report.err != nil || report.view == nil || report.packet.Snapshot != report.view.snapshot {
			violations = append(violations, fmt.Sprintf("recall %d failed or lost pin: %v", report.index, report.err))
			continue
		}
		certified := report.packet.BayesianShadow.SelectionSupportCertified && report.packet.BayesianShadow.OmittedInfluenceCertified
		if certified {
			certifiedRecalls++
		}
		if !firstOutcomePublished.IsZero() && !report.offered.Before(firstOutcomePublished) {
			afterOutcomeRecalls++
			if certified {
				afterOutcomeCertified++
			}
		}
		journal, err := gate.store.GetBayesianJournal(ctx, "tenant-a", report.packet.BayesianShadow.JournalID)
		if err != nil || journal.Snapshot != report.packet.Snapshot || len(journal.Report.Decisions) != 150 || report.packet.Recalled != 150 {
			violations = append(violations, fmt.Sprintf("recall %d journal/frontier mismatch: %v", report.index, err))
			continue
		}
		rows := append([]denseOracleRowV7(nil), seedRows...)
		for _, ack := range writer.acks {
			if ack.version <= report.packet.Snapshot.RuntimeVersion {
				var index int
				if _, err := fmt.Sscanf(strings.TrimPrefix(ack.id, "visible-"), "%d", &index); err == nil {
					rows = append(rows, denseOracleRowV7{ack.id, 0.00213 * float64(index+1)})
				}
			}
		}
		want := make(map[string]bool, 150)
		for _, row := range sortedDenseOracleV7(rows)[:150] {
			want[row.id] = true
		}
		for _, decision := range journal.Report.Decisions {
			if !want[decision.EventID] || strings.HasPrefix(decision.EventID, "future") {
				violations = append(violations, fmt.Sprintf("recall %d bad nominee %s", report.index, decision.EventID))
			}
			delete(want, decision.EventID)
		}
		if len(want) != 0 {
			violations = append(violations, fmt.Sprintf("recall %d omitted %d oracle nominees", report.index, len(want)))
		}
		for _, ack := range writer.acks {
			if !ack.acked.After(report.offered) && report.packet.Snapshot.RuntimeVersion < ack.version {
				violations = append(violations, fmt.Sprintf("recall %d omitted acknowledged write %s", report.index, ack.id))
			}
		}
		for _, outcome := range outcomes {
			if !outcome.published.After(report.offered) && report.packet.Snapshot.RuntimeVersion < outcome.response.Snapshot.RuntimeVersion {
				violations = append(violations, fmt.Sprintf("recall %d omitted published outcome %d", report.index, outcome.index))
			}
		}
		for _, candidate := range report.packet.Candidates {
			if strings.HasPrefix(candidate.Event.ID, "future") || candidate.Event.AvailableAt.After(report.offered) {
				violations = append(violations, fmt.Sprintf("recall %d packed future %s", report.index, candidate.Event.ID))
			}
			if candidate.BayesianApplied {
				learned++
			}
			for _, outcome := range outcomes {
				if outcome.err == nil && outcome.eventID == candidate.Event.ID &&
					!outcome.published.After(report.offered) {
					labelledCandidates++
					if outcome.response.Posterior.Certified {
						certifiedLabelledCandidates++
					}
					if outcome.response.Posterior.EvidenceEpoch == report.packet.Snapshot.EvidenceEpoch {
						epochMatchedLabelledCandidates++
					}
					if outcome.response.Posterior.Certified &&
						outcome.response.Posterior.EvidenceEpoch == report.packet.Snapshot.EvidenceEpoch {
						epochAlignedLabelledCandidates++
					}
				}
			}
		}
	}
	for _, outcome := range outcomes {
		if outcome.err != nil || outcome.response.Duplicate || outcome.view == nil ||
			outcome.view.snapshot != outcome.response.Snapshot || outcome.view.at.After(outcome.published) {
			violations = append(violations, fmt.Sprintf("outcome %d failed publication: %v", outcome.index, outcome.err))
			continue
		}
		outcomeAgeNS = append(outcomeAgeNS, outcome.published.Sub(outcome.offered).Nanoseconds())
	}
	finalView := base.current.Load()
	var finalEpoch uint64
	if finalView != nil {
		finalEpoch = finalView.snapshot.EvidenceEpoch
	}
	if _, ready := gate.capture(ctx); !ready || finalView == nil || base.stales.Load() != 0 {
		violations = append(violations, fmt.Sprintf("final publication unready or stale=%d", base.stales.Load()))
	} else if finalView.snapshot.RuntimeVersion != seedVersion+128+uint64(len(outcomes))+2*uint64(len(writer.refreshNS)) {
		violations = append(violations, fmt.Sprintf("final runtime version=%d want=%d",
			finalView.snapshot.RuntimeVersion, seedVersion+128+uint64(len(outcomes))+2*uint64(len(writer.refreshNS))))
	}
	if feedback && expectLearning && learned == 0 && !raceCorrectnessOnly {
		violations = append(violations, "no loaded Recall used an updated posterior")
	}
	if feedback && !expectLearning && learned != 0 {
		violations = append(violations, "uncertified control unexpectedly used a learned posterior")
	}
	if refresh && len(writer.refreshNS) != len(writer.groups) {
		violations = append(violations, fmt.Sprintf("certificate refreshes=%d writer_groups=%d",
			len(writer.refreshNS), len(writer.groups)))
	}
	if refresh && afterOutcomeCertified != afterOutcomeRecalls {
		violations = append(violations, fmt.Sprintf("later certified recalls=%d/%d", afterOutcomeCertified, afterOutcomeRecalls))
	}
	writeP99, callP99, offerP99 := pinnedPercentile(writeAgeNS, .99), pinnedPercentile(callNS, .99), pinnedPercentile(offerNS, .99)
	outcomeP99, outcomeMax := pinnedPercentile(outcomeAgeNS, .99), pinnedPercentile(outcomeAgeNS, 1)
	viewMax := pinnedPercentile(viewAgeNS, 1)
	if !raceCorrectnessOnly && (writeP99 >= 250*time.Millisecond || callP99 >= 100*time.Millisecond ||
		offerP99 >= 100*time.Millisecond || viewMax >= 250*time.Millisecond ||
		feedback && (outcomeP99 >= 100*time.Millisecond || outcomeMax >= 250*time.Millisecond)) {
		violations = append(violations, "frozen loaded latency/freshness gate failed")
	}
	t.Logf("feedback=%v refresh=%v race_only=%v writes=%d recalls=%d outcomes=%d learned_candidates=%d labelled_candidates=%d certified_labelled_candidates=%d epoch_matched_labelled_candidates=%d epoch_aligned_labelled_candidates=%d certified_recalls=%d after_outcome_recalls=%d after_outcome_certified=%d refreshes=%d refresh_p99=%s seed_evidence_epoch=%d final_evidence_epoch=%d stale=%d write_p99=%s call_p99=%s offer_p99=%s outcome_p99=%s outcome_max=%s view_max=%s",
		feedback, refresh, raceCorrectnessOnly, len(writer.acks), len(reports), len(outcomes), learned,
		labelledCandidates, certifiedLabelledCandidates, epochMatchedLabelledCandidates, epochAlignedLabelledCandidates,
		certifiedRecalls, afterOutcomeRecalls, afterOutcomeCertified,
		len(writer.refreshNS), pinnedPercentile(writer.refreshNS, .99),
		selection.EvidenceEpoch, finalEpoch, base.stales.Load(),
		writeP99, callP99, offerP99, outcomeP99, outcomeMax, viewMax)
	for i, violation := range violations {
		if i < 15 {
			t.Error(violation)
		}
	}
	if len(violations) > 0 {
		t.Fatalf("loaded outcome screen failed with %d violations", len(violations))
	}
}

func TestResearchPublishedOutcomeLoadV18(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_OUTCOME_LOAD_V18") != "1" {
		t.Skip("opt-in live-as-of outcome load screen")
	}
	raceCorrectnessOnly := os.Getenv("EVENTFRAME_RACE_CORRECTNESS_ONLY") == "1"
	trials := 2
	if raceCorrectnessOnly {
		trials = 1
	}
	for trial := 1; trial <= trials; trial++ {
		for _, feedback := range []bool{false, true} {
			t.Run(fmt.Sprintf("trial-%d-feedback-%v", trial, feedback), func(t *testing.T) {
				runPublishedOutcomeLoadV18(t, feedback, raceCorrectnessOnly, false, feedback)
			})
		}
	}
}
