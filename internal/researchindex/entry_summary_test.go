package researchindex

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"
)

func checkEntryOracle(t *testing.T, s EntrySummary, m map[uint32]int) {
	t.Helper()
	var want entryChoice
	found := false
	for id, level := range m {
		if !found || level > want.level || level == want.level && id < want.id {
			want = entryChoice{id, level}
			found = true
		}
	}
	id, level, ok := s.Best()
	if ok != found || ok && (id != want.id || level != want.level) {
		t.Fatal("selection mismatch", id, level, ok, want, found)
	}
}
func TestEntrySummaryOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(65537))
	var s EntrySummary
	m := map[uint32]int{}
	var roots []EntrySummary
	var histories []map[uint32]int
	for i := 0; i < 3000; i++ {
		id := uint32(rng.Intn(256))
		if i%17 == 0 {
			id |= 1 << 31
		}
		level := rng.Intn(9) - 1
		var err error
		s, err = s.With(id, level)
		if err != nil {
			t.Fatal(err)
		}
		if level < 0 {
			delete(m, id)
		} else {
			m[id] = level
		}
		checkEntryOracle(t, s, m)
		if i%100 == 0 {
			copyM := map[uint32]int{}
			for k, v := range m {
				copyM[k] = v
			}
			roots = append(roots, s)
			histories = append(histories, copyM)
		}
	}
	for i := range roots {
		checkEntryOracle(t, roots[i], histories[i])
	}
	for id := range m {
		s, _ = s.With(id, -1)
	}
	if s.root != nil {
		t.Fatal("empty summary not pruned")
	}
	if _, err := s.With(1, 32); err == nil {
		t.Fatal("invalid level accepted")
	}
}
func TestEntrySummaryCapture(t *testing.T) {
	path := os.Getenv("RESEARCH_ENTRY_CAPTURE")
	if path == "" {
		t.Skip("explicit capture required")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []layerRow
	if err = json.Unmarshal(b, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 32 {
		t.Fatal("capture count")
	}
	for _, n := range []int{800, 6400} {
		var s EntrySummary
		initialized := false
		for _, r := range rows {
			if r.N != n {
				continue
			}
			if !initialized {
				for id, v := range r.Before.Nodes {
					s, err = s.With(id, v.Level)
					if err != nil {
						t.Fatal(err)
					}
				}
				initialized = true
			}
			before := s
			prior := map[uint32]int{}
			after := map[uint32]int{}
			for id, v := range r.Before.Nodes {
				prior[id] = v.Level
				if _, ok := r.After.Nodes[id]; !ok {
					s, _ = s.With(id, -1)
				}
			}
			for id, v := range r.After.Nodes {
				after[id] = v.Level
				if old, ok := r.Before.Nodes[id]; !ok || old.Level != v.Level {
					s, err = s.With(id, v.Level)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			checkEntryOracle(t, before, prior)
			checkEntryOracle(t, s, after)
			if r.Kind == "delete" && r.Before.Global != r.After.Global {
				id, _, ok := s.Best()
				if !ok {
					t.Fatal("replacement absent")
				}
				// Backend packs the ordinal high and level+1 low.
				if id != uint32(r.After.Global>>32) {
					t.Fatal("backend replacement mismatch", id, r.After.Global)
				}
			}
		}
	}
}
