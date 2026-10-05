package libravdbstore

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type archiveRequestResultV32 struct {
	Index   int
	Request model.RecallRequest
	Packet  model.ContextPacket
	Error   string
}
type archiveRequestTrialV32 struct {
	Mode               string
	Captures           []model.BayesianJournalEntry
	Responses          []archiveRequestResultV32
	Unique, Duplicates int
	ElapsedNS          int64
}

func TestResearchArchiveRequestsV32(t *testing.T) {
	output := os.Getenv("EVENTFRAME_ARCHIVE_REQUESTS_V32_OUTPUT")
	if output == "" {
		t.Skip("exclusive request-identity diagnostic")
	}
	const n = 16
	ctx := context.Background()
	var trials []archiveRequestTrialV32
	for _, mode := range []string{"native-clock", "distinct-asof", "identical-retry"} {
		t.Run(mode, func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachArchiveV31(t, f)
			defer s.Close()
			prime, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if err != nil {
				t.Fatal(err)
			}
			captures := make(chan *archiveCaptureV29, n)
			resume := make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(resume) }) }
			defer unblock()
			s.afterCapture = func(c *archiveCaptureV29) { captures <- c; <-resume }
			done := make(chan archiveRequestResultV32, n)
			remaining := n
			defer func() {
				unblock()
				for remaining > 0 {
					select {
					case <-done:
						remaining--
					case <-time.After(5 * time.Second):
						t.Error("live archive did not terminate")
						return
					}
				}
			}()
			base := time.Now().UTC().Add(-time.Millisecond)
			started := time.Now()
			for i := 0; i < n; i++ {
				go func(i int) {
					at := time.Now().UTC()
					if mode == "distinct-asof" {
						at = base.Add(time.Duration(i) * time.Nanosecond)
					}
					if mode == "identical-retry" {
						at = base
					}
					req := f.request(at, "request")
					p, _, e := s.recall(ctx, f.svc, req)
					row := archiveRequestResultV32{Index: i, Request: req, Packet: p}
					if e != nil {
						row.Error = e.Error()
					}
					done <- row
				}(i)
			}
			row := archiveRequestTrialV32{Mode: mode}
			var owned []*archiveCaptureV29
			for len(owned) < n {
				select {
				case c := <-captures:
					owned = append(owned, c)
					row.Captures = append(row.Captures, c.entry)
				case <-time.After(5 * time.Second):
					t.Fatal("captures stalled")
				}
			}
			if stats := s.scheduler.Snapshot(); stats.ActiveReaders != 0 || stats.ActiveWriter {
				t.Fatal("capture held admission")
			}
			mutation := make(chan error, 1)
			go func() {
				for _, id := range []string{"first", "second"} {
					if _, e := f.feedback(ctx, prime.BayesianShadow.JournalID, "past100", id, false); e != nil {
						mutation <- e
						return
					}
				}
				w := pinnedWrite("visible-request-control", f.origin)
				w.Vector = denseRowV6(f.query, 7, .0003)
				_, e := s.append(ctx, []ResearchEventWrite{w}, false)
				mutation <- e
			}()
			select {
			case e := <-mutation:
				if e != nil {
					t.Fatal(e)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("mutation blocked")
			}
			select {
			case <-done:
				remaining--
				t.Fatal("early ack")
			default:
			}
			unblock()
			for remaining > 0 {
				select {
				case r := <-done:
					remaining--
					row.Responses = append(row.Responses, r)
				case <-time.After(5 * time.Second):
					t.Fatal("drain stalled")
				}
			}
			row.ElapsedNS = time.Since(started).Nanoseconds()
			wires := map[string]string{}
			for _, c := range owned {
				if old, ok := wires[c.entry.ID]; ok && old != c.wire {
					t.Fatal("same journal ID/different wire")
				}
				wires[c.entry.ID] = c.wire
				got, e := s.gate.store.GetBayesianJournal(ctx, "tenant-a", c.entry.ID)
				if e != nil || witnessHashV23(got) != c.wire {
					t.Fatal("lost/changed archive", e)
				}
			}
			row.Unique = len(wires)
			row.Duplicates = n - row.Unique
			for _, r := range row.Responses {
				if r.Error != "" {
					t.Error(r.Error)
					continue
				}
				id := r.Packet.BayesianShadow.JournalID
				if id == "" || wires[id] == "" {
					t.Error("unbound packet", r.Index, id)
					continue
				}
				got, e := s.gate.store.GetBayesianJournal(ctx, "tenant-a", id)
				if e != nil {
					t.Error(e)
					continue
				}
				if got.AsOf != r.Request.AsOf || got.Snapshot != r.Packet.Snapshot || witnessHashV23(got.Report) != witnessHashV23(r.Packet.BayesianShadow) {
					t.Error("packet/request/capture mismatch", r.Index)
				}
			}
			if mode == "distinct-asof" && row.Unique != n {
				t.Error("distinct requests collapsed")
			}
			if mode == "identical-retry" && row.Unique != 1 {
				t.Error("identical retry did not deduplicate")
			}
			if len(s.state.Journals) != row.Unique+1 {
				t.Error("wrong durable distinct count")
			}
			f.reopen(t)
			if f.adapter.poison.Load() {
				t.Error("request-identity chain failed reopen")
			}
			trials = append(trials, row)
			t.Logf("mode=%s submitted=%d unique=%d duplicates=%d elapsed_ms=%.3f", mode, n, row.Unique, row.Duplicates, float64(row.ElapsedNS)/1e6)
		})
	}
	// Preserve diagnostics even if an invariant failed; runner code stays nonzero.
	file, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(map[string]any{"Type": "request-identity-diagnostic", "Trials": trials, "LoadedLatencyProof": false}); err != nil {
		t.Fatal(err)
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
}
