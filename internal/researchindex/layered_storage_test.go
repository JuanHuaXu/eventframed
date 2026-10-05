package researchindex

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"unsafe"
)

// Diagnostic reachability accounting only. It walks whole retained roots and
// allocates visited sets, so it is not a hot-path admission algorithm. Capacities
// include backing storage but not Go allocator rounding, GC, or diagnostic maps.
type layeredStorage struct {
	TrieNodes, Records, VectorValues, AdjacencyValues, HeuristicValues, LayerHeaders uint64
	LogicalBytes                                                                     uint64
}

func countLayeredStorage(roots []LayeredSnapshot) layeredStorage {
	var out layeredStorage
	tries := map[*layeredTrie]bool{}
	records := map[*LayeredRecord]bool{}
	vectors := map[*float32]bool{}
	arrays := map[*uint32]bool{}
	headers := map[*[]uint32]bool{}
	countArray := func(a []uint32, heuristic bool) {
		if cap(a) == 0 {
			return
		}
		p := unsafe.SliceData(a)
		if arrays[p] {
			return
		}
		arrays[p] = true
		if heuristic {
			out.HeuristicValues += uint64(cap(a))
		} else {
			out.AdjacencyValues += uint64(cap(a))
		}
	}
	var visit func(*layeredTrie)
	visit = func(n *layeredTrie) {
		if n == nil || tries[n] {
			return
		}
		tries[n] = true
		out.TrieNodes++
		visit(n.child[0])
		visit(n.child[1])
		r := n.record
		if r == nil || records[r] {
			return
		}
		records[r] = true
		out.Records++
		if cap(r.Vector) > 0 {
			p := unsafe.SliceData(r.Vector)
			if !vectors[p] {
				vectors[p] = true
				out.VectorValues += uint64(cap(r.Vector))
			}
		}
		for _, layers := range [][][]uint32{r.Links, r.Backlinks} {
			if cap(layers) > 0 {
				p := unsafe.SliceData(layers)
				if !headers[p] {
					headers[p] = true
					out.LayerHeaders += uint64(cap(layers))
				}
			}
			for _, a := range layers {
				countArray(a, false)
			}
		}
		countArray(r.Heuristic, true)
	}
	for _, r := range roots {
		visit(r.tree)
	}
	out.LogicalBytes = out.TrieNodes*uint64(unsafe.Sizeof(layeredTrie{})) + out.Records*uint64(unsafe.Sizeof(LayeredRecord{})) + out.LayerHeaders*uint64(unsafe.Sizeof([]uint32{})) + 4*(out.VectorValues+out.AdjacencyValues+out.HeuristicValues)
	// ID string headers are in LayeredRecord; underlying borrowed immutable string
	// bytes are excluded because the graph does not own their source allocation.
	return out
}

func TestLayeredStorageAccounting(t *testing.T) {
	r := LayeredRecord{ID: "x", Level: 0, Vector: []float32{1, 2}, Links: [][]uint32{{1}}, Backlinks: [][]uint32{{2}}, Heuristic: []uint32{0}}
	ctx := context.Background()
	lim := LayeredLimits{2, 2, 1, 8}
	a, _, e := PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{{0, &r}}, 0, lim)
	if e != nil {
		t.Fatal(e)
	}
	one := countLayeredStorage([]LayeredSnapshot{a})
	dup := countLayeredStorage([]LayeredSnapshot{a, a})
	if one != dup || one.TrieNodes != 33 || one.Records != 1 || one.VectorValues < 2 {
		t.Fatal(one, dup)
	}
	r.Links[0][0] = 3
	b, _, e := PrepareLayered(ctx, a, []LayeredEdit{{0, &r}}, 1, lim)
	if e != nil {
		t.Fatal(e)
	}
	both := countLayeredStorage([]LayeredSnapshot{a, b})
	if both.Records != 2 || both.VectorValues != one.VectorValues || both.AdjacencyValues <= one.AdjacencyValues {
		t.Fatal("sharing count", one, both)
	}
	empty, _, e := PrepareLayered(ctx, b, []LayeredEdit{{0, nil}}, 2, lim)
	if e != nil {
		t.Fatal(e)
	}
	dead := countLayeredStorage([]LayeredSnapshot{empty})
	if dead.TrieNodes != 0 || dead.Records != 0 || dead.VectorValues != 0 {
		t.Fatal("deleted path accounting", dead)
	}
	if countLayeredStorage(nil).LogicalBytes != 0 {
		t.Fatal("empty count")
	}
}

func TestLayeredCapturedStorage(t *testing.T) {
	path := os.Getenv("RESEARCH_LAYER_CAPTURE")
	if path == "" {
		t.Skip("explicit capture path required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var rows []layerRow
	if e = json.Unmarshal(raw, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 32 {
		t.Fatal("capture count")
	}
	type report struct {
		N                                              int
		Latest, RecentEight, OldestEight, AllSeventeen layeredStorage
	}
	var reports []report
	for _, n := range []int{800, 6400} {
		var group []layerRow
		for _, r := range rows {
			if r.N == n {
				group = append(group, r)
			}
		}
		if len(group) != 16 {
			t.Fatal("group count")
		}
		var edits []LayeredEdit
		for id, r := range group[0].Before.Nodes {
			edits = append(edits, LayeredEdit{id, layerInput(t, r)})
		}
		root, _, e := PrepareLayered(context.Background(), LayeredSnapshot{}, edits, group[0].Before.Global, LayeredLimits{len(edits), 768, 32, 1024})
		if e != nil {
			t.Fatal(e)
		}
		history := []LayeredSnapshot{root}
		for _, r := range group {
			edits = nil
			ids := map[uint32]bool{}
			for id := range r.Before.Nodes {
				ids[id] = true
			}
			for id := range r.After.Nodes {
				ids[id] = true
			}
			for id := range ids {
				a, ao := r.Before.Nodes[id]
				b, bo := r.After.Nodes[id]
				if ao == bo && reflect.DeepEqual(a, b) {
					continue
				}
				var record *LayeredRecord
				if bo {
					record = layerInput(t, b)
				}
				edits = append(edits, LayeredEdit{id, record})
			}
			root, _, e = PrepareLayered(context.Background(), root, edits, r.After.Global, LayeredLimits{128, 768, 32, 1024})
			if e != nil {
				t.Fatal(e)
			}
			history = append(history, root)
		}
		latest := countLayeredStorage([]LayeredSnapshot{root})
		recent := countLayeredStorage(history[len(history)-9:]) // current + eight old readers
		oldRoots := append([]LayeredSnapshot{root}, history[:8]...)
		old := countLayeredStorage(oldRoots)
		all := countLayeredStorage(history)
		if latest.LogicalBytes > recent.LogicalBytes || latest.LogicalBytes > old.LogicalBytes || recent.LogicalBytes > all.LogicalBytes || old.LogicalBytes > all.LogicalBytes {
			t.Fatal("nonmonotone accounting")
		}
		if latest.Records != uint64(len(group[len(group)-1].After.Nodes)) {
			t.Fatal("latest record count")
		}
		reports = append(reports, report{n, latest, recent, old, all})
	}
	encoded, e := json.MarshalIndent(reports, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	t.Log(string(encoded))
	if output := os.Getenv("RESEARCH_LAYER_STORAGE_OUTPUT"); output != "" {
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
}
