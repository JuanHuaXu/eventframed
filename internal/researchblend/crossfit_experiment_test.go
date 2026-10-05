package researchblend

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type armV32 struct {
	armV31
	LOO [][3]float64
}
type worldV32 struct {
	Kind, Geometry, Regime              string
	World, Partition                    int
	Seed                                int64
	Base, Coordinates, Rates, LeafMeans []float64
	Labels                              []bool
	Arms                                []armV32
}

func runCrossfitV32(base, coordinates []float64, labels []bool, policy string, seed int64) armV32 {
	start := time.Now()
	a := armV32{armV31: armV31{armV30: armV30{Model: "crossfit", Policy: policy, Trace: make([]traceV30, 0, 32)}, Issued: make([][3]float64, 0)}, LOO: make([][3]float64, 0, 32)}
	m, err := NewCrossfit(base, coordinates)
	if err != nil {
		panic(err)
	}
	a.SetupNS = time.Since(start).Nanoseconds()
	rng := rand.New(rand.NewSource(seed + 303))
	for k := 0; k < 32; k++ {
		s := time.Now()
		choice, e := m.Select(policy, rng)
		if e != nil {
			panic(e)
		}
		q, e := m.Predict(choice.Index)
		if e != nil {
			panic(e)
		}
		a.SelectionNS += time.Since(s).Nanoseconds()
		// Selected labels are unavailable until AFTER today's forecast.
		y := labels[choice.Index]
		a.Trace = append(a.Trace, traceV30{choice.Index, y, q, choice.Probability, choice.Score})
		s = time.Now()
		if e = m.Observe(choice.Index, y); e != nil {
			panic(e)
		}
		a.UpdateNS += time.Since(s).Nanoseconds()
	}
	s := time.Now()
	a.Forecast = make([]float64, len(base))
	for i := range base {
		a.Forecast[i], err = m.Predict(i)
		if err != nil {
			panic(err)
		}
	}
	a.Weights = m.Weights()
	for _, v := range a.Trace {
		row, e := m.trainingRow(v.Index)
		if e != nil {
			panic(e)
		}
		a.LOO = append(a.LOO, row)
	}
	a.FinalNS = time.Since(s).Nanoseconds()
	a.TotalNS = time.Since(start).Nanoseconds()
	return a
}

func makeWorldV32(seedBase int64, g, r, w int) worldV32 {
	old := makeWorldV31(seedBase, g, r, w)
	result := worldV32{Kind: old.Kind, Geometry: old.Geometry, Regime: old.Regime, World: old.World, Partition: old.Partition, Seed: old.Seed, Base: old.Base, Coordinates: old.Coordinates, Rates: old.Rates, LeafMeans: old.LeafMeans, Labels: old.Labels}
	for _, a := range old.Arms {
		result.Arms = append(result.Arms, armV32{armV31: a, LOO: make([][3]float64, 0)})
	}
	for _, policy := range []string{"stratified_random", "random"} {
		a := runCrossfitV32(result.Base, result.Coordinates, result.Labels, policy, result.Seed)
		scoreV30(&a.armV30, result.Rates)
		result.Arms = append(result.Arms, a)
	}
	return result
}

var sourcesV32 = append(append([]string(nil), sourcesV31...), "internal/researchblend/crossfit.go", "internal/researchblend/crossfit_test.go", "internal/researchcalibration/loo.go", "internal/researchcalibration/loo_test.go", "internal/researchpartition/loo.go", "internal/researchpartition/loo_test.go", "internal/researchblend/crossfit_experiment_test.go", "research/crossfit-v32-verify.mjs", "docs/experiments/mmm-crossfit-v32-protocol.md", "docs/experiments/mmm-crossfit-v32-preflight.md")

func TestCrossfitCollectorNoFutureV32(t *testing.T) {
	base, coordinate := fixture(150)
	labels := make([]bool, 150)
	for i := range labels {
		labels[i] = i%3 == 0
	}
	for _, policy := range []string{"random", "stratified_random"} {
		a := runCrossfitV32(base, coordinate, labels, policy, 941)
		changed := append([]bool(nil), labels...)
		seen := make([]bool, 150)
		for _, row := range a.Trace {
			seen[row.Index] = true
		}
		for i := range changed {
			if !seen[i] {
				changed[i] = !changed[i]
			}
		}
		b := runCrossfitV32(base, coordinate, changed, policy, 941)
		if !reflect.DeepEqual(a.Trace, b.Trace) || !reflect.DeepEqual(a.Forecast, b.Forecast) || !reflect.DeepEqual(a.LOO, b.LOO) || a.Weights != b.Weights {
			t.Fatal("unobserved labels changed collected law or LOO rows")
		}
		c := runV30(base, coordinate, labels, "blend", policy, 941)
		for i, row := range a.Trace {
			if row.Index != c.Trace[i].Index || row.Useful != c.Trace[i].Useful || row.Probability != c.Trace[i].Probability {
				t.Fatal("collector evidence differs from matched control")
			}
		}
	}
}

func TestExperimentV32(t *testing.T) {
	out := os.Getenv("EVENTFRAME_CROSSFIT_V32_OUT")
	if out == "" {
		t.Skip("opt-in omitted-member outcome experiment")
	}
	split := os.Getenv("EVENTFRAME_CROSSFIT_V32_SPLIT")
	seed := int64(2026103203)
	if split == "confirmation" {
		seed = 2026103204
	} else if split != "design" {
		t.Fatal("unknown split")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{}
	for _, source := range sourcesV32 {
		b, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		hashes[source] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"Kind": "manifest", "Split": split, "SeedBase": seed, "Worlds": 768, "Sources": hashes}); err != nil {
		t.Fatal(err)
	}
	for g := 0; g < 2; g++ {
		for r := range regimesV30 {
			for w := 0; w < 32; w++ {
				if err = enc.Encode(makeWorldV32(seed, g, r, w)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
}
