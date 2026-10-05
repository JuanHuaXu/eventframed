package researchpublicationstore

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestOptionalVectorsPreservedNotInvented(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	for _, persistent := range []bool{false, true} {
		var backend store.EventStore = memorystore.New()
		var e error
		if persistent {
			backend, e = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/vectors.libravdb", Dimension: 4, EmbeddingModel: "fixture", Quantization: "none", MemoryMapping: true})
			if e != nil {
				t.Fatal(e)
			}
		}
		wrapped, e := Wrap(backend)
		if e != nil {
			t.Fatal(e)
		}
		ev := testutil.Event("vector-fixture", "public fixture", now)
		if _, e = wrapped.Put(ctx, ev, []float32{1, 0, 0, 0}, "digest"); e != nil {
			t.Fatal(e)
		}
		vectors, ok := wrapped.(store.VectorEventStore)
		if !ok {
			t.Fatal("lost vectors")
		}
		got, e := vectors.GetEventsWithVectors(ctx, ev.TenantID, []string{ev.ID}, now)
		if e != nil {
			t.Fatal(e)
		}
		want, e := backend.(store.VectorEventStore).GetEventsWithVectors(ctx, ev.TenantID, []string{ev.ID}, now)
		if e != nil || !reflect.DeepEqual(got, want) || len(got) != 1 || len(got[0].Embedding) != 4 {
			t.Fatal("vector forwarding changed data", e)
		}
		proof, ok := wrapped.(interface {
			ResearchSnapshotCompatible(context.Context, model.Snapshot, time.Time) bool
		})
		if !ok || !proof.ResearchSnapshotCompatible(ctx, wrapped.Snapshot(ctx), now) {
			t.Fatal("lost research compatibility")
		}
		wrapped.Close()
	}
	limited := struct{ store.EventStore }{memorystore.New()}
	wrapped, e := Wrap(limited)
	if e != nil {
		t.Fatal(e)
	}
	defer wrapped.Close()
	if _, ok := wrapped.(store.VectorEventStore); ok {
		t.Fatal("invented vector capability")
	}
}
