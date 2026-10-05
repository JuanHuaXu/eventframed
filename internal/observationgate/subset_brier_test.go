package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func brierPredictiveWeights(w [4]float64) ([4]float64, error) {
	prior := [4]float64{.7, .1, .1, .1}
	if w == [4]float64{} {
		return prior, nil
	}
	sum := 0.
	for _, v := range w {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return w, fmt.Errorf("invalid weight")
		}
		sum += v
	}
	if math.Abs(sum-1) > 1e-9 {
		return w, fmt.Errorf("invalid total")
	}
	for i := range w {
		w[i] = .998*w[i]/sum + .002*prior[i]
	}
	return w, nil
}
func brierSubstitution(w, p [4]float64, strong bool) (float64, error) {
	if !bayes.ValidForecastExperts(p) {
		return 0, fmt.Errorf("invalid expert")
	}
	z0, z1, linear, sum := 0., 0., 0., 0.
	for i, v := range w {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return 0, fmt.Errorf("invalid weight")
		}
		sum += v
		x := math.Max(1e-6, math.Min(1-1e-6, p[i]))
		linear += v * x
		z0 += v * math.Exp(-2*x*x)
		z1 += v * math.Exp(-2*(1-x)*(1-x))
	}
	if math.Abs(sum-1) > 1e-9 {
		return 0, fmt.Errorf("invalid total")
	}
	if !strong {
		return linear, nil
	}
	return math.Max(0, math.Min(1, .5+(math.Log(z1)-math.Log(z0))/4)), nil
}

type brierSubsetTrial struct {
	*fixedBirthTrial
	strong bool
}

func newBrierSubsetTrial(base *observation.Model, scenario int, strong bool) *brierSubsetTrial {
	return &brierSubsetTrial{newFixedBirthTrial(base, scenario, false), strong}
}
func (t *brierSubsetTrial) predictFrom(ref *subsetTrial, m observationpreserved.Models, step int) error {
	w, err := brierPredictiveWeights(t.state.mix.Weights)
	if err != nil {
		return err
	}
	if err := t.fixedBirthTrial.predictFrom(ref, m, step); err != nil {
		return err
	}
	p, err := brierSubstitution(w, t.state.pending.p.Experts, t.strong)
	if err != nil {
		return err
	}
	t.state.pending.p.P = p
	return nil
}
func (t *brierSubsetTrial) feedback(step int, y, authorized bool) error {
	if t.state.pending == nil || step != t.state.next {
		return fmt.Errorf("missing issued Brier forecast")
	}
	w, err := brierPredictiveWeights(t.state.mix.Weights)
	if err != nil {
		return err
	}
	value := 0.
	if y {
		value = 1
	}
	sum := 0.
	for i, p := range t.state.pending.p.Experts {
		x := math.Max(1e-6, math.Min(1-1e-6, p))
		w[i] *= math.Exp(-2 * (x - value) * (x - value))
		sum += w[i]
	}
	for i := range w {
		w[i] /= sum
	}
	revoke := authorized && !t.state.split
	// Retain issued-forecast scoring, inner updates and the original split action;
	// only the outer selector's likelihood update is replaced by Brier weights.
	if err := t.subsetTrial.feedback(step, y, authorized); err != nil {
		return err
	}
	if revoke {
		w[2] = math.Min(.1, w[2])
		sum = 0
		for _, v := range w {
			sum += v
		}
		for i := range w {
			w[i] /= sum
		}
	}
	t.state.mix.Weights = w
	return nil
}

func TestIntegratedBrierMath(t *testing.T) {
	rng := rand.New(rand.NewSource(2026092101))
	negative := false
	for n := 0; n < 2000; n++ {
		var w, p [4]float64
		sum := 0.
		for i := range w {
			w[i] = rng.Float64()
			sum += w[i]
			p[i] = rng.Float64()
		}
		for i := range w {
			w[i] /= sum
		}
		f, err := brierSubstitution(w, p, true)
		if err != nil {
			t.Fatal(err)
		}
		linear, _ := brierSubstitution(w, p, false)
		g := [2]float64{}
		for y := 0; y < 2; y++ {
			z := 0.
			for i, x := range p {
				x = math.Max(1e-6, math.Min(1-1e-6, x))
				z += w[i] * math.Exp(-2*(x-float64(y))*(x-float64(y)))
			}
			g[y] = -math.Log(z)
			if 2*(f-float64(y))*(f-float64(y)) > g[y]+1e-14 {
				t.Fatal("one-step bound")
			}
			if 2*(linear-float64(y))*(linear-float64(y)) > g[y]+1e-6 {
				negative = true
			}
		}
		lo, hi := math.Min(g[0], g[1]), math.Max(g[0], g[1])+2
		for i := 0; i < 80; i++ {
			mid := (lo + hi) / 2
			if math.Max(0, mid-g[0])+math.Max(0, mid-g[1]) > 2 {
				hi = mid
			} else {
				lo = mid
			}
		}
		if math.Abs(math.Max(0, ((lo+hi)/2-g[1])/2)-f) > 1e-14 {
			t.Fatal("literal substitution")
		}
		shared, err := brierPredictiveWeights(w)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := brierSubstitution(shared, p, false)
		if math.Abs(got-(bayes.ForecastMix{Weights: w}).Forecast(p)) > 1e-14 {
			t.Fatal("sharing parity")
		}
	}
	if !negative {
		t.Fatal("linear negative control vacuous")
	}
}

