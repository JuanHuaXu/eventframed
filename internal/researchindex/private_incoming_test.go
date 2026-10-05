package researchindex

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPrivateIncomingBoundsAndOwnership(t *testing.T) {
	ctx := context.Background()
	lim := LayeredLimits{4, 2, 1, 16}
	record := func(id string, links, back []uint32) *LayeredRecord {
		return &LayeredRecord{ID: id, Vector: []float32{1, 2}, Links: [][]uint32{links}, Backlinks: [][]uint32{back}, Heuristic: []uint32{0}}
	}
	s, _, e := PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{{0, record("target", []uint32{1}, []uint32{1, 2})}, {1, record("a", []uint32{0, 2}, nil)}, {2, record("b", []uint32{0}, nil)}}, 0, lim)
	if e != nil {
		t.Fatal(e)
	}
	for _, budget := range [][2]int{{1, 2}, {3, 1}} {
		if out, e := DiscoverIncomingRemoval(ctx, s, 0, budget[0], budget[1]); e == nil || len(out.Edits) != 0 {
			t.Fatal("partial budget acceptance")
		}
	}
	p, e := DiscoverIncomingRemoval(ctx, s, 0, 3, 2)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Edits) != 2 || p.ReadRecords != 3 {
		t.Fatal(p)
	}
	if !reflect.DeepEqual(p.Affected[0], []uint32{1, 2}) {
		t.Fatal(p.Affected)
	}
	p.Edits[0].Record.Vector[0] = 99
	p.Edits[0].Record.Backlinks[0] = []uint32{99}
	old, _ := s.Lookup(1)
	if old.Vector[0] != 1 || old.Links[0][0] != 0 {
		t.Fatal("borrowed output mutated source")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, e := DiscoverIncomingRemoval(cancelled, s, 0, 3, 2); e == nil {
		t.Fatal("cancellation")
	}
}

func TestPrivateIncomingCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_INCOMING_CAPTURE")
	if path == "" {
		t.Skip("explicit incoming-stage capture required")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		N             int
		Kind, ID      string
		Before, After layerCapture
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	tested := 0
	for _, r := range rows {
		if r.Kind != "delete" {
			continue
		}
		var initial []LayeredEdit
		var target uint32
		found := false
		for id, n := range r.Before.Nodes {
			initial = append(initial, LayeredEdit{id, layerInput(t, n)})
			if n.ID == r.ID {
				target = id
				found = true
			}
		}
		if !found {
			t.Fatal("target missing")
		}
		s, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, initial, r.Before.Global, LayeredLimits{len(initial), 768, 32, 1024})
		if e != nil {
			t.Fatal(e)
		}
		p, e := DiscoverIncomingRemoval(context.Background(), s, target, 4096, 128)
		if e != nil {
			t.Fatal(e)
		}
		next, _, e := PrepareLayered(context.Background(), s, p.Edits, r.After.Global, LayeredLimits{128, 768, 32, 1024})
		if e != nil {
			t.Fatal(e)
		}
		if len(r.Before.Nodes) != len(r.After.Nodes) {
			t.Fatal("not an incoming-only fixture")
		}
		for id, w := range r.After.Nodes {
			got, ok := next.Lookup(id)
			if !ok {
				t.Fatal("lost record")
			}
			actual := layerCaptureNode{got.ID, got.Level, layerHash(got.Vector), got.Links, got.Backlinks, got.Heuristic}
			if !reflect.DeepEqual(actual, w) {
				t.Fatal("stage mismatch", r.N, r.ID, id)
			}
			old, _ := s.Lookup(id)
			want := r.Before.Nodes[id]
			if !reflect.DeepEqual(old.Links, want.Links) {
				t.Fatal("source changed")
			}
		}
		tested++
	}
	if tested != 16 {
		t.Fatal("case count", tested)
	}
}
