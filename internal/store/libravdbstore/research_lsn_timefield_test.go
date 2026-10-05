package libravdbstore

import (
	"context"
	"os"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestResearchLSNTimeFieldV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_LSN_TIMEFIELD_V1") != "1" {
		t.Skip("opt-in TimeField drop-in rescue probe")
	}
	ctx := context.Background()
	s, err := Open(Config{Path: t.TempDir() + "/timefield.libravdb", Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	name := collectionName("tenant-a", "research:d4")
	if _, err := s.db.CreateCollection(ctx, name, libra.WithDimension(4),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true),
		libra.WithMetadataSchema(libra.MetadataSchema{"available_at": libra.TimeField})); err != nil {
		t.Fatal("create TimeField collection", err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if _, err := s.PutResearchEventBatch(ctx, []ResearchEventWrite{
		pinnedWrite("early", second.Add(100*time.Millisecond)),
		pinnedWrite("late", second.Add(150*time.Millisecond)),
	}); err != nil {
		t.Fatal("unchanged writer into TimeField", err)
	}
	lsn, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal(err)
	}
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + name +
		" AS OF LSN $snapshot_lsn d WHERE available_at <= $available_by ORDER BY distance LIMIT 10"
	for _, check := range []struct {
		asOf time.Time
		want map[string]bool
	}{
		{second.Add(100 * time.Millisecond), map[string]bool{"early": true}},
		{second.Add(120 * time.Millisecond), map[string]bool{"early": true}},
		{second.Add(150 * time.Millisecond), map[string]bool{"early": true, "late": true}},
	} {
		rows, err := s.db.QueryWithParams(ctx, query, libra.QueryParams{
			"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
			"available_by": check.asOf.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
		})
		if err != nil {
			t.Fatal("TimeField SQL", err)
		}
		sqlIDs := make(map[string]bool)
		for _, row := range rows.Results {
			sqlIDs[row.ID] = true
		}
		current, err := s.Search(ctx, "tenant-a", []float32{1, 0, 0, 0}, check.asOf, 10)
		if err != nil {
			t.Fatal("current Search", err)
		}
		currentIDs := make(map[string]bool)
		for _, row := range current {
			currentIDs[row.Event.ID] = true
		}
		t.Logf("as_of=%s TimeField_SQL=%v Search=%v want=%v", check.asOf.Format(time.RFC3339Nano), sqlIDs, currentIDs, check.want)
		if len(sqlIDs) != len(check.want) || len(currentIDs) != len(check.want) {
			t.Errorf("TimeField availability mismatch at %s", check.asOf.Format(time.RFC3339Nano))
			continue
		}
		for id := range check.want {
			if !sqlIDs[id] || !currentIDs[id] {
				t.Errorf("TimeField availability missing %s at %s", id, check.asOf.Format(time.RFC3339Nano))
			}
		}
	}
}
