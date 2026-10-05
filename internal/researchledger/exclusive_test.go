package researchledger

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestExclusiveSettingsAndReplacement(t *testing.T) {
	path := t.TempDir() + "/exclusive.sqlite"
	l, e := OpenExclusiveResearch(path)
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	check := func() {
		t.Helper()
		c, e := l.db.Conn(context.Background())
		if e != nil {
			t.Fatal(e)
		}
		defer c.Close()
		var lock, journal string
		var sync int
		if e = c.QueryRowContext(context.Background(), "PRAGMA locking_mode").Scan(&lock); e != nil {
			t.Fatal(e)
		}
		if e = c.QueryRowContext(context.Background(), "PRAGMA journal_mode").Scan(&journal); e != nil {
			t.Fatal(e)
		}
		if e = c.QueryRowContext(context.Background(), "PRAGMA synchronous").Scan(&sync); e != nil {
			t.Fatal(e)
		}
		if lock != "exclusive" || journal != "wal" || sync != 2 {
			t.Fatal("unsafe settings", lock, journal, sync)
		}
	}
	check()
	if other, e := Open(path); e == nil {
		other.Close()
		t.Fatal("second owner accepted")
	}
	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	q.Set("_busy_timeout", "10")
	u.RawQuery = q.Encode()
	raw, e := sql.Open("sqlite", u.String())
	if e != nil {
		t.Fatal(e)
	}
	defer raw.Close()
	var n int
	e = raw.QueryRow("SELECT count(*) FROM research_log").Scan(&n)
	var code interface{ Code() int }
	if !errors.As(e, &code) || code.Code()&255 != 5 {
		t.Fatal("external reader not busy", e)
	}
	l.db.SetMaxIdleConns(0)
	for i := 0; i < 3; i++ {
		check()
	}
	if l.db.Stats().MaxIdleClosed < 3 {
		t.Fatal("replacement not exercised")
	}
	if e = l.Close(); e != nil {
		t.Fatal(e)
	}
	if e = raw.QueryRow("SELECT count(*) FROM research_log").Scan(&n); e != nil {
		t.Fatal("reader remains blocked after close", e)
	}
}

func TestExclusiveCrashHelper(t *testing.T) {
	path := os.Getenv("EVENTFRAME_EXCLUSIVE_CHILD")
	if path == "" {
		t.Skip("child only")
	}
	l, e := OpenExclusiveResearch(path)
	if e != nil {
		t.Fatal(e)
	}
	var hook func()
	if os.Getenv("EVENTFRAME_EXCLUSIVE_PHASE") == "before" {
		hook = func() { os.Exit(23) }
	}
	if _, e = l.appendBatchMode(context.Background(), batchFixture(), true, hook); e != nil {
		t.Fatal(e)
	}
	os.Exit(23)
}

func TestExclusiveCrashRecovery(t *testing.T) {
	for _, phase := range []string{"before", "after"} {
		t.Run(phase, func(t *testing.T) {
			path := t.TempDir() + "/crash.sqlite"
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestExclusiveCrashHelper$")
			cmd.Env = append(os.Environ(), "EVENTFRAME_EXCLUSIVE_CHILD="+path, "EVENTFRAME_EXCLUSIVE_PHASE="+phase)
			out, e := cmd.CombinedOutput()
			var ex *exec.ExitError
			if !errors.As(e, &ex) || ex.ExitCode() != 23 {
				t.Fatalf("child %v %s", e, out)
			}
			l, e := Open(path)
			if e != nil {
				t.Fatal(e)
			}
			r, e := l.ReadAfter(ctx, 0, 256)
			if e != nil || (phase == "before" && len(r) != 0) || (phase == "after" && len(r) != 4) {
				t.Fatal("crash rows", r, e)
			}
			if e = l.Close(); e != nil {
				t.Fatal(e)
			}
			l, e = OpenExclusiveResearch(path)
			if e != nil {
				t.Fatal(e)
			}
			defer l.Close()
			ack, e := l.AppendBatchPrepared(ctx, batchFixture())
			if e != nil {
				t.Fatal(e)
			}
			for _, a := range ack {
				if a.Retry != (phase == "after") {
					t.Fatal("crash retry", a)
				}
			}
		})
	}
}
