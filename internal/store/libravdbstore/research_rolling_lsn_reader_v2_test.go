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

type rollingV2Read struct {
	callNS, offerNS int64
	offered         time.Time
	ended           time.Time
	maxBatch        int
	err             error
}

type rollingV2Ack struct {
	ended time.Time
	batch int
}

type rollingV2Trial struct {
	call, offer, gaps, lags, writeAcks []int64
	writerTotal                        time.Duration
	retainedBytes                      uint64
	maxLag                             time.Duration
}

func rollingV2Write(id string, at time.Time, vec []float32) ResearchEventWrite {
	w := pinnedWrite(id, at)
	w.Vector = vec
	return w
}

func runRollingReaderV2Trial(t *testing.T, pinned bool) rollingV2Trial {
	return runRollingReaderTrialWithKey(t, pinned, false, true)
}

func runRollingReaderTrialWithKey(t *testing.T, pinned, sortable, requirePostAck bool) rollingV2Trial {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	config := Config{Path: t.TempDir() + "/rolling-v2.libravdb", Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true}
	s, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if s != nil {
			s.Close()
		}
	}()
	name := collectionName("tenant-a", "research:d4")
	schema := libra.MetadataSchema{"available_at": libra.StringField}
	if sortable {
		schema["available_at_sort"] = libra.StringField
	}
	col, err := s.db.CreateCollection(ctx, name,
		libra.WithDimension(4), libra.WithMetric(libra.CosineDistance),
		libra.WithHNSW(16, 200, 100), libra.WithMemoryMapping(true),
		libra.WithMetadataSchema(schema))
	if err != nil {
		t.Fatal("create private schema-backed collection", err)
	}
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	seedAt, futureAt := asOf.Add(-time.Minute), asOf.Add(time.Hour)
	availabilityField, availabilityParameter := "available_at", "available_by"
	availabilityValue := asOf.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
	put := s.PutResearchEventBatch
	if sortable {
		asOf = asOf.Add(120 * time.Millisecond)
		seedAt, futureAt = asOf.Add(-20*time.Millisecond), asOf.Add(5*time.Millisecond)
		availabilityField, availabilityParameter = "available_at_sort", "available_by_sort"
		availabilityValue = researchAvailabilitySortKey(asOf)
		put = s.PutResearchEventBatchSortable
	}
	for start := 0; start < 200; start += 16 {
		writes := make([]ResearchEventWrite, 0, 16)
		for i := start; i < start+16 && i < 200; i++ {
			writes = append(writes, rollingV2Write(fmt.Sprintf("base-%03d", i), seedAt, []float32{0, 1, 0, 0}))
		}
		if _, err := put(ctx, writes); err != nil {
			t.Fatal("seed base", err)
		}
	}
	future := make([]ResearchEventWrite, 16)
	for i := range future {
		future[i] = rollingV2Write(fmt.Sprintf("future-%03d", i), futureAt, []float32{1, 0, 0, 0})
	}
	if _, err := put(ctx, future); err != nil {
		t.Fatal("seed future", err)
	}
	base := s.Snapshot(ctx)
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + name +
		" AS OF LSN $snapshot_lsn d WHERE " + availabilityField + " <= $" + availabilityParameter + " ORDER BY distance LIMIT 50"
	read := func() rollingV2Read {
		began := time.Now()
		result := rollingV2Read{maxBatch: -1}
		var ids []string
		var captured uint64
		if pinned {
			s.writeMu.RLock()
			captured = s.snapshot.RuntimeVersion
			lsn, lsnErr := s.db.LatestCommitLSN(ctx)
			var lease *libra.TemporalSnapshot
			if lsnErr == nil {
				lease, lsnErr = s.db.SnapshotAtLSN(ctx, lsn)
			}
			s.writeMu.RUnlock()
			if lsnErr != nil {
				result.err = fmt.Errorf("paired snapshot/LSN capture: %w", lsnErr)
			} else {
				rows, queryErr := s.db.QueryWithParams(ctx, query, libra.QueryParams{
					"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
					availabilityParameter: availabilityValue,
				})
				lease.Close()
				if queryErr != nil {
					result.err = fmt.Errorf("as-of SQL: %w", queryErr)
				} else {
					ids = make([]string, len(rows.Results))
					for i, row := range rows.Results {
						ids[i] = row.ID
					}
				}
			}
		} else {
			rows, searchErr := s.Search(ctx, "tenant-a", []float32{1, 0, 0, 0}, asOf, 50)
			if searchErr != nil {
				result.err = fmt.Errorf("current Search: %w", searchErr)
			} else {
				ids = make([]string, len(rows))
				for i, row := range rows {
					ids[i] = row.Event.ID
				}
			}
		}
		if result.err == nil {
			if len(ids) != 50 {
				result.err = fmt.Errorf("returned %d of 50 eligible rows", len(ids))
			} else {
				seen := make(map[string]bool, len(ids))
				for _, id := range ids {
					if seen[id] {
						result.err = fmt.Errorf("duplicate ID %s", id)
						break
					}
					seen[id] = true
					switch {
					case strings.HasPrefix(id, "base-"):
					case strings.HasPrefix(id, "visible-"):
						i, parseErr := strconv.Atoi(strings.TrimPrefix(id, "visible-"))
						if parseErr != nil || i < 0 || i >= 256 || pinned && uint64(i) >= captured-base.RuntimeVersion {
							result.err = fmt.Errorf("visible ID outside captured view: %s captured=%d", id, captured)
						} else if i/16 > result.maxBatch {
							result.maxBatch = i / 16
						}
					default:
						result.err = fmt.Errorf("future or unknown ID %s", id)
					}
					if result.err != nil {
						break
					}
				}
			}
		}
		result.ended = time.Now()
		result.callNS = result.ended.Sub(began).Nanoseconds()
		return result
	}
	for i := 0; i < 4; i++ {
		if r := read(); r.err != nil || r.maxBatch != -1 {
			t.Fatal("warm read", r.err, r.maxBatch)
		}
	}
	jobs := make(chan time.Time, 192)
	results := make(chan rollingV2Read, 192)
	offers := make(chan []time.Time, 1)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for offered := range jobs {
				r := read()
				r.offered = offered
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
		defer func() { offers <- times }()
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
		acks  []rollingV2Ack
		costs []int64
		total time.Duration
		err   error
	}
	written := make(chan writeReport, 1)
	go func() {
		<-start
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		report := writeReport{}
		began := time.Now()
		for b := 0; b < 16; b++ {
			if b > 0 {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					report.err = ctx.Err()
					written <- report
					return
				}
			}
			offered := time.Now()
			batch := make([]ResearchEventWrite, 16)
			for i := range batch {
				batch[i] = rollingV2Write(fmt.Sprintf("visible-%03d", b*16+i), asOf,
					[]float32{1, float32(0.8 - 0.04*float64(b)), 0, 0})
			}
			outcomes, writeErr := put(ctx, batch)
			if writeErr != nil || len(outcomes) != 16 {
				report.err = fmt.Errorf("batch %d: outcomes=%d err=%v", b, len(outcomes), writeErr)
				written <- report
				return
			}
			for _, outcome := range outcomes {
				if outcome.Duplicate {
					report.err = fmt.Errorf("duplicate in batch %d", b)
					written <- report
					return
				}
			}
			ack := time.Now()
			report.acks = append(report.acks, rollingV2Ack{ack, b})
			report.costs = append(report.costs, ack.Sub(offered).Nanoseconds())
		}
		report.total = time.Since(began)
		written <- report
	}()
	close(start)
	allReads := make([]rollingV2Read, 0, 192)
	var firstErr error
	for r := range results {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		allReads = append(allReads, r)
	}
	offerTimes := <-offers
	writer := <-written
	if writer.err != nil && firstErr == nil {
		firstErr = writer.err
	}
	if firstErr != nil {
		t.Fatal("rolling v2 trial pinned=", pinned, firstErr)
	}
	if len(allReads) != 192 || len(offerTimes) != 192 || len(writer.acks) != 16 || s.Snapshot(ctx).RuntimeVersion != base.RuntimeVersion+256 {
		t.Fatal("incomplete rolling v2 trial pinned=", pinned, len(allReads), len(offerTimes), len(writer.acks))
	}
	for version := base.RuntimeVersion + 1; version <= base.RuntimeVersion+256; version++ {
		if !s.ingestMotion[version].Equal(asOf) {
			t.Fatal("missing visible motion", version)
		}
	}
	final := read()
	if final.err != nil || final.maxBatch != 15 {
		t.Fatal("final read hid latest batch pinned=", pinned, final.err, final.maxBatch)
	}
	stats, err := s.db.TemporalStats(ctx)
	if err != nil || stats.ActiveLeaseCount != 0 {
		t.Fatal("temporal lease leaked pinned=", pinned, stats, err)
	}
	out := rollingV2Trial{writerTotal: writer.total, writeAcks: writer.costs, retainedBytes: stats.RetainedBytes}
	for i := 1; i < len(offerTimes); i++ {
		out.gaps = append(out.gaps, offerTimes[i].Sub(offerTimes[i-1]).Nanoseconds())
	}
	for _, r := range allReads {
		out.call = append(out.call, r.callNS)
		out.offer = append(out.offer, r.offerNS)
	}
	for _, ack := range writer.acks {
		var first time.Time
		for _, r := range allReads {
			if !r.offered.Before(ack.ended) && r.maxBatch >= ack.batch && (first.IsZero() || r.ended.Before(first)) {
				first = r.ended
			}
		}
		if first.IsZero() {
			if requirePostAck {
				t.Fatal("no read reflected acknowledged batch pinned=", pinned, ack.batch)
			}
			continue
		}
		lag := first.Sub(ack.ended)
		out.lags = append(out.lags, lag.Nanoseconds())
		if lag > out.maxLag {
			out.maxLag = lag
		}
	}
	if sortable {
		records, err := col.ListAll(ctx)
		if err != nil || len(records) != 472 {
			t.Fatal("incomplete sortable collection", len(records), err)
		}
		for _, record := range records {
			event, decodeErr := decodeStoredEvent(record.Metadata, true)
			if decodeErr != nil || event.ID != record.ID || record.Metadata["available_at_sort"] != researchAvailabilitySortKey(event.AvailableAt) {
				t.Fatal("sort key differs from durable EventFrame", record.ID, decodeErr)
			}
		}
		if err := s.Close(); err != nil {
			t.Fatal("close sortable store", err)
		}
		s = nil
		reopened, err := Open(config)
		if err != nil {
			t.Fatal("reopen sortable store", err)
		}
		defer reopened.Close()
		lsn, err := reopened.db.LatestCommitLSN(ctx)
		if err != nil {
			t.Fatal("reopened LSN", err)
		}
		rows, err := reopened.db.QueryWithParams(ctx, query, libra.QueryParams{
			"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
			availabilityParameter: availabilityValue,
		})
		if err != nil {
			t.Fatal("reopened sortable query", err)
		}
		if len(rows.Results) != 50 {
			t.Fatal("reopened sortable query returned", len(rows.Results), "rows")
		}
		latest := false
		for _, row := range rows.Results {
			if strings.HasPrefix(row.ID, "future-") {
				t.Fatal("reopened query leaked future row", row.ID)
			}
			if strings.HasPrefix(row.ID, "visible-") {
				i, parseErr := strconv.Atoi(strings.TrimPrefix(row.ID, "visible-"))
				if parseErr == nil && i >= 240 && i < 256 {
					latest = true
				}
			}
		}
		if !latest {
			t.Fatal("reopened sortable query hid latest batch")
		}
	}
	return out
}

