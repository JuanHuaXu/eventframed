package libravdbstore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

type joinedJournalV25 struct {
	entry   model.BayesianJournalEntry
	encoded []byte
	id      string
}

// appendJoinedJournalV25 is a research-only atomic group transition. The
// witness callback runs in the marker transaction; callers publish state only
// after its commit. Backend-before-marker interruption remains fail closed.
func (g *incrementalSortGate) appendJoinedJournalV25(ctx context.Context, entries []model.BayesianJournalEntry, stopAt string, failAfter int, witness func(*sql.Tx) error) error {
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
	unique := make(map[string]joinedJournalV25, len(entries))
	pending := make([]joinedJournalV25, 0, len(entries))
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
		item := joinedJournalV25{entry: entry, encoded: encoded, id: bayesianJournalRecordID(entry.TenantID, entry.ID)}
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
		tx, err := g.sidecar.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := witness(tx); err != nil {
			return err
		}
		return tx.Commit()
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
	if err := witness(markerTx); err != nil {
		g.verified = ""
		return err
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
