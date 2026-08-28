package libravdbstore_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

func testConfig(path, modelKey string) libravdbstore.Config {
	return libravdbstore.Config{Path: path, Dimension: 4, Quantization: "none", MemoryMapping: true, EmbeddingModel: modelKey}
}

func TestLegacyMigrationRequiresBackupAndRecoversRecords(t *testing.T) {
	root := t.TempDir()
	path := root + "/legacy.libravdb"
	raw, err := libra.Open(libra.WithStoragePath(path))
	if err != nil {
		t.Fatal(err)
	}
	collection, err := raw.CreateCollection(context.Background(), "events_legacy", libra.WithDimension(4))
	if err != nil {
		t.Fatal(err)
	}
	event := testutil.Event("legacy", "legacy event", time.Now().UTC())
	encoded, _ := json.Marshal(event)
	if err := collection.Insert(context.Background(), event.ID, []float32{1, 0, 0, 0}, map[string]interface{}{"event_json": string(encoded), "content_digest": "old"}); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}
	config := testConfig(path, "model-a:d4")
	if _, err := libravdbstore.Open(config); err == nil {
		t.Fatal("expected legacy migration requirement")
	}
	backup := root + "/legacy-backup.libravdb"
	if err := libravdbstore.MigrateLegacy(context.Background(), config, backup); err != nil {
		t.Fatal(err)
	}
	migrated, err := libravdbstore.Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer migrated.Close()
	results, err := migrated.Search(context.Background(), "tenant-a", []float32{1, 0, 0, 0}, time.Now().Add(time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Event.EmbeddingModel != "model-a:d4" {
		t.Fatalf("migrated results = %+v", results)
	}
}

func TestStoreRoundTripAndAvailabilityGate(t *testing.T) {
	store, err := libravdbstore.Open(libravdbstore.Config{
		Path: t.TempDir() + "/events.libravdb", Dimension: 4, Quantization: "none", MemoryMapping: true, EmbeddingModel: "test:d4",
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
	stats, err := store.Stats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.Tenants != 1 || stats.Events != 2 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestDurableStateSurvivesRestartAndPinsEmbeddingContract(t *testing.T) {
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	event := testutil.Event("one", "durable", time.Now().UTC())
	result, err := first.Put(context.Background(), event, []float32{1, 0, 0, 0}, "digest")
	if err != nil {
		t.Fatal(err)
	}
	if result.Snapshot.RuntimeVersion != 1 || result.Snapshot.EvidenceEpoch != 1 {
		t.Fatalf("snapshot = %+v", result.Snapshot)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	if snapshot := second.Snapshot(context.Background()); snapshot.RuntimeVersion != 1 || snapshot.EvidenceEpoch != 1 {
		t.Fatalf("reopened snapshot = %+v", snapshot)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := libravdbstore.Open(testConfig(path, "model-b:d4")); err == nil {
		t.Fatal("expected embedding contract mismatch")
	}
}

func TestBayesianJournalSurvivesRestartAndRejectsConflict(t *testing.T) {
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	entry := model.BayesianJournalEntry{
		ID: "bj_stable", TenantID: "tenant-a", SessionID: "session-a",
		AsOf: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC), QueryDigest: "digest",
		Snapshot: first.Snapshot(context.Background()),
		Report:   model.BayesianShadowReport{Mode: "shadow", JournalID: "bj_stable", JournalDurable: true, Nominated: 1},
	}
	if err := first.PutBayesianJournal(context.Background(), entry); err != nil {
		t.Fatal(err)
	}
	if err := first.PutBayesianJournal(context.Background(), entry); err != nil {
		t.Fatalf("idempotent journal write: %v", err)
	}
	conflict := entry
	conflict.QueryDigest = "different"
	if err := first.PutBayesianJournal(context.Background(), conflict); err == nil {
		t.Fatal("expected conflicting journal write to fail")
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	got, err := second.GetBayesianJournal(context.Background(), entry.TenantID, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.QueryDigest != entry.QueryDigest || got.Report.Nominated != 1 {
		t.Fatalf("journal after restart = %+v", got)
	}
}

func TestBayesianPosteriorUpdateSurvivesRestart(t *testing.T) {
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	request := model.BayesianOutcomeRequest{IdempotencyKey: "outcome-1", TenantID: "tenant-a", Useful: true, AvailableAt: now}
	result, err := first.ApplyBayesianOutcome(context.Background(), request, "event-1", "digest", 2, bayes.ChangePolicy{Hazard: .05, Threshold: .3, MaxRun: 64})
	if err != nil {
		t.Fatal(err)
	}
	if result.Posterior.Mean() != .75 || result.Snapshot.PosteriorVersion != 2 {
		t.Fatalf("outcome result = %+v", result)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	posterior, err := second.GetBayesianPosterior(context.Background(), "tenant-a", "event-1")
	if err != nil {
		t.Fatal(err)
	}
	if posterior.Mean() != .75 || second.Snapshot(context.Background()).PosteriorVersion != 2 {
		t.Fatalf("posterior after restart = %+v", posterior)
	}
}

func TestBayesianPolicyDigestSurvivesRestartAndAdvancesVersionOnChange(t *testing.T) {
	path := t.TempDir() + "/events.libravdb"
	first, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	bound, err := first.BindBayesianPolicy(context.Background(), "policy-a")
	if err != nil {
		t.Fatal(err)
	}
	if bound.PolicyVersion != 2 {
		t.Fatalf("first policy version = %d", bound.PolicyVersion)
	}
	unchanged, err := first.BindBayesianPolicy(context.Background(), "policy-a")
	if err != nil || unchanged != bound {
		t.Fatalf("same policy bind = %+v, %v", unchanged, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := libravdbstore.Open(testConfig(path, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	reopened, err := second.BindBayesianPolicy(context.Background(), "policy-a")
	if err != nil || reopened != bound {
		t.Fatalf("reopened policy bind = %+v, %v", reopened, err)
	}
	changed, err := second.BindBayesianPolicy(context.Background(), "policy-b")
	if err != nil {
		t.Fatal(err)
	}
	if changed.PolicyVersion != bound.PolicyVersion+1 || changed.RuntimeVersion != bound.RuntimeVersion+1 {
		t.Fatalf("changed policy snapshot = %+v", changed)
	}
}

func TestDeleteRetentionCompactAndBackup(t *testing.T) {
	root := t.TempDir()
	database, err := libravdbstore.Open(testConfig(root+"/events.libravdb", "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	items := []struct {
		value  model.Event
		digest string
	}{
		{testutil.Event("old", "old", now.Add(-2*time.Hour)), "old"},
		{testutil.Event("recent", "recent", now.Add(-time.Minute)), "recent"},
	}
	for _, item := range items {
		if _, err := database.Put(context.Background(), item.value, []float32{1, 0, 0, 0}, item.digest); err != nil {
			t.Fatal(err)
		}
	}
	backup := root + "/backup.libravdb"
	if err := database.Backup(context.Background(), backup); err != nil {
		t.Fatal(err)
	}
	deleted, err := database.Delete(context.Background(), "tenant-a", "recent")
	if err != nil || !deleted.Deleted {
		t.Fatalf("delete = %+v, %v", deleted, err)
	}
	if deleted.Snapshot.GraphVersion != 2 || deleted.Snapshot.PosteriorVersion != 2 {
		t.Fatalf("delete snapshot = %+v", deleted.Snapshot)
	}
	retained, err := database.DeleteBefore(context.Background(), "tenant-a", now.Add(-time.Hour), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(retained.DeletedIDs) != 1 || retained.DeletedIDs[0] != "old" {
		t.Fatalf("retention = %+v", retained)
	}
	if err := database.Compact(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}

	recovered, err := libravdbstore.Open(testConfig(backup, "model-a:d4"))
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	results, err := recovered.Search(context.Background(), "tenant-a", []float32{1, 0, 0, 0}, now, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("backup recovered %d events", len(results))
	}
}
