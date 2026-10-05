package libravdbstore

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

type rollingReadResult struct {
	callNS, offerNS int64
	ended           time.Time
	version         uint64
	retries         int
	err             error
}

type rollingWriteAck struct {
	ended   time.Time
	version uint64
}

type rollingTrial struct {
	call, offer, gaps, lags, writeAcks []int64
	writerTotal                        time.Duration
	retries                            int
	retainedBytes                      uint64
}

func rollingWrite(id string, at time.Time, visible bool) ResearchEventWrite {
	w := pinnedWrite(id, at)
	if !visible {
		w.Vector = []float32{0, 1, 0, 0}
	}
	return w
}

func runRollingReaderTrial(t *testing.T, pinned bool) rollingTrial {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	s, err := Open(Config{Path: t.TempDir() + "/rolling.libravdb", Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	seedAt, futureAt := asOf.Add(-time.Minute), asOf.Add(time.Hour)
	for start := 0; start < 200; start += 16 {
		writes := make([]ResearchEventWrite, 0, 16)
		for i := start; i < start+16 && i < 200; i++ {
			writes = append(writes, rollingWrite(fmt.Sprintf("base-%03d", i), seedAt, false))
		}
		if _, err := s.PutResearchEventBatch(ctx, writes); err != nil {
			t.Fatal("seed", err)
		}
	}
	future := make([]ResearchEventWrite, 16)
	for i := range future {
		future[i] = rollingWrite(fmt.Sprintf("future-%03d", i), futureAt, true)
	}
	if _, err := s.PutResearchEventBatch(ctx, future); err != nil {
		t.Fatal("future seed", err)
	}
	base := s.Snapshot(ctx)
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + collectionName("tenant-a", "research:d4") +
		" AS OF LSN $snapshot_lsn d WHERE available_at <= $available_by ORDER BY distance LIMIT 50"
	read := func() rollingReadResult {
		began := time.Now()
		out := rollingReadResult{}
		rejectedVersions := make([]uint64, 0, 4)
		for attempt := 0; attempt < 4; attempt++ {
			s.writeMu.RLock()
			snap := s.snapshot
			var lsn uint64
			var lease *libra.TemporalSnapshot
			var captureErr error
			if pinned {
				lsn, captureErr = s.db.LatestCommitLSN(ctx)
				if captureErr == nil {
					lease, captureErr = s.db.SnapshotAtLSN(ctx, lsn)
				}
			}
			s.writeMu.RUnlock()
			if captureErr != nil {
				out.err = fmt.Errorf("capture snapshot and LSN: %w", captureErr)
				break
			}
			var ids []string
			if pinned {
				rows, queryErr := s.db.QueryWithParams(ctx, query, libra.QueryParams{
					"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
					"available_by": asOf.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
				})
				lease.Close()
				if queryErr != nil {
					out.err = fmt.Errorf("rolling LSN query: %w", queryErr)
					break
				}
				ids = make([]string, len(rows.Results))
				for i, row := range rows.Results {
					ids[i] = row.ID
				}
			} else {
				rows, queryErr := s.Search(ctx, "tenant-a", []float32{1, 0, 0, 0}, asOf, 50)
				if queryErr != nil {
					out.err = fmt.Errorf("current search: %w", queryErr)
					break
				}
				ids = make([]string, len(rows))
				for i, row := range rows {
					ids[i] = row.Event.ID
				}
			}
			if !s.ResearchSnapshotCompatible(ctx, snap, asOf) {
				out.retries++
				rejectedVersions = append(rejectedVersions, snap.RuntimeVersion)
				continue
			}
			visible := int(snap.RuntimeVersion - base.RuntimeVersion)
			seen := make(map[string]bool, len(ids))
			visibleSeen := false
			if len(ids) != 50 {
				out.err = fmt.Errorf("returned %d of 50 eligible rows", len(ids))
				break
			}
			for _, id := range ids {
				if seen[id] {
					out.err = fmt.Errorf("duplicate row %s", id)
					break
				}
				seen[id] = true
				switch {
				case strings.HasPrefix(id, "base-"):
				case strings.HasPrefix(id, "visible-"):
					i, parseErr := strconv.Atoi(strings.TrimPrefix(id, "visible-"))
					if parseErr != nil || i < 0 || i >= visible {
						out.err = fmt.Errorf("row outside captured version: %s", id)
					} else {
						visibleSeen = true
					}
				default:
					out.err = fmt.Errorf("future or unknown row %s", id)
				}
				if out.err != nil {
					break
				}
			}
			if out.err == nil && visible > 0 && !visibleSeen {
				out.err = fmt.Errorf("captured version %d hid all visible rows", snap.RuntimeVersion)
			}
			out.version = snap.RuntimeVersion
			break
		}
		if out.err == nil && out.version == 0 {
			out.err = fmt.Errorf("exhausted four publication attempts pinned=%t captured=%v latest=%d", pinned, rejectedVersions, s.Snapshot(ctx).RuntimeVersion)
		}
		out.ended = time.Now()
		out.callNS = out.ended.Sub(began).Nanoseconds()
		return out
	}
	for i := 0; i < 4; i++ {
		if r := read(); r.err != nil {
			t.Fatal("warm read", r.err)
		}
	}
	jobs := make(chan time.Time, 192)
	results := make(chan rollingReadResult, 192)
	offerTimes := make(chan []time.Time, 1)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for offered := range jobs {
				r := read()
				r.offerNS = r.ended.Sub(offered).Nanoseconds()
				results <- r
			}
		}()
	}
	go func() {
		<-start
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		defer close(jobs)
		times := make([]time.Time, 0, 192)
		defer func() { offerTimes <- times }()
		for i := 0; i < 192; i++ {
			if i > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					return
				}
			}
			at := time.Now()
			times = append(times, at)
			select {
			case jobs <- at:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { workers.Wait(); close(results) }()
	type writeReport struct {
		acks  []rollingWriteAck
		gaps  []int64
		total time.Duration
		err   error
	}
	writerDone := make(chan writeReport, 1)
	go func() {
		<-start
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		report := writeReport{}
		began := time.Now()
		for first := 0; first < 256; first += 16 {
			if first > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					report.err = ctx.Err()
					writerDone <- report
					return
				}
			}
			offered := time.Now()
			writes := make([]ResearchEventWrite, 16)
			for i := range writes {
				writes[i] = rollingWrite(fmt.Sprintf("visible-%03d", first+i), asOf, true)
			}
			outcomes, writeErr := s.PutResearchEventBatch(ctx, writes)
			if writeErr != nil || len(outcomes) != 16 {
				report.err = fmt.Errorf("batch %d: outcomes=%d err=%v", first/16, len(outcomes), writeErr)
				writerDone <- report
				return
			}
			for _, outcome := range outcomes {
				if outcome.Duplicate {
					report.err = fmt.Errorf("duplicate in batch %d", first/16)
					writerDone <- report
					return
				}
			}
			ended := time.Now()
			report.acks = append(report.acks, rollingWriteAck{ended, outcomes[0].Snapshot.RuntimeVersion})
			report.gaps = append(report.gaps, ended.Sub(offered).Nanoseconds())
		}
		report.total = time.Since(began)
		writerDone <- report
	}()
	close(start)
	allReads := make([]rollingReadResult, 0, 192)
	var firstError error
	for r := range results {
		if r.err != nil && firstError == nil {
			firstError = r.err
		}
		allReads = append(allReads, r)
	}
	times := <-offerTimes
	writer := <-writerDone
	if writer.err != nil && firstError == nil {
		firstError = writer.err
	}
	if firstError != nil {
		t.Fatal("rolling reader trial pinned=", pinned, firstError)
	}
	if len(allReads) != 192 || len(times) != 192 || len(writer.acks) != 16 || s.Snapshot(ctx).RuntimeVersion != base.RuntimeVersion+256 {
		t.Fatal("incomplete rolling reader trial", len(allReads), len(times), len(writer.acks))
	}
	if s.ResearchSnapshotCompatible(ctx, base, asOf) {
		t.Fatal("visible writes did not invalidate old publication")
	}
	for version := base.RuntimeVersion + 1; version <= base.RuntimeVersion+256; version++ {
		if !s.ingestMotion[version].Equal(asOf) {
			t.Fatal("missing visible motion", version)
		}
	}
	final := read()
	if final.err != nil || final.version != base.RuntimeVersion+256 {
		t.Fatal("fresh final read failed", final.err, final.version)
	}
	stats, err := s.db.TemporalStats(ctx)
	if err != nil || stats.ActiveLeaseCount != 0 {
		t.Fatal("temporal snapshot lease leaked", stats, err)
	}
	out := rollingTrial{writerTotal: writer.total, writeAcks: writer.gaps, retainedBytes: stats.RetainedBytes}
	for i := 1; i < len(times); i++ {
		out.gaps = append(out.gaps, times[i].Sub(times[i-1]).Nanoseconds())
	}
	for _, r := range allReads {
		out.call = append(out.call, r.callNS)
		out.offer = append(out.offer, r.offerNS)
		out.retries += r.retries
	}
	for _, ack := range writer.acks {
		var first time.Time
		for _, r := range allReads {
			if !r.ended.Before(ack.ended) && r.version >= ack.version && (first.IsZero() || r.ended.Before(first)) {
				first = r.ended
			}
		}
		if first.IsZero() {
			t.Fatal("no read reflected acknowledged version", ack.version)
		}
		out.lags = append(out.lags, first.Sub(ack.ended).Nanoseconds())
	}
	return out
}

