package researchbatch

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

type readerFactorResult struct {
	callNS, offerNS, offerGaps []int64
	writerTotal                time.Duration
	writes                     int
}

func runBatchReaderFactor(t *testing.T, cell string) readerFactorResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p := openPrototypeForTest(t, t.TempDir(), true)
	defer p.close()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	seedAt, futureAt := asOf.Add(-time.Minute), asOf.Add(time.Hour)
	writeBatch := func(start int, tenant string) error {
		writes := make([]libravdbstore.ResearchEventWrite, 0, 16)
		for i := start; i < start+16; i++ {
			write := prototypeWrite(fmt.Sprintf("future-%03d", i), futureAt)
			write.Event.TenantID = tenant
			writes = append(writes, write)
		}
		_, err := p.apply(ctx, writes, "")
		return err
	}
	for start := 0; start < 200; start += 16 {
		writes := make([]libravdbstore.ResearchEventWrite, 0, 16)
		for i := start; i < start+16 && i < 200; i++ {
			writes = append(writes, prototypeWrite(fmt.Sprintf("base-%03d", i), seedAt))
		}
		if _, err := p.apply(ctx, writes, ""); err != nil {
			t.Fatal("seed event", err)
		}
	}
	base := p.snapshot
	if cell == "expanded_static" {
		for start := 0; start < 256; start += 16 {
			if err := writeBatch(start, "tenant-a"); err != nil {
				t.Fatal("preinsert future events", err)
			}
		}
	}
	readJobs := make(chan time.Time, 192)
	readOut := make(chan batchReadResult, 192)
	readTimes := make(chan []time.Time, 1)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for offered := range readJobs {
				began := time.Now()
				found, err := p.backend.Search(ctx, "tenant-a", []float32{1, 0, 0, 0}, asOf, 200)
				result := batchReadResult{callNS: time.Since(began).Nanoseconds(), offerNS: time.Since(offered).Nanoseconds()}
				if err != nil {
					result.err = err.Error()
				} else if len(found) != 200 {
					result.err = fmt.Sprintf("past eligible count %d, want 200", len(found))
				} else {
					for _, item := range found {
						if item.Event.TenantID != "tenant-a" || !strings.HasPrefix(item.Event.ID, "base-") || item.Event.AvailableAt.After(asOf) {
							result.err = "future or other-tenant event in as-of result"
							break
						}
					}
				}
				readOut <- result
			}
		}()
	}
	go func() {
		<-start
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		defer close(readJobs)
		times := make([]time.Time, 0, 192)
		defer func() { readTimes <- times }()
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
			case readJobs <- at:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { workers.Wait(); close(readOut) }()
	type writeReport struct {
		count int
		total time.Duration
		err   error
	}
	writerDone := make(chan writeReport, 1)
	if cell == "other_tenant_writer" || cell == "same_tenant_writer" {
		go func() {
			<-start
			ticker := time.NewTicker(16 * time.Millisecond)
			defer ticker.Stop()
			report := writeReport{}
			began := time.Now()
			tenant := "tenant-a"
			if cell == "other_tenant_writer" {
				tenant = "tenant-b"
			}
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
				if err := writeBatch(start, tenant); err != nil {
					report.err = err
					writerDone <- report
					return
				}
				report.count += 16
			}
			report.total = time.Since(began)
			writerDone <- report
		}()
	}
	close(start)
	var result readerFactorResult
	var firstError string
	for read := range readOut {
		if read.err != "" && firstError == "" {
			firstError = read.err
		}
		result.callNS = append(result.callNS, read.callNS)
		result.offerNS = append(result.offerNS, read.offerNS)
	}
	result.offerGaps = gaps(<-readTimes)
	if cell == "other_tenant_writer" || cell == "same_tenant_writer" {
		report := <-writerDone
		if report.err != nil && firstError == "" {
			firstError = report.err.Error()
		}
		result.writes, result.writerTotal = report.count, report.total
	}
	if firstError != "" {
		t.Fatal(cell, firstError)
	}
	expectedWrites := 0
	if cell != "base_static" {
		expectedWrites = 256
	}
	if len(result.callNS) != 192 || result.writes != expectedWrites && cell != "expanded_static" || p.snapshot.RuntimeVersion != base.RuntimeVersion+uint64(expectedWrites) || !p.asOf(ctx, base, asOf) {
		t.Fatal("incomplete reader factor cell", cell, len(result.callNS), result.writes, p.snapshot, base)
	}
	for version := base.RuntimeVersion + 1; version <= p.snapshot.RuntimeVersion; version++ {
		if !p.motion[version].Equal(futureAt) {
			t.Fatal("missing future-only motion", cell, version)
		}
	}
	return result
}

func TestBatchReaderFactorsV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_BATCH_READER_FACTOR_V1") != "1" {
		t.Skip("opt-in reader-factor diagnostic")
	}
	orders := [][]string{
		{"base_static", "expanded_static", "other_tenant_writer", "same_tenant_writer"},
		{"same_tenant_writer", "other_tenant_writer", "expanded_static", "base_static"},
		{"expanded_static", "base_static", "same_tenant_writer", "other_tenant_writer"},
	}
	pooled := make(map[string][]int64)
	for block, order := range orders {
		for _, cell := range order {
			r := runBatchReaderFactor(t, cell)
			pooled[cell] = append(pooled[cell], r.callNS...)
			t.Logf("block=%d cell=%s searches=%d writes=%d call_p50=%s call_p99=%s offer_p99=%s offer_gap_p50=%s offer_gap_p99=%s writer_total=%s", block, cell, len(r.callNS), r.writes, percentileDuration(r.callNS, .5), percentileDuration(r.callNS, .99), percentileDuration(r.offerNS, .99), percentileDuration(r.offerGaps, .5), percentileDuration(r.offerGaps, .99), r.writerTotal)
		}
	}
	base := percentileDuration(pooled["base_static"], .99)
	for _, cell := range orders[0] {
		p99 := percentileDuration(pooled[cell], .99)
		t.Logf("pooled cell=%s call_p99=%s ratio_to_base=%.3f", cell, p99, float64(p99)/float64(base))
	}
}
