package researchindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"
)

func TestPrivateLinkBackend(t *testing.T) {
	path := os.Getenv("RESEARCH_LINK_CAPTURE")
	if path == "" {
		t.Skip("explicit backend capture")
	}
	var capture struct {
		MaxM, Capacity int
		Alpha          float32
		Pairs          map[string]float32
		Cases          []struct {
			Original, Links                []uint32
			Heuristic, AfterHeuristic, New uint32
			Accepted                       bool
		}
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &capture); err != nil {
		t.Fatal(err)
	}
	metric := func(a, b uint32) (float32, error) {
		if a > b {
			a, b = b, a
		}
		d, ok := capture.Pairs[fmt.Sprintf("%d/%d", a, b)]
		if !ok {
			return 0, fmt.Errorf("missing pair")
		}
		return d, nil
	}
	if len(capture.Cases) != 24 {
		t.Fatal("case count", len(capture.Cases))
	}
	for i, c := range capture.Cases {
		got, err := DecidePrivateLink(context.Background(), 0, c.New, c.Original, c.Heuristic, capture.MaxM, capture.Capacity, 0, 10000, capture.Alpha, metric)
		if err != nil || got.Accepted != c.Accepted || got.Heuristic != c.AfterHeuristic || !slices.Equal(got.Links, c.Links) {
			t.Fatalf("case %d got=%+v want=%+v err=%v", i, got, c, err)
		}
		var dropped []uint32
		for _, id := range c.Original {
			if !slices.Contains(c.Links, id) {
				dropped = append(dropped, id)
			}
		}
		if !slices.Equal(dropped, got.Dropped) {
			t.Fatal("dropped mismatch", i)
		}
	}
}
