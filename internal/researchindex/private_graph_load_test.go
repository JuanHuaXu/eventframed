package researchindex

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	libra "github.com/xDarkicex/libravdb/libravdb"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

type privateLoadSample struct {
	NS, WaitNS, PrepareNS, PersistNS, AcquireNS, SearchNS int64
	Error                                                 string
	Hit                                                   bool
	Revision                                              uint64
	BatchID                                               uint64
	BatchSize                                             int
}
type privateLoadArm struct {
	Repeat                                int
	Batched                               bool
	Reads, Writes                         []privateLoadSample
	WallNS                                int64
	Acknowledged, Reopened, FailedPresent int
	DurableRevision                       uint64
	AuditErrors                           []string
}

func privateLoadVector(id string) []float32 {
	v := make([]float32, 768)
	for b := 0; b < 24; b++ {
		h := sha256.Sum256([]byte(fmt.Sprintf("%s/block-%d", id, b)))
		for j, x := range h {
			v[b*32+j] = float32(x) - 127.5
		}
	}
	return v
}
func privateLoadLevel(id string) int {
	h := sha256.Sum256([]byte("private-load-level-v1/" + id))
	u := (float64(binary.LittleEndian.Uint64(h[:8])>>11) + 1) / (float64(uint64(1)<<53) + 1)
	return min(31, int(-math.Log(u)))
}

