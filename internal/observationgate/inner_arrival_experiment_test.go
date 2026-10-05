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

type innerArrivalRecord struct {
	Phase, Case        string
	Index              int
	Seed, TrainSeed    int64
	Immediate, Delayed innerArrivalResult
}

func TestInnerArrivalExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_INNER_ARRIVAL")
	if path == "" {
		t.Skip("opt-in consumed factorial")
	}
	parent, err := os.Open("../../docs/experiments/mmm-coupled-arrival-v1.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	h := sha256.New()
	if _, err := io.Copy(h, parent); err != nil {
		t.Fatal(err)
	}
	parentHash := hex.EncodeToString(h.Sum(nil))
	if parentHash != "60df42ab7121e4f0c4d2cd54cc0b86e82371a6ff711ca10d37a53305f5475f7b" {
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
	paths = append(paths, "../../docs/experiments/mmm-inner-arrival-v1-contract.md", "../../research/inner-arrival-summary.mjs", "../../go.mod", "../../go.sum")
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
		var old coupledArrivalRecord
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
		i, err := innerArrivalRun(base, cfg, ci, old.Index, old.Seed, false)
		if err != nil {
			t.Fatal(err)
		}
		d, err := innerArrivalRun(base, cfg, ci, old.Index, old.Seed, true)
		if err != nil {
			t.Fatal(err)
		}
		checkInnerArrivalEndpoints(t, i, old.Immediate)
		checkInnerArrivalEndpoints(t, d, old.Delayed)
		if err := enc.Encode(innerArrivalRecord{old.Phase, old.Case, old.Index, old.Seed, old.TrainSeed, i, d}); err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 192 {
		t.Fatal("incomplete parent", n)
	}
}
