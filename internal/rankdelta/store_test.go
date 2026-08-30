package rankdelta

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestSQLiteStorePersistsAndVersionGatesBatchedDeltas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rank-deltas.sqlite")
	snapshot := model.Snapshot{PolicyVersion: 2, EvidenceEpoch: 3, GraphVersion: 4, PosteriorVersion: 5, ResidualVersion: 6, AbstractionVersion: 7}
	now := time.Now().UTC()
	record := Record{
		TenantID: "tenant", Key: "query:event", EventID: "event", Delta: .12, Reliability: .8,
		PolicyVersion: 2, EvidenceEpoch: 3, GraphVersion: 4, PosteriorVersion: 5, ResidualVersion: 6, AbstractionVersion: 7,
		UpdatedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Hour),
	}
	store, err := Open(path, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.PutBatch(context.Background(), []Record{record}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	store, err = Open(path, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	got, err := store.GetBatch(context.Background(), "tenant", []string{"missing", record.Key}, snapshot, now)
	if err != nil {
		t.Fatal(err)
	}
	if got[record.Key].Delta != record.Delta || len(got) != 1 {
		t.Fatalf("deltas = %#v", got)
	}
	stale := snapshot
	stale.PosteriorVersion++
	got, err = store.GetBatch(context.Background(), "tenant", []string{record.Key}, stale, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("stale delta survived version gate: %#v", got)
	}
}

func TestSQLiteStoreRejectsExpiredAndInvalidDeltas(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "rank-deltas.sqlite"), 2)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	now := time.Now().UTC()
	invalid := Record{TenantID: "tenant", Key: "key", EventID: "event", Delta: 2, UpdatedAt: now}
	if err := store.PutBatch(context.Background(), []Record{invalid}); err == nil {
		t.Fatal("out-of-range delta was accepted")
	}
	expired := Record{TenantID: "tenant", Key: "key", EventID: "event", Delta: .1, Reliability: .5, UpdatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(-time.Minute)}
	if err := store.PutBatch(context.Background(), []Record{expired}); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetBatch(context.Background(), "tenant", []string{"key"}, model.Snapshot{}, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expired delta returned: %#v", got)
	}
}
