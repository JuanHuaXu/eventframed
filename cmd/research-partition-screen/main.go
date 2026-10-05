package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	ri "github.com/JuanHuaXu/eventframed/internal/researchindex"
)

type observation struct {
	ID             string
	NS             int64
	Error          string
	Local          [][]ri.Candidate
	Merged, Oracle []ri.Candidate
}
type result struct {
	Partitions, Repeat int
	Sizes              []int
	BuildNS            []int64
	ReplacementNS      int64
	Queries            []observation
}
type layout struct {
	result  result
	indexes []*ri.HNSWBase
	records [][]ri.Mutation
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func vector(id string) []float32 {
	v := make([]float32, 768)
	for b := 0; b < 24; b++ {
		h := sha256.Sum256([]byte(fmt.Sprintf("%s/block-%d", id, b)))
		for i, x := range h {
			v[b*32+i] = float32(x) - 127.5
		}
	}
	return v
}
func cosine(a, b []float32) float64 {
	var dot, aa, bb float64
	for i, x := range a {
		dot += float64(x) * float64(b[i])
		aa += float64(x) * float64(x)
		bb += float64(b[i]) * float64(b[i])
	}
	return dot / math.Sqrt(aa*bb)
}
func exact(q []float32, records []ri.Mutation) []ri.Candidate {
	c := make([]ri.Candidate, len(records))
	for i, r := range records {
		c[i] = ri.Candidate{ID: r.ID, Score: cosine(q, r.Vector)}
	}
	sort.Slice(c, func(i, j int) bool {
		if c[i].Score == c[j].Score {
			return c[i].ID < c[j].ID
		}
		return c[i].Score > c[j].Score
	})
	return c[:10]
}
func build(ctx context.Context, records []ri.Mutation, path string) (*ri.HNSWBase, int64) {
	d, e := ri.RestoreDurable(768, 64, 1, records, func(context.Context, uint64, []ri.Mutation) error { return errors.New("static fixture") })
	check(e)
	v, e := d.View(ctx)
	check(e)
	start := time.Now()
	h, e := ri.BuildHNSWBase(ctx, v, 768, path)
	check(e)
	return h, time.Since(start).Nanoseconds()
}
func main() {
	if len(os.Args) != 2 {
		panic("provide NEW output JSON")
	}
	runtime.GOMAXPROCS(4)
	f, e := os.OpenFile(os.Args[1], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	check(e)
	defer f.Close()
	root := os.Args[1] + ".stores"
	check(os.Mkdir(root, 0700))
	hashes := map[string]string{}
	for _, p := range []string{"cmd/research-partition-screen/main.go", "internal/researchindex/partition.go", "internal/researchindex/nomination_diagnostic.go", "internal/researchindex/durable.go", "internal/researchindex/generation.go", "research/public-task-pilot/bulk-base-overlay-v1/hnsw_base.go.txt", "research/public-task-pilot/PARTITION_PROTOCOL.md"} {
		b, e := os.ReadFile(p)
		check(e)
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	seed := make([]ri.Mutation, 6400)
	for i := range seed {
		id := fmt.Sprintf("seed-%d", i)
		seed[i] = ri.Mutation{ID: id, Vector: vector(id)}
	}
	oracles := make([][]ri.Candidate, 128)
	for i := range oracles {
		oracles[i] = exact(vector(fmt.Sprintf("probe-%d", i)), seed)
	}
	ctx := context.Background()
	var results []result
	for repeat := 0; repeat < 2; repeat++ {
		var layouts []*layout
		for _, count := range []int{1, 8} {
			l := &layout{result: result{Partitions: count, Repeat: repeat}, records: make([][]ri.Mutation, count)}
			for _, r := range seed {
				p, e := ri.ResearchPartition(r.ID, count)
				check(e)
				l.records[p] = append(l.records[p], r)
			}
			for p, rs := range l.records {
				h, ns := build(ctx, rs, filepath.Join(root, fmt.Sprintf("r%d-p%d-base%d", repeat, count, p)))
				l.indexes = append(l.indexes, h)
				l.result.Sizes = append(l.result.Sizes, len(rs))
				l.result.BuildNS = append(l.result.BuildNS, ns)
			}
			layouts = append(layouts, l)
		}
		for i := 0; i < 1152; i++ {
			id := fmt.Sprintf("seed-%d", i)
			var oracle []ri.Candidate
			if i >= 1024 {
				id = fmt.Sprintf("probe-%d", i-1024)
				oracle = oracles[i-1024]
			}
			q := vector(id)
			for offset := 0; offset < 2; offset++ {
				l := layouts[(i+offset)%2]
				o := observation{ID: id, Oracle: oracle, Local: make([][]ri.Candidate, len(l.indexes))}
				start := time.Now()
				for p, h := range l.indexes {
					c, e := h.ResearchNomination(ctx, q, 10, 0)
					if e != nil {
						o.Error = e.Error()
						break
					}
					o.Local[p] = c
				}
				if o.Error == "" {
					o.Merged, e = ri.MergeResearchPartitions(o.Local, 10)
					if e != nil {
						o.Error = e.Error()
					}
				}
				o.NS = time.Since(start).Nanoseconds()
				l.result.Queries = append(l.result.Queries, o)
			}
		}
		for _, l := range layouts {
			rs := append([]ri.Mutation(nil), l.records[0]...)
			added := 0
			for i := 0; added < 32; i++ {
				id := fmt.Sprintf("write-%d", i)
				p, e := ri.ResearchPartition(id, l.result.Partitions)
				check(e)
				if p == 0 {
					rs = append(rs, ri.Mutation{ID: id, Vector: vector(id)})
					added++
				}
			}
			h, ns := build(ctx, rs, filepath.Join(root, fmt.Sprintf("r%d-p%d-replacement", repeat, l.result.Partitions)))
			l.result.ReplacementNS = ns
			check(h.Close())
			for _, h := range l.indexes {
				check(h.Close())
			}
			results = append(results, l.result)
		}
	}
	check(json.NewEncoder(f).Encode(struct {
		N, Dimension int
		Hashes       map[string]string
		Arms         []result
	}{6400, 768, hashes, results}))
	check(f.Sync())
}
