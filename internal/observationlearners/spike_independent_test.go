package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

const spikeIndependentTransfer int64 = 3000011000
const spikeIndependentBoolean int64 = 3004011000

// Shift timing is an evaluator-side generator parameter, never learner input.
func spikeIndependentData(phase, scenario, index int) (softV120Data, int, int64, error) {
	d, err := softV118Generate(phase, scenario, index, spikeIndependentTransfer, spikeIndependentBoolean)
	if err != nil {
		return d, -1, 0, err
	}
	if scenario < 9 {
		return d, d.Teacher.Change, 0, nil
	}
	if scenario < 19 {
		return d, -1, 0, nil
	}
	seed := d.Seeds[0] + 5
	change := 65 + 2*rand.New(rand.NewSource(seed)).Intn(64)
	ys := rand.New(rand.NewSource(d.Seeds[2]))
	// Preserve the initial outcome draws before redrawing the shifted tape.
	for i := 0; i < 16; i++ {
		ys.Float64()
	}
	for i := range d.Frames {
		name, rule := "majority3", d.Rules[0]
		if scenario == 20 {
			name = "parity4"
		}
		if i >= change {
			rule = d.Rules[1]
			if scenario == 19 {
				name = "parity4"
			} else {
				name = "majority3"
			}
		}
		d.Q[i] = stackV93Truth(d.Frames[i].X, rule, name)
		d.Frames[i].Y = ys.Float64() < d.Q[i]
	}
	return d, change, seed, nil
}

func TestSpikeIndependentGenerator(t *testing.T) {
	for phase := 0; phase < 2; phase++ {
		for c := 0; c < 21; c++ {
			for index := 0; index < 8; index++ {
				d, change, seed, err := spikeIndependentData(phase, c, index)
				if err != nil {
					t.Fatal(err)
				}
				again, ch, s, err := spikeIndependentData(phase, c, index)
				if err != nil || ch != change || s != seed || !reflect.DeepEqual(d, again) {
					t.Fatal("reproducibility", phase, c, index)
				}
				original, err := softV118Generate(phase, c, index, spikeIndependentTransfer, spikeIndependentBoolean)
				if err != nil {
					t.Fatal(err)
				}
				if c < 19 {
					if !reflect.DeepEqual(d, original) {
						t.Fatal("adjacent generator changed")
					}
					continue
				}
				if change < 65 || change > 191 || change%2 != 1 || seed != d.Seeds[0]+5 || d.Initial != original.Initial {
					t.Fatal("shift contract")
				}
				ys := rand.New(rand.NewSource(d.Seeds[2]))
				for j := 0; j < 16; j++ {
					ys.Float64()
				}
				for j, f := range d.Frames {
					stage := 0
					if j >= change {
						stage = 1
					}
					names := [2]string{"majority3", "parity4"}
					if c == 20 {
						names = [2]string{"parity4", "majority3"}
					}
					q := stackV93Truth(f.X, d.Rules[stage], names[stage])
					o := original.Frames[j]
					if q != d.Q[j] || f.Y != (ys.Float64() < q) || f.X != o.X || f.Delay != o.Delay || f.Missing != o.Missing {
						t.Fatal("outcome/shift/tape", phase, c, index, j)
					}
				}
			}
		}
	}
}

