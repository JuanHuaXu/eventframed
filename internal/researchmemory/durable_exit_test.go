package researchmemory

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func durableWrapperFixture(t *testing.T, path, phase string, exit bool) Frozen {
	t.Helper()
	ctx := context.Background()
	d, e := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	func() {
		d.worker.adapter.mu.Lock()
		defer d.worker.adapter.mu.Unlock()
		for id := uint64(1); id <= 64; id++ {
			if _, dup, e := d.Admit(ctx, id, uint16(id), .6, durableFixtureTime); e != nil || dup {
				t.Fatal(id, dup, e)
			}
			if phase == "discarded" {
				if dup, e := d.Discard(ctx, id, durableFixtureTime); e != nil || dup {
					t.Fatal(id, dup, e)
				}
			} else if phase != "pending" {
				if dup, e := d.Feedback(ctx, id, id%3 != 0, durableFixtureTime); e != nil || dup {
					t.Fatal(id, dup, e)
				}
			}
		}
		if exit && phase != "applied" {
			os.Exit(25)
		}
	}()
	if phase != "pending" && phase != "discarded" {
		wait, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if e = d.worker.WaitProcessed(wait, 64); e != nil {
			t.Fatal(e)
		}
	}
	if exit {
		os.Exit(25)
	}
	return d.worker.Snapshot()
}

func TestDurableWrapperExitHelper(t *testing.T) {
	path := os.Getenv("EVENTFRAME_DURABLE_CHILD_PATH")
	if path == "" {
		t.Skip("child only")
	}
	phase := os.Getenv("EVENTFRAME_DURABLE_CHILD_PHASE")
	if phase != "pending" && phase != "queued" && phase != "applied" && phase != "discarded" {
		t.Fatal("invalid phase")
	}
	durableWrapperFixture(t, path, phase, true)
	t.Fatal("exit not reached")
}

func TestDurableWrapperAcknowledgedRecovery(t *testing.T) {
	for _, phase := range []string{"pending", "queued", "applied", "discarded"} {
		t.Run(phase, func(t *testing.T) {
			control := durableWrapperFixture(t, filepath.Join(t.TempDir(), "control.sqlite"), phase, false)
			path := filepath.Join(t.TempDir(), "recovery.sqlite")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDurableWrapperExitHelper$")
			cmd.Env = append(os.Environ(), "EVENTFRAME_DURABLE_CHILD_PATH="+path, "EVENTFRAME_DURABLE_CHILD_PHASE="+phase)
			output, e := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(e, &exit) || exit.ExitCode() != 25 {
				t.Fatalf("boundary not reached: %v %s", e, output)
			}
			d, e := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
			if e != nil {
				t.Fatal(e)
			}
			defer d.Close()
			n, f, p, q := d.worker.Counts()
			wantN, wantP := uint64(64), 0
			if phase == "pending" {
				wantN, wantP = 0, 64
			} else if phase == "discarded" {
				wantN = 0
			}
			if f != 0 || q != 0 || n != wantN || p != wantP {
				t.Fatal(n, f, p, q)
			}
			for x := uint16(0); x < 512; x++ {
				want, e := control.Score(x, .6, 1, durableFixtureTime)
				got, f := d.worker.Snapshot().Score(x, .6, 1, durableFixtureTime)
				if e != nil || f != nil || want != got {
					t.Fatal(x, want, got, e, f)
				}
			}
			for id := uint64(1); id <= 64; id++ {
				if p, dup, e := d.Admit(ctx, id, uint16(id), .6, durableFixtureTime); e != nil || !dup || p.Probability != .6 {
					t.Fatal("ack retry changed", p, dup, e)
				}
				if phase == "discarded" {
					if dup, e := d.Discard(ctx, id, durableFixtureTime); e != nil || !dup {
						t.Fatal("discard retry", dup, e)
					}
					if _, e := d.Feedback(ctx, id, true, durableFixtureTime); e == nil {
						t.Fatal("discarded record learned")
					}
				} else if phase != "pending" {
					if dup, e := d.Feedback(ctx, id, id%3 != 0, durableFixtureTime); e != nil || !dup {
						t.Fatal("feedback retry", dup, e)
					}
				}
			}
			if phase == "pending" {
				if dup, e := d.Feedback(ctx, 1, true, durableFixtureTime); e != nil || dup {
					t.Fatal("pending cannot continue", dup, e)
				}
				if e = d.worker.WaitProcessed(ctx, 1); e != nil {
					t.Fatal(e)
				}
			}
		})
	}
}