func TestPrivateGraphSustainedLoad(t *testing.T) {
	output := os.Getenv("RESEARCH_PRIVATE_LOAD_OUTPUT")
	if output == "" {
		t.Skip("explicit load output")
	}
	raw, e := os.ReadFile(os.Getenv("RESEARCH_PRIVATE_LOAD_CAPTURE"))
	if e != nil {
		t.Fatal(e)
	}
	var rows []layerRow
	if e = json.Unmarshal(raw, &rows); e != nil {
		t.Fatal(e)
	}
	var capture layerCapture
	for _, r := range rows {
		if r.N == 6400 {
			capture = r.Before
			break
		}
	}
	if len(capture.Nodes) != 6400 {
		t.Fatal("initial corpus")
	}
	root := output + ".stores"
	if e = os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	oldCPU := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(oldCPU)
	var arms []privateLoadArm
	for repeat := 0; repeat < 2; repeat++ {
		a := privateLoadArm{Repeat: repeat, Reads: make([]privateLoadSample, 8192), Writes: make([]privateLoadSample, 4096)}
		path := filepath.Join(root, fmt.Sprintf("r%d", repeat))
		db, e := libra.Open(libra.WithStoragePath(path))
		if e != nil {
			t.Fatal(e)
		}
		ctx := context.Background()
		if _, e = db.EnsureCollection(ctx, "records", 768, libra.WithFlat(), libra.WithMetric(libra.CosineDistance)); e != nil {
			t.Fatal(e)
		}
		if _, e = db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly()); e != nil {
			t.Fatal(e)
		}
		persist := func(c context.Context, rev uint64, changes []Mutation) error {
			return db.WithTx(c, func(tx libra.Tx) error {
				for _, m := range changes {
					if e := tx.Upsert(c, "records", m.ID, m.Vector, map[string]interface{}{"public": "Curiosity landed on Mars in 2012"}); e != nil {
						return e
					}
				}
				return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"revision": strconv.FormatUint(rev, 10)})
			})
		}
		for i := 0; i < 6400; i += 16 {
			var ms []Mutation
			for j := i; j < i+16; j++ {
				id := fmt.Sprintf("seed-%d", j)
				ms = append(ms, Mutation{ID: id, Vector: privateLoadVector(id)})
			}
			if e = persist(ctx, 1, ms); e != nil {
				t.Fatal(e)
			}
		}
		var edits []LayeredEdit
		for id, r := range capture.Nodes {
			n := layerInput(t, r)
			n.Vector = privateLoadVector(r.ID)
			edits = append(edits, LayeredEdit{id, n})
		}
		graph, _, e := PrepareLayered(ctx, LayeredSnapshot{}, edits, capture.Global, LayeredLimits{6400, 768, 32, 1024})
		if e != nil {
			t.Fatal(e)
		}
		w, e := NewPrivateGraphWriter(graph, 1, 768, 16, 200, 8, 2, persist)
		if e != nil {
			t.Fatal(e)
		}
		var collector *PrivateBatchCollector
		a.Batched = os.Getenv("RESEARCH_PRIVATE_LOAD_BATCH") == "1"
		if a.Batched {
			collector, e = NewPrivateBatchCollector(w, 64, 10*time.Millisecond)
			if e != nil {
				t.Fatal(e)
			}
		}
		runCtx, cancelRun := context.WithTimeout(ctx, 120*time.Second)
		start := time.Now().Add(20 * time.Millisecond)
		var wg sync.WaitGroup
		slots := make(chan struct{}, 64)
		for i := range a.Reads {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				due := start.Add(time.Duration(i) * 5 * time.Millisecond)
				time.Sleep(time.Until(due))
				c, cancel := context.WithDeadline(runCtx, due.Add(100*time.Millisecond))
				defer cancel()
				r := privateLoadSample{}
				begin := time.Now()
				lease, err := w.Acquire(c)
				r.AcquireNS = time.Since(begin).Nanoseconds()
				if err == nil {
					r.Revision, _ = lease.Revision()
					id := fmt.Sprintf("seed-%d", i%6400)
					begin = time.Now()
					found, _, e := lease.Search(c, privateLoadVector(id), 10, 400, 20000)
					r.SearchNS = time.Since(begin).Nanoseconds()
					err = e
					for _, f := range found {
						if f.ID == id {
							r.Hit = true
						}
					}
					lease.Release()
				}
				r.NS = time.Since(due).Nanoseconds()
				if err != nil {
					r.Error = err.Error()
				}
				a.Reads[i] = r
			}(i)
		}
		for i := range a.Writes {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				due := start.Add(time.Duration(2*i+1) * 5 * time.Millisecond)
				time.Sleep(time.Until(due))
				c, cancel := context.WithDeadline(runCtx, due.Add(100*time.Millisecond))
				defer cancel()
				r := privateLoadSample{}
				var err error
				select {
				case slots <- struct{}{}:
					id := fmt.Sprintf("write-%d", i)
					var timing PrivateWriteTiming
					if collector != nil {
						result, e := collector.Submit(c, PrivateInsertRequest{id, privateLoadVector(id), privateLoadLevel(id)})
						timing = result.Timing
						err = e
						r.BatchID = result.BatchID
						r.BatchSize = result.BatchSize
					} else {
						timing, err = w.Insert(c, id, privateLoadVector(id), privateLoadLevel(id))
					}
					r.WaitNS = timing.Wait.Nanoseconds()
					r.PrepareNS = timing.Prepare.Nanoseconds()
					r.PersistNS = timing.Persist.Nanoseconds()
					<-slots
				default:
					err = ErrCapacity
				}
				r.NS = time.Since(due).Nanoseconds()
				if err != nil {
					r.Error = err.Error()
				}
				a.Writes[i] = r
			}(i)
		}
		wg.Wait()
		a.WallNS = time.Since(start).Nanoseconds()
		cancelRun()
		if collector != nil {
			if e = collector.Close(ctx); e != nil {
				a.AuditErrors = append(a.AuditErrors, e.Error())
			}
		}
		if e = w.Close(ctx); e != nil {
			a.AuditErrors = append(a.AuditErrors, e.Error())
		}
		if e = db.Close(); e != nil {
			a.AuditErrors = append(a.AuditErrors, e.Error())
		}
		db, e = libra.Open(libra.WithStoragePath(path))
		if e != nil {
			t.Fatal(e)
		}
		col, e := db.GetCollection("records")
		if e != nil {
			t.Fatal(e)
		}
		state, e := db.GetCollection("state")
		if e != nil {
			t.Fatal(e)
		}
		rev, e := state.Get(ctx, "revision")
		if e != nil {
			t.Fatal(e)
		}
		a.DurableRevision, e = strconv.ParseUint(rev.Metadata["revision"].(string), 10, 64)
		if e != nil {
			t.Fatal(e)
		}
		for i, s := range a.Writes {
			id := fmt.Sprintf("write-%d", i)
			r, e := col.Get(ctx, id)
			if e != nil && !errors.Is(e, libra.ErrRecordNotFound) {
				a.AuditErrors = append(a.AuditErrors, e.Error())
			}
			if s.Error == "" {
				a.Acknowledged++
				if e == nil && r.ID == id && reflect.DeepEqual(r.Vector, privateLoadVector(id)) {
					a.Reopened++
				} else {
					a.AuditErrors = append(a.AuditErrors, "lost acknowledgement "+id)
				}
			} else if e == nil {
				a.FailedPresent++
			}
		}
		if a.DurableRevision != uint64(1+a.Acknowledged) {
			a.AuditErrors = append(a.AuditErrors, "revision mismatch")
		}
		if e = db.Close(); e != nil {
			a.AuditErrors = append(a.AuditErrors, e.Error())
		}
		f, e := os.OpenFile(fmt.Sprintf("%s.arm%d.json", output, repeat), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			t.Fatal(e)
		}
		if e = json.NewEncoder(f).Encode(a); e != nil {
			t.Fatal(e)
		}
		if e = f.Close(); e != nil {
			t.Fatal(e)
		}
		arms = append(arms, a)
		t.Log("completed arm", repeat, "acknowledgements", a.Acknowledged)
	}
	f, e := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.NewEncoder(f).Encode(arms); e != nil {
		t.Fatal(e)
	}
	if e = f.Close(); e != nil {
		t.Fatal(e)
	}
	for _, a := range arms {
		bad := len(a.AuditErrors) + a.FailedPresent
		for _, r := range a.Reads {
			if r.Error != "" || r.NS > int64(100*time.Millisecond) || !r.Hit {
				bad++
			}
		}
		for _, r := range a.Writes {
			if r.Error != "" || r.NS > int64(100*time.Millisecond) {
				bad++
			}
		}
		if bad > 0 {
			t.Errorf("arm %d failed frozen serving gates: %d failing requests/audits", a.Repeat, bad)
		}
	}
}
