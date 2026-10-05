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

var regimesV36 = []string{"aligned", "independent", "curved", "mean_shared", "symmetric_noise", "mid_shift"}
var schedulesV36 = []string{"immediate", "fixed150", "uniform299", "burst300", "reverse_flush", "outcome_coupled"}
var filesV36 = append(append([]string(nil), filesV35...), "internal/researchdispersion/delayed.go", "internal/researchdispersion/delayed_test.go", "internal/researchdispersion/delayed_experiment_test.go", "docs/experiments/mmm-delayed-v36-preflight.md", "docs/experiments/mmm-delayed-v36-protocol.md")

type delayedSnapshotV36 struct {
	Tick, Arrived                                      int
	Pending                                            int
	Forecast                                           []float64
	Brier, PriorityBrier, PacketUsefulness, PacketBias float64
}

type delayedCostsV36 struct {
	SetupNS, ScheduleNS, IssueNS, ResolveNS, SnapshotNS, AccountedNS, ElapsedNS int64
}

type delayedArmV36 struct {
	Schedule    string
	Issued      []float64
	Receipts    []DelayedReceipt
	Snapshots   []delayedSnapshotV36
	PeakPending int
	IssuedBrier float64
	Costs       delayedCostsV36
}

type delayedWorldV36 struct {
	Kind, Geometry, Regime string
	World                  int
	Seed                   int64
	Base, Before, After    []float64
	Outcomes               [][]bool
	Arms                   []delayedArmV36
}

func makeV36(seed int64, g, r, id int) delayedWorldV36 {
	old := []int{0, 4, 3, 11, 4, 4}[r]
	w := makeV35(seed+int64(r-old)*1000000, g, old, id)
	x := delayedWorldV36{Kind: "world", Geometry: w.Geometry, Regime: regimesV36[r], World: id, Seed: w.Seed, Base: w.Base, Before: w.Rates, After: append([]float64(nil), w.Rates...), Outcomes: w.Outcomes}
	if r == 4 {
		noise := rand.New(rand.NewSource(x.Seed + 404))
		for _, row := range x.Outcomes {
			for i := range row {
				if noise.Float64() < .1 {
					row[i] = !row[i]
				}
			}
		}
	}
	if r == 5 {
		for i := range x.After {
			x.After[i] = 1 - x.Before[i]
		}
		draw := rand.New(rand.NewSource(x.Seed + 202))
		for round, row := range x.Outcomes {
			rates := x.Before
			if round >= 8 {
				rates = x.After
			}
			for i := range row {
				row[i] = draw.Float64() < rates[i]
			}
		}
	}
	return x
}

func scheduleV36(w delayedWorldV36, name string) ([][]int, error) {
	due := make([][]int, 4801)
	rng := rand.New(rand.NewSource(w.Seed + 505))
	for trial := 0; trial < 2400; trial++ {
		delay := 0
		switch name {
		case "immediate":
		case "fixed150":
			delay = 150
		case "uniform299":
			delay = rng.Intn(300)
		case "burst300":
			delay = 299 - trial%300
		case "reverse_flush":
			delay = 4800 - 2*trial
		case "outcome_coupled":
			if !w.Outcomes[trial/150][trial%150] {
				delay = 599
			}
		default:
			return nil, fmt.Errorf("unknown schedule")
		}
		due[trial+delay] = append(due[trial+delay], trial)
	}
	return due, nil
}

