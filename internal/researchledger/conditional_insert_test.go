package researchledger

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"
)

func TestConditionalInsertParity(t *testing.T) {
	ctx := context.Background()
	old, e := Open(t.TempDir() + "/old.sqlite")
	if e != nil {
		t.Fatal(e)
	}
	defer old.Close()
	path := t.TempDir() + "/new.sqlite"
	new, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer func() { new.Close() }()
	a := batchFixture()
	late := AppendRequest{Key{"tenant", "stream", "3", "v1"}, "admit", json.RawMessage(`true`)}
	for _, requests := range [][]AppendRequest{a, a, {a[0], late, late}} {
		x, e := old.AppendBatchPrepared(ctx, requests)
		if e != nil {
			t.Fatal(e)
		}
		y, e := new.AppendBatchConditionalInsert(ctx, requests)
		if e != nil || !reflect.DeepEqual(x, y) {
			t.Fatal("receipt parity", x, y, e)
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := new.AppendBatchConditionalInsert(ctx, a)
			if e != nil {
				t.Error(e)
				return
			}
			for i, v := range r {
				if !v.Retry || v.Sequence != int64(i+1) {
					t.Error("retry receipt", v)
				}
			}
		}()
	}
	wg.Wait()
	want, e := old.ReadAfter(ctx, 0, 256)
	if e != nil {
		t.Fatal(e)
	}
	if e = new.Close(); e != nil {
		t.Fatal(e)
	}
	new, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	got, e := new.ReadAfter(ctx, 0, 256)
	if e != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("reopen parity", e)
	}
}

func TestConditionalInsertRollback(t *testing.T) {
	for _, fault := range []string{"late-conflict", "feedback-first", "cancel", "panic"} {
		t.Run(fault, func(t *testing.T) {
			l, e := Open(t.TempDir() + "/fault.sqlite")
			if e != nil {
				t.Fatal(e)
			}
			defer l.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			r := batchFixture()
			var hook func()
			switch fault {
			case "late-conflict":
				bad := r[0]
				bad.Payload = json.RawMessage(`false`)
				r = append(r, bad)
			case "feedback-first":
				r[0], r[1] = r[1], r[0]
			case "cancel":
				hook = cancel
			case "panic":
				hook = func() { panic("precommit") }
			}
			if fault == "panic" {
				func() {
					defer func() {
						if recover() == nil {
							t.Error("missing panic")
						}
					}()
					_, _ = l.appendBatchStrategy(ctx, r, true, true, hook)
				}()
			} else {
				ack, e := l.appendBatchStrategy(ctx, r, true, true, hook)
				if e == nil || ack != nil {
					t.Fatal("bad acknowledgment", ack, e)
				}
			}
			rows, e := l.ReadAfter(context.Background(), 0, 256)
			if e != nil || len(rows) != 0 {
				t.Fatal("partial commit", rows, e)
			}
		})
	}
}
