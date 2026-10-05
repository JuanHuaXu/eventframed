package libravdbstore

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
)

type denseOracleRowV7 struct {
	id    string
	angle float64
}

func sortedDenseOracleV7(rows []denseOracleRowV7) []denseOracleRowV7 {
	copyRows := append([]denseOracleRowV7(nil), rows...)
	sort.Slice(copyRows, func(i, j int) bool {
		if copyRows[i].angle == copyRows[j].angle {
			return copyRows[i].id < copyRows[j].id
		}
		return copyRows[i].angle < copyRows[j].angle
	})
	return copyRows
}

type denseReadV7 struct {
	offered, started, ended time.Time
	lsn                     uint64
	viewAge                 time.Duration
	items                   []store.SearchResult
	err                     error
}

type denseWriteV7 struct {
	ackAges []int64
	groups  []int
	err     error
}

func runDensePublishedLoadV7(t *testing.T, k int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	query := denseQueryV6()
	rawQuery := make([]float32, len(query))
	for i, value := range query {
		rawQuery[i] = 2 * value
	}
	gate := createDensePublishedGateV6(t, t.TempDir(), query)
	defer gate.close()
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
		if err := appendDenseV6(ctx, gate, batch); err != nil {
			t.Fatal(err)
		}
	}
	for start := 0; start < 16; start += 16 {
		batch := make([]ResearchEventWrite, 0, 16)
		for i := start; i < start+16; i++ {
			write := pinnedWrite(fmt.Sprintf("future-%03d", i), asOf.Add(5*time.Millisecond))
			write.Vector = denseRowV6(query, i, 0.001*float64(i+1))
			batch = append(batch, write)
		}
		if err := appendDenseV6(ctx, gate, batch); err != nil {
			t.Fatal(err)
		}
	}
	reader, err := newSortPublishedReader(ctx, gate)
	if err != nil {
		t.Fatal(err)
	}
	oracleByLSN := map[uint64][]denseOracleRowV7{reader.current.Load().lsn: sortedDenseOracleV7(allRows)}
	writeJobs := make(chan sortLoadWriteJob, 160)
	readJobs := make(chan time.Time, 160)
	writeOffers := make(chan []time.Time, 1)
	readOffers := make(chan []time.Time, 1)
	writeDone := make(chan denseWriteV7, 1)
	readReports := make(chan denseReadV7, 160)
	start := make(chan struct{})
	go func() {
		<-start
		defer close(writeJobs)
		ticker := time.NewTicker(4 * time.Millisecond)
		defer ticker.Stop()
		times := make([]time.Time, 0, 160)
		defer func() { writeOffers <- times }()
		for i := 0; i < 160; i++ {
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
		times := make([]time.Time, 0, 160)
		defer func() { readOffers <- times }()
		for i := 0; i < 160; i++ {
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
		report := denseWriteV7{}
		for {
			group, closed := sortLoadTakeGroup(writeJobs, 16, 16*time.Millisecond)
			if len(group) == 0 {
				writeDone <- report
				return
			}
			writes := make([]ResearchEventWrite, len(group))
			for i, job := range group {
				writes[i] = job.write
			}
			err := appendDenseV6(ctx, gate, writes)
			if err == nil {
				for _, job := range group {
					index := strings.TrimPrefix(job.write.Event.ID, "visible-")
					var i int
					if _, scanErr := fmt.Sscanf(index, "%d", &i); scanErr != nil {
						err = scanErr
						break
					}
					allRows = append(allRows, denseOracleRowV7{job.write.Event.ID, 0.00213 * float64(i+1)})
				}
				if err == nil {
					oracleByLSN[gate.verifiedLSN] = sortedDenseOracleV7(allRows)
					err = reader.publish(ctx)
				}
			}
			if err != nil {
				reader.poison()
				report.err = err
				writeDone <- report
				return
			}
			acked := time.Now()
			report.groups = append(report.groups, len(group))
			for _, job := range group {
				report.ackAges = append(report.ackAges, acked.Sub(job.offered).Nanoseconds())
			}
			if closed {
				writeDone <- report
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
			for offered := range readJobs {
				began := time.Now()
				view := reader.current.Load()
				result := denseReadV7{offered: offered, started: began}
				if view == nil {
					result.err = fmt.Errorf("no published view")
				} else {
					result.lsn = view.lsn
					result.items, result.err = queryDensePublishedV6(ctx, gate, view.lsn, asOf, rawQuery, k)
					result.viewAge = time.Since(view.at)
					if result.viewAge < 0 || result.viewAge >= 250*time.Millisecond {
						result.err = fmt.Errorf("published view age %s", result.viewAge)
					}
				}
				result.ended = time.Now()
				readReports <- result
			}
		}()
	}
	go func() { readers.Wait(); close(readReports) }()
	close(start)
	readCallNS, readOfferNS, viewAgeNS := make([]int64, 0, 160), make([]int64, 0, 160), make([]int64, 0, 160)
	reports := make([]denseReadV7, 0, 160)
	for report := range readReports {
		reports = append(reports, report)
		readCallNS = append(readCallNS, report.ended.Sub(report.started).Nanoseconds())
		readOfferNS = append(readOfferNS, report.ended.Sub(report.offered).Nanoseconds())
		viewAgeNS = append(viewAgeNS, report.viewAge.Nanoseconds())
	}
	writer := <-writeDone
	writeTimes := <-writeOffers
	readTimes := <-readOffers
	violations := make([]string, 0)
	if writer.err != nil {
		violations = append(violations, "writer: "+writer.err.Error())
	}
	if len(writer.ackAges) != 160 || len(reports) != 160 || len(writeTimes) != 160 || len(readTimes) != 160 {
		violations = append(violations, fmt.Sprintf("count ack=%d read=%d write_offer=%d read_offer=%d", len(writer.ackAges), len(reports), len(writeTimes), len(readTimes)))
	}
	for _, report := range reports {
		if report.err != nil {
			violations = append(violations, "read: "+report.err.Error())
			continue
		}
		expected, ok := oracleByLSN[report.lsn]
		if !ok || len(expected) < k || len(report.items) != k {
			violations = append(violations, fmt.Sprintf("lsn=%d expected/actual count %d/%d", report.lsn, len(expected), len(report.items)))
			continue
		}
		want := make(map[string]float64, k)
		for _, row := range expected[:k] {
			want[row.id] = row.angle
		}
		seen := make(map[string]bool, k)
		previous := math.Inf(1)
		for _, item := range report.items {
			angle, exists := want[item.Event.ID]
			if !exists || seen[item.Event.ID] || item.Similarity > previous+1e-5 ||
				math.Abs(item.Similarity-math.Cos(angle)) > 1e-4 ||
				item.Event.Content != "public vector search fixture" ||
				strings.HasPrefix(item.Event.ID, "future") {
				violations = append(violations, fmt.Sprintf("lsn=%d bad row %s score=%g", report.lsn, item.Event.ID, item.Similarity))
			}
			seen[item.Event.ID] = true
			previous = item.Similarity
		}
		for id := range want {
			if !seen[id] {
				violations = append(violations, fmt.Sprintf("lsn=%d omitted %s", report.lsn, id))
			}
		}
	}
	writeP99 := pinnedPercentile(writer.ackAges, .99)
	readCallP99 := pinnedPercentile(readCallNS, .99)
	readOfferP99 := pinnedPercentile(readOfferNS, .99)
	maxAge := pinnedPercentile(viewAgeNS, 1)
	if writeP99 >= 250*time.Millisecond || readCallP99 >= 100*time.Millisecond ||
		readOfferP99 >= 100*time.Millisecond || maxAge >= 250*time.Millisecond {
		violations = append(violations, fmt.Sprintf("latency/freshness write_p99=%s read_call_p99=%s read_offer_p99=%s age_max=%s", writeP99, readCallP99, readOfferP99, maxAge))
	}
	t.Logf("k=%d writes=%d groups=%d reads=%d violations=%d write_gap_p50=%s write_gap_p99=%s read_gap_p50=%s read_gap_p99=%s write_age_p50=%s write_age_p99=%s read_call_p50=%s read_call_p99=%s read_offer_p50=%s read_offer_p99=%s view_age_p50=%s view_age_max=%s",
		k, len(writer.ackAges), len(writer.groups), len(reports), len(violations),
		pinnedPercentile(sortLoadOfferGaps(writeTimes), .5), pinnedPercentile(sortLoadOfferGaps(writeTimes), .99),
		pinnedPercentile(sortLoadOfferGaps(readTimes), .5), pinnedPercentile(sortLoadOfferGaps(readTimes), .99),
		pinnedPercentile(writer.ackAges, .5), writeP99,
		pinnedPercentile(readCallNS, .5), readCallP99,
		pinnedPercentile(readOfferNS, .5), readOfferP99,
		pinnedPercentile(viewAgeNS, .5), maxAge)
	if len(violations) > 0 {
		for i, violation := range violations {
			if i >= 12 {
				break
			}
			t.Error(violation)
		}
		t.Fatalf("dense loaded published view failed with %d violations", len(violations))
	}
}

func TestResearchPublishedDenseLoadV7(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_DENSE_LOAD_V7") != "1" {
		t.Skip("opt-in loaded dense published-LSN top-k transfer test")
	}
	for trial, k := range []int{50, 200, 200, 50} {
		t.Run(fmt.Sprintf("trial-%d-k%d", trial+1, k), func(t *testing.T) {
			runDensePublishedLoadV7(t, k)
		})
	}
}