func runDelayedV36(w delayedWorldV36, name string) (delayedArmV36, error) {
	a := delayedArmV36{Schedule: name, Issued: make([]float64, 2400)}
	started := time.Now()
	phase := time.Now()
	due, err := scheduleV36(w, name)
	a.Costs.ScheduleNS = time.Since(phase).Nanoseconds()
	if err != nil {
		return a, err
	}
	phase = time.Now()
	d, err := NewDelayedShape(w.Base, 1, 2400)
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
			ticket, err := d.Issue(tick%150, int64(tick))
			a.Costs.IssueNS += time.Since(phase).Nanoseconds()
			if err != nil {
				return a, err
			}
			tickets[tick], a.Issued[tick] = ticket, ticket.Forecast()
			if d.Pending() > a.PeakPending {
				a.PeakPending = d.Pending()
			}
		}
		for _, trial := range due[tick] {
			phase = time.Now()
			r, err := d.Resolve(tickets[trial], w.Outcomes[trial/150][trial%150], int64(tick))
			a.Costs.ResolveNS += time.Since(phase).Nanoseconds()
			if err != nil {
				return a, err
			}
			a.Receipts = append(a.Receipts, r)
		}
		if tick == 599 || tick == 1199 || tick == 2399 || tick == last {
			phase = time.Now()
			s := delayedSnapshotV36{Tick: tick, Arrived: len(a.Receipts), Pending: d.Pending(), Forecast: make([]float64, 150)}
			for i := range s.Forecast {
				s.Forecast[i], err = d.Predict(i)
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

func scoreDelayedV36(w delayedWorldV36, a *delayedArmV36) {
	for j := range a.Snapshots {
		s := &a.Snapshots[j]
		rates := w.Before
		if s.Tick >= 1200 {
			rates = w.After
		}
		v := snapV34{Forecast: s.Forecast}
		scoreV34(&v, rates)
		s.Brier, s.PriorityBrier, s.PacketUsefulness, s.PacketBias = v.Brier, v.PriorityBrier, v.PacketUsefulness, v.PacketBias
	}
	for trial, q := range a.Issued {
		rates := w.Before
		if trial >= 1200 {
			rates = w.After
		}
		p := rates[trial%150]
		a.IssuedBrier += (q*q - 2*q*p + p) / 2400
	}
}

func TestExperimentV36(t *testing.T) {
	path, split := os.Getenv("EVENTFRAME_DELAYED_V36_OUT"), os.Getenv("EVENTFRAME_DELAYED_V36_SPLIT")
	if path == "" {
		t.Skip("explicit output and split required")
	}
	seed := int64(2026103603)
	if split == "confirmation" {
		seed++
	} else if split != "design" {
		t.Fatal("invalid split")
	}
	root := rootV34(t)
	hashes := map[string]string{}
	for _, p := range filesV36 {
		b, err := os.ReadFile(filepath.Join(root, p))
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
	if err = e.Encode(manifestV34{Kind: "manifest", Split: split, SeedBase: seed, Worlds: 192, Sources: hashes}); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV36 {
			for id := 0; id < 16; id++ {
				w := makeV36(seed, g, r, id)
				for _, name := range schedulesV36 {
					a, err := runDelayedV36(w, name)
					if err != nil {
						t.Fatal(err)
					}
					scoreDelayedV36(w, &a)
					w.Arms = append(w.Arms, a)
				}
				if err := e.Encode(w); err != nil {
					t.Fatal(err)
				}
			}
			t.Log(wNameV36(g, r))
		}
	}
	if err := b.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

func wNameV36(g, r int) string {
	return []string{"tight", "wide"}[g] + "/" + regimesV36[r]
}

func TestDelayedCollectorFuturePrefixV36(t *testing.T) {
	for r := range regimesV36 {
		for _, name := range schedulesV36 {
			if name == "outcome_coupled" {
				continue // changing the outcome changes its selection schedule
			}
			w, v := makeV36(2026103699, 1, r, 0), makeV36(2026103699, 1, r, 0)
			for round := 8; round < 16; round++ {
				for i := range v.Outcomes[round] {
					v.Outcomes[round][i] = !v.Outcomes[round][i]
				}
			}
			a, err := runDelayedV36(w, name)
			if err != nil {
				t.Fatal(err)
			}
			b, err := runDelayedV36(v, name)
			if err != nil || !reflect.DeepEqual(a.Issued[:1201], b.Issued[:1201]) {
				t.Fatal("future outcomes changed earlier issued forecasts", r, name, err)
			}
			for j, s := range a.Snapshots {
				if s.Tick < 1200 && !reflect.DeepEqual(s, b.Snapshots[j]) {
					t.Fatal("future outcomes changed earlier snapshots")
				}
			}
		}
	}
}
