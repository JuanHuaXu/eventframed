package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type birthSubsetTrial struct {
	*subsetTrial
	enabled    bool
	activation int
}

func newBirthSubsetTrial(base *observation.Model, scenario int, enabled bool) *birthSubsetTrial {
	return &birthSubsetTrial{subsetTrial: newSubsetTrial(base, scenario), enabled: enabled, activation: -1}
}
func localBirthWeights(w [4]float64) ([4]float64, error) {
	sum := 0.
	other := 0.
	for j, v := range w {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return w, fmt.Errorf("invalid weights")
		}
		sum += v
		if j != 2 {
			other += v
		}
	}
	if math.Abs(sum-1) > 1e-9 || other <= 0 {
		return w, fmt.Errorf("invalid weight mass")
	}
	for j := range w {
		if j == 2 {
			w[j] = .1
		} else {
			w[j] *= .9 / other
		}
	}
	return w, nil
}
func (t *birthSubsetTrial) predict(reader observation.Reader, m observationpreserved.Models, step int, seed int64) error {
	if t.state.pending != nil || step != t.state.next {
		return fmt.Errorf("invalid birth prediction order")
	}
	activate := t.enabled && t.activation < 0 && t.state.split && m.Local != nil
	previous := t.state.mix
	if activate {
		w, err := localBirthWeights(previous.Weights)
		if err != nil {
			return err
		}
		t.state.mix.Weights = w
	}
	if err := t.subsetTrial.predict(reader, m, step, seed); err != nil {
		t.state.mix = previous
		return err
	}
	if activate {
		t.activation = step
	}
	return nil
}

func TestLocalBirthWeights(t *testing.T) {
	for _, w := range [][4]float64{{.7, .1, .1, .1}, {.8, .19, .0001, .0099}, {.001, .001, .997, .001}} {
		next, err := localBirthWeights(w)
		if err != nil {
			t.Fatal(err)
		}
		if next[2] != .1 || math.Abs(next[0]+next[1]+next[2]+next[3]-1) > 1e-14 {
			t.Fatal(next)
		}
		for _, j := range []int{1, 3} {
			if math.Abs(next[0]*w[j]-next[j]*w[0]) > 1e-14 {
				t.Fatal("relative history changed")
			}
		}
	}
	for _, w := range [][4]float64{{}, {0, 0, 1, 0}, {math.NaN(), 0, 0, 0}, {.2, .2, .2, .2}} {
		if _, err := localBirthWeights(w); err == nil {
			t.Fatal("invalid accepted")
		}
	}
}

func BenchmarkLocalBirthWeights(b *testing.B) {
	w := [4]float64{.8, .19, .0001, .0099}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := localBirthWeights(w)
		if err != nil || v[2] != .1 {
			b.Fatal(v, err)
		}
	}
}

func TestLocalBirthPairParity(t *testing.T) {
	for scenario := 0; scenario < 5; scenario++ {
		base, err := observationpreserved.Base(scenario == 4)
		if err != nil {
			t.Fatal(err)
		}
		control, oldControl := newSubsetTrial(base, scenario), newSubsetTrial(base, scenario)
		a, err := subsetGatePairRun(base, "unit", scenario, 0, 2026092197, control, oldControl)
		if err != nil {
			t.Fatal(err)
		}
		current, old := newSubsetTrial(base, scenario), newSubsetTrial(base, scenario)
		born, oldBorn := newBirthSubsetTrial(base, scenario, true), newBirthSubsetTrial(base, scenario, false)
		b, err := subsetBirthPairRun(base, "unit", scenario, 0, 2026092197, current, old, born, oldBorn)
		if err != nil || a != b || current.metrics != control.metrics || old.metrics != oldControl.metrics || hex.EncodeToString(current.hash.Sum(nil)) != hex.EncodeToString(control.hash.Sum(nil)) || hex.EncodeToString(old.hash.Sum(nil)) != hex.EncodeToString(oldControl.hash.Sum(nil)) {
			t.Fatal("controls changed", scenario, err)
		}
		if oldBorn.metrics != old.metrics || hex.EncodeToString(oldBorn.hash.Sum(nil)) != hex.EncodeToString(old.hash.Sum(nil)) || oldBorn.activation != -1 {
			t.Fatal("disabled drift")
		}
		if born.metrics.SplitAt != current.metrics.SplitAt || born.fits != current.fits {
			t.Fatal("authority/fit drift")
		}
		if born.activation >= 0 && (born.activation <= born.metrics.SplitAt || born.activation > 511) {
			t.Fatal("activation precedes authorization")
		}
		if current.metrics.SplitAt < 0 && born.activation >= 0 {
			t.Fatal("unauthorized activation")
		}
	}
}

func TestLocalBirthExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_LOCAL_BIRTH_ARTIFACT")
	if path == "" {
		t.Skip("opt-in birth prior pilot")
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
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-local-birth-v1-contract.md")
	sources, hashes := map[string]string{}, map[string]string{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(data)
		sources[path[6:]] = string(data)
		hashes[path[6:]] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Kind": "header", "Sources": sources, "Hashes": hashes, "PerCell": 16, "Seeds": [2]int64{2026092121, 2026092122}}); err != nil {
		t.Fatal(err)
	}
	for phase := 0; phase < 2; phase++ {
		for scenario, name := range observationpreserved.Scenarios {
			for index := 0; index < 16; index++ {
				seed := int64(2026092121 + phase)
				fitSeed := observationpreserved.Seed(seed, scenario, index, 4)
				base, fitHash, err := memberFreshBase(name == "null", fitSeed)
				if err != nil {
					t.Fatal(err)
				}
				current, old := newSubsetTrial(base, scenario), newSubsetTrial(base, scenario)
				born, oldBorn := newBirthSubsetTrial(base, scenario, true), newBirthSubsetTrial(base, scenario, true)
				r, err := subsetBirthPairRun(base, []string{"design", "confirmation"}[phase], scenario, index, seed, current, old, born, oldBorn)
				if err != nil {
					t.Fatal(err)
				}
				if born.metrics.SplitAt != current.metrics.SplitAt || oldBorn.metrics.SplitAt != old.metrics.SplitAt || born.fits != current.fits || oldBorn.fits != current.fits {
					t.Fatal("authority/fit changed")
				}
				for _, arm := range []*birthSubsetTrial{born, oldBorn} {
					if arm.activation >= 0 && arm.activation <= arm.metrics.SplitAt {
						t.Fatal("early activation")
					}
				}
				if err := enc.Encode(struct {
					Original                               memberRecord
					Old, Mixture, BirthOld, BirthMixture   integrationArm
					ActivationOld, ActivationMixture, Fits int
					FitSeed                                int64
					FitHash                                string
					Tapes                                  [4]string
				}{r, old.metrics, current.metrics, oldBorn.metrics, born.metrics, oldBorn.activation, born.activation, current.fits, fitSeed, fitHash, [4]string{hex.EncodeToString(old.hash.Sum(nil)), hex.EncodeToString(current.hash.Sum(nil)), hex.EncodeToString(oldBorn.hash.Sum(nil)), hex.EncodeToString(born.hash.Sum(nil))}}); err != nil {
					t.Fatal(err)
				}
			}
			if err := f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(phase, name, "16 complete")
		}
	}
}
