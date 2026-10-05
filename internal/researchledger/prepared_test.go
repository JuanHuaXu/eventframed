package researchledger

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestPreparedBatchParity(t *testing.T) {
	ctx := context.Background()
	var logs []*Ledger
	for i := 0; i < 2; i++ {
		l, err := Open(t.TempDir() + "/parity.sqlite")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		logs = append(logs, l)
	}
	requests := batchFixture()
	for round := 0; round < 2; round++ {
		a, err := logs[0].AppendBatch(ctx, requests)
		if err != nil {
			t.Fatal(err)
		}
		b, err := logs[1].AppendBatchPrepared(ctx, requests)
		if err != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("ack parity", err, a, b)
		}
	}
	for _, l := range logs {
		late := AppendRequest{Key{"tenant", "stream", "new", "v1"}, "admit", json.RawMessage(`null`)}
		bad := requests[3]
		bad.Payload = json.RawMessage(`false`)
		if got, err := l.AppendBatchPrepared(ctx, []AppendRequest{late, bad}); err == nil || got != nil {
			t.Fatal("conflict accepted")
		}
	}
	a, err := logs[0].ReadAfter(ctx, 0, 256)
	if err != nil {
		t.Fatal(err)
	}
	b, err := logs[1].ReadAfter(ctx, 0, 256)
	if err != nil || len(b) != 4 || !reflect.DeepEqual(a, b) {
		t.Fatal("replay/rollback parity", err)
	}
}

func TestPreparedBatchFailureRelease(t *testing.T) {
	for _, kind := range []string{"order", "cancel", "panic", "prepare"} {
		t.Run(kind, func(t *testing.T) {
			l, err := Open(t.TempDir() + "/failure.sqlite")
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			requests := batchFixture()
			var hook func()
			switch kind {
			case "order":
				requests[0], requests[1] = requests[1], requests[0]
			case "cancel":
				hook = cancel
			case "panic":
				hook = func() { panic("precommit") }
			case "prepare":
				if _, err = l.db.Exec("ALTER TABLE research_log RENAME TO hidden_log"); err != nil {
					t.Fatal(err)
				}
			}
			panicked := false
			func() {
				defer func() {
					if recover() != nil {
						panicked = true
					}
				}()
				got, e := l.appendBatchMode(ctx, requests, true, hook)
				err = e
				if err == nil || got != nil {
					t.Error("failure acknowledged")
				}
			}()
			if panicked != (kind == "panic") {
				t.Fatal("panic mismatch")
			}
			if kind == "cancel" && !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			if kind == "prepare" {
				if _, err = l.db.Exec("ALTER TABLE hidden_log RENAME TO research_log"); err != nil {
					t.Fatal(err)
				}
			}
			rows, err := l.ReadAfter(context.Background(), 0, 256)
			if err != nil || len(rows) != 0 {
				t.Fatal("partial write", err)
			}
			if _, err = l.AppendBatchPrepared(context.Background(), batchFixture()); err != nil {
				t.Fatal("transaction/statement retained", err)
			}
		})
	}
}
