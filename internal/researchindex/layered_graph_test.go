package researchindex

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"slices"
	"sync"
	"testing"
)

func TestLayeredOwnership(t *testing.T) {
	ctx := context.Background()
	limits := LayeredLimits{8, 2, 4, 32}
	r := LayeredRecord{ID: "a", Level: 1, Vector: []float32{1, 2}, Links: [][]uint32{{1}, {2}}, Backlinks: [][]uint32{{3}, {4}}, Heuristic: []uint32{0, 1}}
	first, _, err := PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{{0, &r}}, 7, limits)
	if err != nil {
		t.Fatal(err)
	}
	r.Links[0][0] = 5
	r.Vector[0] = 9
	owned, _ := first.Lookup(0)
	if owned.Vector[0] != 1 || owned.Links[0][0] != 1 {
		t.Fatal("borrowed input")
	}
	owned.Links[0][0] = 8
	second, stats, err := PrepareLayered(ctx, first, []LayeredEdit{{0, &owned}}, 8, limits)
	if err != nil {
		t.Fatal(err)
	}
	a, b := layeredFind(first.tree, 0), layeredFind(second.tree, 0)
	if &a.Vector[0] != &b.Vector[0] || &a.Links[1][0] != &b.Links[1][0] || &a.Links[0][0] == &b.Links[0][0] {
		t.Fatal("sharing mismatch")
	}
	if stats.CopiedLinks != 1 || stats.SharedLinks != 3 {
		t.Fatal(stats)
	}
	owned.Links[0][0] = 99
	v, _ := second.Lookup(0)
	v.Vector[0] = 99
	got, _ := second.Lookup(0)
	if got.Vector[0] != 1 || got.Links[0][0] != 8 {
		t.Fatal("borrowed output")
	}
	if first.Global() != 7 || second.Global() != 8 {
		t.Fatal("global state")
	}
	cases := []struct {
		name   string
		edits  []LayeredEdit
		limits LayeredLimits
	}{
		{"duplicate", []LayeredEdit{{0, nil}, {0, nil}}, limits},
		{"edit cap", []LayeredEdit{{0, nil}, {1, nil}}, LayeredLimits{1, 2, 4, 32}},
		{"link cap", []LayeredEdit{{0, &r}}, LayeredLimits{8, 2, 4, 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, _, e := PrepareLayered(ctx, first, c.edits, 9, c.limits); e == nil {
				t.Fatal("accepted invalid edit")
			}
		})
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, e := PrepareLayered(canceled, first, nil, 9, limits); e == nil {
		t.Fatal("accepted cancellation")
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			v, ok := first.Lookup(0)
			if !ok || v.Vector[0] != 1 || v.Links[0][0] != 1 {
				t.Error("historical mutation")
				return
			}
		}
	}()
	for i := 0; i < 100; i++ {
		owned.Links[0][0] = uint32(i)
		if _, _, e := PrepareLayered(ctx, second, []LayeredEdit{{0, &owned}}, uint64(i), limits); e != nil {
			t.Fatal(e)
		}
	}
	wg.Wait()
	deleted, _, err := PrepareLayered(ctx, second, []LayeredEdit{{0, nil}}, 9, limits)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := deleted.Lookup(0); ok {
		t.Fatal("delete failed")
	}
	if _, ok := second.Lookup(0); !ok {
		t.Fatal("delete mutated history")
	}
}

type layerCaptureNode struct {
	ID               string
	Level            int
	VectorHash       string
	Links, Backlinks [][]uint32
	Heuristic        []uint32
}
type layerCapture struct {
	Global uint64
	Nodes  map[uint32]layerCaptureNode
}
type layerRow struct {
	N             int
	Kind          string
	Changed       int
	Before, After layerCapture
}

func layerVector(id string) []float32 {
	v := make([]float32, 768)
	var norm float64
	for b := 0; b < 24; b++ {
		h := sha256.Sum256([]byte(fmt.Sprintf("%s/block-%d", id, b)))
		for j, x := range h {
			v[b*32+j] = float32(x) - 127.5
			norm += float64(v[b*32+j]) * float64(v[b*32+j])
		}
	}
	for i := range v {
		v[i] /= float32(math.Sqrt(norm))
	}
	return v
}
func layerHash(v []float32) string {
	raw := make([]byte, 4*len(v))
	for i, x := range v {
		binary.LittleEndian.PutUint32(raw[4*i:], math.Float32bits(x))
	}
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}
func layerInput(t testing.TB, r layerCaptureNode) *LayeredRecord {
	t.Helper()
	v := layerVector(r.ID)
	if layerHash(v) != r.VectorHash {
		t.Fatal("vector regeneration mismatch", r.ID)
	}
	return &LayeredRecord{ID: r.ID, Level: r.Level, Vector: v, Links: r.Links, Backlinks: r.Backlinks, Heuristic: r.Heuristic}
}

