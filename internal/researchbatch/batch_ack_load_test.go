package researchbatch

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

type batchAckJob struct {
	write   libravdbstore.ResearchEventWrite
	offered time.Time
}

type batchAckResult struct {
	id, err         string
	offerNS, callNS int64
	duplicate       bool
	ackSnapshot     model.Snapshot
}

type batchReadJob struct{ offered time.Time }
type batchReadResult struct {
	offerNS, callNS int64
	err             string
}

type batchAckLoadResult struct {
	ackOffer, ackCall, readOffer, readCall []int64
	writeGaps, readGaps                    []int64
	batchSizes                             []int
	writerTotal                            time.Duration
	readCount, writes, duplicates          int
}

func percentileDuration(values []int64, fraction float64) time.Duration {
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

func gaps(times []time.Time) []int64 {
	if len(times) < 2 {
		return nil
	}
	result := make([]int64, 0, len(times)-1)
	for i := 1; i < len(times); i++ {
		result = append(result, times[i].Sub(times[i-1]).Nanoseconds())
	}
	return result
}

func runBatchAckLoad(t *testing.T, maxBatch int) batchAckLoadResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	p := openPrototypeForTest(t, t.TempDir(), true)
	defer p.close()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	seedAt, futureAt := asOf.Add(-time.Minute), asOf.Add(time.Hour)
	for i := 0; i < 200; i += 16 {
		writes := make([]libravdbstore.ResearchEventWrite, 0, 16)
		for j := i; j < i+16 && j < 200; j++ {
			writes = append(writes, prototypeWrite(fmt.Sprintf("base-%03d", j), seedAt))
		}
		if _, err := p.apply(ctx, writes, ""); err != nil {
			t.Fatal("seed batch", err)
		}
	}
	base := p.snapshot
	writeJobs := make(chan batchAckJob, 64)
	acks := make(chan batchAckResult, 256)
	writerSizes := make(chan []int, 1)
	writeOffers := make(chan []time.Time, 1)
	go func() {
		var sizes []int
		var stopped error
		for first := range writeJobs {
			batch := []batchAckJob{first}
			closed := false
			if maxBatch > 1 {
				timer := time.NewTimer(4 * time.Millisecond)
			collect:
				for len(batch) < maxBatch {
					select {
					case job, ok := <-writeJobs:
						if !ok {
							closed = true
							break collect
						}
						batch = append(batch, job)
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
			sizes = append(sizes, len(batch))
			writes := make([]libravdbstore.ResearchEventWrite, len(batch))
			for i := range batch {
				writes[i] = batch[i].write
			}
			started := time.Now()
			var outcomes []bool
			var snapshot model.Snapshot
			if stopped == nil {
				results, err := p.apply(ctx, writes, "")
				if err != nil {
					stopped = err
				} else {
					outcomes = make([]bool, len(results))
					for i := range results {
						outcomes[i] = results[i].Duplicate
					}
					if len(results) != len(batch) {
						stopped = fmt.Errorf("short acknowledgement: %d of %d", len(results), len(batch))
					} else {
						snapshot = results[0].Snapshot
					}
				}
			}
			callNS := time.Since(started).Nanoseconds()
			for i, job := range batch {
				result := batchAckResult{id: job.write.Event.ID, offerNS: time.Since(job.offered).Nanoseconds(), callNS: callNS, ackSnapshot: snapshot}
				if stopped != nil {
					result.err = stopped.Error()
				} else {
					result.duplicate = outcomes[i]
				}
				acks <- result
			}
			if closed {
				break
			}
		}
		writerSizes <- sizes
		close(acks)
	}()
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()
		defer close(writeJobs)
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
			at := time.Now()
			times = append(times, at)
			select {
			case writeJobs <- batchAckJob{write: prototypeWrite(fmt.Sprintf("future-%03d", i), futureAt), offered: at}:
			case <-ctx.Done():
				return
			}
		}
	}()
	readJobs := make(chan batchReadJob, 192)
	readResults := make(chan batchReadResult, 192)
	readOffers := make(chan []time.Time, 1)
	var readers sync.WaitGroup
	for i := 0; i < 8; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for job := range readJobs {
				start := time.Now()
				found, err := p.backend.Search(ctx, "tenant-a", []float32{1, 0, 0, 0}, asOf, 200)
				result := batchReadResult{callNS: time.Since(start).Nanoseconds(), offerNS: time.Since(job.offered).Nanoseconds()}
				if err != nil {
					result.err = err.Error()
				} else if len(found) == 0 {
					result.err = "empty past-available search"
				} else {
					for _, item := range found {
						if item.Event.AvailableAt.After(asOf) {
							result.err = "future event leaked into as-of read"
							break
						}
					}
				}
				readResults <- result
			}
		}()
	}
	go func() {
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		defer close(readJobs)
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
			at := time.Now()
			times = append(times, at)
			select {
			case readJobs <- batchReadJob{offered: at}:
			case <-ctx.Done():
				return
			}
		}
	}()
	go func() { readers.Wait(); close(readResults) }()
	var result batchAckLoadResult
	var firstOffer time.Time
	var firstError string
	for ack := range acks {
		if ack.err != "" && firstError == "" {
			firstError = "batch acknowledgement " + ack.id + ": " + ack.err
		}
		if (ack.duplicate || ack.ackSnapshot.RuntimeVersion <= base.RuntimeVersion) && firstError == "" {
			firstError = "false or duplicate acknowledgement " + ack.id
		}
		result.writes++
		result.ackOffer = append(result.ackOffer, ack.offerNS)
		result.ackCall = append(result.ackCall, ack.callNS)
	}
	result.batchSizes = <-writerSizes
	writeTimes := <-writeOffers
	if len(writeTimes) > 0 {
		firstOffer = writeTimes[0]
		result.writerTotal = time.Since(firstOffer)
	}
	result.writeGaps = gaps(writeTimes)
	for read := range readResults {
		if read.err != "" && firstError == "" {
			firstError = "concurrent search: " + read.err
		}
		result.readCount++
		result.readOffer = append(result.readOffer, read.offerNS)
		result.readCall = append(result.readCall, read.callNS)
	}
	result.readGaps = gaps(<-readOffers)
	if firstError != "" {
		t.Fatal(firstError)
	}
	if len(writeTimes) != 256 || result.writes != 256 || result.readCount != 192 || p.snapshot.RuntimeVersion != base.RuntimeVersion+256 || !p.asOf(ctx, base, asOf) {
		t.Fatal("incomplete loaded trial", len(writeTimes), result.writes, result.readCount, p.snapshot, base)
	}
	for version := base.RuntimeVersion + 1; version <= p.snapshot.RuntimeVersion; version++ {
		if !p.motion[version].Equal(futureAt) {
			t.Fatal("missing future-only motion", version)
		}
	}
	return result
}

