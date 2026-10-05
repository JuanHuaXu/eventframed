package researchindex

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestPrivateInsertCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_INSERT_CAPTURE")
	if path == "" {
		t.Skip("explicit insertion capture")
	}
	var rows []layerRow
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(raw, &rows); e != nil {
		t.Fatal(e)
	}
	tested := 0
	for _, row := range rows {
		if row.Kind != "insert" {
			continue
		}
		var edits []LayeredEdit
		var summary EntrySummary
		for id, r := range row.Before.Nodes {
			edits = append(edits, LayeredEdit{id, layerInput(t, r)})
			summary, e = summary.With(id, r.Level)
			if e != nil {
				t.Fatal(e)
			}
		}
		ctx := context.Background()
		s, _, e := PrepareLayered(ctx, LayeredSnapshot{}, edits, row.Before.Global, LayeredLimits{len(edits), 768, 32, 1024})
		if e != nil {
			t.Fatal(e)
		}
		var newID uint32
		var newNode layerCaptureNode
		found := 0
		for id, r := range row.After.Nodes {
			if _, ok := row.Before.Nodes[id]; !ok {
				newID, newNode = id, r
				found++
			}
		}
		if found != 1 {
			t.Fatal("new node count", found)
		}
		got, e := PreparePrivateInsertion(ctx, s, summary, newID, newNode.ID, layerVector(newNode.ID), newNode.Level, len(edits), 16, 200, 100000, 100000, 128, 1)
		if e != nil {
			t.Fatal(row.N, newNode.ID, e)
		}
		mismatches := 0
		for id, want := range row.After.Nodes {
			r, ok := got.Snapshot.Lookup(id)
			if !ok {
				t.Fatal("missing node")
			}
			actual := layerCaptureNode{r.ID, r.Level, layerHash(r.Vector), r.Links, r.Backlinks, r.Heuristic}
			if !reflect.DeepEqual(actual, want) {
				if mismatches < 3 {
					t.Log("mismatch", row.N, newNode.ID, id, "links", r.Links, "want", want.Links)
				}
				mismatches++
			}
		}
		if mismatches != 0 || got.Snapshot.Global() != row.After.Global {
			t.Fatalf("insertion differs: %d records, global %d/%d", mismatches, got.Snapshot.Global(), row.After.Global)
		}
		for id, want := range row.Before.Nodes {
			r, _ := s.Lookup(id)
			actual := layerCaptureNode{r.ID, r.Level, layerHash(r.Vector), r.Links, r.Backlinks, r.Heuristic}
			if !reflect.DeepEqual(actual, want) {
				t.Fatal("source snapshot mutated")
			}
		}
		tested++
	}
	if tested != 16 {
		t.Fatal("case count", tested)
	}
}

func TestPrivateInsertInitialAndAbort(t *testing.T) {
	ctx := context.Background()
	first, err := PreparePrivateInsertion(ctx, LayeredSnapshot{}, EntrySummary{}, 0, "first", []float32{1, 0}, 0, 0, 4, 32, 100, 100, 16, 1)
	if err != nil || first.Snapshot.Global() != 1 {
		t.Fatal(first, err)
	}
	second, err := PreparePrivateInsertion(ctx, first.Snapshot, first.Summary, 1, "second", []float32{0, 1}, 2, 1, 4, 32, 100, 100, 16, 1)
	if err != nil || second.Snapshot.Global() != uint64(1)<<32|3 {
		t.Fatal(second, err)
	}
	a, _ := second.Snapshot.Lookup(0)
	b, _ := second.Snapshot.Lookup(1)
	if len(a.Links[0]) != 1 || a.Links[0][0] != 1 || len(b.Links[0]) != 1 || b.Links[0][0] != 0 || len(b.Links[1]) != 0 {
		t.Fatal("second-node wiring", a, b)
	}
	old, _ := first.Snapshot.Lookup(0)
	if len(old.Links[0]) != 0 {
		t.Fatal("old root mutated")
	}
	for _, limit := range []int{0, 1} {
		failed, err := PreparePrivateInsertion(ctx, first.Snapshot, first.Summary, 1, "second", []float32{0, 1}, 0, 1, 4, 32, 100, 100, limit, 1)
		if err == nil || failed.Snapshot.tree != nil {
			t.Fatal("partial result", failed, err)
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if failed, err := PreparePrivateInsertion(canceled, second.Snapshot, second.Summary, 2, "third", []float32{1, 1}, 0, 2, 4, 32, 100, 100, 16, 1); err == nil || failed.Snapshot.tree != nil {
		t.Fatal("canceled insertion published")
	}
	if failed, err := PreparePrivateInsertion(ctx, second.Snapshot, second.Summary, 1, "duplicate", []float32{1, 1}, 0, 2, 4, 32, 100, 100, 16, 1); err == nil || failed.Snapshot.tree != nil {
		t.Fatal("duplicate insertion")
	}
}
