package researchindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
)

func TestLayeredSearchBudget(t *testing.T) {
	r := func(id string, vector []float32, links []uint32) *LayeredRecord {
		return &LayeredRecord{ID: id, Vector: vector, Links: [][]uint32{links}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}
	}
	s, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, []LayeredEdit{{0, r("a", []float32{1, 0}, []uint32{1})}, {1, r("b", []float32{0, 1}, []uint32{0})}}, 1, LayeredLimits{2, 2, 1, 8})
	if e != nil {
		t.Fatal(e)
	}
	out, _, e := SearchLayered(context.Background(), s, []float32{0, 1}, 1, 2, 8)
	if e != nil || len(out) != 1 || out[0].ID != "b" {
		t.Fatal(out, e)
	}
	if out, _, e = SearchLayered(context.Background(), s, []float32{0, 1}, 1, 2, 1); e == nil || out != nil {
		t.Fatal("partial budget result")
	}
	if _, _, e = SearchLayered(context.Background(), s, []float32{1}, 1, 2, 8); e == nil {
		t.Fatal("dimension accepted")
	}
}
func TestLayeredSearchCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_SEARCH_CAPTURE")
	if path == "" {
		t.Skip("explicit capture")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var rows []layerRow
	if e = json.Unmarshal(raw, &rows); e != nil {
		t.Fatal(e)
	}
	type result struct {
		N                                                 int
		Phase                                             string
		Queries, Hits, Expected, SelfMiss, MaxEvaluations int
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
			t.Fatal("capture length")
		}
		for phase, snap := range []layerCapture{group[0].Before, group[len(group)-1].After} {
			var edits []LayeredEdit
			ids := []uint32{}
			for id, r := range snap.Nodes {
				edits = append(edits, LayeredEdit{id, layerInput(t, r)})
				ids = append(ids, id)
			}
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
			s, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, edits, snap.Global, LayeredLimits{len(edits), 768, 32, 1024})
			if e != nil {
				t.Fatal(e)
			}
			report := result{N: n, Phase: fmt.Sprint(phase)}
			for probe := 0; probe < 8; probe++ {
				id := ids[probe*len(ids)/8]
				q, _ := s.Lookup(id)
				got, calls, e := SearchLayered(context.Background(), s, q.Vector, 10, 100, 20000)
				if e != nil {
					t.Fatal(e)
				}
				report.MaxEvaluations = max(report.MaxEvaluations, calls)
				if len(got) == 0 || got[0].ID != q.ID {
					report.SelfMiss++
				}
				var exact []Candidate
				for _, ordinal := range ids {
					r := layeredFind(s.tree, ordinal)
					exact = append(exact, Candidate{r.ID, cosine(q.Vector, r.Vector)})
				}
				sort.Slice(exact, func(i, j int) bool {
					if exact[i].Score != exact[j].Score {
						return exact[i].Score > exact[j].Score
					}
					return exact[i].ID < exact[j].ID
				})
				want := map[string]bool{}
				for _, v := range exact[:10] {
					want[v.ID] = true
				}
				for _, v := range got {
					if want[v.ID] {
						report.Hits++
					}
				}
				report.Expected += 10
				report.Queries++
			}
			reports = append(reports, report)
		}
	}
	b, _ := json.MarshalIndent(reports, "", "  ")
	t.Log(string(b))
	if output := os.Getenv("RESEARCH_SEARCH_OUTPUT"); output != "" {
		if e = os.WriteFile(output, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}
