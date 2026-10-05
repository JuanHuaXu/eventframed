package libravdbstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func fixedAvailabilityKey(at time.Time) string {
	return at.UTC().Format("2006-01-02T15:04:05.000000000Z")
}

func TestResearchLSNFixedKeyV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_LSN_FIXED_KEY_V1") != "1" {
		t.Skip("opt-in fixed-width availability key probe")
	}
	ctx := context.Background()
	path := t.TempDir() + "/fixed-key.libravdb"
	config := Config{Path: path, Dimension: 4, EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true}
	s, err := Open(config)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if s != nil {
			s.Close()
		}
	}()
	name := collectionName("tenant-a", "research:d4")
	col, err := s.db.CreateCollection(ctx, name, libra.WithDimension(4),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true),
		libra.WithMetadataSchema(libra.MetadataSchema{
			"available_at": libra.StringField, "available_at_sort": libra.StringField,
		}))
	if err != nil {
		t.Fatal("create fixed-key collection", err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	early, late := second.Add(100*time.Millisecond), second.Add(150*time.Millisecond)
	if fixedAvailabilityKey(early) != fixedAvailabilityKey(early.In(time.FixedZone("other", -5*3600))) {
		t.Fatal("sort key depends on time zone representation")
	}
	for _, row := range []struct {
		id string
		at time.Time
	}{{"early", early}, {"late", late}} {
		if err := col.Insert(ctx, row.id, []float32{1, 0, 0, 0}, map[string]interface{}{
			"available_at":      row.at.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
			"available_at_sort": fixedAvailabilityKey(row.at),
		}); err != nil {
			t.Fatal("insert fixed-key row", err)
		}
	}
	lsn, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal(err)
	}
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + name +
		" AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT 10"
	check := func(db *libra.Database, stage string) {
		t.Helper()
		for _, row := range []struct {
			asOf time.Time
			want map[string]bool
		}{
			{early, map[string]bool{"early": true}},
			{second.Add(120 * time.Millisecond), map[string]bool{"early": true}},
			{late, map[string]bool{"early": true, "late": true}},
		} {
			rows, err := db.QueryWithParams(ctx, query, libra.QueryParams{
				"snapshot_lsn": int64(lsn), "query_vec": []float32{1, 0, 0, 0},
				"available_by_sort": fixedAvailabilityKey(row.asOf),
			})
			if err != nil {
				t.Fatal(stage, " SQL", err)
			}
			ids := make(map[string]bool)
			for _, item := range rows.Results {
				if ids[item.ID] {
					t.Fatal(stage, " duplicate", item.ID)
				}
				ids[item.ID] = true
			}
			if len(ids) != len(row.want) {
				t.Fatalf("%s as_of=%s got=%v want=%v", stage, row.asOf.Format(time.RFC3339Nano), ids, row.want)
			}
			for id := range row.want {
				if !ids[id] {
					t.Fatal(fmt.Sprintf("%s as_of=%s missing=%s got=%v", stage, row.asOf.Format(time.RFC3339Nano), id, ids))
				}
			}
			t.Logf("stage=%s as_of=%s IDs=%v", stage, row.asOf.Format(time.RFC3339Nano), ids)
		}
	}
	check(s.db, "initial")
	if err := s.Close(); err != nil {
		t.Fatal("close", err)
	}
	s = nil
	reopened, err := Open(config)
	if err != nil {
		t.Fatal("reopen", err)
	}
	defer reopened.Close()
	check(reopened.db, "reopen")
}
