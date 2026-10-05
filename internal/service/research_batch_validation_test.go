package service

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

type batchValidationProbe struct {
	store.EventStore
	journals, events int
	alter            string
}

func (p *batchValidationProbe) GetBayesianJournal(ctx context.Context, tenant, id string) (model.BayesianJournalEntry, error) {
	p.journals++
	j, err := p.EventStore.GetBayesianJournal(ctx, tenant, id)
	if p.alter == "duplicate-decision" {
		j.Report.Decisions = append(j.Report.Decisions, j.Report.Decisions[0])
	}
	return j, err
}

func (p *batchValidationProbe) GetEvents(ctx context.Context, tenant string, ids []string, at time.Time) ([]model.Event, error) {
	p.events++
	events, err := p.EventStore.GetEvents(ctx, tenant, ids, at)
	if err != nil {
		return nil, err
	}
	switch p.alter {
	case "duplicate-event":
		events[len(events)-1] = events[0]
	case "missing-event":
		events = events[:len(events)-1]
	case "changed-policy":
		_, err = p.EventStore.BindBayesianPolicy(ctx, "batch-interleave")
	}
	return events, err
}

func TestResearchBatchValidationParity(t *testing.T) {
	for _, kind := range []string{"valid", "features", "baseline", "event", "journal", "snapshot", "query", "asof", "tenant", "duplicate", "prediction-id", "empty", "over-cap", "duplicate-decision", "duplicate-event", "missing-event", "changed-policy"} {
		t.Run(kind, func(t *testing.T) {
			s, _, in := bridgeFixture(t)
			ctx := context.Background()
			a := researchmemory.New(1, 42)
			records := make([]researchmemory.RecordedPrediction, 0, len(in.Candidates))
			for _, c := range in.Candidates {
				p, err := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
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
			if len(records) < 2 {
				t.Fatal("multi-record fixture missing")
			}
			request := model.RecallRequest{TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: in.AsOf}
			for _, r := range records {
				if err := s.ValidateResearchAdmission(ctx, r, request); err != nil {
					t.Fatal("single control", err)
				}
			}
			probe := &batchValidationProbe{EventStore: s.store, alter: kind}
			s.store = probe
			switch kind {
			case "features":
				records[1].Prediction.Features ^= 1
			case "baseline":
				records[1].Outer[0] = 1 - records[1].Outer[0]
				records[1].Prediction.Probability = records[1].Outer[0]
			case "event":
				records[1].Binding.EventID = "missing"
			case "journal":
				records[1].Binding.JournalID = "other"
			case "snapshot":
				records[1].Binding.Snapshot.RuntimeVersion++
			case "query":
				request.Query = "other"
			case "asof":
				records[1].At = records[1].At.Add(time.Second)
			case "tenant":
				records[1].Binding.Tenant = "other"
			case "duplicate":
				records[1] = records[0]
			case "prediction-id":
				records[1].Prediction.ID = records[0].Prediction.ID
			case "empty":
				records = nil
			case "over-cap":
				records = make([]researchmemory.RecordedPrediction, 257)
			}
			err := s.ValidateResearchAdmissionBatch(ctx, records, request)
			if kind == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				if probe.journals != 1 || probe.events != 1 {
					t.Fatal("shared reads not amortized", probe.journals, probe.events)
				}
			} else if err == nil {
				t.Fatal("invalid batch accepted", kind)
			}
		})
	}
}
