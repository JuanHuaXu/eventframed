package researchindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestPrivateReconnectCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_RECONNECT_CAPTURE")
	if path == "" {
		t.Skip("explicit backend stage capture required")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Name                    string
		Error                   bool
		BeforeState, AfterState layerCapture
		Distances               map[string]float32
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 5 {
		t.Fatal("case count")
	}
	for _, r := range rows {
		t.Run(r.Name, func(t *testing.T) {
			var edits []LayeredEdit
			for id, n := range r.BeforeState.Nodes {
				edits = append(edits, LayeredEdit{id, layerInput(t, n)})
			}
			lim := LayeredLimits{128, 768, 32, 1024}
			s, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, edits, r.BeforeState.Global, lim)
			if e != nil {
				t.Fatal(e)
			}
			neighbors := []uint32{1, 2, 3, 4, 5, 6, 7, 8}
			if r.Name == "single" {
				neighbors = []uint32{1, 10000}
			}
			ctx := context.Background()
			if r.Name == "cancel" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			metric := func(a, b uint32) (float32, error) {
				if a > b {
					a, b = b, a
				}
				x, ok := r.Distances[fmt.Sprintf("%d/%d", a, b)]
				if !ok {
					return 0, fmt.Errorf("uncaptured pair")
				}
				return x, nil
			}
			p, e := DiscoverReconnect(ctx, s, neighbors, 0, 4, 32, 128, 1024, metric)
			if (e != nil) != r.Error {
				t.Fatal("error mismatch", e)
			}
			if e != nil {
				return
			}
			next, _, e := PrepareLayered(ctx, s, p.Edits, r.AfterState.Global, lim)
			if e != nil {
				t.Fatal(e)
			}
			for id, w := range r.AfterState.Nodes {
				v, ok := next.Lookup(id)
				if !ok {
					t.Fatal("lost node")
				}
				got := layerCaptureNode{v.ID, v.Level, layerHash(v.Vector), v.Links, v.Backlinks, v.Heuristic}
				if !reflect.DeepEqual(got, w) {
					t.Fatal("backend mismatch", id, got.Links, w.Links)
				}
				old, _ := s.Lookup(id)
				if !reflect.DeepEqual(old.Links, r.BeforeState.Nodes[id].Links) {
					t.Fatal("source changed")
				}
			}
		})
	}
}
