package researchindex

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunDeltaOverridesAllRunsAndPreservesViews(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	old, err := BuildImmutableRun(ctx, []Mutation{{ID: "a", Vector: []float32{1, 0}}, {ID: "b", Vector: []float32{.9, .1}}, {ID: "c", Vector: []float32{.8, .2}}}, 2, filepath.Join(root, "old"))
	if err != nil {
		t.Fatal(err)
	}
	defer old.Close()
	newer, err := BuildImmutableRun(ctx, []Mutation{{ID: "a", Vector: []float32{1, .1}}}, 2, filepath.Join(root, "new"))
	if err != nil {
		t.Fatal(err)
	}
	defer newer.Close()
	runs := []*ImmutableRun{newer, old}
	historical, err := NewRunDeltaSearch(ctx, runs, nil, 2, 64, 1)
	if err != nil {
		t.Fatal(err)
	}
	delta := []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{-1, 0}}}
	current, err := NewRunDeltaSearch(ctx, runs, delta, 2, 64, 1)
	if err != nil {
		t.Fatal(err)
	}
	delta[0].ID = "c"
	delta[1].Vector[0] = 1
	runs[0] = old
	for _, test := range []struct {
		s  *RunDeltaSearch
		id string
	}{{historical, "a"}, {current, "c"}} {
		got, err := test.s.Search(ctx, []float32{1, 0})
		if err != nil || len(got) != 1 || got[0].ID != test.id {
			t.Fatal(got, err)
		}
	}
	// A built flush followed by a later overwrite must equal the undrained
	// query state; draining cannot resurrect the earlier value of b.
	flushed, err := BuildImmutableRun(ctx, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{-1, 0}}}, 2, filepath.Join(root, "flush"))
	if err != nil {
		t.Fatal(err)
	}
	defer flushed.Close()
	after, err := NewRunDeltaSearch(ctx, []*ImmutableRun{flushed, newer, old}, []Mutation{{ID: "b", Vector: []float32{1, 0}}}, 2, 64, 3)
	if err != nil {
		t.Fatal(err)
	}
	before, err := NewRunDeltaSearch(ctx, []*ImmutableRun{newer, old}, []Mutation{{ID: "a", Delete: true}, {ID: "b", Vector: []float32{1, 0}}}, 2, 64, 3)
	if err != nil {
		t.Fatal(err)
	}
	x, e := before.Search(ctx, []float32{1, 0})
	y, f := after.Search(ctx, []float32{1, 0})
	if e != nil || f != nil || !reflect.DeepEqual(x, y) {
		t.Fatal(x, y, e, f)
	}
}

func TestRunDeltaBoundsAndEmptyState(t *testing.T) {
	ctx := context.Background()
	if _, err := NewRunDeltaSearch(ctx, nil, []Mutation{{ID: "a", Delete: true}, {ID: "b", Delete: true}}, 2, 1, 1); !errors.Is(err, ErrCapacity) {
		t.Fatal(err)
	}
	if _, err := NewRunDeltaSearch(ctx, make([]*ImmutableRun, 16), nil, 2, 64, 1); !errors.Is(err, ErrCapacity) {
		t.Fatal(err)
	}
	for _, delta := range [][]Mutation{{{ID: "a", Vector: []float32{0, 0}}}, {{ID: "a", Delete: true}, {ID: "a", Delete: true}}} {
		if _, err := NewRunDeltaSearch(ctx, nil, delta, 2, 64, 1); err == nil {
			t.Fatal("invalid delta accepted")
		}
	}
	s, err := NewRunDeltaSearch(ctx, nil, nil, 2, 64, 1)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Search(ctx, []float32{1, 0})
	if err != nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = NewRunDeltaSearch(cancelled, nil, nil, 2, 64, 1); err == nil {
		t.Fatal("cancelled construction")
	}
	if _, err = s.Search(cancelled, []float32{1, 0}); err == nil {
		t.Fatal("cancelled search")
	}
}
