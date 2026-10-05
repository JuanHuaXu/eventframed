package researchindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
)

func TestLayeredHeldout(t *testing.T) {
	path := os.Getenv("RESEARCH_HELDOUT_CAPTURE")
	if path == "" {
		t.Skip("explicit capture")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var rows []layerRow
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	type queryResult struct{ Query, Hits, Evaluations int }
	type result struct {
		N, Phase, EF, Hits, Expected int
		Pass                         bool
		Queries                      []queryResult
	}
	var reports []result
	for _, n := range []int{800, 6400} {
		var group []layerRow
		for _, r := range rows {
			if r.N == n {
				group = append(group, r)
			}
		}
		if len(group) != 16 {
			t.Fatal("capture count")
		}
		for phase, snap := range []layerCapture{group[0].Before, group[15].After} {
			var edits []LayeredEdit
			ids := map[string]bool{}
			for id, r := range snap.Nodes {
				edits = append(edits, LayeredEdit{id, layerInput(t, r)})
				ids[r.ID] = true
			}
			s, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, edits, snap.Global, LayeredLimits{len(edits), 768, 32, 1024})
			if e != nil {
				t.Fatal(e)
			}
			out := []result{{N: n, Phase: phase, EF: 100}, {N: n, Phase: phase, EF: 200}}
			for probe := 0; probe < 32; probe++ {
				name := fmt.Sprintf("heldout-traversal-v1-%d", probe)
				if ids[name] {
					t.Fatal("held-out ID present")
				}
				q := layerVector(name)
				var exact []Candidate
				for id := range snap.Nodes {
					r := layeredFind(s.tree, id)
					exact = append(exact, Candidate{r.ID, cosine(q, r.Vector)})
				}
				sort.Slice(exact, func(i, j int) bool {
					if exact[i].Score != exact[j].Score {
						return exact[i].Score > exact[j].Score
					}
					return exact[i].ID < exact[j].ID
				})
				want := map[string]bool{}
				for _, r := range exact[:10] {
					want[r.ID] = true
				}
				for i := range out {
					got, calls, e := SearchLayered(context.Background(), s, q, 10, out[i].EF, 20000)
					if e != nil {
						t.Fatal(e)
					}
					hits := 0
					for _, r := range got {
						if want[r.ID] {
							hits++
						}
					}
					out[i].Hits += hits
					out[i].Expected += 10
					out[i].Queries = append(out[i].Queries, queryResult{probe, hits, calls})
				}
			}
			for i := range out {
				out[i].Pass = out[i].Hits*100 >= 95*out[i].Expected
			}
			reports = append(reports, out...)
		}
	}
	encoded, e := json.MarshalIndent(reports, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if output := os.Getenv("RESEARCH_HELDOUT_OUTPUT"); output != "" {
		f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		_, e = f.Write(encoded)
		ce := f.Close()
		if e != nil {
			t.Fatal(e)
		}
		if ce != nil {
			t.Fatal(ce)
		}
	}
	for _, r := range reports {
		t.Logf("N%d phase%d ef%d %d/%d pass=%v", r.N, r.Phase, r.EF, r.Hits, r.Expected, r.Pass)
	}
}
