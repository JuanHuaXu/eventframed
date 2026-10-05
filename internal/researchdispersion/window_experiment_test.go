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

var regimesV37 = []string{"aligned", "independent", "curved", "baseline_matched", "abrupt", "late", "recurring", "gradual"}
var modesV37 = []string{"full", "fixed4", "fixed8", "fixed16", "adaptive"}
var schedulesV37 = []string{"immediate", "fixed150"}
var filesV37 = append(append([]string(nil), filesV36...), "internal/researchdispersion/window.go", "internal/researchdispersion/window_test.go", "internal/researchdispersion/window_experiment_test.go", "docs/experiments/mmm-window-v37-preflight.md", "docs/experiments/mmm-window-v37-protocol.md")

type windowSnapshotV37 struct {
	delayedSnapshotV36
	Weights [4]float64
}

type windowArmV37 struct {
	Mode, Schedule              string
	Issued                      []float64
	ExpertIssued                [][4]float64
	Receipts                    []DelayedReceipt
	Snapshots                   []windowSnapshotV37
	PeakPending                 int
	IssuedBrier, IssuedPriority float64
	Recovery                    float64
	Costs                       delayedCostsV36
}

type windowWorldV37 struct {
	Kind, Geometry, Regime string
	World                  int
	Seed                   int64
	Base                   []float64
	Rates                  [][]float64
	Outcomes               [][]bool
	Changes                []int
	Arms                   []windowArmV37
}

func makeV37(seed int64, g, r, id int) windowWorldV37 {
	old := []int{0, 4, 3, 7, 4, 4, 4, 4}[r]
	w := makeV35(seed+int64(r-old)*1000000, g, old, id)
	x := windowWorldV37{Kind: "world", Geometry: w.Geometry, Regime: regimesV37[r], World: id, Seed: w.Seed, Base: w.Base, Rates: make([][]float64, 16), Outcomes: make([][]bool, 16)}
	switch r {
	case 4:
		x.Changes = []int{8}
	case 5:
		x.Changes = []int{12}
	case 6:
		x.Changes = []int{4, 8, 12}
	}
	draw := rand.New(rand.NewSource(x.Seed + 202))
	for round := 0; round < 16; round++ {
		x.Rates[round], x.Outcomes[round] = make([]float64, 150), make([]bool, 150)
		fraction := 0.
		switch r {
		case 4:
			if round >= 8 {
				fraction = 1
			}
		case 5:
			if round >= 12 {
				fraction = 1
			}
		case 6:
			if round/4%2 == 1 {
				fraction = 1
			}
		case 7:
			fraction = float64(max(0, min(8, round-4))) / 8
		}
		for i, p := range w.Rates {
			q := (1-fraction)*p + fraction*(1-p)
			x.Rates[round][i] = q
			x.Outcomes[round][i] = draw.Float64() < q
		}
	}
	return x
}

