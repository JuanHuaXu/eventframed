package researchledger

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestBulkAdmissionParity(t *testing.T) {
	ctx := context.Background()
	a, e := Open(t.TempDir() + "/a.sqlite")
	if e != nil {
		t.Fatal(e)
	}
	defer a.Close()
	path := t.TempDir() + "/b.sqlite"
	b, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer func() { b.Close() }()
	r := make([]AppendRequest, MaxBatchEntries)
	for i := range r {
		r[i] = AppendRequest{Key{"t", "s", fmt.Sprint(i), "v"}, "admit", json.RawMessage(`true`)}
	}
	for _, batch := range [][]AppendRequest{r, r, {r[0], {Key{"t", "s", "next", "v"}, "admit", json.RawMessage(`false`)}, {Key{"t", "s", "next", "v"}, "admit", json.RawMessage(`false`)}}, batchFixture()} {
		x, e := a.AppendBatchPrepared(ctx, batch)
		if e != nil {
			t.Fatal(e)
		}
		y, e := b.AppendBatchBulkAdmissions(ctx, batch)
		if e != nil || !reflect.DeepEqual(x, y) {
			t.Fatal("bulk receipt parity", e)
		}
	}
	if e = b.Close(); e != nil {
		t.Fatal(e)
	}
	b, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	var after int64
	for {
		x, e := a.ReadAfter(ctx, after, 256)
		if e != nil {
			t.Fatal(e)
		}
		y, e := b.ReadAfter(ctx, after, 256)
		if e != nil || !reflect.DeepEqual(x, y) {
			t.Fatal("replay parity", e)
		}
		if len(x) == 0 {
			break
		}
		after = x[len(x)-1].Sequence
	}
}

func TestBulkAdmissionRollback(t *testing.T) {
	for _, fault := range []string{"source-conflict", "duplicate", "cancel", "panic", "feedback-first"} {
		t.Run(fault, func(t *testing.T) {
			l, e := Open(t.TempDir() + "/f.sqlite")
			if e != nil {
				t.Fatal(e)
			}
			defer l.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if e = l.EnableServiceIdentity(ctx); e != nil {
				t.Fatal(e)
			}
			r := []AppendRequest{serviceEntry("1", sourceFixture())}
			var hook func()
			switch fault {
			case "source-conflict":
				for i := 2; i <= 130; i++ {
					k := sourceFixture()
					k.Event = fmt.Sprint(i)
					r = append(r, serviceEntry(fmt.Sprint(i), k))
				}
				r = append(r, serviceEntry("131", sourceFixture()))
			case "duplicate":
				bad := r[0]
				bad.Payload = json.RawMessage(`false`)
				r = append(r, bad)
			case "cancel":
				hook = cancel
			case "panic":
				hook = func() { panic("precommit") }
			case "feedback-first":
				r = batchFixture()
				r[0], r[1] = r[1], r[0]
			}
			if fault == "panic" {
				func() {
					defer func() {
						if recover() == nil {
							t.Error("missing panic")
						}
					}()
					_, _ = l.appendBatchPlan(ctx, r, true, false, true, hook)
				}()
			} else {
				ack, e := l.appendBatchPlan(ctx, r, true, false, true, hook)
				if e == nil || ack != nil {
					t.Fatal("bad ack", ack, e)
				}
			}
			rows, e := l.ReadAfter(context.Background(), 0, 256)
			if e != nil || len(rows) != 0 {
				t.Fatal("partial commit", e)
			}
		})
	}
}
