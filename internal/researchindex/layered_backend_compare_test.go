package researchindex

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
)

func TestLayeredBackendComparison(t *testing.T) {
	path := os.Getenv("RESEARCH_BACKEND_QUERIES")
	if path == "" {
		t.Skip("explicit backend capture")
	}
	read := func(path string, v any) {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(b, v); e != nil {
			t.Fatal(e)
		}
	}
	var backend []struct {
		N, Phase, RequestedEF, EffectiveEF int
		StateHash                          string
		IDs                                [][]string
	}
	read(path, &backend)
	var rows []layerRow
	read(os.Getenv("RESEARCH_GRAPH_CAPTURE"), &rows)
	type report struct{ N, Phase, BackendHits, PrivateHits, Expected, MatchingIDs, MaxEvaluations int }
	var reports []report
	if len(backend) != 4 {
		t.Fatal("backend count")
	}
	for _, b := range backend {
		var group []layerRow
		for _, r := range rows {
			if r.N == b.N {
				group = append(group, r)
			}
		}
		if len(group) != 16 {
			t.Fatal("graph count")
		}
		snap := group[0].Before
		if b.Phase == 1 {
			snap = group[15].After
		}
		raw, _ := json.Marshal(snap)
		if fmt.Sprintf("%x", sha256.Sum256(raw)) != b.StateHash {
			t.Fatal("graph state mismatch")
		}
		if b.EffectiveEF != 400 || len(b.IDs) != 32 {
			t.Fatal("unexpected baseline breadth/queries")
		}
		var edits []LayeredEdit
		for id, r := range snap.Nodes {
			edits = append(edits, LayeredEdit{id, layerInput(t, r)})
		}
		s, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, edits, snap.Global, LayeredLimits{len(edits), 768, 32, 1024})
		if e != nil {
			t.Fatal(e)
		}
		out := report{N: b.N, Phase: b.Phase}
		for i, ids := range b.IDs {
			q := layerVector(fmt.Sprintf("heldout-traversal-v1-%d", i))
			got, calls, e := SearchLayered(context.Background(), s, q, 10, b.EffectiveEF, 20000)
			if e != nil {
				t.Fatal(e)
			}
			out.MaxEvaluations = max(out.MaxEvaluations, calls)
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
			actualBackend := map[string]bool{}
			for _, id := range ids {
				actualBackend[id] = true
				if want[id] {
					out.BackendHits++
				}
			}
			for _, r := range got {
				if want[r.ID] {
					out.PrivateHits++
				}
				if actualBackend[r.ID] {
					out.MatchingIDs++
				}
			}
			out.Expected += 10
		}
		reports = append(reports, out)
	}
	encoded, _ := json.MarshalIndent(reports, "", "  ")
	t.Log(string(encoded))
	if output := os.Getenv("RESEARCH_BACKEND_COMPARE_OUTPUT"); output != "" {
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
}
