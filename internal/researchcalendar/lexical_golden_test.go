package researchcalendar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

func TestLexicalGolden(t *testing.T) {
	root := filepath.Join("..", "..")
	var golden struct {
		Hashes map[string]string
		Rows   []struct {
			Set, Case                   string
			LexicalTrace, CombinedTrace LexicalResult
		}
	}
	read := func(p string, v any) {
		t.Helper()
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, v); e != nil {
			t.Fatal(e)
		}
	}
	read("research/public-task-pilot/what-lexical-v2-results.json", &golden)
	for p, h := range golden.Hashes {
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		sum := sha256.Sum256(b)
		if hex.EncodeToString(sum[:]) != h {
			t.Fatal("stale golden source", p)
		}
	}
	checked := 0
	maxError := 0.0
	for _, g := range golden.Rows {
		file := "task-results.json"
		if g.Set == "landing-transfer-v1" {
			file = "results.json"
		}
		var data struct {
			Results []struct {
				Arm, Case, Effective, Explanation string
				After                             []retrieval.Candidate
			}
		}
		read("research/public-task-pilot/"+g.Set+"/"+file, &data)
		found := false
		for _, r := range data.Results {
			if r.Arm != "priority" || r.Case != g.Case {
				continue
			}
			found = true
			var explanation struct{ Plan TaskPlan }
			if err := json.Unmarshal([]byte(r.Explanation), &explanation); err != nil {
				t.Fatal(err)
			}
			for _, combined := range []bool{false, true} {
				var plan *PriorityPlan
				want := g.LexicalTrace
				if combined {
					plan = &explanation.Plan.PriorityPlan
					want = g.CombinedTrace
				}
				before, _ := json.Marshal(r.After)
				got, err := LexicalOrder(context.Background(), r.Effective, r.After, plan)
				if err != nil {
					t.Fatal(err)
				}
				after, _ := json.Marshal(r.After)
				if string(before) != string(after) {
					t.Fatal("mutated inputs")
				}
				if !reflect.DeepEqual(got.Order, want.Order) || got.Method != want.Method {
					t.Fatal(g.Set, g.Case, combined, got.Order, want.Order)
				}
				for i, v := range got.Scores {
					d := math.Abs(v - want.Scores[i])
					maxError = math.Max(maxError, d)
					if d > 1e-12 {
						t.Fatal("score mismatch", g.Set, g.Case, d)
					}
				}
				checked++
			}
		}
		if !found {
			t.Fatal("missing trace", g.Set, g.Case)
		}
	}
	if checked != 108 {
		t.Fatal("incomplete trace coverage", checked)
	}
	t.Logf("%d orders match; maximum absolute score discrepancy %.17g", checked, maxError)
}
