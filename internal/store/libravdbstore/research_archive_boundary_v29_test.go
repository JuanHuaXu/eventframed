package libravdbstore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type archiveResultV29 struct {
	Kind                                                     string
	CaptureRoot, WireSHA256                                  string
	JournalWire                                              string
	Captured, Current                                        model.Snapshot
	Journal                                                  model.BayesianJournalEntry
	FullStreamResponse                                       model.BayesianOutcomeResponse
	Packet                                                   model.ContextPacket
	Handoff, MutationBegin, MutationEnd, Ack                 time.Time
	CurrentGateAccepts, FullStreamAccepted, SelectedRejected bool
	PostCalls                                                []string
}

// A barrier deliberately lets acknowledged writes and labels land after all
// forecasts are owned but before their historical journal becomes durable.
func TestResearchArchiveBoundaryV29(t *testing.T) {
	output := os.Getenv("EVENTFRAME_ARCHIVE_BOUNDARY_V29_OUTPUT")
	if output == "" {
		t.Skip("exclusive isolated archival experiment")
	}
	ctx := context.Background()
	var results []archiveResultV29
	for _, kind := range []string{"future", "visible", "outcome", "mixed", "cancel"} {
		t.Run(kind, func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachArchiveV29(t, f)
			defer s.Close()
			o := &handoffObserverV27{EventStore: s}
			svc, err := service.New(o, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
			if err != nil {
				t.Fatal(err)
			}
			prime, _, err := s.recall(ctx, svc, f.request(time.Now().UTC(), "prime"))
			if err != nil {
				t.Fatal(err)
			}
			entered := make(chan *archiveCaptureV29, 1)
			resume := make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(resume) }) }
			s.afterCapture = func(c *archiveCaptureV29) { entered <- c; <-resume }
			callCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			type response struct {
				packet model.ContextPacket
				err    error
			}
			done := make(chan response, 1)
			o.arm()
			go func() { p, _, e := s.recall(callCtx, svc, f.request(time.Now().UTC(), "held")); done <- response{p, e} }()
			defer func() {
				unblock()
				select {
				case <-done:
				case <-time.After(3 * time.Second):
					t.Error("archive call did not terminate")
				}
			}()
			var c *archiveCaptureV29
			select {
			case c = <-entered:
			case <-time.After(3 * time.Second):
				t.Fatal("no owned handoff")
			}
			row := archiveResultV29{Kind: kind, CaptureRoot: c.root, WireSHA256: c.wire, Captured: c.entry.Snapshot, Handoff: time.Now().UTC()}
			if stats := s.scheduler.Snapshot(); stats.ActiveReaders != 0 || stats.ActiveWriter {
				t.Fatal("read lease held after owned handoff", stats)
			}
			select {
			case r := <-done:
				t.Fatal("ack before archival durability", r.err)
			default:
			}
			row.MutationBegin = time.Now().UTC()
			mutation := make(chan error, 1)
			go func() {
				if kind == "outcome" || kind == "mixed" || kind == "cancel" {
					if _, err := f.feedback(ctx, prime.BayesianShadow.JournalID, "past100", "intervening", false); err != nil {
						mutation <- err
						return
					}
				}
				if kind != "outcome" && kind != "cancel" {
					at := f.origin.Add(time.Hour)
					if kind == "visible" || kind == "mixed" {
						at = f.origin
					}
					w := pinnedWrite("after-capture-"+kind, at)
					w.Vector = denseRowV6(f.query, 7, .0003)
					if _, err := s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
						mutation <- err
						return
					}
				}
				mutation <- nil
			}()
			select {
			case err := <-mutation:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("mutation blocked by handed-off recall")
			}
			row.MutationEnd = time.Now().UTC()
			row.Current = s.gate.store.Snapshot(ctx)
			row.CurrentGateAccepts = store.JournalSnapshotCompatible(c.entry.Snapshot, row.Current, c.entry.AsOf, s.gate.store.ingestMotion)
			if row.Current == row.Captured || row.CurrentGateAccepts != (kind == "future") {
				t.Fatal("control did not move intended dependencies", row)
			}
			if kind == "cancel" {
				cancel()
			}
			select {
			case r := <-done:
				t.Fatal("early ack after mutation", r.err)
			case <-time.After(5 * time.Millisecond):
			}
			unblock()
			var r response
			select {
			case r = <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("archive append did not finish")
			}
			// Return a terminal token for the cleanup drain; no live operation remains.
			done <- r
			row.Ack = time.Now().UTC()
			if kind == "cancel" {
				if !errors.Is(r.err, context.Canceled) || r.packet.BayesianShadow.JournalID != "" {
					t.Fatal("canceled caller received packet", r.err)
				}
			} else if r.err != nil {
				t.Fatal(r.err)
			}
			saved, err := s.gate.store.GetBayesianJournal(ctx, "tenant-a", c.entry.ID)
			if err != nil || witnessHashV23(saved) != c.wire {
				t.Fatal("owned wire changed", err)
			}
			row.Journal = saved
			wire, _ := json.Marshal(saved)
			row.JournalWire = string(wire)
			row.Packet = r.packet
			if saved.Snapshot != row.Captured || s.state.Journals[saved.ID].Snapshot != row.Captured || s.state.Head != row.Current {
				t.Fatal("snapshot retag/history fork")
			}
			if kind != "cancel" {
				if r.packet.Snapshot != row.Captured || len(r.packet.Candidates) != 10 || len(saved.Report.Decisions) != 150 {
					t.Fatal("wrong captured packet")
				}
				for _, candidate := range r.packet.Candidates {
					found := false
					for _, decision := range saved.Report.Decisions {
						if candidate.Event.ID == decision.EventID {
							found = true
							if witnessHashV23(candidate.Forecast) != witnessHashV23(decision.Forecast) {
								t.Fatal("packed law changed")
							}
						}
					}
					if !found {
						t.Fatal("packing outside captured frontier")
					}
				}
			}
			o.mu.Lock()
			o.armed = false
			row.PostCalls = append([]string(nil), o.post...)
			o.mu.Unlock()
			if len(row.PostCalls) != 0 {
				t.Fatal("late service-facing read", row.PostCalls)
			}
			s.afterCapture = nil
			// Full-stream feedback may learn from a historical commitment. Selected
			// feedback cannot bypass the existing policy/epoch guard after a visible write.
			feedbackResponse, err := f.feedback(ctx, saved.ID, "past100", "archived-label", true)
			if err != nil {
				t.Fatal("historical full-stream feedback", err)
			}
			row.FullStreamResponse = feedbackResponse
			row.FullStreamAccepted = true
			if kind == "visible" || kind == "mixed" {
				req := f.outcomeRequest(saved.ID, "past100", "selected-stale", true, time.Now().UTC())
				req.Source = model.OutcomeSelected
				_, err := svc.ObserveBayesianOutcome(ctx, req)
				if err == nil {
					t.Fatal("stale selection accepted")
				}
				row.SelectedRejected = true
			}
			// Native/witness reopen validates the ORIGINAL captured binding, not
			// a replacement epoch. The archived byte commitment survives without repair.
			f.reopen(t)
			if f.adapter.poison.Load() {
				t.Fatal("archive chain rejected on reopen")
			}
			reopened, err := f.adapter.gate.store.GetBayesianJournal(ctx, "tenant-a", saved.ID)
			if err != nil || witnessHashV23(reopened) != row.WireSHA256 {
				t.Fatal("reopen changed archive", err)
			}
			results = append(results, row)
		})
	}
	if t.Failed() {
		return
	}
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(map[string]any{"Type": "historical-owned-archive-technical", "Cases": results, "LoadOrQualityProof": false}); err != nil {
		t.Fatal(err)
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
}

