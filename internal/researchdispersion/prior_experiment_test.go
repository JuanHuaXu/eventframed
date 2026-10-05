package researchdispersion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/researchprior"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var modesPriorV40 = []string{"full", "adaptive", "local16", "raw2_private", "raw2_shared", "raw4_private", "raw4_shared", "inverse2_private", "inverse2_shared", "inverse4_private", "inverse4_shared"}
var filesPriorV40 = append(append([]string(nil), filesV39...), "internal/researchprior/model.go", "internal/researchprior/model_test.go", "internal/researchprior/path_test.go", "internal/researchpriorref/reference.go", "internal/researchdispersion/prior_experiment_test.go", "internal/researchdispersion/prior_audit_test.go", "internal/researchdispersion/prior_stats_test.go", "docs/experiments/mmm-prior-v40-protocol.md", "research/prior-v40-run.mjs")

func configPriorV40(mode string) (researchprior.Config, error) {
	for _, center := range []string{"raw", "inverse"} {
		for _, strength := range []int{2, 4} {
			for _, shared := range []bool{false, true} {
				suffix := "private"
				if shared {
					suffix = "shared"
				}
				if mode == fmt.Sprintf("%s%d_%s", center, strength, suffix) {
					return researchprior.Config{Center: center, Strength: float64(strength), Hazard: 1. / 16, Shared: shared}, nil
				}
			}
		}
	}
	return researchprior.Config{}, fmt.Errorf("unknown prior mode")
}
func runPriorV40(w windowWorldV37, mode, schedule string) (orientationArmV38, error) {
	if mode == "full" || mode == "adaptive" || mode == "local16" {
		return runContinuingV39(w, mode, schedule)
	}
	a := orientationArmV38{ReadyAt: -1, windowArmV37: windowArmV37{Mode: mode, Schedule: schedule, Issued: make([]float64, 2400), ExpertIssued: make([][4]float64, 2400)}}
	cfg, e := configPriorV40(mode)
	if e != nil {
		return a, e
	}
	start, phase := time.Now(), time.Now()
	due, e := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, schedule)
	a.Costs.ScheduleNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	phase = time.Now()
	m, e := researchprior.New(w.Base, 1, 2400, cfg)
	a.Costs.SetupNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	tickets := make([]researchprior.Ticket, 2400)
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			phase = time.Now()
			ticket, e := m.Issue(tick%150, int64(tick))
			a.Costs.IssueNS += time.Since(phase).Nanoseconds()
			if e != nil {
				return a, e
			}
			tickets[tick], a.Issued[tick] = ticket, ticket.Forecast()
			a.PeakPending = max(a.PeakPending, m.Pending())
		}
		for _, trial := range due[tick] {
			phase = time.Now()
			r, e := m.Resolve(tickets[trial], w.Outcomes[trial/150][trial%150], int64(tick))
			a.Costs.ResolveNS += time.Since(phase).Nanoseconds()
			if e != nil {
				return a, e
			}
			a.Receipts = append(a.Receipts, DelayedReceipt{Member: r.Member, TrialOrdinal: r.TrialOrdinal, Epoch: r.Epoch, IssuedAt: r.IssuedAt, ArrivedAt: r.ArrivedAt, Forecast: r.Forecast, Useful: r.Useful})
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			phase = time.Now()
			s := windowSnapshotV37{delayedSnapshotV36: delayedSnapshotV36{Tick: tick, Arrived: len(a.Receipts), Pending: m.Pending(), Forecast: make([]float64, 150)}}
			for i := range s.Forecast {
				s.Forecast[i], e = m.Predict(i)
				if e != nil {
					return a, e
				}
			}
			a.Costs.SnapshotNS += time.Since(phase).Nanoseconds()
			a.Snapshots = append(a.Snapshots, s)
		}
	}
	a.Costs.AccountedNS = a.Costs.SetupNS + a.Costs.IssueNS + a.Costs.ResolveNS + a.Costs.SnapshotNS
	a.Costs.ElapsedNS = time.Since(start).Nanoseconds()
	return a, nil
}
func priorCohortV40(split string) (int64, int, error) {
	switch split {
	case "diagnostic":
		return 2026104001, 1, nil
	case "design":
		return 2026104003, 16, nil
	case "confirmation":
		return 2026104004, 16, nil
	}
	return 0, 0, fmt.Errorf("invalid prior cohort")
}
func TestPriorExperimentV40(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_PRIOR_V40_OUT"), os.Getenv("EVENTFRAME_PRIOR_V40_SPLIT")
	if path == "" {
		t.Skip("explicit output required")
	}
	seed, worlds, e := priorCohortV40(split)
	if e != nil {
		t.Fatal(e)
	}
	hashes := map[string]string{}
	for _, p := range filesPriorV40 {
		b, e := os.ReadFile(filepath.Join(rootV34(t), p))
		if e != nil {
			t.Fatal(e)
		}
		hashes[p] = hashV34(b)
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	enc := json.NewEncoder(buf)
	if e = enc.Encode(manifestV34{Kind: "manifest", Split: split, SeedBase: seed, Worlds: 28 * worlds, Sources: hashes}); e != nil {
		t.Fatal(e)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV39 {
			for id := 0; id < worlds; id++ {
				w := orientationWorldV38{Population: makeV39(seed, g, r, id)}
				for _, schedule := range schedulesV39 {
					for _, mode := range modesPriorV40 {
						a, e := runPriorV40(w.Population, mode, schedule)
						if e != nil {
							t.Fatal(e)
						}
						scoreWindowV37(w.Population, &a.windowArmV37)
						w.Arms = append(w.Arms, a)
					}
				}
				if e = enc.Encode(w); e != nil {
					t.Fatal(e)
				}
			}
			t.Log([]string{"tight", "wide"}[g] + "/" + regimesV39[r])
		}
	}
	if e = buf.Flush(); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
}
func TestPriorFuturePrefixV40(t *testing.T) {
	for _, schedule := range schedulesV39 {
		for _, mode := range modesPriorV40[3:] {
			w, x := makeV39(730001, 1, 8, 0), makeV39(730001, 1, 8, 0)
			for round := 8; round < 16; round++ {
				for i := range x.Outcomes[round] {
					x.Outcomes[round][i] = !x.Outcomes[round][i]
					x.Rates[round][i] = 1 - x.Rates[round][i]
				}
			}
			a, e := runPriorV40(w, mode, schedule)
			if e != nil {
				t.Fatal(e)
			}
			b, e := runPriorV40(x, mode, schedule)
			if e != nil || !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) {
				t.Fatal("future issued-law leak", e)
			}
			for j, s := range a.Snapshots {
				if s.Tick < 1200 && !reflect.DeepEqual(s, b.Snapshots[j]) {
					t.Fatal("future snapshot leak")
				}
			}
		}
	}
}