func TestResearchRollingLSNReaderV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_ROLLING_LSN_READER_V1") != "1" {
		t.Skip("opt-in rolling-LSN visible-write design screen")
	}
	type totals struct {
		call, offer, lags []int64
		writer            time.Duration
		retries           int
	}
	all := map[bool]*totals{false: {}, true: {}}
	for trial := 0; trial < 3; trial++ {
		order := []bool{false, true}
		if trial%2 == 1 {
			order = []bool{true, false}
		}
		for _, pinned := range order {
			r := runRollingReaderTrial(t, pinned)
			a := all[pinned]
			a.call = append(a.call, r.call...)
			a.offer = append(a.offer, r.offer...)
			a.lags = append(a.lags, r.lags...)
			a.writer += r.writerTotal
			a.retries += r.retries
			t.Logf("trial=%d pinned=%v call_p50=%s call_p99=%s offer_p99=%s gap_p50=%s gap_p99=%s writer_total=%s write_ack_p99=%s lag_p99=%s retries=%d retained_bytes=%d",
				trial, pinned, pinnedPercentile(r.call, .5), pinnedPercentile(r.call, .99), pinnedPercentile(r.offer, .99),
				pinnedPercentile(r.gaps, .5), pinnedPercentile(r.gaps, .99), r.writerTotal, pinnedPercentile(r.writeAcks, .99),
				pinnedPercentile(r.lags, .99), r.retries, r.retainedBytes)
		}
	}
	current, candidate := all[false], all[true]
	readP99 := pinnedPercentile(candidate.call, .99)
	readRatio := float64(readP99) / float64(pinnedPercentile(current.call, .99))
	writerRatio := float64(candidate.writer) / float64(current.writer)
	lagP99 := pinnedPercentile(candidate.lags, .99)
	t.Logf("pooled current_call_p99=%s rolling_call_p99=%s read_ratio=%.3f current_offer_p99=%s rolling_offer_p99=%s writer_total_ratio=%.3f rolling_lag_p99=%s current_retries=%d rolling_retries=%d",
		pinnedPercentile(current.call, .99), readP99, readRatio, pinnedPercentile(current.offer, .99),
		pinnedPercentile(candidate.offer, .99), writerRatio, lagP99, current.retries, candidate.retries)
	if len(current.call) != 576 || len(candidate.call) != 576 || len(candidate.lags) != 48 || readRatio > 1.10 ||
		readP99 > 100*time.Millisecond || writerRatio > 1.25 || lagP99 > 25*time.Millisecond {
		t.Error("rolling-LSN reader failed frozen visible-write design screen")
	}
}

func TestResearchRollingLSNCandidateDiagnostic(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_ROLLING_LSN_DIAGNOSTIC") != "1" {
		t.Skip("opt-in candidate-only diagnostic; not a paired gate")
	}
	r := runRollingReaderTrial(t, true)
	t.Logf("candidate-only call_p99=%s offer_p99=%s lag_p99=%s retries=%d writer_total=%s retained_bytes=%d",
		pinnedPercentile(r.call, .99), pinnedPercentile(r.offer, .99), pinnedPercentile(r.lags, .99),
		r.retries, r.writerTotal, r.retainedBytes)
}
