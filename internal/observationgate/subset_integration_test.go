package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type subsetTrial struct {
	state                    subsetState
	model                    *observationlearners.ConditionalForest
	metrics                  integrationArm
	post, fits, subsetGuides int
	hash                     hash.Hash
}

func newSubsetTrial(base *observation.Model, scenario int) *subsetTrial {
	post := 256
	if scenario == 3 {
		post = 128
	}
	return &subsetTrial{state: subsetState{base: base, enabled: true}, metrics: integrationArm{Arm: "retained_subset", SplitAt: -1}, post: post, hash: sha256.New()}
}
func (t *subsetTrial) predict(reader observation.Reader, m observationpreserved.Models, step int, seed int64) error {
	p, e := t.state.predict(reader, m, t.model, step, seed)
	if e != nil {
		return e
	}
	if p.Cost > 6 || p.P <= 0 || p.P >= 1 || p.Values&^p.Mask != 0 || p.Version != m.Version {
		return fmt.Errorf("invalid subset prediction")
	}
	return nil
}
func (t *subsetTrial) feedback(step int, y, authorized bool) error {
	if t.state.pending == nil {
		return fmt.Errorf("unprepared candidate")
	}
	p := *t.state.pending
	t.metrics.Full.add(p.p, y)
	if step >= t.post {
		t.metrics.Post.add(p.p, y)
	}
	if p.subsetGuide {
		t.subsetGuides++
	}
	if e := json.NewEncoder(t.hash).Encode(struct {
		Step                                  int
		P                                     observationpreserved.Prediction
		Inner, Weights                        [4]float64
		Available, SubsetGuide, Y, Authorized bool
	}{step, p.p, p.inner, p.innerWeights, p.available, p.subsetGuide, y, authorized}); e != nil {
		return e
	}
	if e := t.state.observe(step, y, authorized); e != nil {
		return e
	}
	if t.state.split && t.metrics.SplitAt < 0 {
		t.metrics.SplitAt = step
	}
	return nil
}
func (t *subsetTrial) fit(samples []observation.Sample) error {
	var weights [512]float64
	for i := range weights {
		weights[i] = 1
	}
	next, e := observationlearners.NewSubsetConditional(samples, weights)
	if e != nil {
		return e
	}
	t.model = next
	t.fits++
	return nil
}

func TestSubsetStateContracts(t *testing.T) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		t.Fatal(e)
	}
	control, _ := observationpreserved.New(base, "mix_mmm_ap")
	state := subsetState{base: base}
	samples := make([]observation.Sample, 64)
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i * 7 % 512), Outcome: i%2 == 0}
	}
	count, e := observation.Fit(samples)
	if e != nil {
		t.Fatal(e)
	}
	models := observationpreserved.Models{Short: count, Local: count, Pooled: count, Version: 1}
	for i := 0; i < 128; i++ {
		rd := observationexperiment.Frames(uint16(i*13%512), "control-parity")
		a, e := control.Predict(rd, models, i, 0)
		if e != nil {
			t.Fatal(e)
		}
		b, e := state.predict(rd, models, nil, i, 0)
		if e != nil || a != b {
			t.Fatal("disabled control drift", i, e)
		}
		y, auth := i%3 == 0, i == 32
		if e := control.Observe(i, y, auth); e != nil {
			t.Fatal(e)
		}
		if e := state.observe(i, y, auth); e != nil {
			t.Fatal(e)
		}
	}
	trial := newSubsetTrial(base, 1)
	if e := trial.fit(samples); e != nil {
		t.Fatal(e)
	}
	before := trial.state.mix
	if e := trial.state.observe(0, true, false); e == nil {
		t.Fatal("feedback before prediction")
	}
	if e := trial.predict(observationexperiment.Frames(511, "candidate"), models, 0, 0); e != nil {
		t.Fatal(e)
	}
	if trial.state.mix != before {
		t.Fatal("prediction changed weights")
	}
	if e := trial.predict(observationexperiment.Frames(511, "candidate"), models, 0, 0); e == nil {
		t.Fatal("duplicate prediction")
	}
	if e := trial.state.observe(1, true, true); e == nil || trial.state.mix != before {
		t.Fatal("out-of-order mutation")
	}
	if e := trial.feedback(0, true, true); e != nil {
		t.Fatal(e)
	}
	saved := trial.state.mix
	if e := trial.state.observe(0, true, true); e == nil || trial.state.mix != saved {
		t.Fatal("duplicate feedback mutation")
	}
	if base.Support() != 4096 || !trial.state.split {
		t.Fatal("base/split invariant")
	}
}

