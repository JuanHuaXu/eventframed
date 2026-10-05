package researchindex

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPrivateDeleteCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_DELETE_CAPTURE")
	if path == "" {
		t.Skip("explicit full deletion capture")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		N             int
		ID, Kind      string
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
		var summary EntrySummary
		var target uint32
		found := false
		for id, n := range r.Before.Nodes {
			initial = append(initial, LayeredEdit{id, layerInput(t, n)})
			summary, e = summary.With(id, n.Level)
			if e != nil {
				t.Fatal(e)
			}
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
		p, e := PreparePrivateDeletion(context.Background(), s, summary, target, 16, 4096, 256, 128, 65536)
		if e != nil {
			t.Fatal(r.N, r.ID, e)
		}
		if p.Snapshot.Global() != r.After.Global {
			t.Fatal("global mismatch")
		}
		for id := range r.Before.Nodes {
			v, ok := p.Snapshot.Lookup(id)
			w, wok := r.After.Nodes[id]
			if ok != wok {
				t.Fatal("presence mismatch")
			}
			if ok {
				got := layerCaptureNode{v.ID, v.Level, layerHash(v.Vector), v.Links, v.Backlinks, v.Heuristic}
				if !reflect.DeepEqual(got, w) {
					t.Fatal("full deletion mismatch", r.N, r.ID, id)
				}
			}
			old, _ := s.Lookup(id)
			if !reflect.DeepEqual(old.Links, r.Before.Nodes[id].Links) {
				t.Fatal("source mutation")
			}
		}
		tested++
	}
	if tested != 16 {
		t.Fatal("case count", tested)
	}
}
func TestPrivateDeleteSingletonAndAbort(t *testing.T) {
	r := LayeredRecord{ID: "only", Vector: []float32{1, 2}, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}
	s, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, []LayeredEdit{{0, &r}}, 1, LayeredLimits{1, 2, 1, 8})
	if e != nil {
		t.Fatal(e)
	}
	summary, _ := (EntrySummary{}).With(0, 0)
	p, e := PreparePrivateDeletion(context.Background(), s, summary, 0, 2, 8, 8, 1, 0)
	if e != nil {
		t.Fatal(e)
	}
	if p.Snapshot.tree != nil || p.Snapshot.Global() != 0 {
		t.Fatal("singleton not empty")
	}
	if _, ok := s.Lookup(0); !ok {
		t.Fatal("history")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if p, e := PreparePrivateDeletion(ctx, s, summary, 0, 2, 8, 8, 1, 0); e == nil || p.Snapshot.tree != nil {
		t.Fatal("cancellation")
	}
}
