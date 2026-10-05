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
)

// appendJournal is an explicit metadata-only publication transition. Every
// Store writer must be routed through this gate for its LSN proof to hold.
func (g *incrementalSortGate) appendJournal(ctx context.Context, entry model.BayesianJournalEntry, stopAt string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	marker, err := g.readMarker(ctx)
	if err != nil {
		return err
	}
	if g.verified == "" || marker.phase != "READY" || marker.digest != g.verified ||
		marker.count != g.wantRows || marker.lsn != g.verifiedLSN {
		return errors.New("frontier journal gate is not current")
	}
	g.store.writeMu.RLock()
	beforeSnapshot := g.store.snapshot
	beforeLSN, err := g.store.db.LatestCommitLSN(ctx)
	g.store.writeMu.RUnlock()
	if err != nil {
		g.verified = ""
		return err
	}
	if beforeLSN != marker.lsn {
		g.verified = ""
		return fmt.Errorf("LibraVDB moved outside frontier journal gate: %d != %d", beforeLSN, marker.lsn)
	}
	if err := g.store.PutBayesianJournal(ctx, entry); err != nil {
		g.verified = ""
		return err
	}
	g.store.writeMu.RLock()
	afterSnapshot := g.store.snapshot
	afterLSN, err := g.store.db.LatestCommitLSN(ctx)
	g.store.writeMu.RUnlock()
	if err != nil || afterLSN < beforeLSN || afterSnapshot != beforeSnapshot {
		g.verified = ""
		return fmt.Errorf("frontier journal moved EventFrame state: LSN %d -> %d: %v", beforeLSN, afterLSN, err)
	}
	saved, err := g.store.GetBayesianJournal(ctx, entry.TenantID, entry.ID)
	if err != nil {
		g.verified = ""
		return err
	}
	wantJSON, wantErr := json.Marshal(entry)
	gotJSON, gotErr := json.Marshal(saved)
	if wantErr != nil || gotErr != nil || !bytes.Equal(wantJSON, gotJSON) {
		g.verified = ""
		return fmt.Errorf("frontier journal readback mismatch: want=%v got=%v", wantErr, gotErr)
	}
	if afterLSN == beforeLSN {
		return nil
	}
	if stopAt == "after_db" {
		return errors.New("injected interruption after frontier journal DB commit")
	}
	tx, err := g.sidecar.BeginTx(ctx, nil)
	if err != nil {
		g.verified = ""
		return err
	}
	defer tx.Rollback()
	updated, err := tx.ExecContext(ctx,
		"UPDATE marker SET lsn=? WHERE id=1 AND phase='READY' AND lsn=? AND row_count=? AND digest=?",
		afterLSN, marker.lsn, marker.count, marker.digest)
	if err != nil {
		g.verified = ""
		return err
	}
	if affected, err := updated.RowsAffected(); err != nil || affected != 1 {
		g.verified = ""
		return fmt.Errorf("frontier journal marker update affected %d rows: %v", affected, err)
	}
	if err := tx.Commit(); err != nil {
		g.verified = ""
		return err
	}
	if stopAt == "after_sqlite" {
		return errors.New("injected interruption after frontier journal SQLite commit")
	}
	g.verifiedLSN = afterLSN
	return nil
}

func researchSortJournal(g *incrementalSortGate, id string, at time.Time) model.BayesianJournalEntry {
	return model.BayesianJournalEntry{ID: id, TenantID: "tenant-a", SessionID: "private-lab",
		AsOf: at, QueryDigest: "private-journal-" + id, Snapshot: g.store.Snapshot(context.Background())}
}

