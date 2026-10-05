package service

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type groupReadProbe struct {
	store.EventStore
	reads, journals int
	keys            []string
	alter           string
}

func (p *groupReadProbe) GetEvents(ctx context.Context, tenant string, ids []string, at time.Time) ([]model.Event, error) {
	p.reads++
	p.keys = append([]string(nil), ids...)
	e, err := p.EventStore.GetEvents(ctx, tenant, ids, at)
	if err == nil && len(e) > 1 {
		switch p.alter {
		case "missing":
			e = e[:len(e)-1]
		case "duplicate-return":
			e[len(e)-1] = e[0]
		case "unknown-return":
			e[0].ID = "unknown"
		}
	}
	return e, err
}
func (p *groupReadProbe) GetBayesianJournal(ctx context.Context, tenant, id string) (model.BayesianJournalEntry, error) {
	p.journals++
	return p.EventStore.GetBayesianJournal(ctx, tenant, id)
}
func (p *groupReadProbe) ResearchPublicationCompatible(ctx context.Context, s model.Snapshot, at time.Time) bool {
	return p.EventStore.(interface {
		ResearchPublicationCompatible(context.Context, model.Snapshot, time.Time) bool
	}).ResearchPublicationCompatible(ctx, s, at)
}
func (p *groupReadProbe) WithResearchAsOfSnapshotWait(ctx context.Context, s model.Snapshot, at time.Time, f func() error) error {
	return p.EventStore.(interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	}).WithResearchAsOfSnapshotWait(ctx, s, at, f)
}

func groupValidationFixture(t *testing.T) (*Service, [][]researchmemory.RecordedPrediction, []model.RecallRequest, *groupReadProbe) {
	t.Helper()
	s, now := shadowService(t, ResearchShadowPolicy{})
	tap, err := NewResearchFrontierTap("tenant-a", 8)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tap.Close)
	s.config.ResearchFrontier = tap
	a := researchmemory.New(1, 42)
	var groups [][]researchmemory.RecordedPrediction
	var requests []model.RecallRequest
	for _, query := range []string{"public query", "test fact"} {
		req := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: query, Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: now, RecallK: 3, PackK: 2, TokenBudget: 1000}
		if _, err = s.Recall(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		in, err := tap.Take(ctx)
		cancel()
		if err != nil {
			t.Fatal(err)
		}
		var records []researchmemory.RecordedPrediction
		for _, c := range in.Candidates {
			p, err := a.Predict(c.Features, c.Baseline, 1, now)
			if err != nil {
				t.Fatal(err)
			}
			r, err := a.Record(p.ID)
			if err != nil {
				t.Fatal(err)
			}
			r.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
			records = append(records, r)
		}
		groups = append(groups, records)
		requests = append(requests, req)
	}
	w, err := researchpublicationstore.Wrap(s.store)
	if err != nil {
		t.Fatal(err)
	}
	p := &groupReadProbe{EventStore: w}
	s.store = p
	return s, groups, requests, p
}

func TestResearchGroupValidationParity(t *testing.T) {
	for _, kind := range []string{"overlap", "disjoint", "query", "tenant", "time", "prediction", "features", "baseline", "duplicate-member", "missing", "duplicate-return", "unknown-return", "empty", "count", "record-cap", "stale", "canceled"} {
		t.Run(kind, func(t *testing.T) {
			s, groups, requests, probe := groupValidationFixture(t)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			for i, g := range groups {
				if err := s.ValidateResearchAdmissionBatch(ctx, g, requests[i]); err != nil {
					t.Fatal("single control", err)
				}
			}
			if kind == "disjoint" {
				id := groups[0][0].Binding.EventID
				groups[0] = groups[0][:1]
				var rest []researchmemory.RecordedPrediction
				for _, r := range groups[1] {
					if r.Binding.EventID != id {
						rest = append(rest, r)
					}
				}
				groups[1] = rest
			}
			probe.reads, probe.journals = 0, 0
			probe.alter = kind
			switch kind {
			case "query":
				requests[1].Query = "unrelated"
			case "tenant":
				requests[1].TenantID = "tenant-b"
			case "time":
				requests[1].AsOf = requests[1].AsOf.Add(time.Second)
			case "prediction":
				groups[1][0].Prediction.ID = groups[0][0].Prediction.ID
			case "features":
				groups[1][0].Prediction.Features ^= 1
			case "baseline":
				groups[1][0].Outer[0] = 1 - groups[1][0].Outer[0]
				groups[1][0].Prediction.Probability = groups[1][0].Outer[0]
			case "duplicate-member":
				groups[1][1] = groups[1][0]
				groups[1][1].Prediction.ID = 999
			case "empty":
				groups[1] = nil
			case "count":
				requests = requests[:1]
			case "record-cap":
				groups[1] = make([]researchmemory.RecordedPrediction, 257)
			case "stale":
				if _, err := s.store.BindBayesianPolicy(ctx, "changed"); err != nil {
					t.Fatal(err)
				}
			case "canceled":
				cancel()
			}
			called := false
			err := s.WithValidatedResearchGroupAdmission(ctx, groups, requests, func() error { called = true; return nil })
			valid := kind == "overlap" || kind == "disjoint"
			if valid {
				if err != nil || !called || probe.reads != 1 || probe.journals != 2 || len(probe.keys) != 3 {
					t.Fatal("union parity", err, probe.reads, probe.journals, len(probe.keys))
				}
				// No authority/read cache survives the invocation.
				if _, err = s.store.BindBayesianPolicy(ctx, "after"); err != nil {
					t.Fatal(err)
				}
				called = false
				if err = s.WithValidatedResearchGroupAdmission(ctx, groups, requests, func() error { called = true; return nil }); err == nil || called {
					t.Fatal("stale group reused")
				}
			} else if err == nil || called {
				t.Fatal("invalid group admitted", kind)
			}
		})
	}
}
