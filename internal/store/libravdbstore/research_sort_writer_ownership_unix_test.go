//go:build darwin || linux

package libravdbstore

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
	"golang.org/x/sys/unix"
)

func TestResearchSortWriterOwnershipV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_WRITER_OWNERSHIP_V1") != "1" {
		t.Skip("opt-in private writer ownership bypass probe")
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
		t.Fatal("private initial sortable write", lsn, err)
	}
	if err := g.publish(ctx); err != nil {
		t.Fatal("private READY", err)
	}
	before, ok := g.capture(ctx)
	if !ok {
		t.Fatal("private READY was not captured")
	}
	lockPath := filepath.Join(root, "eventframed-owner.lock")
	owner, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if err := unix.Flock(int(owner.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		owner.Close()
		t.Fatal("first owner lock", err)
	}
	defer owner.Close()
	contender, err := os.OpenFile(lockPath, os.O_RDWR, 0600)
	if err != nil {
		owner.Close()
		t.Fatal(err)
	}
	if err := unix.Flock(int(contender.Fd()), unix.LOCK_EX|unix.LOCK_NB); err == nil {
		contender.Close()
		owner.Close()
		t.Fatal("advisory owner lock admitted a second descriptor")
	}
	if err := contender.Close(); err != nil {
		owner.Close()
		t.Fatal(err)
	}
	other, openErr := Open(Config{Path: filepath.Join(root, "events.libravdb"), Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true})
	if openErr != nil {
		t.Logf("second Store Open refused while first Store owns database: %v", openErr)
	} else {
		legacy := pinnedWrite("unkeyed-direct", second.Add(110*time.Millisecond))
		_, writeErr := other.PutResearchEventBatch(ctx, []ResearchEventWrite{legacy})
		otherLSN, otherLSNErr := other.db.LatestCommitLSN(ctx)
		closeErr := other.Close()
		if writeErr != nil || closeErr != nil {
			t.Logf("second Store direct write did not complete: write=%v close=%v", writeErr, closeErr)
		} else {
			t.Logf("second Store post-write LSN=%d err=%v", otherLSN, otherLSNErr)
			after, err := g.store.db.LatestCommitLSN(ctx)
			if err != nil {
				t.Fatal("first Store latest LSN after direct write", err)
			}
			captured, accepted := g.capture(ctx)
			t.Logf("direct write bypassed advisory owner lock: first_latest_before=%d after=%d captured=%d accepted=%v",
				before, after, captured, accepted)
			if accepted && captured == after && after > before {
				t.Error("READY accepted an unkeyed row at the current LSN")
			}
			if accepted && captured == before && otherLSN > before {
				t.Error("already-open Store kept a stale READY view after another Store acknowledged a write")
			}
			if err := g.close(); err != nil {
				t.Fatal("close first Store", err)
			}
			g = nil
			g, err = openSortPublicationGate(root, false)
			if err != nil {
				t.Fatal("reopen first Store", err)
			}
			reopenedLSN, reopenedErr := g.store.db.LatestCommitLSN(ctx)
			col, colErr := g.store.db.GetCollection(g.name)
			var rowErr error
			if colErr == nil {
				_, rowErr = col.Get(ctx, legacy.Event.ID)
			}
			t.Logf("reopened LSN=%d err=%v external_row_err=%v collection_err=%v", reopenedLSN, reopenedErr, rowErr, colErr)
			if colErr != nil {
				t.Fatal(colErr)
			}
			if rowErr != nil {
				t.Error("second Store reported a successful commit, but its event was lost after the first Store closed:", rowErr)
			} else {
				if _, accepted := g.capture(ctx); accepted {
					t.Error("READY survived a durable unkeyed external write")
				}
				if err := g.publish(ctx); err == nil {
					t.Error("unkeyed direct write was republished")
				}
			}
		}
	}
	if err := owner.Close(); err != nil {
		t.Fatal("release owner lock", err)
	}
	reclaimer, err := os.OpenFile(lockPath, os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer reclaimer.Close()
	if err := unix.Flock(int(reclaimer.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		t.Fatal("owner lock was not released", err)
	}
	if err := unix.Flock(int(reclaimer.Fd()), unix.LOCK_UN); err != nil {
		t.Fatal("release reclaimed lock", err)
	}
}
