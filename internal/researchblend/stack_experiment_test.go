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

type armV31 struct {
	armV30
	Issued [][3]float64
}
type worldV31 struct {
	Kind, Geometry, Regime              string
	World, Partition                    int
	Seed                                int64
	Base, Coordinates, Rates, LeafMeans []float64
	Labels                              []bool
	Arms                                []armV31
}

func runStackV31(base, r []float64, labels []bool, policy string, seed int64) armV31 {
	start := time.Now()
	a := armV31{armV30: armV30{Model: "stack", Policy: policy, Trace: make([]traceV30, 0, 32)}, Issued: make([][3]float64, 0, 32)}
	m, err := NewStack(base, r)
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
		ticket, e := m.Issue(choice.Index)
		if e != nil {
			panic(e)
		}
		a.SelectionNS += time.Since(s).Nanoseconds()
		// Hidden selected outcomes are read only AFTER ticket issuance.
		y := labels[choice.Index]
		a.Trace = append(a.Trace, traceV30{choice.Index, y, ticket.Forecast(), choice.Probability, choice.Score})
		a.Issued = append(a.Issued, m.pending[choice.Index].p)
		s = time.Now()
		if e = m.Resolve(ticket, y); e != nil {
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
	a.FinalNS = time.Since(s).Nanoseconds()
	a.TotalNS = time.Since(start).Nanoseconds()
	return a
}
func makeWorldV31(seedBase int64, g, r, w int) worldV31 {
	old := makeWorldV30(seedBase, g, r, w)
	result := worldV31{Kind: old.Kind, Geometry: old.Geometry, Regime: old.Regime, World: old.World, Partition: old.Partition, Seed: old.Seed, Base: old.Base, Coordinates: old.Coordinates, Rates: old.Rates, LeafMeans: old.LeafMeans, Labels: old.Labels}
	for _, a := range old.Arms {
		result.Arms = append(result.Arms, armV31{armV30: a, Issued: make([][3]float64, 0)})
	}
	for _, policy := range []string{"stratified_random", "random", "uncertainty", "disagreement"} {
		a := runStackV31(result.Base, result.Coordinates, result.Labels, policy, result.Seed)
		scoreV30(&a.armV30, result.Rates)
		result.Arms = append(result.Arms, a)
	}
	return result
}
func TestStackNoFutureAndMatchedV31(t *testing.T) {
	b, r := fixture(150)
	labels := make([]bool, 150)
	for i := range labels {
		labels[i] = i%3 == 0
	}
	for _, policy := range []string{"random", "stratified_random", "uncertainty", "disagreement"} {
		a := runStackV31(b, r, labels, policy, 973)
		changed := append([]bool(nil), labels...)
		seen := make([]bool, 150)
		for _, v := range a.Trace {
			seen[v.Index] = true
		}
		for i := range changed {
			if !seen[i] {
				changed[i] = !changed[i]
			}
		}
		c := runStackV31(b, r, changed, policy, 973)
		if !reflect.DeepEqual(a.Trace, c.Trace) || !reflect.DeepEqual(a.Issued, c.Issued) || !reflect.DeepEqual(a.Forecast, c.Forecast) || a.Weights != c.Weights {
			t.Fatal("unused label entered prequential scoring")
		}
	}
	a := runStackV31(b, r, labels, "stratified_random", 973)
	for _, model := range []string{"local", "old", "partition", "blend"} {
		c := runV30(b, r, labels, model, "stratified_random", 973)
		for i, v := range a.Trace {
			if v.Index != c.Trace[i].Index || v.Useful != c.Trace[i].Useful || v.Probability != c.Trace[i].Probability {
				t.Fatal("matched control evidence differs")
			}
		}
	}
}

var sourcesV31 = []string{"internal/researchblend/stack.go", "internal/researchblend/stack_test.go", "internal/researchblend/stack_experiment_test.go", "internal/researchblend/model.go", "internal/researchblend/experiment_test.go", "internal/researchblend/model_test.go", "internal/researchcalibration/model.go", "internal/researchpartition/model.go", "research/stack-v31-verify.mjs", "docs/experiments/mmm-stack-v31-protocol.md", "docs/experiments/mmm-stack-v31-preflight.md"}

func TestExperimentV31(t *testing.T) {
	out := os.Getenv("EVENTFRAME_STACK_V31_OUT")
	if out == "" {
		t.Skip("opt-in prequential outcome experiment")
	}
	split := os.Getenv("EVENTFRAME_STACK_V31_SPLIT")
	seed := int64(2026103103)
	if split == "confirmation" {
		seed = 2026103104
	} else if split != "design" {
		t.Fatal("unknown split")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	hashes := map[string]string{}
	for _, source := range sourcesV31 {
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
				record := makeWorldV31(seed, g, r, w)
				if err = enc.Encode(record); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
}
