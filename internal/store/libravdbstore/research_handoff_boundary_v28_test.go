package libravdbstore

import (
	"context"
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"os"
	"testing"
	"time"
)

// New prospective control order. V27's failed same-epoch expectation stays
// frozen; altered request controls now exercise genuinely cross-epoch reuse.
func TestResearchHandoffBoundaryV28(t *testing.T) {
	output := os.Getenv("EVENTFRAME_HANDOFF_BOUNDARY_V28_OUTPUT")
	if output == "" {
		t.Skip("exclusive isolated boundary diagnostic")
	}
	ctx := context.Background()
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachCombinedV26(t, f, true)
	o := &handoffObserverV27{EventStore: s}
	svc, err := service.New(o, f.em, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 10000, OverfetchMultiplier: 3})
	if err != nil {
		t.Fatal(err)
	}
	f.svc = svc
	var records []handoffResultV27
	for _, name := range []string{"prime", "learned", "future-only", "query", "vector", "selection", "visible"} {
		r := f.request(time.Now().UTC(), name)
		switch name {
		case "query":
			r.Query = "different"
		case "vector":
			r.Embedding = denseRowV6(f.query, 11, .1)
		case "selection":
			r.PackK = 9
		case "future-only", "visible":
			at := f.origin.Add(time.Hour)
			if name == "visible" {
				at = f.origin
			}
			w := pinnedWrite(name, at)
			w.Vector = denseRowV6(f.query, 7, .0003)
			if _, err := s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
		}
		o.arm()
		packet, _, err := s.recall(ctx, svc, r)
		if err != nil {
			t.Fatal(err)
		}
		o.mu.Lock()
		o.armed = false
		calls := append([]string(nil), o.calls...)
		post := append([]string(nil), o.post...)
		journal := o.journal
		o.mu.Unlock()
		if len(post) != 0 || journal.ID == "" || journal.Snapshot != packet.Snapshot || handoffMethodCountV27 < 31 {
			t.Fatal("incomplete boundary", name, post)
		}
		saved, err := s.gate.store.GetBayesianJournal(ctx, "tenant-a", journal.ID)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := json.Marshal(saved)
		want, _ := json.Marshal(journal)
		if string(got) != string(want) {
			t.Fatal("wire changed")
		}
		for _, c := range packet.Candidates {
			found := false
			for _, d := range journal.Report.Decisions {
				if d.EventID == c.Event.ID {
					a, _ := json.Marshal(c.Forecast)
					b, _ := json.Marshal(d.Forecast)
					if string(a) != string(b) {
						t.Fatal("packing changed law")
					}
					found = true
					break
				}
			}
			if !found {
				t.Fatal("packed outside journal")
			}
		}
		records = append(records, handoffResultV27{name, handoffMethodCountV27, calls, post, journal.ID, len(journal.Report.Decisions), len(packet.Candidates), witnessBeliefsV23(packet), witnessHashV23(journal)})
		wantBelief := name == "learned" || name == "future-only"
		if (witnessBeliefsV23(packet) > 0) != wantBelief {
			t.Fatal("wrong positive/negative belief control", name, witnessBeliefsV23(packet))
		}
		if name == "prime" {
			if _, err := f.feedback(ctx, journal.ID, "past100", "first", false); err != nil {
				t.Fatal(err)
			}
		}
	}
	// Observer must flag a deliberately inserted post-boundary store call.
	o.arm()
	o.mu.Lock()
	o.entered = true
	o.mu.Unlock()
	_ = o.Snapshot(ctx)
	o.mu.Lock()
	if len(o.post) != 1 || o.post[0] != "Snapshot" {
		o.mu.Unlock()
		t.Fatal("observer negative control failed")
	}
	o.armed = false
	o.mu.Unlock()
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err = json.NewEncoder(file).Encode(map[string]any{"Type": "technical-boundary-diagnostic", "Cases": records, "NegativeControl": "post-boundary Snapshot detected", "LeasesUnchanged": true, "NoEarlyRelease": true, "Time": time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if err = file.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("cases=%d interface_methods=%d no post-handoff store calls; not an early-release or latency proof", len(records), handoffMethodCountV27)
}
