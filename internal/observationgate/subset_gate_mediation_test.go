package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type gateMediationFrame struct {
	P, Weight, Pool, Local, WithPool, WithLocal float64
	Available, Split                            bool
	Guide                                       int
}
type gateMediationStep struct {
	Step         int
	Old, Mixture gateMediationFrame
	Y            bool
}
type gateMediation struct{ Steps []gateMediationStep }

func (d *gateMediation) capture(step int, models observationpreserved.Models, current, old *subsetTrial) error {
	if len(d.Steps) != step {
		return fmt.Errorf("diagnostic order")
	}
	row := gateMediationStep{Step: step}
	for i, t := range []*subsetTrial{old, current} {
		if t.state.pending == nil {
			return fmt.Errorf("missing issued forecast")
		}
		p := t.state.pending.p
		mix := t.state.mix
		f := gateMediationFrame{P: p.P, Split: t.state.split, Guide: p.Guide}
		f.Weight = (mix.Forecast([4]float64{0, 0, 1, 0}) - mix.Forecast([4]float64{})) / (1 - 2e-6)
		if f.Weight < 0 || f.Weight > 1 || math.Abs(mix.Forecast(p.Experts)-p.P) > 1e-14 {
			return fmt.Errorf("effective mixture mismatch")
		}
		if models.Pooled != nil && models.Local != nil {
			var err error
			f.Pool, err = models.Pooled.ForecastObserved(p.Mask, p.Values)
			if err != nil {
				return err
			}
			f.Local, err = models.Local.ForecastObserved(p.Mask, p.Values)
			if err != nil {
				return err
			}
			pool, local := p.Experts, p.Experts
			pool[2] = f.Pool
			local[2] = f.Local
			f.WithPool, f.WithLocal = mix.Forecast(pool), mix.Forecast(local)
			clip := func(p float64) float64 { return math.Max(1e-6, math.Min(1-1e-6, p)) }
			if math.Abs(f.WithLocal-f.WithPool-f.Weight*(clip(f.Local)-clip(f.Pool))) > 1e-14 {
				return fmt.Errorf("slot identity")
			}
			actual := f.WithPool
			if f.Split {
				actual = f.WithLocal
			}
			if math.Abs(actual-f.P) > 1e-14 {
				return fmt.Errorf("actual slot mismatch")
			}
			f.Available = true
		}
		if i == 0 {
			row.Old = f
		} else {
			row.Mixture = f
		}
	}
	d.Steps = append(d.Steps, row)
	return nil
}
func (d *gateMediation) reveal(y bool) { d.Steps[len(d.Steps)-1].Y = y }

type gatePairArchived struct {
	Original                      memberRecord
	OldSubset, MixtureSubset      integrationArm
	Fits                          int
	FitSeed                       int64
	FitHash, OldTape, MixtureTape string
}

func TestGateMediationParity(t *testing.T) {
	base, err := observationpreserved.Base(false)
	if err != nil {
		t.Fatal(err)
	}
	a, b := newSubsetTrial(base, 1), newSubsetTrial(base, 1)
	control, err := subsetGatePairRun(base, "unit", 1, 0, 2026092198, a, b)
	if err != nil {
		t.Fatal(err)
	}
	c, e := newSubsetTrial(base, 1), newSubsetTrial(base, 1)
	d := gateMediation{}
	actual, err := subsetGateMediationRun(base, "unit", 1, 0, 2026092198, c, e, &d)
	if err != nil || control != actual || a.metrics != c.metrics || b.metrics != e.metrics || hex.EncodeToString(a.hash.Sum(nil)) != hex.EncodeToString(c.hash.Sum(nil)) || hex.EncodeToString(b.hash.Sum(nil)) != hex.EncodeToString(e.hash.Sum(nil)) {
		t.Fatal("diagnostic changed state", err)
	}
	if len(d.Steps) != 512 || d.Steps[0].Old.Available || !d.Steps[511].Old.Available {
		t.Fatal("availability")
	}
}

func TestGateMediationReplay(t *testing.T) {
	output := os.Getenv("EVENTFRAME_GATE_MEDIATION_ARTIFACT")
	if output == "" {
		t.Skip("opt-in archived diagnostic replay")
	}
	input := "../../docs/experiments/mmm-subset-gate-pair-v1.jsonl"
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
	var archivedHeader struct{ Hashes map[string]string }
	if err := dec.Decode(&archivedHeader); err != nil {
		t.Fatal(err)
	}
	for name, digest := range archivedHeader.Hashes {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != digest {
			t.Fatal("archive source changed", name)
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
	paths = append(paths, "../../go.mod", "../../go.sum", "../../docs/experiments/mmm-gate-mediation-v1-contract.md")
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
		var a gatePairArchived
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
		seed := int64(2026092111)
		if a.Original.Split == "confirmation" {
			seed++
		} else if a.Original.Split != "design" {
			t.Fatal("cohort")
		}
		base, fitHash, err := memberFreshBase(a.Original.Scenario == "null", a.FitSeed)
		if err != nil || fitHash != a.FitHash {
			t.Fatal("fit", err)
		}
		current, old := newSubsetTrial(base, scenario), newSubsetTrial(base, scenario)
		d := gateMediation{}
		r, err := subsetGateMediationRun(base, a.Original.Split, scenario, a.Original.Index, seed, current, old, &d)
		if err != nil || r != a.Original || current.metrics != a.MixtureSubset || old.metrics != a.OldSubset || current.fits != a.Fits || old.fits != a.Fits || hex.EncodeToString(current.hash.Sum(nil)) != a.MixtureTape || hex.EncodeToString(old.hash.Sum(nil)) != a.OldTape || len(d.Steps) != 512 {
			t.Fatal("actual replay drift", n, err)
		}
		if err := enc.Encode(struct {
			Original gatePairArchived
			Steps    []gateMediationStep
		}{a, d.Steps}); err != nil {
			t.Fatal(err)
		}
		n++
		if n%16 == 0 {
			if err := f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Log(n, "trajectories verified")
		}
	}
	if n != 160 {
		t.Fatal("trajectory count", n)
	}
}
