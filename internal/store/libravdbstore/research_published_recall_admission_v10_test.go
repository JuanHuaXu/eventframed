package libravdbstore

import (
	"context"
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

type publishedRecallLoadStoreV10 struct {
	store.EventStore
	gate     *incrementalSortGate
	owner    sync.Mutex
	current  atomic.Pointer[publishedRecallViewV8]
	searches atomic.Int64
	stales   atomic.Int64
}

func (s *publishedRecallLoadStoreV10) publishLocked(ctx context.Context) error {
	lsn, ready := s.gate.capture(ctx)
	if !ready {
		s.current.Store(nil)
		return fmt.Errorf("journaled EventFrame view is not READY")
	}
	s.current.Store(&publishedRecallViewV8{lsn: lsn, snapshot: s.gate.store.Snapshot(ctx), at: time.Now()})
	return nil
}

func (s *publishedRecallLoadStoreV10) appendBatch(ctx context.Context, writes []ResearchEventWrite) (model.Snapshot, error) {
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

func (s *publishedRecallLoadStoreV10) Search(ctx context.Context, tenantID string, vector []float32, availableBy time.Time, limit int) ([]store.SearchResult, error) {
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

func (s *publishedRecallLoadStoreV10) Snapshot(ctx context.Context) model.Snapshot {
	if pin, ok := ctx.Value(publishedRecallPinKeyV8{}).(*publishedRecallPinV8); ok && pin.view != nil {
		return pin.view.snapshot
	}
	return s.gate.store.Snapshot(ctx)
}

func (s *publishedRecallLoadStoreV10) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	s.owner.Lock()
	defer s.owner.Unlock()
	current := s.gate.store.Snapshot(ctx)
	if !store.JournalSnapshotCompatible(entry.Snapshot, current, entry.AsOf, s.gate.store.ingestMotion) {
		s.stales.Add(1)
		return store.ErrStaleSnapshot
	}
	if err := s.gate.appendJournal(ctx, entry, ""); err != nil {
		s.current.Store(nil)
		return err
	}
	return s.publishLocked(ctx)
}

func (s *publishedRecallLoadStoreV10) Close() error { return nil }

type recallLoadOfferV10 struct {
	index   int
	offered time.Time
}

type recallLoadResultV10 struct {
	index                   int
	offered, started, ended time.Time
	admissionWait           time.Duration
	view                    *publishedRecallViewV8
	packet                  model.ContextPacket
	err                     error
}

type recallLoadAckV10 struct {
	id      string
	offered time.Time
	acked   time.Time
	version uint64
}

type recallLoadWriterV10 struct {
	acks           []recallLoadAckV10
	groups         []int
	leaseWaitNanos []int64
	err            error
}

func runPublishedRecallLoadV10(t *testing.T, raceCorrectnessOnly bool) {
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
	adapter := &publishedRecallLoadStoreV10{EventStore: gate.store, gate: gate}
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
	readJobs := make(chan recallLoadOfferV10, 128)
	writeOffers, readOffers := make(chan []time.Time, 1), make(chan []time.Time, 1)
	writerDone := make(chan recallLoadWriterV10, 1)
	readReports := make(chan recallLoadResultV10, 128)
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
			case readJobs <- recallLoadOfferV10{index: i, offered: offered}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		<-start
		report := recallLoadWriterV10{}
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
				report.acks = append(report.acks, recallLoadAckV10{job.write.Event.ID, job.offered, acked, snapshot.RuntimeVersion})
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
				callCtx := context.WithValue(ctx, publishedRecallPinKeyV8{}, pin)
				request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion,
					TenantID: "tenant-a", SessionID: fmt.Sprintf("lab-%03d", offer.index),
					Query: "public vector search fixture", Embedding: rawQuery, EmbeddingModel: em.ModelKey(),
					AsOf: asOf, RecallK: 50, PackK: 10, TokenBudget: 10000}
				waiting := time.Now()
				admission.RLock()
				leaseWait := time.Since(waiting)
				packet, err := svc.Recall(callCtx, request)
				admission.RUnlock()
				readReports <- recallLoadResultV10{index: offer.index, offered: offer.offered, started: began,
					ended: time.Now(), admissionWait: leaseWait, view: pin.view, packet: packet, err: err}
			}
		}()
	}
	go func() { readers.Wait(); close(readReports) }()
	close(start)
	reports := make([]recallLoadResultV10, 0, 128)
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

func TestResearchPublishedRecallLoadV10(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_RECALL_LOAD_V10") != "1" {
		t.Skip("opt-in full Service Recall published-LSN load screen")
	}
	raceCorrectnessOnly := os.Getenv("EVENTFRAME_RACE_CORRECTNESS_ONLY") == "1"
	for trial := 1; trial <= 2; trial++ {
		t.Run(fmt.Sprintf("trial-%d", trial), func(t *testing.T) {
			runPublishedRecallLoadV10(t, raceCorrectnessOnly)
		})
	}
}