func TestResearchSortJournalGateV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_JOURNAL_GATE_V1") != "1" {
		t.Skip("opt-in private journal-through-gate contract")
	}
	ctx := context.Background()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	t.Run("new_duplicate_then_event_and_reopen", func(t *testing.T) {
		root := t.TempDir()
		g := createIncrementalSortGate(t, root)
		defer func() {
			if g != nil {
				g.close()
			}
		}()
		entry := researchSortJournal(g, "gate-journal-001", asOf)
		before, ok := g.capture(ctx)
		if !ok {
			t.Fatal("initial gate not READY")
		}
		if err := g.appendJournal(ctx, entry, ""); err != nil {
			t.Fatal("authorized metadata-only journal", err)
		}
		marker, err := g.readMarker(ctx)
		if err != nil || marker.lsn <= before || marker.count != 3 || marker.digest != g.verified {
			t.Fatal("journal did not advance only the LSN", marker, err)
		}
		if _, ok := g.capture(ctx); !ok {
			t.Fatal("journal-through-gate lost READY")
		}
		if err := g.appendJournal(ctx, entry, ""); err != nil {
			t.Fatal("exact journal duplicate", err)
		}
		if retry, err := g.readMarker(ctx); err != nil || retry != marker {
			t.Fatal("exact journal retry moved marker", retry, err)
		}
		write := pinnedWrite("after-gate-journal", second.Add(110*time.Millisecond))
		if err := g.appendBatch(ctx, []ResearchEventWrite{write}, ""); err != nil {
			t.Fatal("event batch after authorized journal", err)
		}
		ids, err := g.search(ctx, asOf)
		if err != nil || len(ids) != 3 {
			t.Fatal("post-journal exact-LSN as-of read", ids, err)
		}
		if err := g.close(); err != nil {
			t.Fatal(err)
		}
		g = nil
		g, err = openIncrementalSortGate(root)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := g.capture(ctx); !ok || g.wantRows != 4 {
			t.Fatal("reopen lost journal/event publication", ok, g.wantRows)
		}
		conflict := entry
		conflict.QueryDigest = "conflicting-content"
		if err := g.appendJournal(ctx, conflict, ""); err == nil {
			t.Fatal("conflicting frontier journal was accepted")
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("error path remained READY in memory")
		}
	})
	for _, stopAt := range []string{"after_db", "after_sqlite"} {
		t.Run(stopAt, func(t *testing.T) {
			root := t.TempDir()
			g := createIncrementalSortGate(t, root)
			defer func() {
				if g != nil {
					g.close()
				}
			}()
			entry := researchSortJournal(g, "interrupted-"+stopAt, asOf)
			if err := g.appendJournal(ctx, entry, stopAt); err == nil {
				t.Fatal("journal interruption did not fire")
			}
			if _, ok := g.capture(ctx); ok {
				t.Fatal("interrupted metadata-only journal remained READY")
			}
			if err := g.close(); err != nil {
				t.Fatal(err)
			}
			g = nil
			var err error
			g, err = openIncrementalSortGate(root)
			if err != nil {
				t.Fatal(err)
			}
			_, ok := g.capture(ctx)
			if stopAt == "after_db" && ok || stopAt == "after_sqlite" && !ok {
				t.Fatal("reopened interrupted journal had wrong READY state", stopAt, ok)
			}
		})
	}
	t.Run("direct_journal_bypass", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		entry := researchSortJournal(g, "direct-bypass", asOf)
		if err := g.store.PutBayesianJournal(ctx, entry); err != nil {
			t.Fatal(err)
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("direct journal bypass was silently accepted")
		}
		another := researchSortJournal(g, "would-launder-journal", asOf)
		if err := g.appendJournal(ctx, another, ""); err == nil {
			t.Fatal("journal gate laundered a bypassed metadata write")
		}
	})
}

func TestResearchSortJournalGateCostV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_JOURNAL_GATE_COST_V1") != "1" {
		t.Skip("opt-in private journal-through-gate cost diagnostic")
	}
	ctx := context.Background()
	g := createIncrementalSortGate(t, t.TempDir())
	defer g.close()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	var costs, reads []int64
	for i := 0; i < 116; i++ {
		entry := researchSortJournal(g, fmt.Sprintf("cost-journal-%03d", i), asOf)
		start := time.Now()
		if err := g.appendJournal(ctx, entry, ""); err != nil {
			t.Fatal("journal gate cost write", i, err)
		}
		if i >= 16 {
			costs = append(costs, time.Since(start).Nanoseconds())
		}
	}
	r, err := newSortPublishedReader(ctx, g)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		start := time.Now()
		ids, _, err := r.search(ctx, asOf)
		if err != nil || len(ids) != 2 {
			t.Fatal("quiet published view after journals", ids, err)
		}
		reads = append(reads, time.Since(start).Nanoseconds())
	}
	t.Logf("journal p50=%s p99=%s quiet_published_read p50=%s p99=%s",
		pinnedPercentile(costs, .5), pinnedPercentile(costs, .99),
		pinnedPercentile(reads, .5), pinnedPercentile(reads, .99))
}
