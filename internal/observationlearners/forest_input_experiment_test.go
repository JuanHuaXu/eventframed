package observationlearners

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/rand"
	"os"
	"testing"
)

func TestForestInputExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_FOREST_INPUT")
	if path == "" {
		t.Skip("opt-in fixed input-model comparison")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	sources, hashes := map[string]string{}, map[string]string{}
	for _, p := range []string{"forest_input_experiment_test.go", "chow_input_test.go", "chow_forest_test.go", "subset.go", "conditional.go", "../../docs/experiments/mmm-forest-input-v1-contract.md"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		sources[p] = string(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "Masks": [5]uint16{0, 1, 3, 31, 511}, "Arms": [4]string{"uniform", "histogram", "tree", "forest"}}); err != nil {
		t.Fatal(err)
	}
	type row struct {
		Law, Target string
		Noise       float64
		N, Index    int
		Seed        int64
		Risk        [4][5]float64
	}
	for law := 0; law < 3; law++ {
		for target := 0; target < 2; target++ {
			for ni, noise := range []float64{.05, .25, .5} {
				for _, n := range []int{64, 128} {
					for index := 0; index < 16; index++ {
						seed := int64(2026092210 + law*1000000 + target*100000 + ni*10000 + n*20 + index)
						rng := rand.New(rand.NewSource(seed))
						transform := func(x uint16) uint16 {
							switch law {
							case 1:
								return (x &^ 1) | ((x >> 2) & 1)
							case 2:
								return (x &^ 1) | (((x >> 1) ^ (x >> 2)) & 1)
							}
							return x
						}
						truth := func(x uint16) bool {
							if target == 0 {
								return x&4 != 0
							}
							return (((x&2 != 0) != (x&4 != 0)) != (x&8 != 0)) != (x&16 != 0)
						}
						samples := make([]observation.Sample, n)
						var uniform, histogram [512]float64
						for i := range uniform {
							uniform[i] = 1
							histogram[i] = 1. / 512
						}
						for i := range samples {
							x := transform(uint16(rng.Intn(512)))
							y := truth(x)
							if rng.Float64() < noise {
								y = !y
							}
							samples[i] = observation.Sample{Bits: x, Outcome: y}
							histogram[x]++
						}
						tree, err := fitChowInput(samples)
						if err != nil {
							t.Fatal(err)
						}
						forest, err := fitHeldoutInputForest(samples)
						if err != nil {
							t.Fatal(err)
						}
						var models [4]*ConditionalForest
						for i, w := range [4][512]float64{uniform, histogram, tree.weights, forest.weights} {
							models[i], err = NewSubsetConditional(samples, w)
							if err != nil {
								t.Fatal(err)
							}
						}
						r := row{Law: []string{"uniform", "copy", "xor"}[law], Target: []string{"bit", "parity4"}[target], Noise: noise, N: n, Index: index, Seed: seed}
						for raw := uint16(0); raw < 512; raw++ {
							x := transform(raw)
							q := noise
							if truth(x) {
								q = 1 - noise
							}
							var full [4]float64
							for a, m := range models {
								for j, mask := range [5]uint16{0, 1, 3, 31, 511} {
									p, err := m.Forecast(mask, x&mask)
									if err != nil {
										t.Fatal(err)
									}
									r.Risk[a][j] += (q*(1-p)*(1-p) + (1-q)*p*p) / 512
									if mask == 511 {
										full[a] = p
									}
								}
							}
							if math.Abs(full[0]-full[1]) > 1e-12 || math.Abs(full[0]-full[2]) > 1e-12 || math.Abs(full[0]-full[3]) > 1e-12 {
								t.Fatal("full-input drift")
							}
						}
						if err := enc.Encode(r); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
			if err := f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(law, target, "96 complete")
		}
	}
}
