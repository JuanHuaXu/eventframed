package researchindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"
)

func TestConstructionBackendCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_CONSTRUCTION_CAPTURE")
	if path == "" {
		t.Skip("explicit backend capture")
	}
	type candidate struct {
		ID       uint32
		Distance float32
	}
	var capture struct {
		State layerCapture
		Alpha float32
		Pairs map[string]float32
		Cases []struct {
			Query                string
			Start                uint32
			Level, EF, MaxM      int
			Candidates, Selected []candidate
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &capture); err != nil {
		t.Fatal(err)
	}
	var edits []LayeredEdit
	for id, r := range capture.State.Nodes {
		edits = append(edits, LayeredEdit{id, layerInput(t, r)})
	}
	ctx := context.Background()
	s, _, err := PrepareLayered(ctx, LayeredSnapshot{}, edits, capture.State.Global, LayeredLimits{len(edits), 768, 32, 1024})
	if err != nil {
		t.Fatal(err)
	}
	pair := func(a, b uint32) (float32, error) {
		if a > b {
			a, b = b, a
		}
		d, ok := capture.Pairs[fmt.Sprintf("%d/%d", a, b)]
		if !ok {
			return 0, fmt.Errorf("missing pair %d/%d", a, b)
		}
		return d, nil
	}
	report := struct{ Cases, CandidateSetMismatch, SelectedSetMismatch, SelectedOrderMismatch int }{}
	for _, c := range capture.Cases {
		got, _, err := SearchConstructionLayer(ctx, s, layerVector(c.Query), c.Start, c.Level, c.EF, 20000)
		if err != nil {
			t.Fatal(err)
		}
		selected, _, err := SelectPrivateNeighbors(ctx, got, c.MaxM, c.Level, 10000, capture.Alpha, pair)
		if err != nil {
			t.Fatal(err)
		}
		ids := func(a []NeighborCandidate) []uint32 {
			out := make([]uint32, len(a))
			for i, v := range a {
				out[i] = v.Ordinal
			}
			return out
		}
		want := func(a []candidate) []uint32 {
			out := make([]uint32, len(a))
			for i, v := range a {
				out[i] = v.ID
			}
			return out
		}
		setEqual := func(a, b []uint32) bool { slices.Sort(a); slices.Sort(b); return slices.Equal(a, b) }
		report.Cases++
		if !setEqual(ids(got), want(c.Candidates)) {
			report.CandidateSetMismatch++
			t.Logf("candidate mismatch level=%d ef=%d query=%s got=%v want=%v", c.Level, c.EF, c.Query, got, c.Candidates)
		}
		if !setEqual(ids(selected), want(c.Selected)) {
			report.SelectedSetMismatch++
		}
		if !slices.Equal(ids(selected), want(c.Selected)) {
			report.SelectedOrderMismatch++
		}
	}
	if report.Cases == 0 {
		t.Fatal("empty capture")
	}
	b, _ := json.Marshal(report)
	t.Log(string(b))
	if out := os.Getenv("RESEARCH_CONSTRUCTION_COMPARISON"); out != "" {
		f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		_, e = f.Write(b)
		ce := f.Close()
		if e != nil || ce != nil {
			t.Fatal(e, ce)
		}
	}
	// This diagnostic records order discrepancies instead of redefining them as
	// equivalence. Candidate-set disagreement is a separate traversal failure.
	if report.CandidateSetMismatch != 0 || report.SelectedSetMismatch != 0 {
		t.Fatal("construction set mismatch", report)
	}
}
