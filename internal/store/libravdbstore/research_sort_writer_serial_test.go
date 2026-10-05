package libravdbstore

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

func TestResearchSortWriterSerialHandoffV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_WRITER_SERIAL_V1") != "1" {
		t.Skip("opt-in private serial writer handoff control")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	root := t.TempDir()
	g, err := openSortPublicationGate(root, true)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if g != nil {
			g.close()
		}
	}()
	_, err = g.store.db.CreateCollection(ctx, g.name, libra.WithDimension(4),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true),
		libra.WithMetadataSchema(libra.MetadataSchema{
			"available_at": libra.StringField, "available_at_sort": libra.StringField,
		}))
	if err != nil {
		t.Fatal(err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writes := []ResearchEventWrite{
		pinnedWrite("past100", second.Add(100*time.Millisecond)),
		pinnedWrite("at120", second.Add(120*time.Millisecond)),
		pinnedWrite("future125", second.Add(125*time.Millisecond)),
	}
	if _, lsn, err := g.store.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil || lsn == 0 {
		t.Fatal("initial sortable rows", lsn, err)
	}
	if err := g.publish(ctx); err != nil {
		t.Fatal("publish initial READY", err)
	}
	before, ok := g.capture(ctx)
	if !ok {
		t.Fatal("initial READY denied")
	}
	if err := g.close(); err != nil {
		t.Fatal("close first owner", err)
	}
	g = nil
	other, err := Open(Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
	if err != nil {
		t.Fatal("open serial owner", err)
	}
	legacy := pinnedWrite("unkeyed-serial", second.Add(110*time.Millisecond))
	if _, err := other.PutResearchEventBatch(ctx, []ResearchEventWrite{legacy}); err != nil {
		other.Close()
		t.Fatal("serial legacy write", err)
	}
	committed, err := other.db.LatestCommitLSN(ctx)
	if err != nil || committed <= before {
		other.Close()
		t.Fatal("serial write did not advance LSN", before, committed, err)
	}
	if err := other.Close(); err != nil {
		t.Fatal("close serial owner", err)
	}
	g, err = openSortPublicationGate(root, false)
	if err != nil {
		t.Fatal("reopen original gate", err)
	}
	latest, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil || latest != committed {
		t.Fatal("serial commit boundary was lost", latest, committed, err)
	}
	col, err := g.store.db.GetCollection(g.name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := col.Get(ctx, legacy.Event.ID); err != nil {
		t.Fatal("serial event was not durable", err)
	}
	if _, accepted := g.capture(ctx); accepted {
		t.Fatal("stale READY marker admitted a serial unkeyed write")
	}
	if err := g.publish(ctx); err == nil {
		t.Fatal("full scan republished an unkeyed serial write")
	}
	if got := g.store.Snapshot(ctx).RuntimeVersion; got != 4 {
		t.Fatal("serial handoff lost runtime event count", got)
	}
	t.Logf("sequential handoff preserved LSN %d->%d and failed closed", before, committed)
}
