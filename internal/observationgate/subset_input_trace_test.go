package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
	"github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

type inputTraceStep struct {
	Step                         int
	Control, Coupled             observationpreserved.Prediction
	ControlSubset, CoupledSubset bool
	Stop                         string
	ObserverProbability          float64
}

func TestSubsetInputTrace(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SUBSET_INPUT_TRACE")
	if path == "" {
		t.Skip("opt-in consumed-data trace")
	}
	raw, err := os.ReadFile("../../docs/experiments/mmm-subset-input-integration-v1.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Open("../../docs/experiments/mmm-subset-input-integration-v1.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var header struct{ Sources, Hashes map[string]string }
	if err := dec.Decode(&header); err != nil {
		t.Fatal(err)
	}
	for name, want := range header.Hashes {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("source drift", name)
		}
	}
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	h := sha256.Sum256(raw)
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"subset_input_trace_test.go", "subset_input_trace_run_test.go"} {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		x := sha256.Sum256(b)
		sources[name] = string(b)
		hashes[name] = hex.EncodeToString(x[:])
	}
	if err := enc.Encode(map[string]any{"ArchiveHash": hex.EncodeToString(h[:]), "Sources": sources, "Hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	count := 0
	for i := 0; i < 320; i++ {
		var row subsetInputRecord
		if err := dec.Decode(&row); err != nil {
			t.Fatal(err)
		}
		if row.Case != "parity4_uniform" {
			continue
		}
		cfg := subsetBreadthCases[5]
		base, err := breadthBase(cfg, row.TrainSeed)
		if err != nil {
			t.Fatal(err)
		}
		c, a, b := newSubsetTrial(base, 1), newSubsetTrial(base, 1), newSubsetTrial(base, 1)
		var steps []inputTraceStep
		hook := func(step int, m observationpreserved.Models, reader observation.Reader) error {
			if step < 256 {
				return nil
			}
			cp, bp := *c.state.pending, *b.state.pending
			tr := inputTraceStep{Step: step, Control: cp.p, Coupled: bp.p, ControlSubset: cp.subsetGuide, CoupledSubset: bp.subsetGuide}
			var view observation.Result
			var err error
			if bp.subsetGuide {
				view, err = observationlearners.RunConditionalObserver(b.model, reader, reader.Epoch())
			} else {
				models := [3]*observation.Model{base, m.Short, m.Pooled}
				if b.state.split {
					models[2] = m.Local
				}
				view, err = observation.Run(models[bp.p.Guide], reader, reader.Epoch(), "mmm", 0)
			}
			if err != nil {
				return err
			}
			last := view.Trace[len(view.Trace)-1]
			if last.Observed != bp.p.Mask || last.Values != bp.p.Values || view.Cost != bp.p.Cost {
				return fmt.Errorf("path replay mismatch")
			}
			tr.Stop = view.Stop
			tr.ObserverProbability = view.Probability
			steps = append(steps, tr)
			return nil
		}
		actual, err := subsetInputTraceRun(base, row.Phase, 5, row.Index, row.StreamSeed, c, cfg, a, b, hook)
		if err != nil || actual != row.Original || c.metrics != row.Control || a.metrics != row.Fixed || b.metrics != row.Coupled {
			t.Fatal("metric replay", err)
		}
		for j, v := range []*subsetTrial{c, a, b} {
			if hex.EncodeToString(v.hash.Sum(nil)) != row.Tapes[j] {
				t.Fatal("tape replay")
			}
		}
		if len(steps) != 256 {
			t.Fatal("trace shape")
		}
		if err := enc.Encode(struct {
			Phase string
			Index int
			Steps []inputTraceStep
		}{row.Phase, row.Index, steps}); err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 32 {
		t.Fatal("missing parity rows")
	}
	if err := out.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Log("32 exact trajectory/tape replays;8192 pre-label path checks")
}
