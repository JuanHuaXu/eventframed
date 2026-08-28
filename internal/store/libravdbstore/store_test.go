package libravdbstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestStoreRoundTripAndAvailabilityGate(t *testing.T) {
	store, err := libravdbstore.Open(libravdbstore.Config{
		Path: t.TempDir() + "/events.libravdb", Dimension: 4, Quantization: "none", MemoryMapping: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	now := time.Now().UTC()
	past := testutil.Event("past", "remember this", now.Add(-time.Minute))
	future := testutil.Event("future", "remember this", now.Add(time.Hour))
	vector := []float32{1, 0, 0, 0}
	if _, err := store.Put(context.Background(), past, vector, "past-digest"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Put(context.Background(), future, vector, "future-digest"); err != nil {
		t.Fatal(err)
	}
	results, err := store.Search(context.Background(), "tenant-a", vector, now, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Event.ID != "past" {
		t.Fatalf("results = %+v", results)
	}
}
