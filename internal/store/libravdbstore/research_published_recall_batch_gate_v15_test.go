package libravdbstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

type batchJournalV15 struct {
	entry   model.BayesianJournalEntry
	encoded []byte
	id      string
}

// appendJournalBatchV15 is a research-only atomic group transition. The
// injected failure and interruption points are used only by contract tests.
func (g *incrementalSortGate) appendJournalBatchV15(ctx context.Context, entries []model.BayesianJournalEntry, stopAt string, failAfter int) error {
	if len(entries) == 0 || len(entries) > 4 {
		return errors.New("journal batch size must be in [1,4]")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	marker, err := g.readMarker(ctx)
	if err != nil {
		return err
	}
	if g.verified == "" || marker.phase != "READY" || marker.digest != g.verified ||
		marker.count != g.wantRows || marker.lsn != g.verifiedLSN {
		return errors.New("frontier journal batch gate is not current")
	}
	g.store.writeMu.RLock()
	defer g.store.writeMu.RUnlock()
	beforeSnapshot := g.store.snapshot
	beforeLSN, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil || beforeLSN != marker.lsn {
		g.verified = ""
		return fmt.Errorf("frontier journal batch LSN moved: latest=%d marker=%d: %w", beforeLSN, marker.lsn, err)
	}
	unique := make(map[string]batchJournalV15, len(entries))
	pending := make([]batchJournalV15, 0, len(entries))
	for _, entry := range entries {
		if entry.ID == "" || entry.TenantID == "" || entry.AsOf.IsZero() ||
			!store.JournalSnapshotCompatible(entry.Snapshot, beforeSnapshot, entry.AsOf, g.store.ingestMotion) {
			g.verified = ""
			return store.ErrStaleSnapshot
		}
		encoded, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		item := batchJournalV15{entry: entry, encoded: encoded, id: bayesianJournalRecordID(entry.TenantID, entry.ID)}
		if previous, ok := unique[item.id]; ok {
			if !bytes.Equal(previous.encoded, item.encoded) {
				g.verified = ""
				return store.ErrJournalConflict
			}
			continue
		}
		unique[item.id] = item
		existing, err := g.store.bayesian.Get(ctx, item.id)
		if err == nil {
			current, _ := existing.Metadata["journal_json"].(string)
			if current != string(encoded) {
				g.verified = ""
				return store.ErrJournalConflict
			}
			continue
		}
		if !errors.Is(err, libra.ErrRecordNotFound) {
			g.verified = ""
			return err
		}
		pending = append(pending, item)
	}
	if len(pending) == 0 {
		return nil
	}
	receipt, err := g.store.db.WithTxReceipt(ctx, func(tx libra.ReceiptTx) error {
		for index, item := range pending {
			metadata := map[string]interface{}{
				"record_type": "frontier_journal", "tenant_id": item.entry.TenantID,
				"journal_id": item.entry.ID, "as_of": item.entry.AsOf.UTC().Format(time.RFC3339Nano),
				"journal_json": string(item.encoded),
			}
			if err := tx.Insert(ctx, bayesianCollection, item.id, nil, metadata); err != nil {
				return err
			}
			if failAfter == index+1 {
				return errors.New("injected journal batch transaction failure")
			}
		}
		return nil
	})
	if err != nil {
		g.verified = ""
		return err
	}
	afterLSN, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil || receipt.CommitLSN == 0 || afterLSN != receipt.CommitLSN || g.store.snapshot != beforeSnapshot {
		g.verified = ""
		return fmt.Errorf("frontier journal batch receipt mismatch: receipt=%d latest=%d: %v", receipt.CommitLSN, afterLSN, err)
	}
	for _, item := range unique {
		saved, err := g.store.GetBayesianJournal(ctx, item.entry.TenantID, item.entry.ID)
		if err != nil {
			g.verified = ""
			return err
		}
		got, err := json.Marshal(saved)
		if err != nil || !bytes.Equal(got, item.encoded) {
			g.verified = ""
			return fmt.Errorf("frontier batch journal readback mismatch: %v", err)
		}
	}
	if stopAt == "after_db" {
		return errors.New("injected interruption after journal batch DB commit")
	}
	markerTx, err := g.sidecar.BeginTx(ctx, nil)
	if err != nil {
		g.verified = ""
		return err
	}
	defer markerTx.Rollback()
	updated, err := markerTx.ExecContext(ctx,
		"UPDATE marker SET lsn=? WHERE id=1 AND phase='READY' AND lsn=? AND row_count=? AND digest=?",
		afterLSN, marker.lsn, marker.count, marker.digest)
	if err != nil {
		g.verified = ""
		return err
	}
	if affected, err := updated.RowsAffected(); err != nil || affected != 1 {
		g.verified = ""
		return fmt.Errorf("frontier batch marker update affected %d rows: %v", affected, err)
	}
	if err := markerTx.Commit(); err != nil {
		g.verified = ""
		return err
	}
	if stopAt == "after_sqlite" {
		return errors.New("injected interruption after journal batch marker commit")
	}
	g.verifiedLSN = afterLSN
	return nil
}

func TestResearchPublishedBatchJournalGateV15(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_PUBLISHED_BATCH_JOURNAL_V15") != "1" {
		t.Skip("opt-in native batch-journal contract")
	}
	ctx := context.Background()
	asOf := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC).Add(120 * time.Millisecond)
	t.Run("atomic_duplicate_conflict", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		one := researchSortJournal(g, "batch-one", asOf)
		two := researchSortJournal(g, "batch-two", asOf)
		if err := g.appendJournalBatchV15(ctx, []model.BayesianJournalEntry{one, two}, "", 0); err != nil {
			t.Fatal("commit batch", err)
		}
		before, ok := g.capture(ctx)
		if !ok {
			t.Fatal("committed batch is not READY")
		}
		if err := g.appendJournalBatchV15(ctx, []model.BayesianJournalEntry{one, two, one}, "", 0); err != nil {
			t.Fatal("exact batch duplicate", err)
		}
		if after, ok := g.capture(ctx); !ok || after != before {
			t.Fatal("duplicate moved the gate", before, after, ok)
		}
		conflict := one
		conflict.QueryDigest = "changed"
		if !errors.Is(g.appendJournalBatchV15(ctx, []model.BayesianJournalEntry{one, conflict}, "", 0), store.ErrJournalConflict) {
			t.Fatal("intra-batch conflict was not rejected")
		}
	})
	t.Run("mixed_horizon_rejects_whole_batch", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		captured := g.store.Snapshot(ctx)
		write := pinnedWrite("batch-horizon-late", asOf.Add(5*time.Millisecond))
		if err := g.appendBatch(ctx, []ResearchEventWrite{write}, ""); err != nil {
			t.Fatal("publish distinguishing event", err)
		}
		compatible := researchSortJournal(g, "batch-horizon-compatible", asOf)
		compatible.Snapshot = captured
		incompatible := researchSortJournal(g, "batch-horizon-incompatible", asOf.Add(10*time.Millisecond))
		incompatible.Snapshot = captured
		if !errors.Is(g.appendJournalBatchV15(ctx, []model.BayesianJournalEntry{compatible, incompatible}, "", 0), store.ErrStaleSnapshot) {
			t.Fatal("mixed-horizon batch was accepted")
		}
		for _, entry := range []model.BayesianJournalEntry{compatible, incompatible} {
			if _, err := g.store.GetBayesianJournal(ctx, entry.TenantID, entry.ID); !errors.Is(err, store.ErrJournalNotFound) {
				t.Fatal("mixed-horizon batch partially wrote", entry.ID, err)
			}
		}
	})
	for _, failure := range []string{"tx", "after_db", "after_sqlite"} {
		t.Run(failure, func(t *testing.T) {
			root := t.TempDir()
			g := createIncrementalSortGate(t, root)
			one := researchSortJournal(g, "batch-failure-one", asOf)
			two := researchSortJournal(g, "batch-failure-two", asOf)
			stopAt, failAfter := failure, 0
			if failure == "tx" {
				stopAt, failAfter = "", 1
			}
			if err := g.appendJournalBatchV15(ctx, []model.BayesianJournalEntry{one, two}, stopAt, failAfter); err == nil {
				t.Fatal("injected failure returned success")
			}
			if err := g.close(); err != nil {
				t.Fatal(err)
			}
			g, err := openIncrementalSortGate(root)
			if err != nil {
				t.Fatal(err)
			}
			defer g.close()
			_, ready := g.capture(ctx)
			if ready != (failure != "after_db") {
				t.Fatal("wrong READY state after reopen", failure, ready)
			}
			for _, entry := range []model.BayesianJournalEntry{one, two} {
				_, err := g.store.GetBayesianJournal(ctx, entry.TenantID, entry.ID)
				if (err == nil) != (failure != "tx") {
					t.Fatal("journal batch was partially committed", failure, entry.ID, err)
				}
			}
		})
	}
}
