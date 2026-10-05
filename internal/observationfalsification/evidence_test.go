package observationfalsification

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
)

func TestEvidenceRecomputesAndReplays(t *testing.T) {
	root := filepath.Join("..", "..")
	f, e := os.Open(filepath.Join(root, "docs/experiments/mmm-falsification-v4.json.gz"))
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
		t.Fatal("incomplete experiment")
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
		if len(r.Ticks) != 512 {
			t.Fatal("missing live evaluations")
		}
		version, audits := 0, 0
		for _, tick := range r.Ticks {
			for _, p := range tick.Predictions {
				if p.Version != version || p.Cost > 6 || p.Values&^p.Mask != 0 || math.IsNaN(p.P) || p.P <= 0 || p.P >= 1 {
					t.Fatal("invalid prediction")
				}
			}
			if tick.Audit {
				audits++
				for i, p := range tick.Probes {
					if p.Choice < 0 || p.Choice >= 4 || p.Probability < .0625 || p.Bits != tick.Pool[p.Choice] {
						t.Fatal("invalid selection")
					}
					for k := range p.Forecasts {
						for j, v := range p.Forecasts[k] {
							if v <= 0 || v >= 1 || math.IsNaN(v) || p.InspectionCosts[k][j] > 6 {
								t.Fatal("invalid observer")
							}
						}
					}
					for j, q := range tick.Probes {
						if j < i && q.Choice == p.Choice && (q.Outcome != p.Outcome || q.ReferenceOutcome != p.ReferenceOutcome) {
							t.Fatal("shared candidate outcome mismatch")
						}
					}
				}
				if audits >= 32 && (audits-32)%16 == 0 {
					version++
				}
			}
		}
		if r.ProbeCost != audits*36 || r.OutcomeQueries != audits*2 {
			t.Fatal("budget mismatch")
		}
		post := 256
		if r.Scenario == "recurring" {
			post = 128
		}
		for i := 0; i < 3; i++ {
			if !reflect.DeepEqual(score(r.Ticks, i, 0), r.Full[i]) || !reflect.DeepEqual(score(r.Ticks, i, post), r.Post[i]) || !reflect.DeepEqual(score(r.Ticks, i, 384), r.Late[i]) {
				t.Fatal("metric mismatch")
			}
		}
	}
	copy := Output{Records: out.Records}
	Summarize(&copy)
	if !reflect.DeepEqual(copy.Comparisons, out.Comparisons) || copy.OverallPass != out.OverallPass {
		t.Fatal("comparison mismatch")
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
		t.Fatal("full replay mismatch")
	}
}
