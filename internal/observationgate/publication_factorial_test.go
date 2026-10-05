package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type factorialRecord struct {
	Phase, Case        string
	Index              int
	Seed, TrainSeed    int64
	Immediate, Delayed factorialResult
}

func checkFactorialEndpoints(t testing.TB, got factorialResult, want publicationResult) {
	t.Helper()
	for a, b := range map[int]int{0: 0, 3: 1} {
		if got.Arms[a].Full != want.Arms[b].Full || got.Arms[a].Post != want.Arms[b].Post || got.Arms[a].SplitAt != want.Arms[b].SplitAt || got.Applied[a] != want.Applied[b] || got.Stale[a] != want.Stale[b] || got.Censored[a] != want.Censored[b] {
			t.Fatal("factorial endpoint metrics")
		}
		for n, f := range got.Frames {
			g := want.Frames[n]
			if f.Predictions[a] != g.Predictions[b] || f.X != g.X || f.RX != g.RX || f.Y != g.Y || f.RY != g.RY || f.Arrival != g.Arrival || f.Missing != g.Missing || f.Audit != g.Audit {
				t.Fatal("factorial endpoint forecast/tape")
			}
		}
	}
	if len(got.Fits) != len(want.Fits) {
		t.Fatal("fit cadence")
	}
	for n, f := range got.Fits {
		if f.Clock != want.Fits[n].Clock {
			t.Fatal("fit clock")
		}
	}
}

func TestPublicationFactorialContracts(t *testing.T) {
	cfg := forestDelayCases()[0]
	base, err := forestDependenceBase(cfg, 2026092191)
	if err != nil {
		t.Fatal(err)
	}
	for _, delay := range []bool{false, true} {
		want, err := publicationRun(base, cfg, 0, 0, 2026092192, delay)
		if err != nil {
			t.Fatal(err)
		}
		got, err := factorialRun(base, cfg, 0, 0, 2026092192, delay)
		if err != nil {
			t.Fatal(err)
		}
		checkFactorialEndpoints(t, got, want)
		for _, fit := range got.Fits {
			if fit.Clock+1 < 512 {
				for _, a := range []int{1, 3} {
					if got.Frames[fit.Clock+1].Predictions[a].Weights != [4]float64{.7, .1, .1, .1} {
						t.Fatal("reset did not reach forecast")
					}
				}
			}
		}
	}
}

func TestPublicationFactorialExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_PUBLICATION_FACTORIAL")
	if path == "" {
		t.Skip("opt-in consumed factorial")
	}
	parent, err := os.Open("../../docs/experiments/mmm-publication-v1.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	h := sha256.New()
	if _, err := io.Copy(h, parent); err != nil {
		t.Fatal(err)
	}
	parentHash := hex.EncodeToString(h.Sum(nil))
	if parentHash != "566710fe18c30d2fa905b4c5a9203874bf25d965b73da846c15cff2efcddc596" {
		t.Fatal("parent changed")
	}
	if _, err := parent.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	paths, err := filepath.Glob("../../internal/*/*.go")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, "../../docs/experiments/mmm-publication-factorial-v1-contract.md", "../../research/publication-factorial-summary.mjs", "../../go.mod", "../../go.sum")
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
	enc := json.NewEncoder(out)
	if err := enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "ParentHash": parentHash, "Consumed": true}); err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(parent)
	var header json.RawMessage
	if err := dec.Decode(&header); err != nil {
		t.Fatal(err)
	}
	n := 0
	for {
		var old publicationRecord
		if err := dec.Decode(&old); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		ci := -1
		for i, cfg := range forestDelayCases() {
			if cfg.Name == old.Case {
				ci = i
			}
		}
		if ci < 0 {
			t.Fatal("unknown parent case")
		}
		cfg := forestDelayCases()[ci]
		base, err := forestDependenceBase(cfg, old.TrainSeed)
		if err != nil {
			t.Fatal(err)
		}
		i, err := factorialRun(base, cfg, ci, old.Index, old.Seed, false)
		if err != nil {
			t.Fatal(err)
		}
		d, err := factorialRun(base, cfg, ci, old.Index, old.Seed, true)
		if err != nil {
			t.Fatal(err)
		}
		checkFactorialEndpoints(t, i, old.Immediate)
		checkFactorialEndpoints(t, d, old.Delayed)
		if err := enc.Encode(factorialRecord{old.Phase, old.Case, old.Index, old.Seed, old.TrainSeed, i, d}); err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 192 {
		t.Fatal("incomplete parent", n)
	}
}
