package researchmemory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

type durableFixtureFeedback struct {
	ID        uint64
	Useful    bool
	Available time.Time
}

var durableFixtureTime = time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)

func durableBackgroundFixture(t *testing.T, path, phase string) Frozen {
	t.Helper()
	ctx := context.Background()
	log, e := researchledger.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer log.Close()
	b, e := NewBackground(1, 42, 256)
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	batches := 1
	if strings.HasPrefix(phase, "warm-") {
		batches = 2
	}
	for batch := 0; batch < batches; batch++ {
		func() {
			b.adapter.mu.Lock()
			defer b.adapter.mu.Unlock()
			for i := batch * 64; i < (batch+1)*64; i++ {
				p, e := b.Predict(uint16(i), .6, 1, durableFixtureTime)
				if e != nil {
					t.Fatal(e)
				}
				r, e := b.Record(p.ID)
				if e != nil {
					t.Fatal(e)
				}
				if r.Ready != (batch > 0) {
					t.Fatal("fixture readiness mismatch", batch, r.Ready)
				}
				raw, e := json.Marshal(r)
				if e != nil {
					t.Fatal(e)
				}
				key := researchledger.Key{Tenant: "fixture", Journal: "journal", Event: fmt.Sprint(p.ID), Contract: RecordContract}
				if _, _, e = log.Append(ctx, key, "admit", raw); e != nil {
					t.Fatal(e)
				}
				f := durableFixtureFeedback{p.ID, i%3 != 0, durableFixtureTime}
				raw, e = json.Marshal(f)
				if e != nil {
					t.Fatal(e)
				}
				if _, _, e = log.Append(ctx, key, "feedback", raw); e != nil {
					t.Fatal(e)
				}
				if e = b.Feedback(p.ID, f.Useful, 1, f.Available); e != nil {
					t.Fatal(e)
				}
			}
			if batch == batches-1 && (phase == "queued" || phase == "warm-queued") {
				os.Exit(24)
			}
		}()
		wait, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if e = b.WaitProcessed(wait, uint64((batch+1)*64)); e != nil {
			t.Fatal(e)
		}
		if n, f, _, _ := b.Counts(); n != uint64((batch+1)*64) || f != 0 {
			t.Fatal(n, f)
		}
		if batch == batches-1 && (phase == "applied" || phase == "warm-applied") {
			os.Exit(24)
		}
	}
	return b.Snapshot()
}

func TestDurableBackgroundExitHelper(t *testing.T) {
	path := os.Getenv("EVENTFRAME_BACKGROUND_CHILD_PATH")
	if path == "" {
		t.Skip("subprocess only")
	}
	phase := os.Getenv("EVENTFRAME_BACKGROUND_CHILD_PHASE")
	if phase != "queued" && phase != "applied" && phase != "warm-queued" && phase != "warm-applied" {
		t.Fatal("invalid phase")
	}
	durableBackgroundFixture(t, path, phase)
	t.Fatal("child failed to terminate")
}

func TestDurableBackgroundRecoveryParity(t *testing.T) {
	cold := durableBackgroundFixture(t, filepath.Join(t.TempDir(), "control.sqlite"), "")
	warm := durableBackgroundFixture(t, filepath.Join(t.TempDir(), "control.sqlite"), "warm-control")
	for _, phase := range []string{"queued", "applied", "warm-queued", "warm-applied"} {
		t.Run(phase, func(t *testing.T) {
			control, labels := cold, 64
			if strings.HasPrefix(phase, "warm-") {
				control, labels = warm, 128
			}
			path := filepath.Join(t.TempDir(), "recovery.sqlite")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDurableBackgroundExitHelper$")
			cmd.Env = append(os.Environ(), "EVENTFRAME_BACKGROUND_CHILD_PATH="+path, "EVENTFRAME_BACKGROUND_CHILD_PHASE="+phase)
			output, e := cmd.CombinedOutput()
			var exit *exec.ExitError
			if !errors.As(e, &exit) || exit.ExitCode() != 24 {
				t.Fatalf("child boundary not reached: %v %s", e, output)
			}
			log, e := researchledger.Open(path)
			if e != nil {
				t.Fatal(e)
			}
			defer log.Close()
			restored := New(1, 42)
			var after int64
			for {
				rows, e := log.ReadAfter(ctx, after, 17)
				if e != nil {
					t.Fatal(e)
				}
				if len(rows) == 0 {
					break
				}
				for _, r := range rows {
					if r.Kind == "admit" {
						var p RecordedPrediction
						if e = json.Unmarshal(r.Payload, &p); e != nil {
							t.Fatal(e)
						}
						e = restored.RestorePrediction(p)
					} else {
						var f durableFixtureFeedback
						if e = json.Unmarshal(r.Payload, &f); e != nil {
							t.Fatal(e)
						}
						e = restored.Feedback(f.ID, f.Useful, 1, f.Available)
					}
					if e != nil {
						t.Fatal(e)
					}
					// Retry the stored bytes as an acknowledgment-loss control. They must not
					// create a second log entry or trigger another model update.
					seq, duplicate, e := log.Append(ctx, r.Key, r.Kind, r.Payload)
					if e != nil || !duplicate || seq != r.Sequence {
						t.Fatal(seq, duplicate, e)
					}
					after = r.Sequence
				}
			}
			if n, p := restored.Counts(); n != labels || p != 0 || after != int64(2*labels) {
				t.Fatal(n, p, after)
			}
			apiRestored, e := ReplayLedger(ctx, log, "fixture", "journal", 1, 42)
			if e != nil {
				t.Fatal(e)
			}
			for x := uint16(0); x < 512; x++ {
				want, e := control.Score(x, .6, 1, durableFixtureTime)
				got, f := restored.Freeze().Score(x, .6, 1, durableFixtureTime)
				if e != nil || f != nil || want != got {
					t.Fatal("recovery differs", x, want, got, e, f)
				}
				apiGot, apiErr := apiRestored.Freeze().Score(x, .6, 1, durableFixtureTime)
				if apiErr != nil || apiGot != want {
					t.Fatal("replay API differs", x, apiGot, want, apiErr)
				}
			}
		})
	}
}
