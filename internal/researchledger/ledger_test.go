package researchledger

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
)

func TestDurableIdentityAndReplay(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "ledger.sqlite")
	l, e := Open(path)
	if e != nil {
		t.Fatal(e)
	}
	key := Key{"tenant", "journal", "event", "contract-v1"}
	p := json.RawMessage(`{"original_experts":[0.2,0.7],"epoch":1}`)
	if _, _, e = l.Append(ctx, key, "feedback", json.RawMessage(`true`)); e == nil {
		t.Fatal("unbound feedback accepted")
	}
	seq, dup, e := l.Append(ctx, key, "admit", p)
	if e != nil || dup || seq != 1 {
		t.Fatal(seq, dup, e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, d, e := l.Append(ctx, key, "admit", p)
			if e != nil || !d || s != seq {
				t.Error(s, d, e)
			}
		}()
	}
	wg.Wait()
	if _, _, e = l.Append(ctx, key, "admit", json.RawMessage(`{}`)); e == nil {
		t.Fatal("conflict accepted")
	}
	if _, _, e = l.Append(ctx, key, "feedback", json.RawMessage(`{"useful":true}`)); e != nil {
		t.Fatal(e)
	}
	if e = l.Close(); e != nil {
		t.Fatal(e)
	}
	l, e = Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	rows, e := l.ReadAfter(ctx, 0, 256)
	if e != nil || len(rows) != 2 || string(rows[0].Payload) != string(p) || rows[1].Kind != "feedback" {
		t.Fatal(rows, e)
	}
	if s, d, e := l.Append(ctx, key, "feedback", json.RawMessage(`{"useful":true}`)); e != nil || !d || s != 2 {
		t.Fatal(s, d, e)
	}
	if _, _, e = l.Append(ctx, key, "feedback", json.RawMessage(`{"useful":false}`)); e == nil {
		t.Fatal("conflicting outcome accepted")
	}
	other := key
	other.Contract = "contract-v2"
	if _, _, e = l.Append(ctx, other, "feedback", json.RawMessage(`true`)); e == nil {
		t.Fatal("cross-contract admission inherited")
	}
	page, e := l.ReadAfter(ctx, 1, 1)
	if e != nil || len(page) != 1 || page[0].Sequence != 2 {
		t.Fatal(page, e)
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if _, _, e = l.Append(c, key, "admit", p); e == nil {
		t.Fatal("cancel ignored")
	}
}
