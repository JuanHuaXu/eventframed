package service

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"testing"
	"time"
)

func TestDurableAdmissionValidatesActualServiceJournal(t *testing.T) {
	s, _, in := bridgeFixture(t)
	ctx := context.Background()
	c := in.Candidates[0]
	a := researchmemory.New(1, 42)
	p, e := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
	if e != nil {
		t.Fatal(e)
	}
	r, e := a.Record(p.ID)
	if e != nil {
		t.Fatal(e)
	}
	binding := researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
	r.Binding = &binding
	req := model.RecallRequest{TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: in.AsOf}
	if e = s.ValidateResearchAdmission(ctx, r, req); e != nil {
		t.Fatal("valid journal rejected", e)
	}
	for _, kind := range []string{"features", "baseline", "event", "journal", "snapshot", "query", "asof", "tenant"} {
		bad := r
		b := binding
		bad.Binding = &b
		q := req
		switch kind {
		case "features":
			bad.Prediction.Features ^= 1
		case "baseline":
			bad.Outer[0] = 1 - c.Baseline
			bad.Prediction.Probability = bad.Outer[0]
		case "event":
			b.EventID = "missing"
		case "journal":
			b.JournalID = "missing"
		case "snapshot":
			b.Snapshot.EvidenceEpoch++
		case "query":
			q.Query = "unrelated query"
		case "asof":
			q.AsOf = q.AsOf.Add(time.Second)
		case "tenant":
			q.TenantID = "other"
		}
		if e = s.ValidateResearchAdmission(ctx, bad, q); e == nil {
			t.Fatal("invalid input accepted", kind)
		}
	}
	if _, e = s.store.BindBayesianPolicy(ctx, "changed-policy"); e != nil {
		t.Fatal(e)
	}
	if e = s.ValidateResearchAdmission(ctx, r, req); e == nil {
		t.Fatal("stale binding accepted")
	}
}
