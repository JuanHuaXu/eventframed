package observationlearners

import (
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math/rand"
	"os"
	"testing"
)

func TestCoherentCountStress(t *testing.T) {
	path := os.Getenv("EVENTFRAME_COHERENT_COUNT_STRESS")
	if path == "" {
		t.Skip("opt-in finite sample screen")
	}
	type row struct {
		Noise         float64
		Dependent     bool
		Target        string
		N, Index      int
		Seed          int64
		Old, Coherent [4]float64
	}
	var rows []row
	for noiseIndex, noise := range []float64{.25, .5} {
		for dep := 0; dep < 2; dep++ {
			for target := 0; target < 2; target++ {
				for _, n := range []int{64, 128, 4096} {
					for index := 0; index < 16; index++ {
						seed := int64(2026092180 + noiseIndex*10000000 + dep*1000000 + target*100000 + n*20 + index)
						rng := rand.New(rand.NewSource(seed))
						samples := make([]observation.Sample, n)
						transform := func(x uint16) uint16 {
							if dep == 1 {
								return (x &^ 1) | ((x >> 2) & 1)
							}
							return x
						}
						truth := func(x uint16) bool {
							if target == 0 {
								return x&4 != 0
							}
							return (((x&2 != 0) != (x&4 != 0)) != (x&8 != 0)) != (x&16 != 0)
						}
						for i := range samples {
							x := transform(uint16(rng.Intn(512)))
							y := truth(x)
							if rng.Float64() < noise {
								y = !y
							}
							samples[i] = observation.Sample{Bits: x, Outcome: y}
						}
						old, err := observation.Fit(samples)
						if err != nil {
							t.Fatal(err)
						}
						c, err := fitCoherentCount(samples)
						if err != nil {
							t.Fatal(err)
						}
						r := row{Noise: noise, Dependent: dep == 1, Target: []string{"bit", "parity4"}[target], N: n, Index: index, Seed: seed}
						for raw := uint16(0); raw < 512; raw++ {
							x := transform(raw)
							q := noise
							if truth(x) {
								q = 1 - noise
							}
							for i, mask := range []uint16{0, 1, 31, 511} {
								a, err := old.ForecastObserved(mask, x&mask)
								if err != nil {
									t.Fatal(err)
								}
								b, err := c.Forecast(mask, x&mask)
								if err != nil {
									t.Fatal(err)
								}
								if mask == 0 && a != b {
									t.Fatal("root drift")
								}
								r.Old[i] += (q*(1-a)*(1-a) + (1-q)*a*a) / 512
								r.Coherent[i] += (q*(1-b)*(1-b) + (1-q)*b*b) / 512
							}
						}
						rows = append(rows, r)
					}
				}
			}
		}
	}
	b, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err = f.Write(append(b, '\n')); err != nil {
		t.Fatal(err)
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Log("384 independent fits; exact population risks at four fixed masks")
}