func TestResearchRollingLSNReaderV2(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_ROLLING_LSN_READER_V2") != "1" {
		t.Skip("opt-in rolling LSN captured-view screen")
	}
	type totals struct {
		call, offer, lags []int64
		writer            time.Duration
	}
	all := map[bool]*totals{false: {}, true: {}}
	for trial := 0; trial < 3; trial++ {
		order := []bool{false, true}
		if trial%2 == 1 {
			order = []bool{true, false}
		}
		for _, pinned := range order {
			r := runRollingReaderV2Trial(t, pinned)
			a := all[pinned]
			a.call = append(a.call, r.call...)
			a.offer = append(a.offer, r.offer...)
			a.lags = append(a.lags, r.lags...)
			a.writer += r.writerTotal
			t.Logf("trial=%d pinned=%v call_p50=%s call_p99=%s offer_p99=%s gap_p50=%s gap_p99=%s writer_total=%s write_ack_p99=%s lag_p99=%s max_lag=%s retained_bytes=%d",
				trial, pinned, pinnedPercentile(r.call, .5), pinnedPercentile(r.call, .99), pinnedPercentile(r.offer, .99),
				pinnedPercentile(r.gaps, .5), pinnedPercentile(r.gaps, .99), r.writerTotal, pinnedPercentile(r.writeAcks, .99),
				pinnedPercentile(r.lags, .99), r.maxLag, r.retainedBytes)
		}
	}
	current, candidate := all[false], all[true]
	readP99 := pinnedPercentile(candidate.call, .99)
	readRatio := float64(readP99) / float64(pinnedPercentile(current.call, .99))
	writerRatio := float64(candidate.writer) / float64(current.writer)
	lagP99 := pinnedPercentile(candidate.lags, .99)
	t.Logf("pooled current_call_p99=%s rolling_call_p99=%s read_ratio=%.3f current_offer_p99=%s rolling_offer_p99=%s writer_total_ratio=%.3f current_lag_p99=%s rolling_lag_p99=%s",
		pinnedPercentile(current.call, .99), readP99, readRatio, pinnedPercentile(current.offer, .99),
		pinnedPercentile(candidate.offer, .99), writerRatio, pinnedPercentile(current.lags, .99), lagP99)
	if len(current.call) != 576 || len(candidate.call) != 576 || len(candidate.lags) != 48 || readRatio > 1.10 ||
		readP99 > 100*time.Millisecond || writerRatio > 1.25 || lagP99 > 25*time.Millisecond {
		t.Error("rolling-LSN v2 failed frozen captured-view design screen")
	}
}
