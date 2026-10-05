package researchindex

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPrivateConnectCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_CONNECT_CAPTURE")
	if path == "" {
		t.Skip("explicit backend capture")
	}
	var cases []struct {
		Before, After         layerCapture
		Node                  uint32
		Neighbors             []uint32
		Level, MaxM, Capacity int
		Alpha                 float32
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(raw, &cases); e != nil {
		t.Fatal(e)
	}
	if len(cases) != 6 {
		t.Fatal("case count")
	}
	for i, c := range cases {
		var edits []LayeredEdit
		for id, r := range c.Before.Nodes {
			edits = append(edits, LayeredEdit{id, layerInput(t, r)})
		}
		ctx := context.Background()
		s, _, e := PrepareLayered(ctx, LayeredSnapshot{}, edits, c.Before.Global, LayeredLimits{len(edits), 768, 32, 1024})
		if e != nil {
			t.Fatal(e)
		}
		got, e := PreparePrivateConnections(ctx, s, c.Node, c.Neighbors, c.Level, c.MaxM, c.Capacity, 128, 10000, c.Alpha)
		if e != nil {
			t.Fatal(i, e)
		}
		for id, want := range c.After.Nodes {
			r, ok := got.Snapshot.Lookup(id)
			if !ok {
				t.Fatal("missing", id)
			}
			actual := layerCaptureNode{r.ID, r.Level, layerHash(r.Vector), r.Links, r.Backlinks, r.Heuristic}
			if !reflect.DeepEqual(actual, want) {
				t.Fatalf("case %d node %d got=%+v want=%+v", i, id, actual, want)
			}
			old, _ := s.Lookup(id)
			orig := layerCaptureNode{old.ID, old.Level, layerHash(old.Vector), old.Links, old.Backlinks, old.Heuristic}
			if !reflect.DeepEqual(orig, c.Before.Nodes[id]) {
				t.Fatal("source mutation")
			}
		}
		if got.Snapshot.Global() != c.After.Global {
			t.Fatal("global changed")
		}
		for _, budget := range []int{0, 1} {
			failed, err := PreparePrivateConnections(ctx, s, c.Node, c.Neighbors, c.Level, c.MaxM, c.Capacity, 1, budget, c.Alpha)
			if err == nil || failed.Snapshot.tree != nil {
				t.Fatal("partial failed edit")
			}
		}
	}
}