func batchSizeDistribution(sizes []int) []int {
	counts := make([]int, 17)
	for _, size := range sizes {
		if size >= 1 && size <= 16 {
			counts[size]++
		}
	}
	return counts
}

func TestBatchIntentAckLoadV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_BATCH_ACK_LOAD_V1") != "1" {
		t.Skip("opt-in loaded batch acknowledgement screen")
	}
	type arm struct {
		ack, read []int64
		total     time.Duration
	}
	all := map[int]*arm{1: {}, 16: {}}
	for trial := 0; trial < 3; trial++ {
		order := []int{1, 16}
		if trial%2 == 1 {
			order = []int{16, 1}
		}
		for _, maxBatch := range order {
			r := runBatchAckLoad(t, maxBatch)
			a := all[maxBatch]
			a.ack = append(a.ack, r.ackOffer...)
			a.read = append(a.read, r.readCall...)
			a.total += r.writerTotal
			t.Logf("trial=%d batch=%d writes=%d reads=%d batches=%d size_counts_1_to_16=%v ack_p50=%s ack_p99=%s call_p99=%s read_call_p99=%s read_offer_p99=%s writer_total=%s write_gap_p50=%s write_gap_p99=%s read_gap_p50=%s read_gap_p99=%s", trial, maxBatch, r.writes, r.readCount, len(r.batchSizes), batchSizeDistribution(r.batchSizes)[1:], percentileDuration(r.ackOffer, .5), percentileDuration(r.ackOffer, .99), percentileDuration(r.ackCall, .99), percentileDuration(r.readCall, .99), percentileDuration(r.readOffer, .99), r.writerTotal, percentileDuration(r.writeGaps, .5), percentileDuration(r.writeGaps, .99), percentileDuration(r.readGaps, .5), percentileDuration(r.readGaps, .99))
		}
	}
	control, candidate := all[1], all[16]
	ackRatio := float64(percentileDuration(candidate.ack, .99)) / float64(percentileDuration(control.ack, .99))
	readRatio := float64(percentileDuration(candidate.read, .99)) / float64(percentileDuration(control.read, .99))
	totalRatio := float64(candidate.total) / float64(control.total)
	t.Logf("pooled ack_p99_control=%s candidate=%s ratio=%.3f read_call_p99_control=%s candidate=%s ratio=%.3f writer_total_control=%s candidate=%s ratio=%.3f", percentileDuration(control.ack, .99), percentileDuration(candidate.ack, .99), ackRatio, percentileDuration(control.read, .99), percentileDuration(candidate.read, .99), readRatio, control.total, candidate.total, totalRatio)
	if len(control.ack) != 768 || len(candidate.ack) != 768 || len(control.read) != 576 || len(candidate.read) != 576 || totalRatio >= 1 || ackRatio > 1.25 || readRatio > 1.10 {
		t.Error("frozen batch acknowledgement load screen failed")
	}
}
