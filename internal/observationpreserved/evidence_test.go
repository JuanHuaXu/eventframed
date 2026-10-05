package observationpreserved

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
)

func TestEvidenceRecomputesAndReplays(t *testing.T) {
	root := filepath.Join("..", "..")
	f, e := os.Open(filepath.Join(root, "docs/experiments/mmm-preserved-v3.json.gz"))
	if os.IsNotExist(e) {
		t.Skip("not run yet")
	}
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	gz, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer gz.Close()
	var out Output
	if e = json.NewDecoder(gz).Decode(&out); e != nil {
		t.Fatal(e)
	}
	if len(out.Records) != 320 {
		t.Fatal("incomplete evidence")
	}
	for path, want := range out.Hashes {
		b, e := os.ReadFile(filepath.Join(root, path))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatalf("source mismatch: %s", path)
		}
	}
	for _, r := range out.Records {
		if len(r.Ticks) != 512 || len(r.Arms) != 6 {
			t.Fatal("missing rows")
		}
		y := make([]bool, 512)
		audits, version := 0, 0
		for i, tick := range r.Ticks {
			y[i] = tick.Outcome
			for _, a := range r.Arms {
				p := a.Predictions[i]
				if p.Version != version || p.Cost > 6 || p.Values&^p.Mask != 0 || p.P <= 0 || p.P >= 1 || math.IsNaN(p.P) {
					t.Fatal("invalid forecast")
				}
				want := (bayes.ForecastMix{Weights: p.Weights}).Forecast(p.Experts)
				if a.Arm == "frozen_mmm" {
					want = p.Experts[0]
				}
				if math.Abs(want-p.P) > 1e-14 {
					t.Fatal("mixture score mismatch")
				}
				if i > 0 && p.Split && !a.Predictions[i-1].Split && !r.Ticks[i-1].Authorized {
					t.Fatal("uncertified sharing revocation")
				}
			}
			if tick.Audit {
				audits++
			}
			fit := tick.Audit && audits >= 32 && (audits-32)%16 == 0
			if fit != tick.Fit {
				t.Fatal("fit timing mismatch")
			}
			if fit {
				version++
			}
		}
		post := 256
		if r.Scenario == "recurring" {
			post = 128
		}
		for _, a := range r.Arms {
			if !reflect.DeepEqual(metrics(a.Predictions, y, 0), a.Full) || !reflect.DeepEqual(metrics(a.Predictions, y, post), a.Post) {
				t.Fatal("score mismatch")
			}
		}
	}
	copy := Output{Records: out.Records}
	Summarize(&copy)
	if !reflect.DeepEqual(copy.Summaries, out.Summaries) || !reflect.DeepEqual(copy.Comparisons, out.Comparisons) || copy.OverallPass != out.OverallPass {
		t.Fatal("summary mismatch")
	}
	replay, e := Run(nil)
	if e != nil {
		t.Fatal(e)
	}
	for i := range out.Records {
		out.Records[i].FitNS = 0
		replay.Records[i].FitNS = 0
	}
	if !reflect.DeepEqual(out.Records, replay.Records) {
		t.Fatal("deterministic replay mismatch")
	}
}