func TestSpikeIndependentSeeds(t *testing.T) {
	used := map[int64]bool{}
	for p := 0; p < 2; p++ {
		for c := 0; c < 21; c++ {
			for i := 0; i < 8; i++ {
				d, _, s, err := spikeIndependentData(p, c, i)
				if err != nil {
					t.Fatal(err)
				}
				seeds := append([]int64(nil), d.Seeds[:]...)
				if s != 0 {
					seeds = append(seeds, s)
				}
				for _, seed := range seeds {
					key := seed % 2147483647
					if used[key] {
						t.Fatal("new cohort collision", seed)
					}
					used[key] = true
				}
			}
		}
	}
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300, 2042119400, 2046119500, 2050119600, 2054119700, 2062119900, 2066110000, 2078110100, 2082110200, 2090110300, 2100110400, 2110110600, 2120110800, 2130110900, 2140111100, 2144111200, 2150111300, 2160111500, 2176111600, 2180111700, 2184111800, 2188111900, 2192112000, 2196112100, 2200112200, 2204112300} {
		for p := int64(0); p < 2; p++ {
			for c := int64(0); c < 30; c++ {
				for i := int64(0); i < 128; i++ {
					for role := int64(0); role < 5; role++ {
						if used[(base+p*1000000+c*10000+i*10+role)%2147483647] {
							t.Fatal("archived quality collision", base)
						}
					}
				}
			}
		}
	}
	for _, base := range []int64{3070119900, 3090110000, 3110110100, 3130110200} {
		for mode := int64(0); mode < 2; mode++ {
			for family := int64(0); family < 4096; family++ {
				for test := int64(0); test < 64; test++ {
					if used[(base+mode*10000000+family*1000+test)%2147483647] {
						t.Fatal("archived null collision", base)
					}
				}
			}
		}
	}
}

func TestSpikeIndependentAsOf(t *testing.T) {
	for _, c := range []int{0, 5, 8, 12, 19, 20} {
		d, _, _, err := spikeIndependentData(0, c, 0)
		if err != nil {
			t.Fatal(err)
		}
		r, err := softV120Run(d, 0, c, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		for _, fit := range r.Fits {
			for window, cap := range []int{64, 32} {
				var want []int
				for i := -16; i < fit.Clock; i++ {
					if i < 0 || !d.Frames[i].Missing && i+int(d.Frames[i].Delay) <= fit.Clock {
						want = append(want, i)
					}
				}
				if len(want) > cap {
					want = want[len(want)-cap:]
				}
				if !reflect.DeepEqual(want, fit.Origins[window]) {
					t.Fatal("as-of origins", c, fit.Clock)
				}
			}
		}
		changed := d
		for i := range changed.Q {
			changed.Q[i] = .5
			if i >= 160 || changed.Frames[i].Missing {
				changed.Frames[i].Y = !changed.Frames[i].Y
			}
		}
		other, err := softV120Run(changed, 0, c, 0, 1)
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i <= 160; i++ {
			if r.Steps[i].P != other.Steps[i].P {
				t.Fatal("unavailable evidence leaked", c, i)
			}
		}
	}
}

func TestSpikeIndependentCollect(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SPIKE_INDEPENDENT_SOURCE")
	if path == "" {
		t.Skip("explicit new artifact required")
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	hashes := transferV116Hashes(t)
	for _, name := range []string{"docs/experiments/mmm-spike-independent-v1-protocol.md", "internal/observationlearners/spike_independent_test.go", "research/delayed-fixed-share.mjs", "research/spike-budget-core.mjs"} {
		raw, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		hashes[name] = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	enc := json.NewEncoder(f)
	header := struct {
		softV120Artifact
		Cohort string
	}{softV120Artifact: softV120Artifact{Version: "soft-learners-v120", TransferBase: spikeIndependentTransfer, BooleanBase: spikeIndependentBoolean, Workers: 1, Hazard: .01, GenericMass: .95, Hashes: hashes}, Cohort: "spike-independent-v1"}
	if err := enc.Encode(header); err != nil {
		t.Fatal(err)
	}
	for p := 0; p < 2; p++ {
		for c := 0; c < 21; c++ {
			for i := 0; i < 8; i++ {
				d, change, seed, err := spikeIndependentData(p, c, i)
				if err != nil {
					t.Fatal(err)
				}
				for schedule := 0; schedule < 2; schedule++ {
					r, err := softV120Run(d, p, c, i, schedule)
					if err != nil {
						t.Fatal(err)
					}
					row := struct {
						softV120Record
						Change     int
						ChangeSeed int64
					}{r, change, seed}
					if err := enc.Encode(row); err != nil {
						t.Fatal(err)
					}
				}
			}
			t.Log("completed phase/case", p, c)
		}
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
