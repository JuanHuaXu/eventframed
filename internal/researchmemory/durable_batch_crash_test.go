package researchmemory

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type exitBatchLog struct {
	durableLog
	commit bool
}

func (l exitBatchLog) AppendBatch(ctx context.Context, requests []researchledger.AppendRequest) ([]researchledger.AppendResult, error) {
	if l.commit {
		if _, err := l.durableLog.(batchDurableLog).AppendBatch(ctx, requests); err != nil {
			return nil, err
		}
	}
	os.Exit(23)
	return nil, nil
}

func TestDurableBatchExitHelper(t *testing.T) {
	path := os.Getenv("EVENTFRAME_DURABLE_BATCH_PATH")
	if path == "" {
		t.Skip("subprocess only")
	}
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	d, err := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	requests := []AdmissionRequest{{ID: 1, Features: 1, Baseline: .6, At: now}, {ID: 2, Features: 2, Baseline: .6, At: now}}
	discard := os.Getenv("EVENTFRAME_DURABLE_BATCH_KIND") == "discard"
	if discard {
		if _, err = d.AdmitBatch(ctx, requests); err != nil {
			t.Fatal(err)
		}
	}
	d.log = exitBatchLog{d.log, os.Getenv("EVENTFRAME_DURABLE_BATCH_COMMIT") == "yes"}
	if discard {
		_, err = d.DiscardBatch(ctx, []DiscardRequest{{1, now}, {2, now}})
	} else {
		_, err = d.AdmitBatch(ctx, requests)
	}
	t.Fatal("exit boundary not reached", err)
}

func TestDurableBatchProcessRecovery(t *testing.T) {
	for _, kind := range []string{"admit", "discard"} {
		for _, phase := range []string{"no", "yes"} {
			t.Run(kind+"/committed="+phase, func(t *testing.T) {
				path := t.TempDir() + "/crash.sqlite"
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDurableBatchExitHelper$")
				cmd.Env = append(os.Environ(), "EVENTFRAME_DURABLE_BATCH_PATH="+path, "EVENTFRAME_DURABLE_BATCH_KIND="+kind, "EVENTFRAME_DURABLE_BATCH_COMMIT="+phase)
				out, err := cmd.CombinedOutput()
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 23 {
					t.Fatalf("child missed boundary: %v %s", err, out)
				}
				d, err := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				defer d.Close()
				n, f, p, q := d.worker.Counts()
				want := 0
				if (kind == "admit" && phase == "yes") || (kind == "discard" && phase == "no") {
					want = 2
				}
				if n != 0 || f != 0 || p != want || q != 0 {
					t.Fatal("partial recovery or invented label", n, f, p, q)
				}
				now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
				if kind == "admit" {
					r, err := d.AdmitBatch(ctx, []AdmissionRequest{{ID: 1, Features: 1, Baseline: .6, At: now}, {ID: 2, Features: 2, Baseline: .6, At: now}})
					if err != nil {
						t.Fatal(err)
					}
					for _, v := range r {
						if v.Retry != (phase == "yes") {
							t.Fatal("admission retry mismatch")
						}
					}
				} else {
					r, err := d.DiscardBatch(ctx, []DiscardRequest{{1, now}, {2, now}})
					if err != nil {
						t.Fatal(err)
					}
					for _, v := range r {
						if v != (phase == "yes") {
							t.Fatal("discard retry mismatch")
						}
					}
				}
			})
		}
	}
}
