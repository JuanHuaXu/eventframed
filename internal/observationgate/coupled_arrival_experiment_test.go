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

type coupledArrivalRecord struct {
	Phase, Case        string
	Index              int
	Seed, TrainSeed    int64
	Immediate, Delayed coupledArrivalResult
}

func checkCoupledParent(t testing.TB, got coupledArrivalResult, want publicationResult) {
	t.Helper()
	if got.Arms[0].Full != want.Arms[0].Full || got.Arms[0].Post != want.Arms[0].Post || got.Arms[0].SplitAt != want.Arms[0].SplitAt {
		t.Fatal("parent metric drift")
	}
	for i, f := range got.Frames {
		g := want.Frames[i]
		if f.Predictions[0] != g.Predictions[0] || f.X != g.X || f.RX != g.RX || f.Y != g.Y || f.RY != g.RY || f.Missing != g.Missing || f.Arrival != g.Arrival || f.Audit != g.Audit {
			t.Fatal("parent tape drift", i)
		}
	}
}
func TestCoupledArrivalExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_COUPLED_ARRIVAL")
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
	paths = append(paths, "../../docs/experiments/mmm-coupled-arrival-v1-contract.md", "../../research/coupled-arrival-summary.mjs", "../../go.mod", "../../go.sum")
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
		i, err := coupledArrivalRun(base, cfg, ci, old.Index, old.Seed, false)
		if err != nil {
			t.Fatal(err)
		}
		d, err := coupledArrivalRun(base, cfg, ci, old.Index, old.Seed, true)
		if err != nil {
			t.Fatal(err)
		}
		checkCoupledParent(t, i, old.Immediate)
		checkCoupledParent(t, d, old.Delayed)
		if err := enc.Encode(coupledArrivalRecord{old.Phase, old.Case, old.Index, old.Seed, old.TrainSeed, i, d}); err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 192 {
		t.Fatal("incomplete parent", n)
	}
}
