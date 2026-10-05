package researchdispersion

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var regimesV39 = append(append([]string(nil), regimesV38...), "stationary_noise10", "partial_noise10")
var modesV39 = []string{"full", "adaptive", "anchor", "orientation", "local0", "local32", "local16"}
var schedulesV39 = []string{"immediate", "fixed150", "uniform299"}
var filesV39 = append(append([]string(nil), filesV38...),
	"internal/researchdispersion/continuing.go", "internal/researchdispersion/continuing_test.go",
	"internal/researchdispersion/continuing_experiment_test.go", "internal/researchdispersion/continuing_audit_test.go",
	"internal/researchdispersion/orientation_audit_test.go", "internal/researchdispersion/window_audit_test.go",
	"internal/researchdispersion/delayed_audit_test.go", "internal/researchdispersion/shape_audit_test.go", "internal/researchdispersion/audit_test.go",
	"docs/experiments/mmm-continuing-v39-protocol.md", "research/continuing-v39-run.mjs")

func makeV39(seed int64, g, r, id int) windowWorldV37 {
	if r < 12 {
		return makeV38(seed, g, r, id)
	}
	old := 1
	if r == 13 {
		old = 8
	}
	w := makeV38(seed+int64(r)*2000000, g, old, id)
	w.Regime = regimesV39[r]
	noise := rand.New(rand.NewSource(w.Seed + 707))
	for _, row := range w.Outcomes {
		for i := range row {
			if noise.Float64() < .1 {
				row[i] = !row[i]
			}
		}
	}
	// Rates retain true usefulness; Outcomes here are the noisy evidence labels.
	return w
}

func hazardV39(mode string) (float64, error) {
	switch mode {
	case "local0":
		return 0, nil
	case "local32":
		return 1. / 32, nil
	case "local16":
		return 1. / 16, nil
	}
	return 0, fmt.Errorf("unknown continuing mode %q", mode)
}

func runContinuingV39(w windowWorldV37, mode, schedule string) (orientationArmV38, error) {
	for _, control := range modesV38 {
		if mode == control {
			return runOrientationV38(w, mode, schedule)
		}
	}
	a := orientationArmV38{ReadyAt: -1, windowArmV37: windowArmV37{Mode: mode, Schedule: schedule, Issued: make([]float64, 2400), ExpertIssued: make([][4]float64, 2400)}}
	h, err := hazardV39(mode)
	if err != nil {
		return a, err
	}
	started, phase := time.Now(), time.Now()
	due, err := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, schedule)
	a.Costs.ScheduleNS = time.Since(phase).Nanoseconds()
	if err != nil {
		return a, err
	}
	phase = time.Now()
	m, err := NewContinuingObserver(w.Base, 1, 2400, h)
	a.Costs.SetupNS = time.Since(phase).Nanoseconds()
	if err != nil {
		return a, err
	}
	tickets := make([]DelayedTicket, 2400)
	last := len(due) - 1
	for len(due[last]) == 0 {
		last--
	}
	for tick := 0; tick <= last; tick++ {
		if tick < 2400 {
			phase = time.Now()
			ticket, err := m.Issue(tick%150, int64(tick))
			a.Costs.IssueNS += time.Since(phase).Nanoseconds()
			if err != nil {
				return a, err
			}
			tickets[tick], a.Issued[tick] = ticket, ticket.Forecast()
			a.PeakPending = max(a.PeakPending, m.Pending())
		}
		for _, trial := range due[tick] {
			phase = time.Now()
			r, err := m.Resolve(tickets[trial], w.Outcomes[trial/150][trial%150], int64(tick))
			a.Costs.ResolveNS += time.Since(phase).Nanoseconds()
			if err != nil {
				return a, err
			}
			a.Receipts = append(a.Receipts, r)
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			phase = time.Now()
			s := windowSnapshotV37{delayedSnapshotV36: delayedSnapshotV36{Tick: tick, Arrived: len(a.Receipts), Pending: m.Pending(), Forecast: make([]float64, 150)}}
			for i := range s.Forecast {
				s.Forecast[i], err = m.Predict(i)
				if err != nil {
					return a, err
				}
			}
			a.Costs.SnapshotNS += time.Since(phase).Nanoseconds()
			a.Snapshots = append(a.Snapshots, s)
		}
	}
	a.Costs.AccountedNS = a.Costs.SetupNS + a.Costs.IssueNS + a.Costs.ResolveNS + a.Costs.SnapshotNS
	a.Costs.ElapsedNS = time.Since(started).Nanoseconds()
	return a, nil
}

func TestExperimentV39(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_CONTINUING_V39_OUT"), os.Getenv("EVENTFRAME_CONTINUING_V39_SPLIT")
	if path == "" {
		t.Skip("explicit output and split required")
	}
	seed := int64(2026103903)
	if split == "confirmation" {
		seed++
	} else if split != "design" {
		t.Fatal("invalid split")
	}
	hashes := map[string]string{}
	for _, p := range filesV39 {
		b, err := os.ReadFile(filepath.Join(rootV34(t), p))
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = hashV34(b)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	e := json.NewEncoder(b)
	if err := e.Encode(manifestV34{Kind: "manifest", Split: split, SeedBase: seed, Worlds: 448, Sources: hashes}); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV39 {
			for id := 0; id < 16; id++ {
				w := orientationWorldV38{Population: makeV39(seed, g, r, id)}
				for _, schedule := range schedulesV39 {
					for _, mode := range modesV39 {
						a, err := runContinuingV39(w.Population, mode, schedule)
						if err != nil {
							t.Fatal(err)
						}
						scoreWindowV37(w.Population, &a.windowArmV37)
						w.Arms = append(w.Arms, a)
					}
				}
				if err := e.Encode(w); err != nil {
					t.Fatal(err)
				}
			}
			t.Log([]string{"tight", "wide"}[g] + "/" + regimesV39[r])
		}
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestContinuingCollectorFuturePrefixV39(t *testing.T) {
	for _, schedule := range schedulesV39 {
		for _, mode := range modesV39[4:] {
			w, x := makeV39(123003, 1, 8, 0), makeV39(123003, 1, 8, 0)
			for r := 8; r < 16; r++ {
				for i := range x.Outcomes[r] {
					x.Outcomes[r][i] = !x.Outcomes[r][i]
					x.Rates[r][i] = 1 - x.Rates[r][i]
				}
			}
			a, err := runContinuingV39(w, mode, schedule)
			if err != nil {
				t.Fatal(err)
			}
			b, err := runContinuingV39(x, mode, schedule)
			if err != nil || !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) {
				t.Fatal("future issue law leaked", err)
			}
			for j, s := range a.Snapshots {
				if s.Tick < 1200 && !reflect.DeepEqual(s, b.Snapshots[j]) {
					t.Fatal("future snapshot leaked")
				}
			}
		}
	}
}
