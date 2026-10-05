package observationrescueexperiment

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
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/observationrescue"
)

func TestRescueEvidenceRecomputesAndReplays(t *testing.T) {
	root := filepath.Join("..", "..")
	f, err := os.Open(filepath.Join(root, "docs/experiments/mmm-antipigeon-v2.json.gz"))
	if os.IsNotExist(err) {
		t.Skip("experiment not run yet")
	}
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	var out Output
	if err := json.NewDecoder(gz).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Records) != 2*len(Scenarios)*Streams {
		t.Fatal("incomplete evidence")
	}
	for path, expected := range out.Hashes {
		b, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != expected {
			t.Fatalf("source changed: %s", path)
		}
	}
	for _, r := range out.Records {
		if len(r.Outcomes) != Steps || len(r.Policies) != 6 {
			t.Fatal("missing predictions")
		}
		for _, p := range r.Policies {
			if len(p.Ticks) != Steps {
				t.Fatal("missing trace")
			}
			var sum float64
			for i, tick := range p.Ticks {
				y := 0.
				if r.Outcomes[i] {
					y = 1
				}
				if math.IsNaN(tick.P) || tick.P <= 0 || tick.P >= 1 || tick.Cost > 6 {
					t.Fatal("invalid forecast")
				}
				sum += (tick.P - y) * (tick.P - y)
				if tick.Transition.Refitted && !r.Audit[i] {
					t.Fatal("fit without audit")
				}
				if i+1 < Steps {
					expected := tick.Epoch
					if tick.Transition.Invalidated {
						expected++
					}
					if tick.Transition.Refitted {
						expected++
					}
					if p.Ticks[i+1].Epoch != expected {
						t.Fatal("epoch or timing mismatch")
					}
				}
			}
			if math.Abs(sum/Steps-p.Full.Brier) > 1e-14 {
				t.Fatal("Brier mismatch")
			}
			copy := p
			score(&copy, r.Outcomes, r.Scenario)
			if !reflect.DeepEqual(copy, p) {
				t.Fatal("per-stream metrics mismatch")
			}
		}
	}
	copy := Output{Records: out.Records}
	Summarize(&copy)
	if !reflect.DeepEqual(copy.Summaries, out.Summaries) || !reflect.DeepEqual(copy.Comparisons, out.Comparisons) || copy.OverallPassed != out.OverallPassed {
		t.Fatal("summary mismatch")
	}
	// Regenerate all observation/outcome streams and action histories; wall-clock
	// fitting time is deliberately excluded from deterministic equality.
	replayed, err := Run(nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := range out.Records {
		for j := range out.Records[i].Policies {
			out.Records[i].Policies[j].FitNanoseconds = 0
			replayed.Records[i].Policies[j].FitNanoseconds = 0
		}
	}
	if !reflect.DeepEqual(out.Records, replayed.Records) {
		t.Fatal("full replay mismatch")
	}
	// Trace, rather than infer from aggregate loss, why stable revocations arose.
	instant := 0
	for _, r := range out.Records {
		if r.Split != "confirmation" || r.Scenario != "stable" {
			continue
		}
		var posterior model.BayesianPosterior
		for i, y := range r.Outcomes {
			var cp bool
			posterior, cp = bayes.ApplyOutcomeAuthorized(posterior, (r.Policies[0].Ticks[i].P >= .5) == y, 1, observationrescue.ChangePolicy(), true)
			if i == r.Policies[5].FirstInvalidation {
				if !cp {
					t.Fatal("stable revocation was not a changepoint")
				}
				if posterior.ChangePointProbability >= observationrescue.ChangePolicy().Threshold {
					instant++
				}
				break
			}
		}
	}
	t.Logf("stable confirmation first invalidations with instantaneous changepoint probability crossing: %d/%d", instant, Streams)
}
