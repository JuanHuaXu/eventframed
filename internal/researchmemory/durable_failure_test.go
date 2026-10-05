package researchmemory

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

var errInjectedWrite = errors.New("injected uncertain write")

type failingDurableLog struct {
	durableLog
	commit bool
}

func (f failingDurableLog) Append(ctx context.Context, key researchledger.Key, kind string, payload json.RawMessage) (int64, bool, error) {
	if f.commit {
		if _, _, e := f.durableLog.Append(ctx, key, kind, payload); e != nil {
			return 0, false, e
		}
	}
	return 0, false, errInjectedWrite
}

func TestUncertainDurableWriteStopsUntilReplay(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	for _, kind := range []string{"admit", "feedback", "discard"} {
		for _, commit := range []bool{false, true} {
			name := "before"
			if commit {
				name = "after"
			}
			t.Run(kind+"/"+name, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "failure.sqlite")
				d, e := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
				if e != nil {
					t.Fatal(e)
				}
				if kind != "admit" {
					if _, _, e = d.Admit(ctx, 1, 1, .6, now); e != nil {
						t.Fatal(e)
					}
				}
				d.log = failingDurableLog{d.log, commit}
				if kind == "admit" {
					_, _, e = d.Admit(ctx, 1, 1, .6, now)
				} else if kind == "discard" {
					_, e = d.Discard(ctx, 1, now)
				} else {
					_, e = d.Feedback(ctx, 1, true, now)
				}
				if !errors.Is(e, errInjectedWrite) || !d.stopped {
					t.Fatal("uncertainty not quarantined", e, d.stopped)
				}
				if _, _, e = d.Admit(ctx, 2, 2, .6, now); e == nil {
					t.Fatal("continued after uncertain write")
				}
				if _, e = d.Feedback(ctx, 1, true, now); e == nil {
					t.Fatal("feedback continued after uncertainty")
				}
				if _, e = d.Discard(ctx, 1, now); e == nil {
					t.Fatal("discard continued after uncertainty")
				}
				if e = d.Close(); e != nil {
					t.Fatal(e)
				}
				d, e = OpenDurable(ctx, path, "tenant", "stream", 1, 42)
				if e != nil {
					t.Fatal(e)
				}
				defer d.Close()
				n, f, p, q := d.worker.Counts()
				wantN, wantP := uint64(0), 0
				if kind == "feedback" {
					if commit {
						wantN = 1
					} else {
						wantP = 1
					}
				} else if kind == "discard" {
					if !commit {
						wantP = 1
					}
				} else if commit {
					wantP = 1
				}
				if n != wantN || p != wantP || f != 0 || q != 0 {
					t.Fatal("wrong recovered state", n, f, p, q)
				}
				if kind == "admit" {
					_, duplicate, e := d.Admit(ctx, 1, 1, .6, now)
					if e != nil || duplicate != commit {
						t.Fatal("admission retry", duplicate, e)
					}
				} else if kind == "discard" {
					duplicate, e := d.Discard(ctx, 1, now)
					if e != nil || duplicate != commit {
						t.Fatal("discard retry", duplicate, e)
					}
					if n, f, p, q := d.worker.Counts(); n != 0 || f != 0 || p != 0 || q != 0 {
						t.Fatal("discard changed learning", n, f, p, q)
					}
					if _, e = d.Feedback(ctx, 1, true, now); e == nil {
						t.Fatal("recovered discard accepted label")
					}
				} else {
					duplicate, e := d.Feedback(ctx, 1, true, now)
					if e != nil || duplicate != commit {
						t.Fatal("feedback retry", duplicate, e)
					}
					wait, cancel := context.WithTimeout(ctx, time.Second)
					defer cancel()
					if e = d.worker.WaitProcessed(wait, 1); e != nil {
						t.Fatal(e)
					}
					if n, f, _, _ := d.worker.Counts(); n != 1 || f != 0 {
						t.Fatal("duplicate/lost learning", n, f)
					}
				}
			})
		}
	}
}
