package libravdbstore

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/testutil"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

type pinnedReadResult struct {
	offerNS, callNS int64
	err             string
}

type pinnedReaderTrial struct {
	readCall, readOffer, offerGaps []int64
	writerTotal                    time.Duration
	retainedBytes                  uint64
	lsn                            uint64
}

func pinnedPercentile(values []int64, fraction float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	copyValues := append([]int64(nil), values...)
	sort.Slice(copyValues, func(i, j int) bool { return copyValues[i] < copyValues[j] })
	index := int(float64(len(copyValues))*fraction+.999999) - 1
	if index < 0 {
		index = 0
	}
	return time.Duration(copyValues[index])
}

func pinnedWrite(id string, at time.Time) ResearchEventWrite {
	return ResearchEventWrite{Event: testutil.Event(id, "public vector search fixture", at), Vector: []float32{1, 0, 0, 0}, Digest: id}
}

func runPinnedReaderTrial(t *testing.T, pinned bool) pinnedReaderTrial {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	s, err := Open(Config{Path: t.TempDir() + "/pinned.libravdb", Dimension: 4,
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
			writes = append(writes, pinnedWrite(fmt.Sprintf("base-%03d", i), seedAt))
		}
		if _, err := s.PutResearchEventBatch(ctx, writes); err != nil {
			t.Fatal("seed", err)
		}
	}
	base := s.Snapshot(ctx)
	lsn, err := s.db.LatestCommitLSN(ctx)
	if err != nil || lsn == 0 {
		t.Fatal("capture durable LSN", lsn, err)
	}
	lease, err := s.db.SnapshotAtLSN(ctx, lsn)
	if err != nil {
		t.Fatal("pin durable LSN", err)
	}
	defer lease.Close()
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + collectionName("tenant-a", "research:d4") +
		" AS OF LSN $snapshot_lsn d ORDER BY distance LIMIT 200"
	params := libra.QueryParams{"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0}}
	read := func() ([]string, error) {
		if !s.ResearchSnapshotCompatible(ctx, base, asOf) {
			return nil, fmt.Errorf("pre-read publication guard rejected future-only history")
		}
		var ids []string
		if pinned {
			rows, err := s.db.QueryWithParams(ctx, query, params)
			if err != nil {
				return nil, err
			}
			ids = make([]string, len(rows.Results))
			for i, row := range rows.Results {
				ids[i] = row.ID
			}
		} else {
			rows, err := s.Search(ctx, "tenant-a", []float32{1, 0, 0, 0}, asOf, 200)
			if err != nil {
				return nil, err
			}
			ids = make([]string, len(rows))
			for i, row := range rows {
				ids[i] = row.Event.ID
			}
		}
		if !s.ResearchSnapshotCompatible(ctx, base, asOf) {
			return nil, fmt.Errorf("post-read publication guard rejected future-only history")
		}
		return ids, nil
	}
	for i := 0; i < 4; i++ {
		if _, err := read(); err != nil {
			t.Fatal("warm read", err)
		}
	}
	jobs := make(chan time.Time, 192)
	results := make(chan pinnedReadResult, 192)
	offerTimes := make(chan []time.Time, 1)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for offered := range jobs {
				began := time.Now()
				ids, err := read()
				r := pinnedReadResult{callNS: time.Since(began).Nanoseconds(), offerNS: time.Since(offered).Nanoseconds()}
				if err != nil {
					r.err = err.Error()
				} else if len(ids) != 200 {
					r.err = fmt.Sprintf("read %d of 200 past events", len(ids))
				} else {
					seen := make(map[string]bool, len(ids))
					for _, id := range ids {
						if !strings.HasPrefix(id, "base-") || seen[id] {
							r.err = "future or duplicate event in pinned result"
							break
						}
						seen[id] = true
					}
				}
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
		count int
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
		for start := 0; start < 256; start += 16 {
			if start > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					report.err = ctx.Err()
					writerDone <- report
					return
				}
			}
			writes := make([]ResearchEventWrite, 16)
			for i := range writes {
				writes[i] = pinnedWrite(fmt.Sprintf("future-%03d", start+i), futureAt)
			}
			outcomes, err := s.PutResearchEventBatch(ctx, writes)
			if err != nil {
				report.err = fmt.Errorf("batch %d write: %w", start/16, err)
				writerDone <- report
				return
			}
			if len(outcomes) != 16 {
				report.err = fmt.Errorf("batch %d returned %d outcomes", start/16, len(outcomes))
				writerDone <- report
				return
			}
			for _, outcome := range outcomes {
				if outcome.Duplicate {
					report.err = fmt.Errorf("false duplicate in batch %d", start/16)
					writerDone <- report
					return
				}
			}
			report.count += 16
		}
		report.total = time.Since(began)
		writerDone <- report
	}()
	close(start)
	var result pinnedReaderTrial
	var firstError string
	for r := range results {
		if r.err != "" && firstError == "" {
			firstError = r.err
		}
		result.readCall = append(result.readCall, r.callNS)
		result.readOffer = append(result.readOffer, r.offerNS)
	}
	times := <-offerTimes
	for i := 1; i < len(times); i++ {
		result.offerGaps = append(result.offerGaps, times[i].Sub(times[i-1]).Nanoseconds())
	}
	writer := <-writerDone
	if writer.err != nil && firstError == "" {
		firstError = writer.err.Error()
	}
	result.writerTotal, result.lsn = writer.total, lsn
	if firstError != "" {
		t.Fatal("pinned reader trial", firstError)
	}
	if len(result.readCall) != 192 || writer.count != 256 || s.Snapshot(ctx).RuntimeVersion != base.RuntimeVersion+256 || !s.ResearchSnapshotCompatible(ctx, base, asOf) {
		t.Fatal("incomplete future-only reader trial", len(result.readCall), writer.count)
	}
	for version := base.RuntimeVersion + 1; version <= s.snapshot.RuntimeVersion; version++ {
		if !s.ingestMotion[version].Equal(futureAt) {
			t.Fatal("missing future-only motion", version)
		}
	}
	stats, err := s.db.TemporalStats(ctx)
	if err != nil || stats.ActiveLeaseCount < 1 {
		t.Fatal("missing pinned temporal lease", stats, err)
	}
	result.retainedBytes = stats.RetainedBytes
	if _, err := s.PutResearchEventBatch(ctx, []ResearchEventWrite{pinnedWrite("visible", asOf)}); err != nil {
		t.Fatal("visible guard control", err)
	}
	if s.ResearchSnapshotCompatible(ctx, base, asOf) {
		t.Fatal("pinned LSN was allowed to hide a visible mutation")
	}
	if pinned {
		rows, err := s.db.QueryWithParams(ctx, query, params)
		if err != nil || len(rows.Results) != 200 {
			t.Fatal("pinned LSN did not preserve historical view", err)
		}
	}
	lease.Close()
	stats, err = s.db.TemporalStats(ctx)
	if err != nil || stats.ActiveLeaseCount != 0 {
		t.Fatal("temporal snapshot lease leaked", stats, err)
	}
	return result
}