func runWindowV37(w windowWorldV37, mode, schedule string) (windowArmV37, error) {
	a := windowArmV37{Mode: mode, Schedule: schedule, Issued: make([]float64, 2400), ExpertIssued: make([][4]float64, 2400)}
	started, phase := time.Now(), time.Now()
	due, err := scheduleV36(delayedWorldV36{Seed: w.Seed, Outcomes: w.Outcomes}, schedule)
	a.Costs.ScheduleNS = time.Since(phase).Nanoseconds()
	if err != nil {
		return a, err
	}
	phase = time.Now()
	m, err := NewWindowObserver(w.Base, 1, 2400, mode)
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
			tickets[tick], a.Issued[tick], a.ExpertIssued[tick] = ticket, ticket.Forecast(), m.original[ticket.slot]
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
			s := windowSnapshotV37{delayedSnapshotV36: delayedSnapshotV36{Tick: tick, Arrived: len(a.Receipts), Pending: m.Pending(), Forecast: make([]float64, 150)}, Weights: m.weights}
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

func scoreWindowV37(w windowWorldV37, a *windowArmV37) {
	for j := range a.Snapshots {
		s := &a.Snapshots[j]
		rates := w.Rates[min(15, s.Tick/150)]
		v := snapV34{Forecast: s.Forecast}
		scoreV34(&v, rates)
		s.Brier, s.PriorityBrier, s.PacketUsefulness, s.PacketBias = v.Brier, v.PriorityBrier, v.PacketUsefulness, v.PacketBias
	}
	for trial, q := range a.Issued {
		p := w.Rates[trial/150][trial%150]
		loss, weight := (q-p)*(q-p)+p*(1-p), 1.
		if trial%150 < 10 {
			weight = 3
		}
		a.IssuedBrier += loss / 2400
		a.IssuedPriority += loss * weight / (16 * 170)
	}
	for j, start := range w.Changes {
		end := 16
		if j+1 < len(w.Changes) {
			end = w.Changes[j+1]
		}
		delay, consecutive := end-start+1, 0
		for _, s := range a.Snapshots {
			if s.Tick >= 2400 {
				continue
			}
			round := s.Tick / 150
			if round < start || round >= end {
				continue
			}
			if s.Brier <= .20 && s.PacketUsefulness >= .75 {
				consecutive++
			} else {
				consecutive = 0
			}
			if consecutive == 2 {
				delay = round - start + 1
				break
			}
		}
		a.Recovery += float64(delay) / float64(len(w.Changes))
	}
}

func TestExperimentV37(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_WINDOW_V37_OUT"), os.Getenv("EVENTFRAME_WINDOW_V37_SPLIT")
	if path == "" {
		t.Skip("explicit output and split required")
	}
	seed := int64(2026103703)
	if split == "confirmation" {
		seed++
	} else if split != "design" {
		t.Fatal("invalid split")
	}
	hashes := map[string]string{}
	for _, p := range filesV37 {
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
	if err := e.Encode(manifestV34{Kind: "manifest", Split: split, SeedBase: seed, Worlds: 256, Sources: hashes}); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV37 {
			for id := 0; id < 16; id++ {
				w := makeV37(seed, g, r, id)
				for _, schedule := range schedulesV37 {
					for _, mode := range modesV37 {
						a, err := runWindowV37(w, mode, schedule)
						if err != nil {
							t.Fatal(err)
						}
						scoreWindowV37(w, &a)
						w.Arms = append(w.Arms, a)
					}
				}
				if err := e.Encode(w); err != nil {
					t.Fatal(err)
				}
			}
			t.Log([]string{"tight", "wide"}[g] + "/" + regimesV37[r])
		}
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestWindowCollectorFuturePrefixV37(t *testing.T) {
	for _, schedule := range schedulesV37 {
		for _, mode := range modesV37 {
			w, v := makeV37(2026103799, 1, 6, 0), makeV37(2026103799, 1, 6, 0)
			for round := 8; round < 16; round++ {
				for i := range v.Outcomes[round] {
					v.Outcomes[round][i] = !v.Outcomes[round][i]
				}
			}
			a, err := runWindowV37(w, mode, schedule)
			if err != nil {
				t.Fatal(err)
			}
			b, err := runWindowV37(v, mode, schedule)
			if err != nil || !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) || !reflect.DeepEqual(a.ExpertIssued[:1201], b.ExpertIssued[:1201]) {
				t.Fatal("future outcomes changed original forecasts", err)
			}
			for j, s := range a.Snapshots {
				if s.Tick < 1200 && !reflect.DeepEqual(s, b.Snapshots[j]) {
					t.Fatal("future outcome changed earlier snapshot")
				}
			}
		}
	}
}

func clearWindowV37(w *windowWorldV37) {
	for j := range w.Arms {
		w.Arms[j].Costs = delayedCostsV36{}
	}
}

func wNameV37(w windowWorldV37) string { return fmt.Sprintf("%s/%s/%d", w.Geometry, w.Regime, w.World) }
