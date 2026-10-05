package researchindex

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"
)

func TestPrivateLinkDecisions(t *testing.T) {
	ctx := context.Background()
	// Larger IDs are closer to target; pair distances otherwise avoid occlusion.
	metric := func(a, b uint32) (float32, error) {
		if a == 0 {
			return 1 - float32(b)/100, nil
		}
		return 2, nil
	}
	original := []uint32{1, 2, 3, 4, 5, 6}
	keep := slices.Clone(original)
	slack, err := DecidePrivateLink(ctx, 0, 7, original[:5], 0, 2, 6, 0, 100, 1, metric)
	if err != nil || !slack.Accepted || len(slack.Links) != 6 || slack.Heuristic != 0 || slack.PairCalls != 1 {
		t.Fatal(slack, err)
	}
	full, err := DecidePrivateLink(ctx, 0, 7, original, 0, 2, 6, 0, 100, 1, metric)
	if err != nil || !full.Accepted || !slices.Equal(full.Links, []uint32{7, 6}) || len(full.Dropped) != 5 || full.Heuristic != 2 {
		t.Fatal(full, err)
	}
	sorted := []uint32{6, 5, 4, 3, 2, 1}
	optimized, err := DecidePrivateLink(ctx, 0, 7, sorted, 6, 2, 6, 0, 100, 1, metric)
	if err != nil || !optimized.Accepted || !slices.Equal(optimized.Links, []uint32{7, 6}) {
		t.Fatal(optimized, err)
	}
	duplicate, err := DecidePrivateLink(ctx, 0, 3, original, 0, 2, 6, 0, 0, 1, metric)
	if err != nil || duplicate.Accepted || duplicate.PairCalls != 0 || !slices.Equal(duplicate.Links, original) {
		t.Fatal(duplicate, err)
	}
	for _, budget := range []int{0, 1, 4} {
		got, err := DecidePrivateLink(ctx, 0, 7, original, 0, 2, 6, 0, budget, 1, metric)
		if !errors.Is(err, ErrCapacity) || !reflect.DeepEqual(got, PrivateLinkDecision{}) {
			t.Fatal(got, err)
		}
	}
	if !slices.Equal(original, keep) {
		t.Fatal("mutated input")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if got, err := DecidePrivateLink(canceled, 0, 7, original, 0, 2, 6, 0, 100, 1, metric); err == nil || !reflect.DeepEqual(got, PrivateLinkDecision{}) {
		t.Fatal(got, err)
	}
}
