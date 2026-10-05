package researchindex

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	libra "github.com/xDarkicex/libravdb/libravdb"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestPrivateGraphCrashChild(t *testing.T) {
	path := os.Getenv("RESEARCH_GRAPH_CRASH_PATH")
	if path == "" {
		t.Skip("child only")
	}
	mode := os.Getenv("RESEARCH_GRAPH_CRASH_MODE")
	if mode != "staged" && mode != "committed" {
		t.Fatal("mode")
	}
	ctx := context.Background()
	db, e := libra.Open(libra.WithStoragePath(path))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.EnsureCollection(ctx, "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance)); e != nil {
		t.Fatal(e)
	}
	if _, e = db.EnsureCollection(ctx, "state", 0, libra.WithMetadataOnly()); e != nil {
		t.Fatal(e)
	}
	if e = db.WithTx(ctx, func(tx libra.Tx) error {
		if e := tx.Upsert(ctx, "records", "survivor", []float32{1, 0}, nil); e != nil {
			return e
		}
		return tx.Upsert(ctx, "state", "revision", nil, map[string]interface{}{"value": "1"})
	}); e != nil {
		t.Fatal(e)
	}
	graph, _, e := PrepareLayered(ctx, LayeredSnapshot{}, []LayeredEdit{{0, &LayeredRecord{ID: "survivor", Vector: []float32{1, 0}, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}}}, 1, LayeredLimits{1, 2, 1, 8})
	if e != nil {
		t.Fatal(e)
	}
	barrier := func() {
		fmt.Println("RESEARCH_GRAPH_CRASH_READY")
		for {
			time.Sleep(time.Hour)
		}
	}
	w, e := NewPrivateGraphWriter(graph, 1, 2, 4, 32, 4, 2, func(c context.Context, rev uint64, ms []Mutation) error {
		err := db.WithTx(c, func(tx libra.Tx) error {
			for _, m := range ms {
				if m.Delete {
					t.Fatal("unexpected deletion")
				}
				if e := tx.Upsert(c, "records", m.ID, m.Vector, nil); e != nil {
					return e
				}
			}
			if mode == "staged" {
				barrier()
			}
			return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"value": fmt.Sprint(rev)})
		})
		if err != nil {
			return err
		}
		barrier()
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	batch, err := strconv.Atoi(os.Getenv("RESEARCH_GRAPH_CRASH_BATCH"))
	if err != nil || (batch != 1 && batch != 4) {
		t.Fatal("invalid batch")
	}
	if batch == 1 {
		_, e = w.Insert(ctx, "inserted-0", []float32{0, 1}, 0)
	} else {
		var requests []PrivateInsertRequest
		for i := 0; i < batch; i++ {
			requests = append(requests, PrivateInsertRequest{fmt.Sprintf("inserted-%d", i), []float32{float32(i), 1}, 0})
		}
		_, e = w.InsertBatch(ctx, requests)
	}
	if e != nil {
		t.Fatal(e)
	}
	t.Fatal("barrier returned")
}

func TestPrivateGraphAbruptRecovery(t *testing.T) {
	for _, batch := range []int{1, 4} {
		for _, mode := range []string{"staged", "committed"} {
			t.Run(fmt.Sprintf("batch%d/%s", batch, mode), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "authority")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPrivateGraphCrashChild$", "-test.count=1")
				cmd.Env = append(os.Environ(), "RESEARCH_GRAPH_CRASH_PATH="+path, "RESEARCH_GRAPH_CRASH_MODE="+mode, "RESEARCH_GRAPH_CRASH_BATCH="+strconv.Itoa(batch))
				var stderr bytes.Buffer
				cmd.Stderr = &stderr
				stdout, e := cmd.StdoutPipe()
				if e != nil {
					t.Fatal(e)
				}
				if e = cmd.Start(); e != nil {
					t.Fatal(e)
				}
				scanner := bufio.NewScanner(stdout)
				ready := false
				for scanner.Scan() {
					if scanner.Text() == "RESEARCH_GRAPH_CRASH_READY" {
						ready = true
						break
					}
				}
				if !ready {
					_ = cmd.Process.Kill()
					_ = cmd.Wait()
					t.Fatalf("no barrier: %v %s", scanner.Err(), stderr.String())
				}
				if e = cmd.Process.Kill(); e != nil {
					_ = cmd.Wait()
					t.Fatal(e)
				}
				if e = cmd.Wait(); e == nil || ctx.Err() != nil {
					t.Fatal("not controlled kill", e, ctx.Err())
				}
				db, e := libra.Open(libra.WithStoragePath(path))
				if e != nil {
					t.Fatal(e)
				}
				defer db.Close()
				c := context.Background()
				col, e := db.EnsureCollection(c, "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance))
				if e != nil {
					t.Fatal(e)
				}
				state, e := db.EnsureCollection(c, "state", 0, libra.WithMetadataOnly())
				if e != nil {
					t.Fatal(e)
				}
				wantCount, wantRev := 1, "1"
				if mode == "committed" {
					wantCount, wantRev = 1+batch, strconv.Itoa(1+batch)
				}
				count, e := col.Count(c)
				if e != nil || count != wantCount {
					t.Fatal(count, e)
				}
				rev, e := state.Get(c, "revision")
				if e != nil || rev.Metadata["value"] != wantRev {
					t.Fatal(rev, e)
				}
				survivor, e := col.Get(c, "survivor")
				if e != nil || !reflect.DeepEqual(survivor.Vector, []float32{1, 0}) {
					t.Fatal(survivor, e)
				}
				for i := 0; i < batch; i++ {
					inserted, e := col.Get(c, fmt.Sprintf("inserted-%d", i))
					if mode == "staged" {
						if e == nil {
							t.Fatal("uncommitted insertion visible")
						}
					} else {
						if e != nil || !reflect.DeepEqual(inserted.Vector, []float32{float32(i), 1}) {
							t.Fatal("committed insertion lost", inserted, e)
						}
					}
				}
			})
		}
	}
}
