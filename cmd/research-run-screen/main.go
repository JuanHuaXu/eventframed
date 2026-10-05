package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	ri "github.com/JuanHuaXu/eventframed/internal/researchindex"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

type Query struct {
	Seed               string
	Oracle, Runs, Full []ri.Candidate
	RunsNS, FullNS     int64
}
type Stage struct {
	Repeat, Step, Records, Runs                             int
	RunBuildNS, FullBuildNS, RunPlanNS, FullPlanNS, CloseNS int64
	Queries                                                 []Query
}
type Output struct {
	Hashes       map[string]string
	Stages       []Stage
	FinalCloseNS int64
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func vector(seed string) []float32 {
	v := make([]float32, 768)
	for block := 0; block < 24; block++ {
		h := sha256.Sum256([]byte(fmt.Sprintf("%s/block-%d", seed, block)))
		for j, x := range h {
			v[block*32+j] = float32(x) - 127.5
		}
	}
	return v
}
func score(a, b []float32) float64 {
	var d, aa, bb float64
	for i, x := range a {
		d += float64(x) * float64(b[i])
		aa += float64(x) * float64(x)
		bb += float64(b[i]) * float64(b[i])
	}
	return d / math.Sqrt(aa*bb)
}
func records(state map[string][]float32) []ri.Mutation {
	ids := make([]string, 0, len(state))
	for id := range state {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]ri.Mutation, 0, len(ids))
	for _, id := range ids {
		out = append(out, ri.Mutation{ID: id, Vector: state[id]})
	}
	return out
}
func oracle(state map[string][]float32, q []float32) []ri.Candidate {
	var out []ri.Candidate
	for id, v := range state {
		out = append(out, ri.Candidate{ID: id, Score: score(q, v)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].ID < out[j].ID
		}
		return out[i].Score > out[j].Score
	})
	return out[:10]
}
func main() {
	if len(os.Args) != 2 {
		panic("new output path required")
	}
	runtime.GOMAXPROCS(4)
	f, err := os.OpenFile(os.Args[1], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	check(err)
	root := os.Args[1] + ".stores"
	check(os.Mkdir(root, 0700))
	ctx := context.Background()
	out := Output{Hashes: map[string]string{}}
	for _, p := range []string{"cmd/research-run-screen/main.go", "internal/researchindex/immutable_run.go", "internal/researchindex/run_merge.go", "internal/researchindex/hnsw_base.go", "research/public-task-pilot/IMMUTABLE_RUN_SCREEN_PROTOCOL.md", "research/public-task-pilot/candidate-only-overlay-v1/hnsw_base.go.txt", "research/public-task-pilot/candidate-only-overlay-v1/collection.go.txt", "research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json", "research-candidate-only.mod", "research-candidate-only.sum"} {
		b, e := os.ReadFile(p)
		check(e)
		h := sha256.Sum256(b)
		out.Hashes[p] = hex.EncodeToString(h[:])
	}
	for repeat := 0; repeat < 2; repeat++ {
		state := map[string][]float32{}
		for i := 0; i < 800; i++ {
			id := fmt.Sprintf("seed-%d", i)
			state[id] = vector(id)
		}
		base, e := ri.BuildImmutableRun(ctx, records(state), 768, filepath.Join(root, fmt.Sprintf("r%d-base", repeat)))
		check(e)
		runs := []*ri.ImmutableRun{base}
		for step := 0; step < 8; step++ {
			var changes []ri.Mutation
			for i := 0; i < 32; i++ {
				id := fmt.Sprintf("new-%d-%d", step, i)
				m := ri.Mutation{ID: id, Vector: vector(id)}
				if i < 6 {
					m.ID = fmt.Sprintf("seed-%d", step*6+i)
					m.Vector = vector(fmt.Sprintf("update-%d-%d", step, i))
					m.Delete = i >= 4
				}
				changes = append(changes, m)
				if m.Delete {
					delete(state, m.ID)
				} else {
					state[m.ID] = m.Vector
				}
			}
			s := Stage{Repeat: repeat, Step: step, Records: len(state), Runs: len(runs) + 1}
			var added, full *ri.ImmutableRun
			buildRun := func() {
				start := time.Now()
				var e error
				added, e = ri.BuildImmutableRun(ctx, changes, 768, filepath.Join(root, fmt.Sprintf("r%d-s%d-small", repeat, step)))
				check(e)
				s.RunBuildNS = time.Since(start).Nanoseconds()
			}
			buildFull := func() {
				snapshot := records(state)
				start := time.Now()
				var e error
				full, e = ri.BuildImmutableRun(ctx, snapshot, 768, filepath.Join(root, fmt.Sprintf("r%d-s%d-full", repeat, step)))
				check(e)
				s.FullBuildNS = time.Since(start).Nanoseconds()
			}
			if (repeat+step)%2 == 0 {
				buildRun()
				buildFull()
			} else {
				buildFull()
				buildRun()
			}
			runs = append([]*ri.ImmutableRun{added}, runs...)
			start := time.Now()
			rs, e := ri.NewImmutableRunSearch(runs, 10)
			check(e)
			s.RunPlanNS = time.Since(start).Nanoseconds()
			start = time.Now()
			fs, e := ri.NewImmutableRunSearch([]*ri.ImmutableRun{full}, 10)
			check(e)
			s.FullPlanNS = time.Since(start).Nanoseconds()
			for probe := 0; probe < 32; probe++ {
				seed := fmt.Sprintf("probe-%d-%d", step, probe)
				q := vector(seed)
				row := Query{Seed: seed, Oracle: oracle(state, q)}
				runQ := func() {
					start := time.Now()
					var e error
					row.Runs, e = rs.Search(ctx, q)
					check(e)
					row.RunsNS = time.Since(start).Nanoseconds()
				}
				fullQ := func() {
					start := time.Now()
					var e error
					row.Full, e = fs.Search(ctx, q)
					check(e)
					row.FullNS = time.Since(start).Nanoseconds()
				}
				if (probe+step+repeat)%2 == 0 {
					runQ()
					fullQ()
				} else {
					fullQ()
					runQ()
				}
				s.Queries = append(s.Queries, row)
			}
			start = time.Now()
			check(full.Close())
			s.CloseNS = time.Since(start).Nanoseconds()
			out.Stages = append(out.Stages, s)
		}
		start := time.Now()
		for _, r := range runs {
			check(r.Close())
		}
		out.FinalCloseNS += time.Since(start).Nanoseconds()
	}
	check(json.NewEncoder(f).Encode(out))
	check(f.Close())
}
