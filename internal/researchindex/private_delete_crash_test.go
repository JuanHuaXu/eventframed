package researchindex

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

// The child stops at an explicit transaction boundary. The parent kills only
// this test subprocess, without graceful DB Close or writer publication.
func TestPrivateDeleteCrashChild(t *testing.T) {
	path := os.Getenv("RESEARCH_DELETE_CRASH_PATH")
	if path == "" {
		t.Skip("subprocess only")
	}
	mode := os.Getenv("RESEARCH_DELETE_CRASH_MODE")
	if mode != "staged" && mode != "committed" {
		t.Fatal("invalid crash mode")
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
		if e := tx.Upsert(ctx, "records", "target", []float32{1, 0}, nil); e != nil {
			return e
		}
		if e := tx.Upsert(ctx, "records", "survivor", []float32{0, 1}, nil); e != nil {
			return e
		}
		return tx.Upsert(ctx, "state", "revision", nil, map[string]interface{}{"value": "1"})
	}); e != nil {
		t.Fatal(e)
	}
	var edits []LayeredEdit
	var summary EntrySummary
	for i, id := range []string{"target", "survivor"} {
		v := []float32{1, 0}
		if i == 1 {
			v = []float32{0, 1}
		}
		edits = append(edits, LayeredEdit{uint32(i), &LayeredRecord{ID: id, Vector: v, Links: [][]uint32{{}}, Backlinks: [][]uint32{{}}, Heuristic: []uint32{0}}})
		summary, _ = summary.With(uint32(i), 0)
	}
	graph, _, e := PrepareLayered(ctx, LayeredSnapshot{}, edits, 1, LayeredLimits{2, 2, 1, 8})
	if e != nil {
		t.Fatal(e)
	}
	barrier := func() {
		fmt.Println("RESEARCH_DELETE_CRASH_READY")
		for {
			time.Sleep(time.Hour)
		}
	}
	w, e := NewPrivateDeleteWriter(graph, summary, 1, 2, func(c context.Context, revision uint64, ms []Mutation) error {
		e := db.WithTx(c, func(tx libra.Tx) error {
			for _, m := range ms {
				if e := tx.Delete(c, "records", m.ID); e != nil {
					return e
				}
			}
			if mode == "staged" {
				barrier()
			}
			return tx.Upsert(c, "state", "revision", nil, map[string]interface{}{"value": fmt.Sprint(revision)})
		})
		if e != nil {
			return e
		}
		barrier()
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	if e = w.Delete(ctx, 0); e != nil {
		t.Fatal(e)
	}
	t.Fatal("crash barrier unexpectedly returned")
}

func TestPrivateDeleteAbruptRecovery(t *testing.T) {
	for _, mode := range []string{"staged", "committed"} {
		t.Run(mode, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "authority")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPrivateDeleteCrashChild$", "-test.count=1")
			cmd.Env = append(os.Environ(), "RESEARCH_DELETE_CRASH_PATH="+path, "RESEARCH_DELETE_CRASH_MODE="+mode)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			stdout, e := cmd.StdoutPipe()
			if e != nil {
				t.Fatal(e)
			}
			if e = cmd.Start(); e != nil {
				t.Fatal(e)
			}
			scan := bufio.NewScanner(stdout)
			ready := false
			for scan.Scan() {
				if scan.Text() == "RESEARCH_DELETE_CRASH_READY" {
					ready = true
					break
				}
			}
			if !ready {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				t.Fatalf("barrier not reached: %v %s", scan.Err(), stderr.String())
			}
			if e = cmd.Process.Kill(); e != nil {
				_ = cmd.Wait()
				t.Fatal(e)
			}
			waitErr := cmd.Wait()
			if waitErr == nil || ctx.Err() != nil {
				t.Fatal("not a controlled abrupt exit", waitErr, ctx.Err())
			}
			db, e := libra.Open(libra.WithStoragePath(path))
			if e != nil {
				t.Fatal(e)
			}
			defer db.Close()
			col, e := db.EnsureCollection(context.Background(), "records", 2, libra.WithFlat(), libra.WithMetric(libra.CosineDistance))
			if e != nil {
				t.Fatal(e)
			}
			state, e := db.EnsureCollection(context.Background(), "state", 0, libra.WithMetadataOnly())
			if e != nil {
				t.Fatal(e)
			}
			wantCount, wantRev := 1, "2"
			if mode == "staged" {
				wantCount, wantRev = 2, "1"
			}
			count, e := col.Count(context.Background())
			if e != nil || count != wantCount {
				t.Fatal("recovered count", count, e)
			}
			r, e := state.Get(context.Background(), "revision")
			if e != nil || r.Metadata["value"] != wantRev {
				t.Fatal("recovered revision", r, e)
			}
			s, e := col.Get(context.Background(), "survivor")
			if e != nil || !reflect.DeepEqual(s.Vector, []float32{0, 1}) {
				t.Fatal("survivor", s, e)
			}
			target, e := col.Get(context.Background(), "target")
			if mode == "staged" {
				if e != nil || !reflect.DeepEqual(target.Vector, []float32{1, 0}) {
					t.Fatal("uncommitted deletion leaked", target, e)
				}
			} else if e == nil {
				t.Fatal("committed deletion lost")
			}
		})
	}
}
