package libravdbstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

type sortLoadWriteJob struct {
	write   ResearchEventWrite
	offered time.Time
}

type sortLoadAck struct {
	id  string
	age int64
}

type sortLoadWriterReport struct {
	acks       []sortLoadAck
	groupCalls []int64
	groupSizes []int
	err        error
}

type sortLoadReadReport struct {
	offered, started, ended time.Time
	ids                     []string
	phases                  [4]int64
	pubAge                  int64
	err                     error
}

type sortLoadTrial struct {
	writeAges, writeGaps, readCalls, readOffers, readGaps, groupCalls, pubAges []int64
	groupSizes                                                                 []int
	violations                                                                 []string
	phases                                                                     [4][]int64
	phaseResiduals                                                             []int64
}

type sortLoadReadMode uint8

const (
	sortLoadGateRead sortLoadReadMode = iota
	sortLoadPhaseRead
	sortLoadPublishedRead
)

func sortLoadSearchPhased(ctx context.Context, g *incrementalSortGate, at time.Time) ([]string, [4]int64, error) {
	var phases [4]int64
	started := time.Now()
	lsn, ok := g.capture(ctx)
	phases[0] = time.Since(started).Nanoseconds()
	if !ok {
		return nil, phases, errors.New("sort-key migration is not published")
	}
	started = time.Now()
	lease, err := g.store.db.SnapshotAtLSN(ctx, lsn)
	phases[1] = time.Since(started).Nanoseconds()
	if err != nil {
		return nil, phases, err
	}
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + g.name +
		" AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT 10"
	started = time.Now()
	rows, err := g.store.db.QueryWithParams(ctx, query, libra.QueryParams{
		"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
		"available_by_sort": researchAvailabilitySortKey(at),
	})
	phases[2] = time.Since(started).Nanoseconds()
	if err != nil {
		started = time.Now()
		lease.Close()
		phases[3] = time.Since(started).Nanoseconds()
		return nil, phases, err
	}
	ids := make([]string, len(rows.Results))
	for i, row := range rows.Results {
		ids[i] = row.ID
	}
	started = time.Now()
	lease.Close()
	phases[3] = time.Since(started).Nanoseconds()
	return ids, phases, nil
}

func sortLoadOfferGaps(times []time.Time) []int64 {
	gaps := make([]int64, 0, max(0, len(times)-1))
	for i := 1; i < len(times); i++ {
		gaps = append(gaps, times[i].Sub(times[i-1]).Nanoseconds())
	}
	return gaps
}