// Measures preparation of known logical edits only, excluding their discovery,
// durable publication, graph traversal, and full-state verification.
func BenchmarkLayeredCapturedPrepare(b *testing.B) {
	path := os.Getenv("RESEARCH_LAYER_CAPTURE")
	if path == "" {
		b.Skip("explicit capture path required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}
	var rows []layerRow
	if err = json.Unmarshal(raw, &rows); err != nil {
		b.Fatal(err)
	}
	if len(rows) != 32 {
		b.Fatal("capture count")
	}
	r := rows[len(rows)-1]
	if r.N != 6400 || r.Kind != "delete" {
		b.Fatal("unexpected benchmark scenario")
	}
	var initial, edits []LayeredEdit
	ids := map[uint32]bool{}
	for id, node := range r.Before.Nodes {
		initial = append(initial, LayeredEdit{id, layerInput(b, node)})
		ids[id] = true
	}
	for id := range r.After.Nodes {
		ids[id] = true
	}
	for id := range ids {
		a, ao := r.Before.Nodes[id]
		z, zo := r.After.Nodes[id]
		if ao == zo && reflect.DeepEqual(a, z) {
			continue
		}
		var record *LayeredRecord
		if zo {
			record = layerInput(b, z)
		}
		edits = append(edits, LayeredEdit{id, record})
	}
	sortEdits := func(es []LayeredEdit) {
		slices.SortFunc(es, func(a, z LayeredEdit) int {
			if a.Ordinal < z.Ordinal {
				return -1
			}
			if a.Ordinal > z.Ordinal {
				return 1
			}
			return 0
		})
	}
	sortEdits(initial)
	sortEdits(edits)
	root, _, err := PrepareLayered(context.Background(), LayeredSnapshot{}, initial, r.Before.Global, LayeredLimits{len(initial), 768, 32, 1024})
	if err != nil {
		b.Fatal(err)
	}
	if len(edits) != r.Changed {
		b.Fatal("edit count")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		candidate, _, e := PrepareLayered(context.Background(), root, edits, r.After.Global, LayeredLimits{128, 768, 32, 1024})
		if e != nil || candidate.Global() != r.After.Global {
			b.Fatal("prepare failed", e)
		}
	}
}
func TestLayeredCapturedReplay(t *testing.T) {
	path := os.Getenv("RESEARCH_LAYER_CAPTURE")
	if path == "" {
		t.Skip("explicit capture path required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rows []layerRow
	if err = json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 32 {
		t.Fatal("capture count")
	}
	type report struct {
		N       int
		Kind    string
		Changed int
		LayeredStats
	}
	var reports []report
	for _, n := range []int{800, 6400} {
		var group []layerRow
		ids := map[uint32]bool{}
		for _, r := range rows {
			if r.N == n {
				group = append(group, r)
				for id := range r.Before.Nodes {
					ids[id] = true
				}
				for id := range r.After.Nodes {
					ids[id] = true
				}
			}
		}
		if len(group) != 16 {
			t.Fatal("group count")
		}
		check := func(s LayeredSnapshot, want layerCapture) {
			t.Helper()
			if s.Global() != want.Global {
				t.Fatal("global mismatch")
			}
			for id := range ids {
				v, ok := s.Lookup(id)
				w, exists := want.Nodes[id]
				if ok != exists {
					t.Fatal("presence mismatch", id)
				}
				if ok {
					got := layerCaptureNode{v.ID, v.Level, layerHash(v.Vector), v.Links, v.Backlinks, v.Heuristic}
					if !reflect.DeepEqual(got, w) {
						t.Fatal("record mismatch", id)
					}
				}
			}
		}
		var edits []LayeredEdit
		for id, r := range group[0].Before.Nodes {
			edits = append(edits, LayeredEdit{id, layerInput(t, r)})
		}
		slices.SortFunc(edits, func(a, b LayeredEdit) int {
			if a.Ordinal < b.Ordinal {
				return -1
			}
			if a.Ordinal > b.Ordinal {
				return 1
			}
			return 0
		})
		root, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, edits, group[0].Before.Global, LayeredLimits{n, 768, 32, 1024})
		if e != nil {
			t.Fatal(e)
		}
		history := []LayeredSnapshot{root}
		expected := []layerCapture{group[0].Before}
		for _, r := range group {
			check(root, r.Before)
			edits = nil
			for id := range ids {
				a, aok := r.Before.Nodes[id]
				b, bok := r.After.Nodes[id]
				if aok == bok && reflect.DeepEqual(a, b) {
					continue
				}
				var record *LayeredRecord
				if bok {
					record = layerInput(t, b)
				}
				edits = append(edits, LayeredEdit{id, record})
			}
			if len(edits) != r.Changed {
				t.Fatal("changed count")
			}
			candidate, stats, e := PrepareLayered(context.Background(), root, edits, r.After.Global, LayeredLimits{128, 768, 32, 1024})
			if e != nil {
				t.Fatal(e)
			}
			check(root, r.Before)
			check(candidate, r.After)
			root = candidate
			history = append(history, root)
			expected = append(expected, r.After)
			reports = append(reports, report{n, r.Kind, r.Changed, stats})
		}
		for i := range history {
			check(history[i], expected[i])
		}
	}
	result := struct {
		Passed      bool
		InputSHA256 string
		Reports     []report
	}{true, fmt.Sprintf("%x", sha256.Sum256(raw)), reports}
	encoded, e := json.MarshalIndent(result, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if output := os.Getenv("RESEARCH_LAYER_OUTPUT"); output != "" {
		f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		_, e = f.Write(encoded)
		ce := f.Close()
		if e != nil {
			t.Fatal(e)
		}
		if ce != nil {
			t.Fatal(ce)
		}
	}
	t.Log(string(encoded))
}
