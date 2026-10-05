package researchledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func batchFixture() []AppendRequest {
	a, b := Key{"tenant", "stream", "1", "v1"}, Key{"tenant", "stream", "2", "v1"}
	return []AppendRequest{{a, "admit", json.RawMessage(`{"p":0.2}`)}, {a, "feedback", json.RawMessage(`{"discard":true}`)}, {b, "admit", json.RawMessage(`{"p":0.8}`)}, {b, "feedback", json.RawMessage(`{"discard":true}`)}}
}

func TestBatchSequentialParityAndRetries(t *testing.T) {
	ctx := context.Background()
	l, err := Open(t.TempDir() + "/batch.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	single, err := Open(t.TempDir() + "/single.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer single.Close()
	requests := batchFixture()
	got, err := l.AppendBatch(ctx, requests)
	if err != nil {
		t.Fatal(err)
	}
	for i, request := range requests {
		seq, retry, err := single.Append(ctx, request.Key, request.Kind, request.Payload)
		if err != nil || got[i] != (AppendResult{seq, retry}) {
			t.Fatal("single parity", i, err, got)
		}
	}
	rows, err := l.ReadAfter(ctx, 0, 256)
	if err != nil {
		t.Fatal(err)
	}
	want, err := single.ReadAfter(ctx, 0, 256)
	if err != nil || !reflect.DeepEqual(rows, want) {
		t.Fatal("replay parity", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := l.AppendBatch(ctx, requests)
			if err != nil {
				t.Error(err)
				return
			}
			for i, r := range results {
				if !r.Retry || r.Sequence != int64(i+1) {
					t.Error("retry changed sequence", r)
				}
			}
		}()
	}
	wg.Wait()
	late := AppendRequest{Key{"tenant", "stream", "3", "v1"}, "admit", json.RawMessage(`null`)}
	conflict := requests[3]
	conflict.Payload = json.RawMessage(`false`)
	if results, err := l.AppendBatch(ctx, []AppendRequest{late, conflict}); err == nil || results != nil {
		t.Fatal("late conflict acknowledged", results, err)
	}
	after, err := l.ReadAfter(ctx, 0, 256)
	if err != nil || !reflect.DeepEqual(after, rows) {
		t.Fatal("late conflict leaked insert", err)
	}
	results, err := l.AppendBatch(ctx, []AppendRequest{requests[0], late})
	if err != nil || results[0] != (AppendResult{1, true}) || results[1] != (AppendResult{5, false}) {
		t.Fatal("mixed retry/new", results, err)
	}
}

func TestBatchRejectionAndCancellation(t *testing.T) {
	for _, kind := range []string{"empty", "count", "bytes", "payload", "identity", "order", "cancel-before", "cancel-commit"} {
		t.Run(kind, func(t *testing.T) {
			l, err := Open(t.TempDir() + "/reject.sqlite")
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			requests := batchFixture()
			var hook func()
			switch kind {
			case "empty":
				requests = nil
			case "count":
				requests = make([]AppendRequest, MaxBatchEntries+1)
			case "bytes":
				requests = nil
				for i := 0; i < 9; i++ {
					requests = append(requests, AppendRequest{Key{"tenant", "stream", fmt.Sprint(i), "v1"}, "admit", json.RawMessage(`"` + strings.Repeat("x", (1<<20)-2) + `"`)})
				}
			case "payload":
				requests[3].Payload = json.RawMessage(`{`)
			case "identity":
				requests[3].Key.Tenant = ""
			case "order":
				requests[0], requests[1] = requests[1], requests[0]
			case "cancel-before":
				cancel()
			case "cancel-commit":
				hook = cancel
			}
			results, err := l.appendBatch(ctx, requests, hook)
			if err == nil || results != nil {
				t.Fatal("invalid batch acknowledged", results, err)
			}
			if strings.HasPrefix(kind, "cancel") && !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
			rows, err := l.ReadAfter(context.Background(), 0, 256)
			if err != nil || len(rows) != 0 {
				t.Fatal("rejection left partial rows", rows, err)
			}
		})
	}
}

func TestBatchExactCapsAndSameBatchRetry(t *testing.T) {
	l, err := Open(t.TempDir() + "/caps.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	ctx := context.Background()
	requests := make([]AppendRequest, MaxBatchEntries)
	for i := range requests {
		requests[i] = AppendRequest{Key{"tenant", "caps", fmt.Sprint(i), "v1"}, "admit", json.RawMessage(`null`)}
	}
	results, err := l.AppendBatch(ctx, requests)
	if err != nil || len(results) != MaxBatchEntries {
		t.Fatal("count boundary rejected", err)
	}
	dup := AppendRequest{Key{"tenant", "dup", "1", "v1"}, "admit", json.RawMessage(`true`)}
	results, err = l.AppendBatch(ctx, []AppendRequest{dup, dup})
	if err != nil || results[0].Retry || !results[1].Retry || results[0].Sequence != results[1].Sequence {
		t.Fatal("within-batch retry", results, err)
	}
	large := make([]AppendRequest, 8)
	overhead := 0
	for i := range large {
		key := Key{"tenant", "bytes", fmt.Sprint(i), "v1"}
		encoded, _ := json.Marshal(key)
		overhead += len(encoded) + len("admit")
		large[i] = AppendRequest{key, "admit", json.RawMessage(`"` + strings.Repeat("x", (1<<20)-2) + `"`)}
	}
	large[7].Payload = json.RawMessage(`"` + strings.Repeat("x", (1<<20)-2-overhead) + `"`)
	if _, err = l.AppendBatch(ctx, large); err != nil {
		t.Fatal("exact byte cap rejected", err)
	}
	large[7].Payload = json.RawMessage(`"` + strings.Repeat("x", (1<<20)-1-overhead) + `"`)
	if results, err = l.AppendBatch(ctx, large); err == nil || results != nil || !strings.Contains(err.Error(), "byte cap") {
		t.Fatal("one byte excess accepted", err)
	}
}

func TestBatchPanicRollsBack(t *testing.T) {
	l, err := Open(t.TempDir() + "/panic.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	func() {
		defer func() {
			if recover() == nil {
				t.Error("fault hook did not panic")
			}
		}()
		_, _ = l.appendBatch(context.Background(), batchFixture(), func() { panic("before commit") })
	}()
	rows, err := l.ReadAfter(context.Background(), 0, 256)
	if err != nil || len(rows) != 0 {
		t.Fatal("panic leaked transaction", rows, err)
	}
	if _, err = l.AppendBatch(context.Background(), batchFixture()); err != nil {
		t.Fatal("panic retained transaction", err)
	}
}