func sortLoadTakeGroup(jobs <-chan sortLoadWriteJob, capSize int, dwell time.Duration) ([]sortLoadWriteJob, bool) {
	first, ok := <-jobs
	if !ok {
		return nil, true
	}
	group := []sortLoadWriteJob{first}
	closed := false
	for len(group) < capSize {
		select {
		case next, ok := <-jobs:
			if !ok {
				closed = true
				goto drained
			}
			group = append(group, next)
		default:
			goto drained
		}
	}
drained:
	if len(group) < capSize && !closed {
		if remaining := time.Until(first.offered.Add(dwell)); remaining > 0 {
			timer := time.NewTimer(remaining)
		collect:
			for len(group) < capSize {
				select {
				case next, ok := <-jobs:
					if !ok {
						closed = true
						break collect
					}
					group = append(group, next)
				case <-timer.C:
					break collect
				}
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}
	}
	return group, closed
}

func runSortBatchOpenLoopTrial(t *testing.T, batched bool, readMode sortLoadReadMode) sortLoadTrial {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	root := t.TempDir()
	g := createIncrementalSortGate(t, root)
	defer func() {
		if g != nil {
			g.close()
		}
	}()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	for start := 0; start < 197; start += 16 {
		group := make([]ResearchEventWrite, 0, min(16, 197-start))
		for i := start; i < min(197, start+16); i++ {
			write := pinnedWrite(fmt.Sprintf("base-%03d", i), second.Add(100*time.Millisecond))
			write.Vector = []float32{0, 1, 0, 0}
			group = append(group, write)
		}
		if err := g.appendBatch(ctx, group, ""); err != nil {
			t.Fatal("seed eligible corpus", err)
		}
	}
	future := make([]ResearchEventWrite, 4)
	for i := range future {
		future[i] = pinnedWrite(fmt.Sprintf("future-%02d", i), second.Add(125*time.Millisecond))
	}
	if err := g.appendBatch(ctx, future, ""); err != nil {
		t.Fatal("seed future sentinels", err)
	}
	if g.wantRows != 204 {
		t.Fatal("seed row count", g.wantRows)
	}
	for i := 0; i < 4; i++ {
		ids, err := g.search(ctx, asOf)
		if err != nil || len(ids) != 10 {
			t.Fatal("warm exact-LSN search", ids, err)
		}
		for _, id := range ids {
			if id == "future125" || strings.HasPrefix(id, "future-") {
				t.Fatal("future sentinel leaked before offered load", id)
			}
		}
	}
	var published *sortPublishedReader
	if readMode == sortLoadPublishedRead {
		var err error
		published, err = newSortPublishedReader(ctx, g)
		if err != nil {
			t.Fatal("initial published read view", err)
		}
	}
	writeJobs := make(chan sortLoadWriteJob, 256)
	readJobs := make(chan time.Time, 192)
	writeOffers := make(chan []time.Time, 1)
	readOffers := make(chan []time.Time, 1)
	writerDone := make(chan sortLoadWriterReport, 1)
	readReports := make(chan sortLoadReadReport, 192)
	start := make(chan struct{})
	go func() {
		<-start
		defer close(writeJobs)
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		times := make([]time.Time, 0, 256)
		defer func() { writeOffers <- times }()
		for i := 0; i < 256; i++ {
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
			write.Vector = []float32{0, 1, 0, 0}
			select {
			case writeJobs <- sortLoadWriteJob{write, offered}:
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
		times := make([]time.Time, 0, 192)
		defer func() { readOffers <- times }()
		for i := 0; i < 192; i++ {
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
			case readJobs <- offered:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() {
		<-start
		report := sortLoadWriterReport{}
		capSize, dwell := 1, time.Duration(0)
		if batched {
			capSize, dwell = 16, 16*time.Millisecond
		}
		for {
			group, closed := sortLoadTakeGroup(writeJobs, capSize, dwell)
			if len(group) == 0 {
				writerDone <- report
				return
			}
			writes := make([]ResearchEventWrite, len(group))
			for i, job := range group {
				writes[i] = job.write
			}
			began := time.Now()
			var err error
			if batched {
				err = g.appendBatch(ctx, writes, "")
			} else {
				err = g.append(ctx, writes[0], "")
			}
			if published != nil {
				if err == nil {
					err = published.publish(ctx)
				}
				if err != nil {
					published.poison()
				}
			}
			acked := time.Now()
			if err != nil {
				report.err = fmt.Errorf("publish group after %d acknowledgements: %w", len(report.acks), err)
				writerDone <- report
				return
			}
			report.groupCalls = append(report.groupCalls, acked.Sub(began).Nanoseconds())
			report.groupSizes = append(report.groupSizes, len(group))
			for _, job := range group {
				report.acks = append(report.acks, sortLoadAck{job.write.Event.ID, acked.Sub(job.offered).Nanoseconds()})
			}
			if closed {
				writerDone <- report
				return
			}
		}
	}()
	var readers sync.WaitGroup
	for i := 0; i < 8; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			<-start
			for offered := range readJobs {
				began := time.Now()
				var ids []string
				var phases [4]int64
				var pubAge time.Duration
				var err error
				switch readMode {
				case sortLoadPhaseRead:
					ids, phases, err = sortLoadSearchPhased(ctx, g, asOf)
				case sortLoadPublishedRead:
					ids, pubAge, err = published.search(ctx, asOf)
				default:
					ids, err = g.search(ctx, asOf)
				}
				ended := time.Now()
				readReports <- sortLoadReadReport{offered, began, ended, ids, phases, pubAge.Nanoseconds(), err}
			}
		}()
	}
	go func() { readers.Wait(); close(readReports) }()
	close(start)
	out := sortLoadTrial{}
	for report := range readReports {
		callNS := report.ended.Sub(report.started).Nanoseconds()
		out.readCalls = append(out.readCalls, callNS)
		out.readOffers = append(out.readOffers, report.ended.Sub(report.offered).Nanoseconds())
		if readMode == sortLoadPublishedRead {
			out.pubAges = append(out.pubAges, report.pubAge)
		}
		if readMode == sortLoadPhaseRead {
			residual := callNS
			for i, phase := range report.phases {
				out.phases[i] = append(out.phases[i], phase)
				residual -= phase
			}
			out.phaseResiduals = append(out.phaseResiduals, residual)
		}
		if report.err != nil {
			out.violations = append(out.violations, fmt.Sprintf("search: %v", report.err))
			continue
		}
		if len(report.ids) != 10 {
			out.violations = append(out.violations, fmt.Sprintf("search returned %d of 10", len(report.ids)))
		}
		seen := make(map[string]bool, len(report.ids))
		for _, id := range report.ids {
			if seen[id] {
				out.violations = append(out.violations, "duplicate search ID: "+id)
			}
			seen[id] = true
			if id != "past100" && id != "at120" &&
				!strings.HasPrefix(id, "base-") && !strings.HasPrefix(id, "visible-") {
				out.violations = append(out.violations, "future or unknown search ID: "+id)
			}
		}
	}
	writer := <-writerDone
	writeTimes, readTimes := <-writeOffers, <-readOffers
	out.writeGaps, out.readGaps = sortLoadOfferGaps(writeTimes), sortLoadOfferGaps(readTimes)
	out.groupCalls, out.groupSizes = writer.groupCalls, writer.groupSizes
	for _, ack := range writer.acks {
		out.writeAges = append(out.writeAges, ack.age)
	}
	if writer.err != nil {
		out.violations = append(out.violations, writer.err.Error())
	}
	if len(writeTimes) != 256 || len(writer.acks) != 256 || len(readTimes) != 192 || len(out.readCalls) != 192 {
		out.violations = append(out.violations, fmt.Sprintf("incomplete offers/acks/reads: %d/%d/%d/%d",
			len(writeTimes), len(writer.acks), len(readTimes), len(out.readCalls)))
	}
	ackIDs := make(map[string]bool, len(writer.acks))
	for _, ack := range writer.acks {
		if ackIDs[ack.id] {
			out.violations = append(out.violations, "duplicate acknowledgement: "+ack.id)
		}
		ackIDs[ack.id] = true
	}
	if _, ok := g.capture(ctx); !ok || g.wantRows != 460 {
		out.violations = append(out.violations, fmt.Sprintf("final READY/row count: %v/%d", ok, g.wantRows))
	}
	col, err := g.store.db.GetCollection(g.name)
	if err != nil {
		out.violations = append(out.violations, fmt.Sprintf("get collection: %v", err))
	} else {
		for i := 0; i < 256; i++ {
			id := fmt.Sprintf("visible-%03d", i)
			record, err := col.Get(ctx, id)
			if err != nil || record.Metadata["available_at_sort"] != researchAvailabilitySortKey(asOf) {
				out.violations = append(out.violations, fmt.Sprintf("missing or unsortable offered ID %s: %v", id, err))
			}
		}
	}
	if err := g.close(); err != nil {
		out.violations = append(out.violations, fmt.Sprintf("close owner: %v", err))
	}
	g = nil
	g, err = openIncrementalSortGate(root)
	if err != nil {
		out.violations = append(out.violations, fmt.Sprintf("reopen owner: %v", err))
	} else if _, ok := g.capture(ctx); !ok || g.wantRows != 460 {
		out.violations = append(out.violations, fmt.Sprintf("reopened journal READY/row count: %v/%d", ok, g.wantRows))
	}
	return out
}

func TestResearchIncrementalSortBatchOpenLoopV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_BATCH_OPENLOOP_V1") != "1" {
		t.Skip("opt-in private 4 ms batch publication load diagnostic")
	}
	all := map[bool]*sortLoadTrial{false: {}, true: {}}
	for pair := 0; pair < 3; pair++ {
		order := []bool{false, true}
		if pair%2 == 1 {
			order = []bool{true, false}
		}
		for _, batched := range order {
			r := runSortBatchOpenLoopTrial(t, batched, sortLoadGateRead)
			a := all[batched]
			a.writeAges = append(a.writeAges, r.writeAges...)
			a.writeGaps = append(a.writeGaps, r.writeGaps...)
			a.readCalls = append(a.readCalls, r.readCalls...)
			a.readOffers = append(a.readOffers, r.readOffers...)
			a.readGaps = append(a.readGaps, r.readGaps...)
			a.groupCalls = append(a.groupCalls, r.groupCalls...)
			a.groupSizes = append(a.groupSizes, r.groupSizes...)
			a.violations = append(a.violations, r.violations...)
			t.Logf("pair=%d batched=%v groups=%d writer_offer_gap_p50=%s writer_age_p50=%s writer_age_p95=%s writer_age_p99=%s read_offer_gap_p50=%s read_call_p50=%s read_call_p95=%s read_call_p99=%s read_offer_p99=%s violations=%d",
				pair, batched, len(r.groupCalls), pinnedPercentile(r.writeGaps, .5),
				pinnedPercentile(r.writeAges, .5), pinnedPercentile(r.writeAges, .95), pinnedPercentile(r.writeAges, .99),
				pinnedPercentile(r.readGaps, .5), pinnedPercentile(r.readCalls, .5), pinnedPercentile(r.readCalls, .95),
				pinnedPercentile(r.readCalls, .99), pinnedPercentile(r.readOffers, .99), len(r.violations))
			if len(r.violations) > 0 {
				t.Logf("pair=%d batched=%v first_violation=%s", pair, batched, r.violations[0])
			}
		}
	}
	control, candidate := all[false], all[true]
	readRatio := float64(pinnedPercentile(candidate.readCalls, .99)) / float64(pinnedPercentile(control.readCalls, .99))
	t.Logf("pooled control_writer_p99=%s candidate_writer_p99=%s control_read_p99=%s candidate_read_p99=%s read_ratio=%.3f candidate_groups=%d violations_control=%d violations_candidate=%d",
		pinnedPercentile(control.writeAges, .99), pinnedPercentile(candidate.writeAges, .99),
		pinnedPercentile(control.readCalls, .99), pinnedPercentile(candidate.readCalls, .99),
		readRatio, len(candidate.groupCalls), len(control.violations), len(candidate.violations))
	if len(control.violations) > 0 || len(candidate.violations) > 0 ||
		len(control.writeAges) != 3*256 || len(candidate.writeAges) != 3*256 ||
		len(control.readCalls) != 3*192 || len(candidate.readCalls) != 3*192 ||
		pinnedPercentile(candidate.writeAges, .99) >= 250*time.Millisecond ||
		pinnedPercentile(candidate.readCalls, .99) >= 100*time.Millisecond || readRatio > 1.10 {
		t.Error("batch open-loop backend component failed frozen screen")
	}
}

func TestResearchSortBatchReadPhasesV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_BATCH_READ_PHASES_V1") != "1" {
		t.Skip("opt-in private read-phase attribution diagnostic")
	}
	for pair := 0; pair < 2; pair++ {
		order := []bool{false, true}
		if pair%2 == 1 {
			order = []bool{true, false}
		}
		for _, batched := range order {
			r := runSortBatchOpenLoopTrial(t, batched, sortLoadPhaseRead)
			t.Logf("pair=%d batched=%v read_p50=%s read_p99=%s capture_p50=%s capture_p99=%s lease_p50=%s lease_p99=%s sql_p50=%s sql_p99=%s close_p50=%s close_p99=%s residual_p99=%s writer_age_p99=%s violations=%d",
				pair, batched, pinnedPercentile(r.readCalls, .5), pinnedPercentile(r.readCalls, .99),
				pinnedPercentile(r.phases[0], .5), pinnedPercentile(r.phases[0], .99),
				pinnedPercentile(r.phases[1], .5), pinnedPercentile(r.phases[1], .99),
				pinnedPercentile(r.phases[2], .5), pinnedPercentile(r.phases[2], .99),
				pinnedPercentile(r.phases[3], .5), pinnedPercentile(r.phases[3], .99),
				pinnedPercentile(r.phaseResiduals, .99), pinnedPercentile(r.writeAges, .99), len(r.violations))
			if len(r.violations) > 0 {
				t.Errorf("phase diagnostic violated invariant: %s", r.violations[0])
			}
			if len(r.phases[0]) != 192 || pinnedPercentile(r.phaseResiduals, .99) > time.Millisecond {
				t.Error("read phase measurements do not account for calls")
			}
		}
	}
}

