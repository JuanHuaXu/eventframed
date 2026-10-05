package researchindex

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestPrivateSelectionCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_SELECTION_CAPTURE")
	if path == "" {
		t.Skip("explicit capture")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	type candidate struct {
		ID       uint32
		Distance float32
	}
	var rows []struct {
		Level, MaxM     int
		Alpha           float32
		Input, Selected []candidate
		Pairs           map[string]float32
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 6 {
		t.Fatal("case count")
	}
	for _, r := range rows {
		var input, want []NeighborCandidate
		for _, c := range r.Input {
			input = append(input, NeighborCandidate{c.ID, c.Distance})
		}
		for _, c := range r.Selected {
			want = append(want, NeighborCandidate{c.ID, c.Distance})
		}
		metric := func(a, b uint32) (float32, error) {
			if a > b {
				a, b = b, a
			}
			d, ok := r.Pairs[fmt.Sprintf("%d/%d", a, b)]
			if !ok {
				return 0, fmt.Errorf("missing pair")
			}
			return d, nil
		}
		got, _, e := SelectPrivateNeighbors(context.Background(), input, r.MaxM, r.Level, 4096, r.Alpha, metric)
		if e != nil || !reflect.DeepEqual(got, want) {
			t.Fatal("selection mismatch", r.Level, len(r.Input), got, want, e)
		}
	}
}
