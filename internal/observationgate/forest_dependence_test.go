package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

type forestDependenceCase struct {
	subsetBreadthCase
	Pattern int
}

func forestDependenceCases() []forestDependenceCase {
	var out []forestDependenceCase
	for p, name := range []string{"noise10", "noise30", "appears", "disappears", "reverses"} {
		for _, target := range []string{"bit", "xor2"} {
			out = append(out, forestDependenceCase{subsetBreadthCase{Name: name + "_" + target, Mode: "member_shift", Target: target}, p + 1})
		}
	}
	return append(out, forestDependenceCase{subsetBreadthCase{Name: "stable_noise10", Mode: "stable", Target: "bit"}, 1}, forestDependenceCase{subsetBreadthCase{Name: "null_uniform", Mode: "null", Target: "bit"}, 0})
}
func forestDependenceInput(x uint16, p, step int, rng *rand.Rand) uint16 {
	switch p {
	case 0:
		return x
	case 1, 2:
		v := breadthInput(x, true)
		rate := .1
		if p == 2 {
			rate = .3
		}
		if rng.Float64() < rate {
			v ^= 1
		}
		if rng.Float64() < rate {
			v ^= 16
		}
		return v
	case 3:
		if step < 256 {
			return x
		}
		return breadthInput(x, true)
	case 4:
		if step < 256 {
			return breadthInput(x, true)
		}
		return x
	case 5:
		if step < 256 {
			return breadthInput(x, true)
		}
		return breadthInput(x, true) ^ 17
	case 6:
		return breadthInput(x, true)
	default:
		panic("undeclared input pattern")
	}
}
func forestDependenceBase(cfg forestDependenceCase, seed int64) (*observation.Model, error) {
	rng := rand.New(rand.NewSource(seed))
	samples := make([]observation.Sample, 4096)
	for i := range samples {
		x := forestDependenceInput(uint16(rng.Intn(512)), cfg.Pattern, -1, rng)
		samples[i] = observation.Sample{Bits: x, Outcome: breadthLabel(x, false, cfg.Mode == "null", rng, cfg.Target)}
	}
	return observation.Fit(samples)
}
func TestForestDependenceContracts(t *testing.T) {
	seen := map[int64]bool{}
	for phase := 0; phase < 2; phase++ {
		for ci := range forestDependenceCases() {
			for index := 0; index < 16; index++ {
				for role := 0; role < 7; role++ {
					seed := observationpreserved.Seed(int64(2026092231+100*phase), ci, index, role)
					if seen[seed] {
						t.Fatal("seed collision")
					}
					seen[seed] = true
				}
			}
		}
	}
	for x := uint16(0); x < 512; x++ {
		for _, step := range []int{255, 256} {
			b := breadthInput(x, true)
			a := forestDependenceInput(x, 3, step, nil)
			d := forestDependenceInput(x, 4, step, nil)
			v := forestDependenceInput(x, 5, step, nil)
			if step == 255 {
				if a != x || d != b || v != b {
					t.Fatal("preboundary")
				}
			} else {
				if a != b || d != x || v != b^17 {
					t.Fatal("postboundary")
				}
			}
		}
	}
	cfg := forestDependenceCase{subsetBreadthCase{Name: "parity", Mode: "member_shift", Target: "bit", Dependent: true}, 6}
	base, err := breadthBase(cfg.subsetBreadthCase, 2026092296)
	if err != nil {
		t.Fatal(err)
	}
	c, a, b := newSubsetTrial(base, 1), newSubsetTrial(base, 1), newSubsetTrial(base, 1)
	want, err := forestInputRun(base, "unit", 0, 0, 2026092297, c, cfg.subsetBreadthCase, a, b)
	if err != nil {
		t.Fatal(err)
	}
	cc, aa, bb := newSubsetTrial(base, 1), newSubsetTrial(base, 1), newSubsetTrial(base, 1)
	got, err := forestDependenceRun(base, "unit", 0, 0, 2026092297, cc, cfg, aa, bb)
	if err != nil || got != want {
		t.Fatal("legacy parity", err)
	}
	for i, x := range []*subsetTrial{c, a, b} {
		y := []*subsetTrial{cc, aa, bb}[i]
		if x.metrics != y.metrics || hex.EncodeToString(x.hash.Sum(nil)) != hex.EncodeToString(y.hash.Sum(nil)) {
			t.Fatal("tape parity")
		}
	}
}
func TestForestDependenceExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_FOREST_DEPENDENCE")
	if path == "" {
		t.Skip("opt-in input integration")
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
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-forest-dependence-v1-contract.md")
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
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "Seeds": [2]int64{2026092231, 2026092331}, "PerCell": 16}); err != nil {
		t.Fatal(err)
	}
	for phase, label := range []string{"design", "confirmation"} {
		for ci, cfg := range forestDependenceCases() {
			for index := 0; index < 16; index++ {
				seed := int64(2026092231 + 100*phase)
				train := observationpreserved.Seed(seed, ci, index, 4)
				base, err := forestDependenceBase(cfg, train)
				if err != nil {
					t.Fatal(err)
				}
				c, a, b := newSubsetTrial(base, 1), newSubsetTrial(base, 1), newSubsetTrial(base, 1)
				original, err := forestDependenceRun(base, label, ci, index, seed, c, cfg, a, b)
				if err != nil {
					t.Fatal(err)
				}
				for _, v := range []*subsetTrial{a, b} {
					if v.fits != c.fits || v.metrics.SplitAt != c.metrics.SplitAt {
						t.Fatal("fit/split mismatch")
					}
				}
				if a.metrics.Full.Cost != c.metrics.Full.Cost || a.metrics.Post.Cost != c.metrics.Post.Cost {
					t.Fatal("acquisition mismatch")
				}
				row := forestInputRecord{label, cfg.Name, index, train, seed, original, c.metrics, a.metrics, b.metrics, c.fits, [3]string{hex.EncodeToString(c.hash.Sum(nil)), hex.EncodeToString(a.hash.Sum(nil)), hex.EncodeToString(b.hash.Sum(nil))}}
				if err := enc.Encode(row); err != nil {
					t.Fatal(err)
				}
			}
			if err := f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(label, cfg.Name, "16 complete")
		}
	}
}