func TestSubsetIntegrationParity(t *testing.T) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		t.Fatal(e)
	}
	a, e := memberRun(base, "unit", 1, 0, 2026118299)
	if e != nil {
		t.Fatal(e)
	}
	trial := newSubsetTrial(base, 1)
	b, e := subsetIntegrationRun(base, "unit", 1, 0, 2026118299, trial)
	if e != nil || a != b {
		t.Fatal("candidate changed controls", e)
	}
	if trial.state.next != 512 || trial.metrics.Full.N != 512 || trial.fits*3 != b.Fits || trial.metrics.SplitAt != b.Arms[3].SplitAt {
		t.Fatal("integration shape")
	}
}

type subsetRecord struct {
	Original                 memberRecord
	Candidate                integrationArm
	SubsetFits, SubsetGuides int
	CandidateTape            string
}

func TestSubsetV82Experiment(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_SUBSET_OUT"), os.Getenv("EVENTFRAME_SUBSET_REPLAY")
	if output == "" && replay == "" {
		t.Skip("opt-in")
	}
	if output != "" && replay != "" {
		t.Fatal("choose write or replay")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	var paths []string
	for _, dir := range []string{"observationgate", "observation", "observationlearners", "observationpreserved", "observationrescue", "observationexperiment", "bayes", "model"} {
		ps, e := filepath.Glob(filepath.Join(root, "internal", dir, "*.go"))
		if e != nil {
			t.Fatal(e)
		}
		paths = append(paths, ps...)
	}
	for _, p := range []string{"docs/experiments/mmm-retained-subset-v82-protocol.md", "research/retained-subset-v82-summary.mjs", "go.mod", "go.sum"} {
		paths = append(paths, filepath.Join(root, p))
	}
	hashes := map[string]string{}
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		rel, e := filepath.Rel(root, p)
		if e != nil {
			t.Fatal(e)
		}
		hashes[rel] = hex.EncodeToString(h[:])
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(struct {
		Version string
		Hashes  map[string]string
	}{"v82", hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, name := range observationpreserved.Scenarios {
			base, e := observationpreserved.Base(name == "null")
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 64; i++ {
				trial := newSubsetTrial(base, scenario)
				r, e := subsetIntegrationRun(base, split, scenario, i, int64(2026118201+phase), trial)
				if e != nil {
					t.Fatal(e)
				}
				if trial.state.next != 512 || trial.state.pending != nil || trial.fits*3 != r.Fits || trial.metrics.SplitAt != r.Arms[3].SplitAt {
					t.Fatal("candidate accounting")
				}
				if e := enc.Encode(subsetRecord{r, trial.metrics, trial.fits, trial.subsetGuides, hex.EncodeToString(trial.hash.Sum(nil))}); e != nil {
					t.Fatal(e)
				}
			}
			t.Log(split, name, "complete")
		}
	}
	if replay != "" {
		want, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(want, buf.Bytes()) {
			t.Fatal("replay/source mismatch")
		}
		return
	}
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e := f.Write(buf.Bytes()); e != nil {
		t.Fatal(e)
	}
	if e := f.Sync(); e != nil {
		t.Fatal(e)
	}
}

func BenchmarkSubsetIntegration(b *testing.B) {
	base, e := observationpreserved.Base(false)
	if e != nil {
		b.Fatal(e)
	}
	samples := make([]observation.Sample, 64)
	for i := range samples {
		samples[i] = observation.Sample{Bits: uint16(i * 7 % 512), Outcome: i%3 == 0}
	}
	trial := newSubsetTrial(base, 1)
	if e := trial.fit(samples); e != nil {
		b.Fatal(e)
	}
	count, e := observation.Fit(samples)
	if e != nil {
		b.Fatal(e)
	}
	models := observationpreserved.Models{Short: count, Local: count, Pooled: count, Version: 1}
	b.Run("fit64", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if e := trial.fit(samples); e != nil {
				b.Fatal(e)
			}
		}
	})
	for _, enabled := range []bool{false, true} {
		b.Run(fmt.Sprintf("foreground-%t", enabled), func(b *testing.B) {
			s := subsetState{base: base, enabled: enabled}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if i%512 == 0 {
					s = subsetState{base: base, enabled: enabled}
				}
				rd := observationexperiment.Frames(uint16(i*13%512), "bench")
				if _, e := s.predict(rd, models, trial.model, i%512, 0); e != nil {
					b.Fatal(e)
				}
				if e := s.observe(i%512, i%3 == 0, false); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