func TestResearchSortPublishedOpenLoopV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_PUBLISHED_OPENLOOP_V1") != "1" {
		t.Skip("opt-in three-arm published-view load diagnostic")
	}
	type arm struct {
		name    string
		batched bool
		mode    sortLoadReadMode
	}
	arms := []arm{
		{"single_gate", false, sortLoadGateRead},
		{"batch_gate", true, sortLoadGateRead},
		{"batch_published", true, sortLoadPublishedRead},
	}
	orders := [][]int{{0, 1, 2}, {2, 0, 1}, {1, 2, 0}}
	all := make([]*sortLoadTrial, len(arms))
	for i := range all {
		all[i] = &sortLoadTrial{}
	}
	for pair, order := range orders {
		for _, armIndex := range order {
			a := arms[armIndex]
			r := runSortBatchOpenLoopTrial(t, a.batched, a.mode)
			total := all[armIndex]
			total.writeAges = append(total.writeAges, r.writeAges...)
			total.writeGaps = append(total.writeGaps, r.writeGaps...)
			total.readCalls = append(total.readCalls, r.readCalls...)
			total.readOffers = append(total.readOffers, r.readOffers...)
			total.readGaps = append(total.readGaps, r.readGaps...)
			total.groupCalls = append(total.groupCalls, r.groupCalls...)
			total.pubAges = append(total.pubAges, r.pubAges...)
			total.violations = append(total.violations, r.violations...)
			t.Logf("pair=%d arm=%s groups=%d write_gap_p50=%s writer_age_p50=%s writer_age_p95=%s writer_age_p99=%s read_gap_p50=%s read_call_p50=%s read_call_p95=%s read_call_p99=%s read_offer_p99=%s pub_age_p99=%s pub_age_max=%s violations=%d",
				pair, a.name, len(r.groupCalls), pinnedPercentile(r.writeGaps, .5),
				pinnedPercentile(r.writeAges, .5), pinnedPercentile(r.writeAges, .95), pinnedPercentile(r.writeAges, .99),
				pinnedPercentile(r.readGaps, .5), pinnedPercentile(r.readCalls, .5), pinnedPercentile(r.readCalls, .95),
				pinnedPercentile(r.readCalls, .99), pinnedPercentile(r.readOffers, .99),
				pinnedPercentile(r.pubAges, .99), pinnedPercentile(r.pubAges, 1), len(r.violations))
			if len(r.violations) > 0 {
				t.Logf("pair=%d arm=%s first_violation=%s", pair, a.name, r.violations[0])
			}
		}
	}
	control, originalBatch, rescued := all[0], all[1], all[2]
	readRatio := float64(pinnedPercentile(rescued.readCalls, .99)) / float64(pinnedPercentile(control.readCalls, .99))
	originalRatio := float64(pinnedPercentile(originalBatch.readCalls, .99)) / float64(pinnedPercentile(control.readCalls, .99))
	t.Logf("pooled single_read_p99=%s original_batch_read_p99=%s rescued_read_p99=%s original_ratio=%.3f rescued_ratio=%.3f rescued_writer_age_p99=%s rescued_pub_age_max=%s violations=%d/%d/%d",
		pinnedPercentile(control.readCalls, .99), pinnedPercentile(originalBatch.readCalls, .99),
		pinnedPercentile(rescued.readCalls, .99), originalRatio, readRatio,
		pinnedPercentile(rescued.writeAges, .99), pinnedPercentile(rescued.pubAges, 1),
		len(control.violations), len(originalBatch.violations), len(rescued.violations))
	if len(control.violations) > 0 || len(originalBatch.violations) > 0 || len(rescued.violations) > 0 ||
		len(rescued.writeAges) != 3*256 || len(rescued.readCalls) != 3*192 || len(rescued.pubAges) != 3*192 ||
		pinnedPercentile(rescued.writeAges, .99) >= 250*time.Millisecond ||
		pinnedPercentile(rescued.readCalls, .99) >= 100*time.Millisecond ||
		pinnedPercentile(rescued.pubAges, 1) >= 250*time.Millisecond || readRatio > 1.10 {
		t.Error("last-published-LSN read rescue failed frozen component screen")
	}
}

func TestResearchSortPublishedOpenLoopRaceV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_PUBLISHED_OPENLOOP_RACE_V1") != "1" {
		t.Skip("opt-in published-view concurrent correctness under race instrumentation")
	}
	r := runSortBatchOpenLoopTrial(t, true, sortLoadPublishedRead)
	t.Logf("race diagnostic acknowledgements=%d reads=%d groups=%d max_published_age=%s violations=%d",
		len(r.writeAges), len(r.readCalls), len(r.groupCalls), pinnedPercentile(r.pubAges, 1), len(r.violations))
	if len(r.violations) > 0 || len(r.writeAges) != 256 || len(r.readCalls) != 192 || len(r.pubAges) != 192 {
		if len(r.violations) > 0 {
			t.Logf("first violation: %s", r.violations[0])
		}
		t.Error("published-view race diagnostic lost events or violated as-of integrity")
	}
}
