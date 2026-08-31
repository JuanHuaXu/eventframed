package memorystore_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestVectorHydrationIsExplicit(t *testing.T) {
	ctx := context.Background()
	runtime := memorystore.New()
	event := testutil.Event("vector", "stored vector", time.Now().UTC())
	want := []float32{1, 2, 3, 4}
	if _, err := runtime.Put(ctx, event, want, "digest"); err != nil {
		t.Fatal(err)
	}
	ordinary, err := runtime.GetEvents(ctx, event.TenantID, []string{event.ID}, time.Now().Add(time.Minute))
	if err != nil || len(ordinary[0].Embedding) != 0 {
		t.Fatalf("ordinary event leaked vector: %+v, %v", ordinary, err)
	}
	hydrated, err := runtime.GetEventsWithVectors(ctx, event.TenantID, []string{event.ID}, time.Now().Add(time.Minute))
	if err != nil || !reflect.DeepEqual(hydrated[0].Embedding, want) {
		t.Fatalf("hydrated vector = %v, %v", hydrated[0].Embedding, err)
	}
}
