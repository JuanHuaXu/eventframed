package researchledger

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// A child exits without Close or deferred rollback. This is process termination,
// not power loss; test-only SQL places the uncommitted case at the boundary.
func TestLedgerAbruptExitHelper(t *testing.T) {
	path := os.Getenv("EVENTFRAME_LEDGER_CHILD_PATH")
	if path == "" {
		t.Skip("subprocess only")
	}
	kind := os.Getenv("EVENTFRAME_LEDGER_CHILD_KIND")
	committed := os.Getenv("EVENTFRAME_LEDGER_CHILD_COMMITTED") == "yes"
	if kind != "admit" && kind != "feedback" {
		t.Fatal("invalid child phase")
	}
	l, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	key := Key{"tenant", "journal", "event", "fixture-v1"}
	if kind == "feedback" {
		if _, _, e = l.Append(ctx, key, "admit", json.RawMessage(`{"forecast":0.6}`)); e != nil {
			t.Fatal(e)
		}
	}
	payload := json.RawMessage(`{"fixture":true}`)
	if committed {
		if _, _, e = l.Append(ctx, key, kind, payload); e != nil {
			t.Fatal(e)
		}
	} else {
		tx, e := l.db.BeginTx(ctx, nil)
		if e != nil {
			t.Fatal(e)
		}
		identity, _ := json.Marshal(key)
		if _, e = tx.ExecContext(ctx, "INSERT INTO research_log(identity,kind,payload) VALUES(?,?,?)", string(identity), kind, []byte(payload)); e != nil {
			t.Fatal(e)
		}
	}
	os.Exit(23)
}

func TestAbruptExitCommitBoundary(t *testing.T) {
	for _, kind := range []string{"admit", "feedback"} {
		for _, commit := range []bool{false, true} {
			phase := "uncommitted"
			if commit {
				phase = "committed"
			}
			t.Run(kind+"/"+phase, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "crash.sqlite")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLedgerAbruptExitHelper$")
				flag := "no"
				if commit {
					flag = "yes"
				}
				cmd.Env = append(os.Environ(), "EVENTFRAME_LEDGER_CHILD_PATH="+path, "EVENTFRAME_LEDGER_CHILD_KIND="+kind, "EVENTFRAME_LEDGER_CHILD_COMMITTED="+flag)
				out, e := cmd.CombinedOutput()
				var exit *exec.ExitError
				if !errors.As(e, &exit) || exit.ExitCode() != 23 {
					t.Fatalf("child did not reach exit boundary: %v %s", e, out)
				}
				l, e := Open(path)
				if e != nil {
					t.Fatal(e)
				}
				defer l.Close()
				rows, e := l.ReadAfter(context.Background(), 0, 256)
				if e != nil {
					t.Fatal(e)
				}
				want := 0
				if kind == "feedback" {
					want++
				}
				if commit {
					want++
				}
				if len(rows) != want {
					t.Fatalf("recovered %d records, want %d", len(rows), want)
				}
				key := Key{"tenant", "journal", "event", "fixture-v1"}
				seq, duplicate, e := l.Append(context.Background(), key, kind, json.RawMessage(`{"fixture":true}`))
				if e != nil || duplicate != commit {
					t.Fatal(seq, duplicate, e)
				}
				if kind == "feedback" && seq != 2 || kind == "admit" && seq != 1 {
					t.Fatal("unexpected replay order", seq)
				}
				var integrity string
				if e = l.db.QueryRow("PRAGMA integrity_check").Scan(&integrity); e != nil || integrity != "ok" {
					t.Fatal(integrity, e)
				}
			})
		}
	}
}
