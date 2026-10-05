package observationlearners

import (
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math"
	"math/rand"
	"os"
	"testing"
)

// One joint Dirichlet prior with total mass2, uniform inputs and symmetric
// binary outcomes. Every coarser state sums the same full-cell posterior mass.
// This changes fine-context regularization; it is not old-count forecast parity.
func fitCoherentCount(samples []observation.Sample) (*ConditionalForest, error) {
	if len(samples) == 0 || len(samples) > 8192 {
		return nil, fmt.Errorf("invalid sample count")
	}
	m := new(ConditionalForest)
	for x := uint16(0); x < 512; x++ {
		m.cells[partialIndex(511, x)] = conditionalCell{2. / 512, 1. / 512}
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return nil, fmt.Errorf("invalid bits")
		}
		c := &m.cells[partialIndex(511, s.Bits)]
		c.mass++
		if s.Outcome {
			c.weighted++
		}
	}
	for i := len(m.cells) - 1; i >= 0; i-- {
		v, p := i, 1
		for bit := 0; bit < 9; bit++ {
			if v%3 == 0 {
				a, b := m.cells[i+p], m.cells[i+2*p]
				m.cells[i] = conditionalCell{a.mass + b.mass, a.weighted + b.weighted}
				break
			}
			v /= 3
			p *= 3
		}
	}
	return m, nil
}

func TestCoherentCountContracts(t *testing.T) {
	samples := []observation.Sample{{Bits: 0, Outcome: true}, {Bits: 1, Outcome: true}}
	m, err := fitCoherentCount(samples)
	if err != nil {
		t.Fatal(err)
	}
	for mask := uint16(0); mask < 512; mask++ {
		for value := mask; ; value = (value - 1) & mask {
			mass, yes := 0., 0.
			for x := uint16(0); x < 512; x++ {
				if x&mask == value {
					mass += 2. / 512
					yes += 1. / 512
				}
			}
			for _, s := range samples {
				if s.Bits&mask == value {
					mass++
					if s.Outcome {
						yes++
					}
				}
			}
			c := m.cells[partialIndex(mask, value)]
			p, err := m.Forecast(mask, value)
			if err != nil || c.mass != mass || c.weighted != yes || p != yes/mass {
				t.Fatal("literal marginal", mask, value, err)
			}
			for bit := uint16(1); bit < 512; bit <<= 1 {
				if mask&bit != 0 {
					continue
				}
				a, b := m.cells[partialIndex(mask|bit, value)], m.cells[partialIndex(mask|bit, value|bit)]
				if a.mass+b.mass != c.mass || a.weighted+b.weighted != c.weighted {
					t.Fatal("partition identity")
				}
			}
			if value == 0 {
				break
			}
		}
	}
	for _, v := range []uint16{0, 1} {
		p, _ := m.Forecast(1, v)
		if p != .75 {
			t.Fatal("witness repair")
		}
	}
	samples[0].Outcome = false
	p, _ := m.Forecast(0, 0)
	if p != .75 {
		t.Fatal("aliased fit")
	}
	for _, bad := range [][]observation.Sample{nil, {{Bits: 512}}, make([]observation.Sample, 8193)} {
		if _, err := fitCoherentCount(bad); err == nil {
			t.Fatal("invalid accepted")
		}
	}
	if _, err := m.Forecast(0, 1); err == nil {
		t.Fatal("invalid partial")
	}
	// Equal positive/negative evidence remains neutral everywhere.
	neutral, _ := fitCoherentCount([]observation.Sample{{Bits: 1}, {Bits: 1, Outcome: true}})
	for _, c := range neutral.cells {
		if c.weighted/c.mass != .5 {
			t.Fatal("neutrality")
		}
	}
}

func TestCoherentCountExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_COHERENT_COUNT")
	if path == "" {
		t.Skip("opt-in finite sample screen")
	}
	type row struct {
		Dependent     bool
		Target        string
		N, Index      int
		Seed          int64
		Old, Coherent [4]float64
	}
	var rows []row
	for dep := 0; dep < 2; dep++ {
		for target := 0; target < 2; target++ {
			for _, n := range []int{64, 128, 4096} {
				for index := 0; index < 16; index++ {
					seed := int64(2026092170 + dep*1000000 + target*100000 + n*20 + index)
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
						if rng.Float64() < .05 {
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
					r := row{Dependent: dep == 1, Target: []string{"bit", "parity4"}[target], N: n, Index: index, Seed: seed}
					for raw := uint16(0); raw < 512; raw++ {
						x := transform(raw)
						q := .05
						if truth(x) {
							q = .95
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
	t.Log("192 independent fits; exact population risks at four fixed masks")
}

var coherentCountSink *ConditionalForest

func BenchmarkCoherentCountFit(b *testing.B) {
	samples := make([]observation.Sample, 64)
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i * 7 % 512), Outcome: i%3 == 0}
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m, err := fitCoherentCount(samples)
		if err != nil {
			b.Fatal(err)
		}
		coherentCountSink = m
	}
}

func TestCoherentCountBounded(t *testing.T) {
	m, _ := fitCoherentCount([]observation.Sample{{Bits: 0, Outcome: true}})
	for _, c := range m.cells {
		p := c.weighted / c.mass
		if math.IsNaN(p) || p <= 0 || p >= 1 {
			t.Fatal("invalid probability")
		}
	}
}
