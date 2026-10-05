package researchledger

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestMaterializedBatchCapsAndCancellation(t *testing.T) {
	l, err := Open(t.TempDir() + "/batch.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err := materializedSourceSchema(l); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	k := sourceFixture()
	bad := k
	bad.Event = string([]byte{255})
	for _, req := range [][]ServiceIdentity{nil, make([]ServiceIdentity, 513), {ServiceIdentity{}}, {bad}} {
		if r, e := materializedSourceBatch(ctx, l, req, nil); e == nil || r != nil {
			t.Fatal("invalid input", e)
		}
	}
	full := make([]ServiceIdentity, 512)
	for i := range full {
		full[i] = k
	}
	if r, e := materializedSourceBatch(ctx, l, full, nil); e != nil || len(r) != 512 {
		t.Fatal("count limit", e)
	}
	field := strings.Repeat("\x00", 4096)
	for i := range full {
		full[i] = ServiceIdentity{field, field, field, field, field}
	}
	if r, e := materializedSourceBatch(ctx, l, full, nil); e == nil || r != nil || !strings.Contains(e.Error(), "request byte cap") {
		t.Fatal("request cap", e)
	}
	entry := serviceEntry("1", k)
	// Valid bound JSON exactly at the individual materialization cap.
	var value map[string]any
	if err := json.Unmarshal(entry.Payload, &value); err != nil {
		t.Fatal(err)
	}
	value["padding"] = ""
	raw, _ := json.Marshal(value)
	value["padding"] = strings.Repeat("x", (1<<20)-len(raw))
	entry.Payload, _ = json.Marshal(value)
	if len(entry.Payload) != 1<<20 {
		t.Fatal("fixture size")
	}
	if _, err := materializedSourceAppend(ctx, l, []AppendRequest{entry}, nil); err != nil {
		t.Fatal(err)
	}
	req := make([]ServiceIdentity, 8)
	for i := range req {
		req[i] = k
	}
	if r, e := materializedSourceBatch(ctx, l, req, nil); e != nil || len(r) != 8 {
		t.Fatal("exact payload cap", e)
	}
	missing := k
	missing.Event = "absent"
	if _, e := materializedSourceBatch(ctx, l, append(req, missing), nil); e != nil {
		t.Fatal("zero-budget miss", e)
	}
	if r, e := materializedSourceBatch(ctx, l, append(req, k), nil); e == nil || r != nil {
		t.Fatal("payload overflow", e)
	}
	c, cancel := context.WithCancel(ctx)
	if r, e := materializedSourceBatch(c, l, []ServiceIdentity{k, k}, func(int) { cancel() }); !errors.Is(e, context.Canceled) || r != nil {
		t.Fatal("partial canceled batch", e)
	}
	if r, e := materializedSourceBatch(c, l, []ServiceIdentity{k}, nil); !errors.Is(e, context.Canceled) || r != nil {
		t.Fatal("precancel", e)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("missing panic")
			}
		}()
		_, _ = materializedSourceBatch(ctx, l, []ServiceIdentity{k}, func(int) { panic("read boundary") })
	}()
	if _, e := materializedSourceBatch(ctx, l, []ServiceIdentity{k}, nil); e != nil {
		t.Fatal("transaction leak", e)
	}
}
