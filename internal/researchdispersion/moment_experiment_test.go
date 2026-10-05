package researchdispersion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchmoment"
)

var modesMomentV41 = []string{"full", "adaptive", "v40raw2", "narrow_density2", "narrow_moment2", "narrow_density4", "narrow_moment4", "rich_density2", "rich_moment2", "rich_density4", "rich_moment4"}
var filesMomentV41 = append(append([]string(nil), filesPriorV40...), "go.mod", "go.sum", "internal/researchmoment/model.go", "internal/researchmoment/model_test.go", "internal/researchmoment/reference_test.go", "internal/researchmomentref/reference.go", "internal/researchdispersion/moment_experiment_test.go", "internal/researchdispersion/moment_audit_test.go", "internal/researchdispersion/moment_stats_test.go", "docs/experiments/mmm-moment-v41-protocol.md", "research/moment-v41-run.mjs")

func configMomentV41(mode string) (researchmoment.Config, error) {
	for _, family := range []string{"narrow", "rich"} {
		for _, prior := range []string{"density", "moment"} {
			for _, s := range []int{2, 4} {
				if mode == fmt.Sprintf("%s_%s%d", family, prior, s) {
					return researchmoment.Config{Family: family, Prior: prior, Strength: float64(s), Hazard: 1. / 16, Shared: true}, nil
				}
			}
		}
	}
	return researchmoment.Config{}, fmt.Errorf("unknown moment mode")
}
func runMomentV41(w windowWorldV37, mode, schedule string) (orientationArmV38, error) {
	if mode == "full" || mode == "adaptive" {
		return runPriorV40(w, mode, schedule)
	}
	if mode == "v40raw2" {
		a, e := runPriorV40(w, "raw2_shared", schedule)
		a.Mode = mode
		return a, e
	}
	start := time.Now()
	a := orientationArmV38{ReadyAt: -1, windowArmV37: windowArmV37{Mode: mode, Schedule: schedule, Issued: make([]float64, 2400), ExpertIssued: make([][4]float64, 2400)}}
	cfg, e := configMomentV41(mode)
	if e != nil {
		return a, e
	}
	phase := time.Now()
	due, e := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, schedule)
	a.Costs.ScheduleNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	phase = time.Now()
	m, e := researchmoment.New(w.Base, 1, 2400, cfg)
	a.Costs.SetupNS = time.Since(phase).Nanoseconds()
	if e != nil {
		return a, e
	}
	tickets := make([]researchmoment.Ticket, 2400)
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
func momentCohortV41(split string) (int64, int, error) {
	switch split {
	case "diagnostic":
		return 2026104101, 1, nil
	case "design":
		return 2026104103, 16, nil
	case "confirmation":
		return 2026104104, 16, nil
	}
	return 0, 0, fmt.Errorf("invalid moment cohort")
}
func TestMomentExperimentV41(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_MOMENT_V41_OUT"), os.Getenv("EVENTFRAME_MOMENT_V41_SPLIT")
	if path == "" {
		t.Skip("explicit output required")
	}
	seed, worlds, e := momentCohortV41(split)
	if e != nil {
		t.Fatal(e)
	}
	hashes := map[string]string{}
	for _, p := range filesMomentV41 {
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
					for _, mode := range modesMomentV41 {
						a, e := runMomentV41(w.Population, mode, schedule)
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
func TestMomentFuturePrefixV41(t *testing.T) {
	for _, schedule := range schedulesV39 {
		for _, mode := range modesMomentV41[3:] {
			w, x := makeV39(740001, 1, 8, 0), makeV39(740001, 1, 8, 0)
			for round := 8; round < 16; round++ {
				for i := range x.Outcomes[round] {
					x.Outcomes[round][i] = !x.Outcomes[round][i]
					x.Rates[round][i] = 1 - x.Rates[round][i]
				}
			}
			a, e := runMomentV41(w, mode, schedule)
			if e != nil {
				t.Fatal(e)
			}
			b, e := runMomentV41(x, mode, schedule)
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
func TestMomentSeedSeparationV41(t *testing.T) {
	seen := map[int64]bool{}
	for _, seed := range []int64{2026103903, 2026103904, 2026104001, 2026104003, 2026104004, 2026104101, 2026104103, 2026104104} {
		for g := 0; g < 2; g++ {
			for r := range regimesV39 {
				for id := 0; id < 16; id++ {
					s := makeV39(seed, g, r, id).Seed
					if seen[s] {
						t.Fatal("cohort collision", s)
					}
					seen[s] = true
				}
			}
		}
	}
}
