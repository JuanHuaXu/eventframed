package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
)

// The preview is only an admission-validation envelope. The stored original
// comes from the owner's worker, and retries have no caller-controlled learner ID.
func TestResearchSourceOwnerGuardedRetry(t *testing.T) {
	ctx := context.Background()
	s, _, in := bridgeFixture(t)
	wrapped, err := researchpublicationstore.Wrap(s.store)
	if err != nil {
		t.Fatal(err)
	}
	s.store = wrapped
	path := t.TempDir() + "/source.sqlite"
	o, err := researchmemory.OpenSourceOwner(ctx, path, "tenant-a", "source", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { o.Close() }()
	c := in.Candidates[0]
	a := researchmemory.New(1, 42)
	p, err := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
	if err != nil {
		t.Fatal(err)
	}
	r, err := a.Record(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	r.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
	req := model.RecallRequest{TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: in.AsOf}
	admit := func() ([]researchmemory.AdmissionResult, error) {
		var result []researchmemory.AdmissionResult
		err := s.WithValidatedResearchAdmission(ctx, r, req, func() error {
			var e error
			result, e = o.Admit(ctx, []researchmemory.SourceAdmissionRequest{{Features: c.Features, Baseline: c.Baseline, At: in.AsOf, Binding: *r.Binding}})
			return e
		})
		return result, err
	}
	first, err := admit()
	if err != nil || len(first) != 1 || first[0].Retry {
		t.Fatal(first, err)
	}
	for i := 0; i < 3; i++ {
		if err = o.Close(); err != nil {
			t.Fatal(err)
		}
		o, err = researchmemory.OpenSourceOwner(ctx, path, "tenant-a", "source", 1, 42)
		if err != nil {
			t.Fatal(err)
		}
		got, e := admit()
		if e != nil || !got[0].Retry || !reflect.DeepEqual(got[0].Record, first[0].Record) {
			t.Fatal(got, e)
		}
	}
	if _, err = s.store.BindBayesianPolicy(ctx, "changed"); err != nil {
		t.Fatal(err)
	}
	if got, err := admit(); err == nil || got != nil {
		t.Fatal("stale journal authorized retry", got, err)
	}
	// Reading the immutable original still works, but is not fresh authority.
	if got, err := o.Lookup(ctx, in.JournalID, c.EventID); err != nil || !reflect.DeepEqual(got, first[0].Record) {
		t.Fatal(got, err)
	}
	if _, err = o.Discard(ctx, in.JournalID, c.EventID, in.AsOf); err != nil {
		t.Fatal(err)
	}
}
