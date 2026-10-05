package researchmemory

import (
	"context"
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"path/filepath"
	"testing"
	"time"
)

func TestReplayBoundary(t *testing.T) {
	now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	a := New(1, 42)
	p, e := a.Predict(3, .6, 1, now)
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Record(p.ID)
	if e != nil {
		t.Fatal(e)
	}
	admit, _ := json.Marshal(r)
	feedback, _ := json.Marshal(RecordedFeedback{1, true, now})
	for _, kind := range []string{"valid", "tenant", "stream", "contract", "binding", "missing-label", "unknown-field", "early", "duplicate-json"} {
		t.Run(kind, func(t *testing.T) {
			log, e := researchledger.Open(filepath.Join(t.TempDir(), "replay.sqlite"))
			if e != nil {
				t.Fatal(e)
			}
			defer log.Close()
			key := researchledger.Key{Tenant: "tenant", Journal: "stream", Event: "1", Contract: RecordContract}
			out := append([]byte(nil), feedback...)
			switch kind {
			case "tenant":
				key.Tenant = "other"
			case "stream":
				key.Journal = "other"
			case "contract":
				key.Contract = "other"
			case "binding":
				key.Event = "2"
			case "missing-label":
				out = []byte(`{"ID":1,"Available":"2026-09-12T00:00:00Z"}`)
			case "unknown-field":
				out = []byte(`{"ID":1,"Useful":true,"Available":"2026-09-12T00:00:00Z","extra":1}`)
			case "early":
				out, _ = json.Marshal(RecordedFeedback{1, true, now.Add(-time.Second)})
			case "duplicate-json":
				out = []byte(`{"ID":1,"Useful":false,"Useful":true,"Available":"2026-09-12T00:00:00Z"}`)
			}
			if _, _, e = log.Append(context.Background(), key, "admit", admit); e != nil {
				t.Fatal(e)
			}
			if _, _, e = log.Append(context.Background(), key, "feedback", out); e != nil {
				t.Fatal(e)
			}
			got, e := ReplayLedger(context.Background(), log, "tenant", "stream", 1, 42)
			if kind == "valid" {
				if e != nil {
					t.Fatal(e)
				}
				if n, p := got.Counts(); n != 1 || p != 0 {
					t.Fatal(n, p)
				}
			} else if e == nil || got != nil {
				t.Fatal("invalid replay exposed model", e)
			}
		})
	}
}
