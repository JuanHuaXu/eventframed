package researchdispersion

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var regimesV38 = append(append([]string(nil), regimesV37...), "partial", "unrelated", "asynchronous", "early")
var modesV38 = []string{"full", "adaptive", "anchor", "orientation"}
var filesV38 = append(append([]string(nil), filesV37...), "internal/researchdispersion/orientation.go", "internal/researchdispersion/orientation_test.go", "internal/researchdispersion/orientation_experiment_test.go", "docs/experiments/mmm-orientation-v38-preflight.md", "docs/experiments/mmm-orientation-v38-protocol.md")

type orientationArmV38 struct {
	windowArmV37
	ReadyAt  int
	Template []float64
}

type orientationWorldV38 struct {
	Population windowWorldV37
	Arms       []orientationArmV38
}

func makeV38(seed int64, g, r, id int) windowWorldV37 {
	if r < 8 {
		return makeV37(seed, g, r, id)
	}
	w := makeV37(seed+int64(r-4)*1000000, g, 4, id)
	w.Regime = regimesV38[r]
	first := append([]float64(nil), w.Rates[0]...)
	changed := append([]float64(nil), first...)
	starts := make([]int, 150)
	for i := range starts {
		starts[i] = 8
	}
	rng := rand.New(rand.NewSource(w.Seed + 303))
	switch r {
	case 8:
		for _, i := range rng.Perm(150)[:75] {
			changed[i] = 1 - first[i]
		}
	case 9:
		for i, j := range rng.Perm(150) {
			changed[i] = first[j]
		}
	case 10:
		w.Changes = nil // member changes have no single declared phase boundary
		for i := range starts {
			starts[i], changed[i] = 4+rng.Intn(9), 1-first[i]
		}
	case 11:
		w.Changes = []int{2}
		for i := range starts {
			starts[i], changed[i] = 2, 1-first[i]
		}
	}
	draw := rand.New(rand.NewSource(w.Seed + 202))
	for round := range w.Rates {
		for i := range first {
			p := first[i]
			if round >= starts[i] {
				p = changed[i]
			}
			w.Rates[round][i], w.Outcomes[round][i] = p, draw.Float64() < p
		}
	}
	return w
}

func runOrientationV38(w windowWorldV37, mode, schedule string) (orientationArmV38, error) {
	a := orientationArmV38{ReadyAt: -1}
	if mode == "full" || mode == "adaptive" {
		v, err := runWindowV37(w, mode, schedule)
		a.windowArmV37 = v
		return a, err
	}
	a.windowArmV37 = windowArmV37{Mode: mode, Schedule: schedule, Issued: make([]float64, 2400), ExpertIssued: make([][4]float64, 2400)}
	started, phase := time.Now(), time.Now()
	due, err := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, schedule)
	a.Costs.ScheduleNS = time.Since(phase).Nanoseconds()
	if err != nil {
		return a, err
	}
	phase = time.Now()
	m, err := NewOrientationObserver(w.Base, 1, 2400, mode)
	a.Costs.SetupNS = time.Since(phase).Nanoseconds()
	if err != nil {
		return a, err
	}
	tickets := make([]DelayedTicket, 2400)
	last := 4800
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
			if a.ReadyAt == -1 && m.filter != nil {
				a.ReadyAt, a.Template = tick, append([]float64(nil), m.filter.template...)
			}
		}
		if (tick < 2400 && tick%150 == 149) || tick == last {
			phase = time.Now()
			s := windowSnapshotV37{delayedSnapshotV36: delayedSnapshotV36{Tick: tick, Arrived: len(a.Receipts), Pending: m.Pending(), Forecast: make([]float64, 150)}}
			if m.filter != nil {
				s.Weights[0] = m.filter.next()
			}
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

func TestExperimentV38(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_ORIENTATION_V38_OUT"), os.Getenv("EVENTFRAME_ORIENTATION_V38_SPLIT")
	if path == "" {
		t.Skip("explicit output and split required")
	}
	seed := int64(2026103803)
	if split == "confirmation" {
		seed++
	} else if split != "design" {
		t.Fatal("invalid split")
	}
	hashes := map[string]string{}
	for _, p := range filesV38 {
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
	if err := e.Encode(manifestV34{Kind: "manifest", Split: split, SeedBase: seed, Worlds: 384, Sources: hashes}); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV38 {
			for id := 0; id < 16; id++ {
				w := orientationWorldV38{Population: makeV38(seed, g, r, id)}
				for _, schedule := range schedulesV37 {
					for _, mode := range modesV38 {
						a, err := runOrientationV38(w.Population, mode, schedule)
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
			t.Log([]string{"tight", "wide"}[g] + "/" + regimesV38[r])
		}
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestOrientationCollectorFuturePrefixV38(t *testing.T) {
	for _, schedule := range schedulesV37 {
		for _, mode := range []string{"anchor", "orientation"} {
			w := makeV38(123003, 1, 6, 0)
			x := makeV38(123003, 1, 6, 0)
			for r := 8; r < 16; r++ {
				for i := range x.Outcomes[r] {
					x.Outcomes[r][i] = !x.Outcomes[r][i]
					x.Rates[r][i] = 1 - x.Rates[r][i]
				}
			}
			a, err := runOrientationV38(w, mode, schedule)
			if err != nil {
				t.Fatal(err)
			}
			b, err := runOrientationV38(x, mode, schedule)
			if err != nil || !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) || !reflect.DeepEqual(a.Template, b.Template) || a.ReadyAt != b.ReadyAt {
				t.Fatal("future outcomes leaked", err)
			}
			for j, s := range a.Snapshots {
				if s.Tick < 1200 && !reflect.DeepEqual(s, b.Snapshots[j]) {
					t.Fatal("future snapshot leaked")
				}
			}
		}
	}
}
