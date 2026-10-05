package researchmemory

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"path/filepath"
	"testing"
)

func TestDurableServiceBindingPersistsAndRejectsRebinding(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "binding.sqlite")
	d, e := OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if e != nil {
		t.Fatal(e)
	}
	binding := ServiceBinding{Tenant: "tenant", JournalID: "service-journal", EventID: "event-a", Snapshot: model.Snapshot{RuntimeVersion: 1, ContractVersion: 1}}
	p, dup, e := d.AdmitBound(ctx, 1, 3, .6, durableFixtureTime, binding)
	if e != nil || dup {
		t.Fatal(p, dup, e)
	}
	for _, mutate := range []func(*ServiceBinding){func(b *ServiceBinding) { b.Tenant = "other" }, func(b *ServiceBinding) { b.JournalID = "other" }, func(b *ServiceBinding) { b.EventID = "other" }, func(b *ServiceBinding) { b.Snapshot.EvidenceEpoch++ }} {
		changed := binding
		mutate(&changed)
		if _, _, e = d.AdmitBound(ctx, 1, 3, .6, durableFixtureTime, changed); e == nil {
			t.Fatal("rebound original record", changed)
		}
	}
	if _, _, e = d.Admit(ctx, 1, 3, .6, durableFixtureTime); e == nil {
		t.Fatal("binding removed on retry")
	}
	r, e := d.Admission(ctx, 1)
	if e != nil || r.Binding == nil || *r.Binding != binding {
		t.Fatal(r, e)
	}
	r.Binding.EventID = "caller mutation"
	r, e = d.Admission(ctx, 1)
	if e != nil || r.Binding.EventID != binding.EventID {
		t.Fatal("returned binding aliases storage", e)
	}
	if e = d.Close(); e != nil {
		t.Fatal(e)
	}
	d, e = OpenDurable(ctx, path, "tenant", "stream", 1, 42)
	if e != nil {
		t.Fatal(e)
	}
	defer d.Close()
	r, e = d.Admission(ctx, 1)
	if e != nil || r.Binding == nil || *r.Binding != binding {
		t.Fatal("binding lost on restart", r, e)
	}
	q, dup, e := d.AdmitBound(ctx, 1, 3, .6, durableFixtureTime, binding)
	if e != nil || !dup || q != p {
		t.Fatal("bound retry changed forecast", p, q, dup, e)
	}
}
