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
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchindex"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

type sample struct {
	NS       int64
	Error    string
	Hit      bool
	Revision uint64
}
type build struct {
	Mode                                                        string
	BeforeRuns, AfterRuns                                       int
	Partition, BeforeTotal, AfterTotal, BeforeLocal, AfterLocal int
	BeforeRevision, AfterRevision                               uint64
	NS                                                          int64
	Error                                                       string
}
type arm struct {
	RetirementDrainNS                     int64
	N, GapMS, Repeat                      int
	Reads, Writes                         []sample
	Builds                                []build
	WallNS                                int64
	Acknowledged, Reopened, FailedPresent int
	DurableRevision                       uint64
	AuditErrors                           []string
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
func run(root string, n, gap, repeat int) arm {
	a := arm{N: n, GapMS: gap, Repeat: repeat, Reads: make([]sample, 1024), Writes: make([]sample, 512)}
	dir := filepath.Join(root, fmt.Sprintf("n%d-gap%d-r%d", n, gap, repeat))
	check(os.Mkdir(dir, 0700))
	path := filepath.Join(dir, "authoritative.libravdb")
	db, err := libra.Open(libra.WithStoragePath(path))
	check(err)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	_, err = db.EnsureCollection(ctx, "records", 768, libra.WithFlat(), libra.WithMetric(libra.CosineDistance))
	check(err)
	_, err = db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly())
	check(err)
	persist := func(c context.Context, revision uint64, changes []researchindex.Mutation) error {
		return db.WithTx(c, func(tx libra.Tx) error {
			for _, m := range changes {
				if err := tx.Upsert(c, "records", m.ID, m.Vector, map[string]interface{}{"public": "Curiosity landed on Mars in 2012"}); err != nil {
					return err
				}
			}
			return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"revision": strconv.FormatUint(revision, 10)})
		})
	}
	seed := make([]researchindex.Mutation, n)
	for i := range seed {
		id := fmt.Sprintf("seed-%d", i)
		seed[i] = researchindex.Mutation{ID: id, Vector: vector(id)}
	}
	// Setup is unserved; all bounded chunks belong to initial revision1.
	for i := 0; i < len(seed); i += 16 {
		check(persist(ctx, 1, seed[i:min(i+16, len(seed))]))
	}
	base, err := researchindex.BuildImmutableRun(ctx, seed, 768, filepath.Join(dir, "base-initial"))
	check(err)
	w, err := researchindex.NewDurableRunWriter(ctx, []*researchindex.ImmutableRun{base}, 1, 768, 64, 10, persist)
	check(err)
	pool, err := researchindex.NewRunLeasePool(8, 2)
	check(err)
	stop, stopped := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		count := 0
		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			v, err := w.Status(ctx)
			if err != nil {
				a.Builds = append(a.Builds, build{Error: err.Error()})
				return
			}
			if v.Delta < 32 {
				continue
			}
			count++
			start := time.Now()
			mode := "flush"
			if v.Runs >= 2 {
				mode = "consolidate"
				c, e := w.PrepareRunConsolidation(ctx, filepath.Join(dir, fmt.Sprintf("base-%d", count)))
				err = e
				if err == nil {
					err = pool.PublishConsolidation(ctx, c)
					if err != nil {
						err = errors.Join(err, c.Abort())
					}
				}
			} else {
				f, e := w.PrepareRunFlush(ctx, filepath.Join(dir, fmt.Sprintf("base-%d", count)))
				err = e
				if err == nil {
					err = f.Publish(ctx)
				}
			}
			b := build{Mode: mode, BeforeRuns: v.Runs, NS: time.Since(start).Nanoseconds(), BeforeTotal: v.Delta, BeforeLocal: v.Delta, BeforeRevision: v.Revision}
			after, viewErr := w.Status(ctx)
			if viewErr == nil {
				b.AfterTotal = after.Delta
				b.AfterLocal = after.Delta
				b.AfterRevision = after.Revision
				b.AfterRuns = after.Runs
			} else if err == nil {
				err = viewErr
			}
			if err != nil {
				b.Error = err.Error()
			}
			a.Builds = append(a.Builds, b)
			if err != nil {
				return
			}
		}
	}()
	start := time.Now().Add(20 * time.Millisecond)
	var wg sync.WaitGroup
	for i := 0; i < 1024; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			due := start.Add(time.Duration(i*gap) * time.Millisecond)
			time.Sleep(time.Until(due))
			c, cancel := context.WithDeadline(ctx, due.Add(100*time.Millisecond))
			defer cancel()
			id := fmt.Sprintf("seed-%d", i%n)
			lease, err := pool.AcquireCurrent(c, w)
			r := sample{}
			if err == nil {
				r.Revision = lease.Revision()
				found, e := lease.Search(c, vector(id))
				err = e
				for _, f := range found {
					if f.ID == id {
						r.Hit = true
					}
				}
				releaseErr := lease.Release()
				if err == nil {
					err = releaseErr
				}
			}
			r.NS = time.Since(due).Nanoseconds()
			if err != nil {
				r.Error = err.Error()
			}
			a.Reads[i] = r
		}(i)
	}
	for i := 0; i < 512; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			due := start.Add(time.Duration((2*i+1)*gap) * time.Millisecond)
			time.Sleep(time.Until(due))
			c, cancel := context.WithDeadline(ctx, due.Add(100*time.Millisecond))
			defer cancel()
			id := fmt.Sprintf("write-%d", i)
			err := w.Apply(c, []researchindex.Mutation{{ID: id, Vector: vector(id)}})
			r := sample{NS: time.Since(due).Nanoseconds()}
			if err != nil {
				r.Error = err.Error()
			}
			a.Writes[i] = r
		}(i)
	}
	wg.Wait()
	close(stop)
	<-stopped
	a.WallNS = time.Since(start).Nanoseconds()
	cleanupStart := time.Now()
	cleanupErr := pool.Close()
	if cleanupErr == nil {
		cleanupErr = w.CloseRunsAfterReaders(ctx)
	}
	a.RetirementDrainNS = time.Since(cleanupStart).Nanoseconds()
	if err := cleanupErr; err != nil {
		a.AuditErrors = append(a.AuditErrors, "close serving: "+err.Error())
	}
	if err := db.Close(); err != nil {
		a.AuditErrors = append(a.AuditErrors, "close store: "+err.Error())
	}
	db, err = libra.Open(libra.WithStoragePath(path))
	check(err)
	defer db.Close()
	collection, err := db.GetCollection("records")
	check(err)
	state, err := db.GetCollection("state")
	check(err)
	rev, err := state.Get(ctx, "revision")
	check(err)
	a.DurableRevision, err = strconv.ParseUint(rev.Metadata["revision"].(string), 10, 64)
	check(err)
	for i, w := range a.Writes {
		id := fmt.Sprintf("write-%d", i)
		r, err := collection.Get(ctx, id)
		present := err == nil
		if err != nil && !errors.Is(err, libra.ErrRecordNotFound) {
			a.AuditErrors = append(a.AuditErrors, err.Error())
		}
		if w.Error == "" {
			a.Acknowledged++
			if present && r.ID == id {
				a.Reopened++
			} else {
				a.AuditErrors = append(a.AuditErrors, "missing acknowledgement "+id)
			}
		} else if present {
			a.FailedPresent++
		}
	}
	if a.DurableRevision != uint64(1+a.Acknowledged) {
		a.AuditErrors = append(a.AuditErrors, "revision mismatch")
	}
	return a
}
func main() {
	if len(os.Args) != 2 {
		panic("NEW output required")
	}
	out := os.Args[1]
	f, err := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	check(err)
	defer f.Close()
	root := out + ".stores"
	check(os.Mkdir(root, 0700))
	runtime.GOMAXPROCS(4)
	hashes := map[string]string{}
	files := []string{"cmd/research-run-load/main.go", "research/public-task-pilot/RUN_LOAD_PROTOCOL.md", "internal/researchindex/immutable_run.go", "internal/researchindex/run_merge.go", "internal/researchindex/run_delta.go", "internal/researchindex/run_writer.go", "internal/researchindex/run_flush.go", "internal/researchindex/run_consolidation.go", "internal/researchindex/run_leases.go", "internal/researchindex/run_publication.go", "internal/researchindex/run_load_support.go", "research-candidate-only.mod", "research-candidate-only.sum", "research/public-task-pilot/CANDIDATE_ONLY_PROTOCOL.md", "research/public-task-pilot/candidate-only-overlay-v1/collection.go.txt", "research/public-task-pilot/candidate-only-overlay-v1/hnsw_base.go.txt", "research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json", "internal/researchindex/partition_async_serving.go", "internal/researchindex/retirement_queue.go", "research/public-task-pilot/ASYNC_RETIREMENT_PROTOCOL.md", "research/public-task-pilot/DERIVED_NOSYNC_PROTOCOL.md", "research/public-task-pilot/derived-nosync-overlay-v1/hnsw_base.go.txt", "research/public-task-pilot/derived-nosync-overlay-v1/overlay.json", "research/public-task-pilot/PARTITION_LOAD_PROTOCOL.md", "internal/researchindex/partition.go", "internal/researchindex/partition_durable.go", "internal/researchindex/partition_serving.go", "research/public-task-pilot/GROWTH_PROTOCOL.md", "research/public-task-pilot/WIDE_VECTOR_PROTOCOL.md", "research/public-task-pilot/bulk-base-overlay-v1/hnsw_base.go.txt", "research/public-task-pilot/bulk-base-overlay-v1/overlay.json", "research/public-task-pilot/BULK_BASE_PROTOCOL.md", "research/public-task-pilot/COMPACTION_LOAD_SETUP_CORRECTION.md", "internal/researchindex/serving.go", "internal/researchindex/generation.go", "internal/researchindex/durable.go", "internal/researchindex/durable_compaction.go", "internal/researchindex/hnsw_base.go", "internal/researchindex/merge.go", "go.mod", "research/public-task-pilot/COMPACTION_LOAD_PROTOCOL.md"}
	for _, p := range files {
		b, e := os.ReadFile(p)
		check(e)
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	var arms []arm
	for r := 0; r < 2; r++ {
		for _, n := range []int{800, 3200, 6400} {
			for _, gap := range []int{5} {
				a := run(root, n, gap, r)
				arms = append(arms, a)
				partial, e := os.OpenFile(fmt.Sprintf("%s.arm-n%d-gap%d-r%d.json", out, n, gap, r), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
				check(e)
				check(json.NewEncoder(partial).Encode(a))
				check(partial.Close())
				fmt.Println("completed", n, gap, r)
			}
		}
	}
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	check(encoder.Encode(struct {
		Dimension int
		Hashes    map[string]string
		Arms      []arm
	}{768, hashes, arms}))
}
