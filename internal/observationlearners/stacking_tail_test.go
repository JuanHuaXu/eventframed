package observationlearners

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// This is explicitly a consumed v92 witness, never a fresh confirmation case.
func TestStackingConsumedTail(t *testing.T) {
	parent, out := os.Getenv("EVENTFRAME_STACK_TAIL_PARENT"), os.Getenv("EVENTFRAME_STACK_TAIL_OUT")
	if parent == "" && out == "" {
		t.Skip("explicit consumed diagnostic paths required")
	}
	if parent == "" || out == "" {
		t.Fatal("both paths required")
	}
	raw, err := os.ReadFile(parent)
	if err != nil {
		t.Fatal(err)
	}
	if priorV92SHA(raw) != "fa583cd78260deb9cd86a42f85404b8564d14fb84d822d7d9b4bc6fb3d0997ac" {
		t.Fatal("unexpected parent artifact")
	}
	var a priorV92Artifact
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatal(err)
	}
	for path, h := range a.Hashes {
		b, e := os.ReadFile(filepath.Join("..", "..", path))
		if e != nil || priorV92SHA(b) != h {
			t.Fatal("parent source mismatch", path, e)
		}
	}
	var original *priorV92Record
	for i := range a.Records {
		r := &a.Records[i]
		if r.Phase == "design" && r.Case == "majority3" && r.Index == 37 && r.N == 16 {
			original = r
			break
		}
	}
	if original == nil {
		t.Fatal("missing witness")
	}
	replay, err := priorV92Score(0, 5, 37, 16)
	if err != nil || replay != *original {
		t.Fatal("original witness mismatch", err)
	}
	rule, samples := priorV92Data(0, 5, 37)
	samples = samples[:16]
	f, err := fitStackingEvidence(samples)
	if err != nil {
		t.Fatal(err)
	}
	result := struct {
		Classification, ParentSHA256 string
		Hashes                       map[string]string
		Original                     priorV92Record
		Lambdas, Weights             [2]float64
		Metrics                      [2][2]priorV92Metric
	}{Classification: "post-hoc consumed v92 tail; not confirmation", ParentSHA256: priorV92SHA(raw), Hashes: a.Hashes, Original: *original, Lambdas: [2]float64{0, 1}}
	for _, path := range []string{"internal/observationlearners/stacking.go", "internal/observationlearners/stacking_test.go", "internal/observationlearners/stacking_tail_test.go", "docs/experiments/mmm-stacking-v93-protocol.md"} {
		b, e := os.ReadFile(filepath.Join("..", "..", path))
		if e != nil {
			t.Fatal(e)
		}
		result.Hashes[path] = priorV92SHA(b)
	}
	for arm, lambda := range result.Lambdas {
		m, w, err := compileStacking(f, lambda)
		if err != nil {
			t.Fatal(err)
		}
		result.Weights[arm] = w
		for view, mask := range []uint16{511, 63} {
			for x := uint16(0); x < 512; x++ {
				truth := priorV92Truth(x, rule, "majority3")
				p, e := m.Forecast(mask, x&mask)
				if e != nil || math.IsNaN(p) || p < 0 || p > 1 {
					t.Fatal("invalid forecast", p, e)
				}
				r := &result.Metrics[arm][view]
				r.Brier += ((p-truth)*(p-truth) + truth*(1-truth)) / 512
				accuracy := 1 - truth
				if p >= .5 {
					accuracy = truth
				}
				r.Accuracy += accuracy / 512
			}
		}
	}
	file, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = json.NewEncoder(file).Encode(result)
	ce := file.Close()
	if err != nil || ce != nil {
		t.Fatal(err, ce)
	}
	t.Log("weights", result.Weights, "scores", result.Metrics)
}