func TestResearchPinnedLSNReaderV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PINNED_LSN_READER_V1") != "1" {
		t.Skip("opt-in pinned-LSN reader design screen")
	}
	type totals struct {
		call, offer []int64
		writer      time.Duration
	}
	all := map[bool]*totals{false: {}, true: {}}
	for trial := 0; trial < 3; trial++ {
		order := []bool{false, true}
		if trial%2 == 1 {
			order = []bool{true, false}
		}
		for _, pinned := range order {
			r := runPinnedReaderTrial(t, pinned)
			a := all[pinned]
			a.call = append(a.call, r.readCall...)
			a.offer = append(a.offer, r.readOffer...)
			a.writer += r.writerTotal
			t.Logf("trial=%d pinned=%v reads=%d lsn=%d call_p50=%s call_p99=%s offer_p99=%s gap_p50=%s gap_p99=%s writer_total=%s retained_bytes=%d", trial, pinned, len(r.readCall), r.lsn, pinnedPercentile(r.readCall, .5), pinnedPercentile(r.readCall, .99), pinnedPercentile(r.readOffer, .99), pinnedPercentile(r.offerGaps, .5), pinnedPercentile(r.offerGaps, .99), r.writerTotal, r.retainedBytes)
		}
	}
	current, candidate := all[false], all[true]
	readRatio := float64(pinnedPercentile(candidate.call, .99)) / float64(pinnedPercentile(current.call, .99))
	writerRatio := float64(candidate.writer) / float64(current.writer)
	t.Logf("pooled current_call_p99=%s pinned_call_p99=%s read_ratio=%.3f current_offer_p99=%s pinned_offer_p99=%s writer_total_ratio=%.3f", pinnedPercentile(current.call, .99), pinnedPercentile(candidate.call, .99), readRatio, pinnedPercentile(current.offer, .99), pinnedPercentile(candidate.offer, .99), writerRatio)
	if len(current.call) != 576 || len(candidate.call) != 576 || readRatio > 1.10 || writerRatio > 1.25 {
		t.Error("pinned-LSN reader failed frozen design screen")
	}
}