func TestIntegratedBrierParity(t *testing.T) {
	base, err := observationpreserved.Base(false)
	if err != nil {
		t.Fatal(err)
	}
	a, b := newSubsetTrial(base, 1), newSubsetTrial(base, 1)
	control, err := subsetGatePairRun(base, "unit", 1, 0, 2026092195, a, b)
	if err != nil {
		t.Fatal(err)
	}
	c, d := newSubsetTrial(base, 1), newSubsetTrial(base, 1)
	strong, linear := newBrierSubsetTrial(base, 1, true), newBrierSubsetTrial(base, 1, false)
	actual, err := subsetBrierRun(base, "unit", 1, 0, 2026092195, c, d, strong, linear)
	if err != nil || control != actual || c.metrics != a.metrics || d.metrics != b.metrics || hex.EncodeToString(c.hash.Sum(nil)) != hex.EncodeToString(a.hash.Sum(nil)) || hex.EncodeToString(d.hash.Sum(nil)) != hex.EncodeToString(b.hash.Sum(nil)) {
		t.Fatal("control drift", err)
	}
	for _, s := range []*brierSubsetTrial{strong, linear} {
		if s.metrics.Full.Cost != c.metrics.Full.Cost || s.metrics.SplitAt != c.metrics.SplitAt || s.fits != c.fits {
			t.Fatal("contract drift")
		}
	}
}

func TestIntegratedBrierExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_INTEGRATED_BRIER_ARTIFACT")
	if path == "" {
		t.Skip("opt-in proper-loss pilot")
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
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-integrated-brier-v1-contract.md")
	sources, hashes := map[string]string{}, map[string]string{}
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		sources[path[6:]] = string(b)
		hashes[path[6:]] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Kind": "header", "Sources": sources, "Hashes": hashes, "Seeds": [2]int64{2026092131, 2026092132}, "PerCell": 16}); err != nil {
		t.Fatal(err)
	}
	for phase := 0; phase < 2; phase++ {
		for scenario, name := range observationpreserved.Scenarios {
			for index := 0; index < 16; index++ {
				seed := int64(2026092131 + phase)
				fitSeed := observationpreserved.Seed(seed, scenario, index, 4)
				base, fitHash, err := memberFreshBase(name == "null", fitSeed)
				if err != nil {
					t.Fatal(err)
				}
				current, old := newSubsetTrial(base, scenario), newSubsetTrial(base, scenario)
				strong, linear := newBrierSubsetTrial(base, scenario, true), newBrierSubsetTrial(base, scenario, false)
				r, err := subsetBrierRun(base, []string{"design", "confirmation"}[phase], scenario, index, seed, current, old, strong, linear)
				if err != nil {
					t.Fatal(err)
				}
				for _, s := range []*brierSubsetTrial{strong, linear} {
					if s.metrics.Full.Cost != current.metrics.Full.Cost || s.metrics.SplitAt != current.metrics.SplitAt || s.fits != current.fits {
						t.Fatal("contract drift")
					}
				}
				if err := enc.Encode(struct {
					Original                memberRecord
					Control, Strong, Linear integrationArm
					Fits                    int
					FitSeed                 int64
					FitHash                 string
					Tapes                   [3]string
				}{r, current.metrics, strong.metrics, linear.metrics, current.fits, fitSeed, fitHash, [3]string{hex.EncodeToString(current.hash.Sum(nil)), hex.EncodeToString(strong.hash.Sum(nil)), hex.EncodeToString(linear.hash.Sum(nil))}}); err != nil {
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
