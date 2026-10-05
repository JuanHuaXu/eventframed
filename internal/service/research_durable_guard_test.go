package service

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"testing"
)

func TestGuardedServiceAdmissionCommitsBoundLedger(t *testing.T) {
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
	r.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
	req := model.RecallRequest{TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: in.AsOf}
	called := false
	if e = s.WithValidatedResearchAdmission(ctx, r, req, func() error { called = true; return nil }); e == nil || called {
		t.Fatal("unsupported guard accepted", e)
	}
	wrapped, e := researchpublicationstore.Wrap(s.store)
	if e != nil {
		t.Fatal(e)
	}
	s.store = wrapped
	d, e := researchmemory.OpenDurable(ctx, t.TempDir()+"/guard.sqlite", "tenant-a", "stream", 1, 42)
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	e = s.WithValidatedResearchAdmission(ctx, r, req, func() error {
		called = true
		_, _, e := d.AdmitBound(ctx, 1, c.Features, c.Baseline, in.AsOf, *r.Binding)
		return e
	})
	if e != nil || !called {
		t.Fatal("guarded ledger commit failed", e)
	}
	stored, e := d.Admission(ctx, 1)
	if e != nil || stored.Binding == nil || *stored.Binding != *r.Binding {
		t.Fatal("binding not committed", stored, e)
	}
	if _, e = s.store.BindBayesianPolicy(ctx, "after-guard"); e != nil {
		t.Fatal(e)
	}
	called = false
	if e = s.WithValidatedResearchAdmission(ctx, r, req, func() error { called = true; return nil }); e == nil || called {
		t.Fatal("stale guarded operation entered", e)
	}
}
