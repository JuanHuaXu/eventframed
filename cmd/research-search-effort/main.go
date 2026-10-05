package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchindex"
)

type query struct {
	Ef              int
	DefaultAgrees   bool
	ID              string
	NS              int64
	OwnedExact, Hit bool
	Error           string
	Candidates      []researchindex.Candidate
}
type arm struct {
	Repeat  int
	BuildNS int64
	Queries []query
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
func vector(id string) []float32 {
	v := make([]float32, 768)
	for block := 0; block < 24; block++ {
		h := sha256.Sum256([]byte(fmt.Sprintf("%s/block-%d", id, block)))
		for i, b := range h {
			v[block*32+i] = float32(b) - 127.5
		}
	}
	return v
}

func main() {
	if len(os.Args) != 2 {
		panic("usage: research-search-effort NEW-output.json")
	}
	runtime.GOMAXPROCS(4)
	output := os.Args[1]
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	check(err)
	defer f.Close()
	root := output + ".stores"
	check(os.Mkdir(root, 0700))
	hashes := map[string]string{}
	for _, p := range []string{"cmd/research-search-effort/main.go", "internal/researchindex/nomination_diagnostic.go", "research/public-task-pilot/SEARCH_EFFORT_PROTOCOL.md", "research/public-task-pilot/bulk-base-overlay-v1/hnsw_base.go.txt", "internal/researchindex/merge.go", "internal/researchindex/generation.go", "internal/researchindex/durable.go"} {
		b, e := os.ReadFile(p)
		check(e)
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	var arms []arm
	for repeat := 0; repeat < 2; repeat++ {
		ctx := context.Background()
		seed := make([]researchindex.Mutation, 6400)
		for i := range seed {
			id := fmt.Sprintf("seed-%d", i)
			seed[i] = researchindex.Mutation{ID: id, Vector: vector(id)}
		}
		d, e := researchindex.RestoreDurable(768, 64, 1, seed, func(context.Context, uint64, []researchindex.Mutation) error {
			return errors.New("static fixture forbids persistence")
		})
		check(e)
		v, e := d.View(ctx)
		check(e)
		start := time.Now()
		base, e := researchindex.BuildHNSWBase(ctx, v, 768, filepath.Join(root, fmt.Sprintf("r%d", repeat)))
		check(e)
		a := arm{Repeat: repeat, BuildNS: time.Since(start).Nanoseconds()}
		for i := 0; i < 1024; i++ {
			id := fmt.Sprintf("seed-%d", i)
			input := vector(id)
			owned, ok := v.Lookup(id)
			control, controlErr := base.Search(ctx, v, input, 10)
			check(controlErr)
			for offset := 0; offset < 4; offset++ {
				ef := []int{0, 100, 200, 400}[(i+offset)%4]
				q := query{Ef: ef, ID: id, OwnedExact: ok && reflect.DeepEqual(owned, input)}
				start = time.Now()
				q.Candidates, e = base.ResearchNomination(ctx, input, 10, ef)
				q.NS = time.Since(start).Nanoseconds()
				if e != nil {
					q.Error = e.Error()
				}
				for _, c := range q.Candidates {
					if c.ID == id {
						q.Hit = true
					}
				}
				if ef == 0 {
					q.DefaultAgrees = reflect.DeepEqual(control, q.Candidates)
				}
				a.Queries = append(a.Queries, q)
			}
		}
		check(base.Close())
		arms = append(arms, a)
	}
	check(json.NewEncoder(f).Encode(struct {
		N, Dimension int
		Go           string
		Hashes       map[string]string
		Arms         []arm
	}{6400, 768, runtime.Version(), hashes, arms}))
	check(f.Sync())
}
