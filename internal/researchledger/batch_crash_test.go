package researchledger

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

func TestBatchAbruptExitHelper(t *testing.T) {
	path := os.Getenv("EVENTFRAME_BATCH_CHILD_PATH")
	if path == "" {
		t.Skip("subprocess only")
	}
	l, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	var hook func()
	if os.Getenv("EVENTFRAME_BATCH_CHILD_PHASE") == "before" {
		hook = func() { os.Exit(23) }
	}
	if _, err = l.appendBatchMode(context.Background(), batchFixture(), os.Getenv("EVENTFRAME_BATCH_CHILD_PREPARED") == "true", hook); err != nil {
		t.Fatal(err)
	}
	os.Exit(23)
}

// Process exit on either side of COMMIT exercises the actual batch entry point.
// This is not hardware power-loss testing or a cross-database atomicity proof.
func TestBatchProcessExitBoundary(t *testing.T) {
	for _, prepared := range []string{"false", "true"} {
		for _, phase := range []string{"before", "after"} {
			t.Run(prepared+"/"+phase, func(t *testing.T) {
				path := t.TempDir() + "/crash.sqlite"
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestBatchAbruptExitHelper$")
				cmd.Env = append(os.Environ(), "EVENTFRAME_BATCH_CHILD_PATH="+path, "EVENTFRAME_BATCH_CHILD_PHASE="+phase, "EVENTFRAME_BATCH_CHILD_PREPARED="+prepared)
				out, err := cmd.CombinedOutput()
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 23 {
					t.Fatalf("child failed boundary: %v %s", err, out)
				}
				l, err := Open(path)
				if err != nil {
					t.Fatal(err)
				}
				defer l.Close()
				rows, err := l.ReadAfter(ctx, 0, 256)
				if err != nil {
					t.Fatal(err)
				}
				want := 0
				if phase == "after" {
					want = 4
				}
				if len(rows) != want {
					t.Fatal("partial recovery", len(rows), want)
				}
				requests := batchFixture()
				for i, row := range rows {
					if row.Key != requests[i].Key || row.Kind != requests[i].Kind || !reflect.DeepEqual(row.Payload, requests[i].Payload) {
						t.Fatal("original bytes changed", i)
					}
				}
				results, err := l.appendBatchMode(ctx, requests, prepared == "true", nil)
				if err != nil {
					t.Fatal(err)
				}
				for i, r := range results {
					if r.Sequence != int64(i+1) || r.Retry != (phase == "after") {
						t.Fatal("retry boundary", r)
					}
				}
				var integrity string
				if err = l.db.QueryRow("PRAGMA integrity_check").Scan(&integrity); err != nil || integrity != "ok" {
					t.Fatal(integrity, err)
				}
			})
		}
	}
}
