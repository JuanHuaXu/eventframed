package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func subsetInputFit(t *subsetTrial, samples []observation.Sample) error {
	var w [512]float64
	for i := range w {
		w[i] = 1. / 512
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return fmt.Errorf("invalid input")
		}
		w[s.Bits]++
	}
	m, err := observationlearners.NewSubsetConditional(samples, w)
	if err != nil {
		return err
	}
	t.model = m
	t.fits++
	return nil
}

// No reader access: reuse only the control's issued evidence and version.
// Different input marginals require recomputing the subset expert, not copying
// the control's forecast or retroactively substituting a fitted outcome.
func subsetInputFixedPredict(dst, src *subsetTrial) error {
	if src.state.pending == nil || dst.state.pending != nil || src.state.next != dst.state.next || src.state.split != dst.state.split {
		return fmt.Errorf("fixed input state mismatch")
	}
	p := *src.state.pending
	if p.available {
		if dst.model == nil {
			return fmt.Errorf("missing fitted input model")
		}
		v, err := dst.model.Forecast(p.p.Mask, p.p.Values)
		if err != nil {
			return err
		}
		p.inner[1] = v
		p.inner[2] = v
		p.inner[3] = v
		p.p.Experts[1] = dst.state.inner.Forecast(p.inner)
	}
	p.p.P = dst.state.mix.Forecast(p.p.Experts)
	p.p.Weights = dst.state.mix.Weights
	if p.p.Weights == [4]float64{} {
		p.p.Weights = [4]float64{.7, .1, .1, .1}
	}
	p.innerWeights = dst.state.inner.Weights
	if p.innerWeights == [4]float64{} {
		p.innerWeights = [4]float64{.7, .1, .1, .1}
	}
	dst.state.pending = &p
	return nil
}

func TestSubsetInputIntegrationParity(t *testing.T) {
	for _, ci := range []int{1, 6, 8} {
		cfg := subsetBreadthCases[ci]
		base, err := breadthBase(cfg, 2026092191+int64(ci))
		if err != nil {
			t.Fatal(err)
		}
		ref := newSubsetTrial(base, 1)
		want, err := subsetBreadthRun(base, "unit", ci, 0, 2026092192, ref, cfg)
		if err != nil {
			t.Fatal(err)
		}
		c, f, j := newSubsetTrial(base, 1), newSubsetTrial(base, 1), newSubsetTrial(base, 1)
		got, err := subsetInputRun(base, "unit", ci, 0, 2026092192, c, cfg, f, j)
		if err != nil || got != want || c.metrics != ref.metrics || hex.EncodeToString(c.hash.Sum(nil)) != hex.EncodeToString(ref.hash.Sum(nil)) {
			t.Fatal("control parity", ci, err)
		}
		for _, v := range []*subsetTrial{f, j} {
			if v.fits != c.fits || v.metrics.SplitAt != c.metrics.SplitAt {
				t.Fatal("shared evidence contract")
			}
		}
		if f.metrics.Full.Cost != c.metrics.Full.Cost || f.metrics.Post.Cost != c.metrics.Post.Cost {
			t.Fatal("fixed cost")
		}
	}
}

type subsetInputRecord struct {
	Phase                   string
	Case                    string
	Index                   int
	TrainSeed, StreamSeed   int64
	Original                memberRecord
	Control, Fixed, Coupled integrationArm
	Fits                    int
	Tapes                   [3]string
}

func TestSubsetInputIntegrationExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SUBSET_INPUT_INTEGRATION")
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
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-subset-input-integration-v1-contract.md")
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
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "Seeds": [2]int64{2026092151, 2026092152}, "PerCell": 16}); err != nil {
		t.Fatal(err)
	}
	for phase, label := range []string{"design", "confirmation"} {
		for ci, cfg := range subsetBreadthCases {
			for index := 0; index < 16; index++ {
				seed := int64(2026092151 + phase)
				train := observationpreserved.Seed(seed, ci, index, 4)
				base, err := breadthBase(cfg, train)
				if err != nil {
					t.Fatal(err)
				}
				c, a, b := newSubsetTrial(base, 1), newSubsetTrial(base, 1), newSubsetTrial(base, 1)
				original, err := subsetInputRun(base, label, ci, index, seed, c, cfg, a, b)
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
				row := subsetInputRecord{label, cfg.Name, index, train, seed, original, c.metrics, a.metrics, b.metrics, c.fits, [3]string{hex.EncodeToString(c.hash.Sum(nil)), hex.EncodeToString(a.hash.Sum(nil)), hex.EncodeToString(b.hash.Sum(nil))}}
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
