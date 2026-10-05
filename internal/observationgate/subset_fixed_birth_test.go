package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type fixedBirthTrial struct{ *birthSubsetTrial }

func newFixedBirthTrial(base *observation.Model, scenario int, enabled bool) *fixedBirthTrial {
	return &fixedBirthTrial{newBirthSubsetTrial(base, scenario, enabled)}
}

// This arm cannot access a Reader. Its detached inputs come exclusively from
// the corresponding control's already issued, same-step observed forecast.
func (t *fixedBirthTrial) predictFrom(ref *subsetTrial, models observationpreserved.Models, step int) error {
	if ref == nil || ref.state.pending == nil || ref.state.next != step || t.state.pending != nil || t.state.next != step || t.state.split != ref.state.split || t.model != ref.model {
		return fmt.Errorf("invalid fixed observation source")
	}
	activate := t.enabled && t.activation < 0 && t.state.split && models.Local != nil
	if activate {
		w, err := localBirthWeights(t.state.mix.Weights)
		if err != nil {
			return err
		}
		t.state.mix.Weights = w
	}
	pending := *ref.state.pending
	if pending.available {
		pending.p.Experts[1] = t.state.inner.Forecast(pending.inner)
	}
	pending.p.P = t.state.mix.Forecast(pending.p.Experts)
	pending.p.Weights = t.state.mix.Weights
	if pending.p.Weights == [4]float64{} {
		pending.p.Weights = [4]float64{.7, .1, .1, .1}
	}
	pending.innerWeights = t.state.inner.Weights
	if pending.innerWeights == [4]float64{} {
		pending.innerWeights = [4]float64{.7, .1, .1, .1}
	}
	t.state.pending = &pending
	if activate {
		t.activation = step
	}
	return nil
}

type localBirthArchived struct {
	Original                               memberRecord
	Old, Mixture, BirthOld, BirthMixture   integrationArm
	ActivationOld, ActivationMixture, Fits int
	FitSeed                                int64
	FitHash                                string
	Tapes                                  [4]string
}

func TestFixedBirthParity(t *testing.T) {
	base, err := observationpreserved.Base(false)
	if err != nil {
		t.Fatal(err)
	}
	a, b := newSubsetTrial(base, 1), newSubsetTrial(base, 1)
	c, d := newBirthSubsetTrial(base, 1, true), newBirthSubsetTrial(base, 1, true)
	control, err := subsetBirthPairRun(base, "unit", 1, 0, 2026092196, a, b, c, d)
	if err != nil {
		t.Fatal(err)
	}
	e, f := newSubsetTrial(base, 1), newSubsetTrial(base, 1)
	g, h := newBirthSubsetTrial(base, 1, true), newBirthSubsetTrial(base, 1, true)
	x, y := newFixedBirthTrial(base, 1, false), newFixedBirthTrial(base, 1, false)
	if err := x.predictFrom(e, observationpreserved.Models{}, 0); err == nil {
		t.Fatal("unissued observations accepted")
	}
	actual, err := subsetFixedBirthRun(base, "unit", 1, 0, 2026092196, e, f, g, h, x, y)
	if err != nil || actual != control {
		t.Fatal("original drift", err)
	}
	for _, pair := range [][2]*subsetTrial{{a, e}, {b, f}, {c.subsetTrial, g.subsetTrial}, {d.subsetTrial, h.subsetTrial}, {e, x.subsetTrial}, {f, y.subsetTrial}} {
		if pair[0].metrics != pair[1].metrics || hex.EncodeToString(pair[0].hash.Sum(nil)) != hex.EncodeToString(pair[1].hash.Sum(nil)) {
			t.Fatal("parity drift")
		}
	}
}

func TestFixedBirthReplay(t *testing.T) {
	output := os.Getenv("EVENTFRAME_FIXED_BIRTH_ARTIFACT")
	if output == "" {
		t.Skip("opt-in fixed-observer replay")
	}
	input := "../../docs/experiments/mmm-local-birth-v1.jsonl"
	raw, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	archiveHash := sha256.Sum256(raw)
	file, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	var header struct{ Hashes map[string]string }
	if err := dec.Decode(&header); err != nil {
		t.Fatal(err)
	}
	for name, digest := range header.Hashes {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != digest {
			t.Fatal("archive changed", name)
		}
	}
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-fixed-observer-birth-v1-contract.md")
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
	if err := enc.Encode(map[string]any{"Kind": "header", "ArchiveHash": hex.EncodeToString(archiveHash[:]), "Sources": sources, "Hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	n := 0
	for {
		var a localBirthArchived
		err := dec.Decode(&a)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		scenario := -1
		for i, name := range observationpreserved.Scenarios {
			if name == a.Original.Scenario {
				scenario = i
			}
		}
		if scenario < 0 {
			t.Fatal("scenario")
		}
		seed := int64(2026092121)
		if a.Original.Split == "confirmation" {
			seed++
		} else if a.Original.Split != "design" {
			t.Fatal("cohort")
		}
		base, fh, err := memberFreshBase(a.Original.Scenario == "null", a.FitSeed)
		if err != nil || fh != a.FitHash {
			t.Fatal("fit", err)
		}
		current, old := newSubsetTrial(base, scenario), newSubsetTrial(base, scenario)
		born, oldBorn := newBirthSubsetTrial(base, scenario, true), newBirthSubsetTrial(base, scenario, true)
		fixed, oldFixed := newFixedBirthTrial(base, scenario, true), newFixedBirthTrial(base, scenario, true)
		r, err := subsetFixedBirthRun(base, a.Original.Split, scenario, a.Original.Index, seed, current, old, born, oldBorn, fixed, oldFixed)
		if err != nil || r != a.Original || current.metrics != a.Mixture || old.metrics != a.Old || born.metrics != a.BirthMixture || oldBorn.metrics != a.BirthOld || born.activation != a.ActivationMixture || oldBorn.activation != a.ActivationOld {
			t.Fatal("replay drift", n, err)
		}
		for i, trial := range []*subsetTrial{old, current, oldBorn.subsetTrial, born.subsetTrial} {
			if trial.fits != a.Fits || hex.EncodeToString(trial.hash.Sum(nil)) != a.Tapes[i] {
				t.Fatal("tape/fit drift")
			}
		}
		if fixed.activation != born.activation || oldFixed.activation != oldBorn.activation || fixed.metrics.Full.Cost != current.metrics.Full.Cost || oldFixed.metrics.Full.Cost != old.metrics.Full.Cost {
			t.Fatal("authority/acquisition changed")
		}
		if err := enc.Encode(struct {
			Original               localBirthArchived
			FixedOld, FixedMixture integrationArm
			Tapes                  [2]string
		}{a, oldFixed.metrics, fixed.metrics, [2]string{hex.EncodeToString(oldFixed.hash.Sum(nil)), hex.EncodeToString(fixed.hash.Sum(nil))}}); err != nil {
			t.Fatal(err)
		}
		n++
		if n%16 == 0 {
			if err := f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(n, "verified")
		}
	}
	if n != 160 {
		t.Fatal("count", n)
	}
}
