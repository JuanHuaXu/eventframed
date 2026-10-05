package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/bits"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

var arrivalSwitchNames = []string{"stable_majority3", "stable_parity4", "majority_to_parity", "parity_to_majority"}

func arrivalSwitchMasks(k, phase int) []uint16 {
	var out []uint16
	n := 0
	for m := uint16(1); m < 512; m++ {
		v := m | 1
		cost := bits.Len16(v&7) + bits.Len16((v>>3)&7) + bits.Len16((v>>6)&7)
		if bits.OnesCount16(m) == k && cost <= 6 {
			if n%2 == phase {
				out = append(out, m)
			}
			n++
		}
	}
	return out
}

func arrivalSwitchTruth(x uint16, masks [2]uint16, scenario, clock int, reference bool) bool {
	majority := scenario == 0 || scenario == 2
	m := masks[0]
	if scenario >= 2 && clock >= 256 && !reference {
		majority = !majority
		m = masks[1]
	}
	if majority {
		return bits.OnesCount16(x&m) >= 2
	}
	return bits.OnesCount16(x&m)%2 == 1
}

func arrivalSwitchSetup(phase, scenario, index int) (*observation.Model, [2]uint16, error) {
	seed := int64(2026092431 + 100*phase)
	rules := rand.New(rand.NewSource(observationpreserved.Seed(seed, scenario, index, 8)))
	k := 3
	if scenario == 1 || scenario == 3 {
		k = 4
	}
	var masks [2]uint16
	for i := range masks {
		a := k
		if i == 1 && scenario >= 2 {
			a = 7 - k
		}
		pool := arrivalSwitchMasks(a, phase)
		masks[i] = pool[rules.Intn(len(pool))]
	}
	if scenario < 2 {
		masks[1] = masks[0]
	}
	rng := rand.New(rand.NewSource(observationpreserved.Seed(seed, scenario, index, 4)))
	samples := make([]observation.Sample, 4096)
	for i := range samples {
		x := uint16(rng.Intn(512))
		y := arrivalSwitchTruth(x, masks, scenario, -1, false)
		if rng.Float64() < .05 {
			y = !y
		}
		samples[i] = observation.Sample{Bits: x, Outcome: y}
	}
	base, err := observation.Fit(samples)
	return base, masks, err
}

func TestArrivalSwitchContracts(t *testing.T) {
	seen := map[int64]bool{}
	for phase := 0; phase < 2; phase++ {
		for scenario := 0; scenario < 4; scenario++ {
			for i := 0; i < 16; i++ {
				for role := 0; role < 9; role++ {
					s := observationpreserved.Seed(int64(2026092431+100*phase), scenario, i, role)
					if seen[s] {
						t.Fatal("seed collision")
					}
					seen[s] = true
				}
			}
		}
	}
	for _, k := range []int{3, 4} {
		pool := map[uint16]bool{}
		for phase := 0; phase < 2; phase++ {
			for _, m := range arrivalSwitchMasks(k, phase) {
				if pool[m] || bits.OnesCount16(m) != k {
					t.Fatal("pool")
				}
				pool[m] = true
				v := m | 1
				if bits.Len16(v&7)+bits.Len16((v>>3)&7)+bits.Len16((v>>6)&7) > 6 {
					t.Fatal("inaccessible")
				}
			}
		}
	}
	for s := 0; s < 4; s++ {
		_, m, err := arrivalSwitchSetup(0, s, 0)
		if err != nil {
			t.Fatal(err)
		}
		changed := 0
		for x := uint16(0); x < 512; x++ {
			old := arrivalSwitchTruth(x, m, s, 255, false)
			next := arrivalSwitchTruth(x, m, s, 256, false)
			if arrivalSwitchTruth(x, m, s, 256, true) != old {
				t.Fatal("reference changed")
			}
			if old != next {
				changed++
			}
		}
		if (s < 2 && changed != 0) || (s >= 2 && changed == 0) {
			t.Fatal("boundary")
		}
	}
	// Supplying the old label rule explicitly must leave both schedules identical.
	cfg := forestDependenceCases()[0]
	base, err := forestDependenceBase(cfg, 123)
	if err != nil {
		t.Fatal(err)
	}
	for _, delay := range []bool{false, true} {
		want, err := innerArrivalRun(base, cfg, 0, 0, 123, delay)
		if err != nil {
			t.Fatal(err)
		}
		got, err := innerArrivalRunOutcome(base, cfg, 0, 0, 123, delay, func(x uint16, c int, ref bool, r *rand.Rand) bool {
			return breadthLabel(x, !ref && c >= 256, false, r, cfg.Target)
		})
		if err != nil || !reflect.DeepEqual(want, got) {
			t.Fatal("callback parity", err)
		}
	}
}

type arrivalSwitchRecord struct {
	Phase, Case        string
	Index              int
	Masks              [2]uint16
	Immediate, Delayed innerArrivalResult
}

func TestArrivalSwitchExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ARRIVAL_SWITCH")
	if path == "" {
		t.Skip("opt-in exploratory transfer")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-arrival-switch-transfer-v1-contract.md", "../../research/arrival-switch-summary.mjs")
	sources, hashes := map[string]string{}, map[string]string{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		sources[p[6:]] = string(b)
		hashes[p[6:]] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "Exploratory": true, "Seeds": [2]int64{2026092431, 2026092531}, "PerCell": 16}); err != nil {
		t.Fatal(err)
	}
	for phase := 0; phase < 2; phase++ {
		for s, name := range arrivalSwitchNames {
			for index := 0; index < 16; index++ {
				base, m, err := arrivalSwitchSetup(phase, s, index)
				if err != nil {
					t.Fatal(err)
				}
				cfg := forestDependenceCase{subsetBreadthCase{Name: name, Mode: "stable", Target: "bit"}, 0}
				outcome := func(x uint16, c int, ref bool, r *rand.Rand) bool {
					y := arrivalSwitchTruth(x, m, s, c, ref)
					if r.Float64() < .05 {
						y = !y
					}
					return y
				}
				row := arrivalSwitchRecord{Phase: []string{"cohort1", "cohort2"}[phase], Case: name, Index: index, Masks: m}
				row.Immediate, err = innerArrivalRunOutcome(base, cfg, s, index, int64(2026092431+100*phase), false, outcome)
				if err != nil {
					t.Fatal(err)
				}
				row.Delayed, err = innerArrivalRunOutcome(base, cfg, s, index, int64(2026092431+100*phase), true, outcome)
				if err != nil {
					t.Fatal(err)
				}
				if err := enc.Encode(row); err != nil {
					t.Fatal(err)
				}
			}
			if err := f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(phase, name, "complete")
		}
	}
}
