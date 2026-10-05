package observationlearners

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Population truth is evaluator-only. Every fitted arm receives identical past
// samples; oracle input weights diagnose headroom and are not deployable evidence.
func TestSubsetInputDiagnostic(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SUBSET_INPUT_DIAGNOSTIC")
	if path == "" {
		t.Skip("opt-in population diagnostic")
	}
	type record struct {
		Generator          string
		N, Index           int
		Seed               int64
		Partial, Full      [3]float64
		OraclePartialFloor float64
	}
	var rows []record
	for _, dependent := range []bool{false, true} {
		for _, n := range []int{64, 128} {
			for index := 0; index < 16; index++ {
				seed := int64(2026092140 + n*100 + index)
				if dependent {
					seed += 100000
				}
				rng := rand.New(rand.NewSource(seed))
				transform := func(x uint16) uint16 {
					if dependent {
						x = (x &^ 1) | ((x >> 2) & 1)
					}
					return x
				}
				samples := make([]observation.Sample, n)
				var uniform, empirical, oracle [512]float64
				for x := range uniform {
					uniform[x] = 1
					empirical[x] = 1. / 512
					// Positive support is required by the existing constructor.
					oracle[x] = 1e-12
				}
				for x := uint16(0); x < 512; x++ {
					oracle[transform(x)] += 1. / 512
				}
				for i := range samples {
					x := transform(uint16(rng.Intn(512)))
					y := x&4 != 0
					if rng.Float64() < .05 {
						y = !y
					}
					samples[i] = observation.Sample{Bits: x, Outcome: y}
					empirical[x]++
				}
				r := record{Generator: "fair", N: n, Index: index, Seed: seed, OraclePartialFloor: .25}
				if dependent {
					r.Generator = "copy_bit2_to_bit0"
					r.OraclePartialFloor = .0475
				}
				var models [3]*ConditionalForest
				for a, w := range [3][512]float64{uniform, empirical, oracle} {
					m, err := NewSubsetConditional(samples, w)
					if err != nil {
						t.Fatal(err)
					}
					models[a] = m
				}
				for raw := uint16(0); raw < 512; raw++ {
					x := transform(raw)
					q := .05
					if x&4 != 0 {
						q = .95
					}
					var full [3]float64
					for a, m := range models {
						p, err := m.Forecast(1, x&1)
						if err != nil {
							t.Fatal(err)
						}
						f, err := m.Forecast(511, x)
						if err != nil {
							t.Fatal(err)
						}
						full[a] = f
						r.Partial[a] += (q*(1-p)*(1-p) + (1-q)*p*p) / 512
						r.Full[a] += (q*(1-f)*(1-f) + (1-q)*f*f) / 512
					}
					if math.Abs(full[0]-full[1]) > 1e-12 || math.Abs(full[0]-full[2]) > 1e-12 {
						t.Fatal("full-input negative control changed")
					}
				}
				for _, risk := range r.Partial {
					if risk < r.OraclePartialFloor-1e-12 {
						t.Fatal("below population floor")
					}
				}
				rows = append(rows, r)
			}
		}
	}
	sources := map[string]string{}
	hashes := map[string]string{}
	for _, name := range []string{"subset_input_diagnostic_test.go", "subset.go", "conditional.go", "../../docs/experiments/mmm-subset-input-diagnostic-v1-contract.md"} {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(struct {
		Sources, Hashes map[string]string
		Arms            [3]string
		Records         []record
	}{sources, hashes, [3]string{"uniform", "empirical", "oracle_input"}, rows}); err != nil {
		t.Fatal(err)
	}
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Log("64 fits evaluated on exact finite populations; full-input parity passed")
}
