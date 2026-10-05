package researchindex

import (
	"context"
	"reflect"
	"testing"
)

func TestPrivateSelectionDiversityAndFill(t *testing.T) {
	in := []NeighborCandidate{{3, 3}, {2, 2}, {1, 1}}
	pair := func(a, b uint32) (float32, error) {
		if a == 2 {
			return .5, nil
		}
		return 4, nil
	}
	out, _, e := SelectPrivateNeighbors(context.Background(), in, 2, 0, 20, 1, pair)
	if e != nil || !reflect.DeepEqual(out, []NeighborCandidate{{1, 1}, {3, 3}}) {
		t.Fatal(out, e)
	}
	if in[0].Ordinal != 3 {
		t.Fatal("input sorted in place")
	}
	out[0].Ordinal = 99
	if in[2].Ordinal != 1 {
		t.Fatal("borrowed output")
	}
	out, _, e = SelectPrivateNeighbors(context.Background(), in, 2, 0, 20, 1, func(uint32, uint32) (float32, error) { return 0, nil })
	if e != nil || !reflect.DeepEqual(out, []NeighborCandidate{{1, 1}, {2, 2}}) {
		t.Fatal("fill order", out, e)
	}
	if out, _, e = SelectPrivateNeighbors(context.Background(), in, 2, 0, 0, 1, pair); e == nil || out != nil {
		t.Fatal("partial result on budget exhaustion")
	}
}
func TestPrivateSelectionSmallInputOrder(t *testing.T) {
	in := []NeighborCandidate{{9, 2}, {1, 1}}
	out, calls, e := SelectPrivateNeighbors(context.Background(), in, 2, 0, 0, 1, func(uint32, uint32) (float32, error) { t.Fatal("unneeded metric"); return 0, nil })
	if e != nil || calls != 0 || !reflect.DeepEqual(out, in) {
		t.Fatal(out, e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, e = SelectPrivateNeighbors(ctx, in, 2, 0, 0, 1, func(uint32, uint32) (float32, error) { return 0, nil }); e == nil {
		t.Fatal("cancellation")
	}
}