func TestResearchArchiveAuthorityV29(t *testing.T) {
	if os.Getenv("EVENTFRAME_ARCHIVE_BOUNDARY_V29_OUTPUT") == "" {
		t.Skip("isolated authority controls")
	}
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachArchiveV29(t, f)
	defer s.Close()
	var captured *archiveCaptureV29
	s.afterCapture = func(c *archiveCaptureV29) { captured = c }
	if _, _, err := s.recall(context.Background(), f.svc, f.request(time.Now().UTC(), "capture")); err != nil {
		t.Fatal(err)
	}
	s.owner.Lock()
	defer s.owner.Unlock()
	for _, name := range []string{"owner", "wire", "snapshot", "binding", "root", "chain"} {
		t.Run(name, func(t *testing.T) {
			copy := *captured
			encoded, _ := json.Marshal(captured.entry)
			_ = json.Unmarshal(encoded, &copy.entry)
			switch name {
			case "owner":
				copy.owner = nil
			case "wire":
				copy.entry.Report.Decisions[0].Forecast.RankScore += 1
			case "snapshot":
				copy.entry.Snapshot.RuntimeVersion++
			case "binding":
				copy.binding.Frontier = "forged"
			case "root":
				copy.root = "forged"
			case "chain":
				if _, err := s.gate.sidecar.Exec("UPDATE witness_v23 SET prior='forged' WHERE seq=(SELECT MAX(seq) FROM witness_v23)"); err != nil {
					t.Fatal(err)
				}
			}
			if s.validate(context.Background(), &copy, s.state.Head) == nil {
				t.Fatal("authority corruption accepted")
			}
			if !reflect.DeepEqual(copy.entry, captured.entry) && name != "wire" && name != "snapshot" {
				t.Fatal("control changed wrong field")
			}
		})
	}
}
