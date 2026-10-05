package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

// appendBatch publishes a receipt-bound group only after every new row has
// been read back and journaled. A DB-only commit leaves the old marker stale.
func (g *incrementalSortGate) appendBatch(ctx context.Context, writes []ResearchEventWrite, stopAt string) error {
	return g.appendBatchWithHook(ctx, writes, stopAt, nil)
}

func (g *incrementalSortGate) appendBatchWithHook(ctx context.Context, writes []ResearchEventWrite, stopAt string, afterPrecheck func() error) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	marker, err := g.readMarker(ctx)
	if err != nil {
		return err
	}
	if g.verified == "" || marker.phase != "READY" || marker.digest != g.verified ||
		marker.lsn != g.verifiedLSN || marker.count != g.wantRows {
		return errors.New("incremental batch gate is not current")
	}
	g.store.writeMu.RLock()
	beforeSnapshot := g.store.snapshot
	before, err := g.store.db.LatestCommitLSN(ctx)
	g.store.writeMu.RUnlock()
	if err != nil {
		g.verified = ""
		return err
	}
	if before != marker.lsn {
		g.verified = ""
		return fmt.Errorf("LibraVDB moved outside the batch gate: %d != %d", before, marker.lsn)
	}
	if afterPrecheck != nil {
		if err := afterPrecheck(); err != nil {
			return err
		}
	}
	results, receipt, err := g.store.PutResearchEventBatchSortableReceipt(ctx, writes)
	if err != nil {
		g.verified = ""
		return err
	}
	if len(results) != len(writes) {
		g.verified = ""
		return fmt.Errorf("batch result count %d != %d", len(results), len(writes))
	}
	if receipt == 0 {
		for _, result := range results {
			if !result.Duplicate {
				g.verified = ""
				return errors.New("zero batch receipt with a new event")
			}
		}
		g.store.writeMu.RLock()
		afterSnapshot := g.store.snapshot
		after, err := g.store.db.LatestCommitLSN(ctx)
		g.store.writeMu.RUnlock()
		if err != nil || after != before || afterSnapshot != beforeSnapshot {
			g.verified = ""
			return fmt.Errorf("duplicate-only batch moved LibraVDB state: LSN %d -> %d: %v", before, after, err)
		}
		return nil
	}
	g.store.writeMu.RLock()
	afterSnapshot := g.store.snapshot
	latest, err := g.store.db.LatestCommitLSN(ctx)
	g.store.writeMu.RUnlock()
	if err != nil || latest != receipt {
		g.verified = ""
		return fmt.Errorf("batch receipt is not exact latest commit: %d != %d: %v", receipt, latest, err)
	}
	col, err := g.store.db.GetCollection(g.name)
	if err != nil {
		g.verified = ""
		return err
	}
	type journalRow struct {
		id   string
		hash [sha256.Size]byte
	}
	rows := make([]journalRow, 0, len(writes))
	for i, result := range results {
		if result.Duplicate {
			continue
		}
		record, err := col.Get(ctx, writes[i].Event.ID)
		if err != nil {
			g.verified = ""
			return err
		}
		hash, err := incrementalRowHash(record.ID, record.Metadata, record.Vector)
		if err != nil {
			g.verified = ""
			return err
		}
		rows = append(rows, journalRow{record.ID, hash})
	}
	if len(rows) == 0 || len(rows) > 16 {
		g.verified = ""
		return fmt.Errorf("nonzero receipt has %d new rows", len(rows))
	}
	expectedSnapshot := beforeSnapshot
	expectedSnapshot.RuntimeVersion += uint64(len(rows))
	expectedSnapshot.EvidenceEpoch += uint64(len(rows))
	if afterSnapshot != expectedSnapshot {
		g.verified = ""
		return fmt.Errorf("batch snapshot moved outside journaled events: before=%+v after=%+v new=%d",
			beforeSnapshot, afterSnapshot, len(rows))
	}
	if stopAt == "after_db" {
		return errors.New("injected interruption after LibraVDB batch commit")
	}
	prior, err := hex.DecodeString(marker.digest)
	if err != nil || len(prior) != sha256.Size {
		g.verified = ""
		return fmt.Errorf("invalid prior batch chain: length=%d err=%v", len(prior), err)
	}
	var chain [sha256.Size]byte
	copy(chain[:], prior)
	tx, err := g.sidecar.BeginTx(ctx, nil)
	if err != nil {
		g.verified = ""
		return err
	}
	defer tx.Rollback()
	for i, row := range rows {
		seq := marker.count + i + 1
		chain = incrementalChainNext(chain, row.hash, receipt, uint64(seq))
		if _, err := tx.ExecContext(ctx, "INSERT INTO journal(seq,id,row_hash,lsn) VALUES(?,?,?,?)",
			seq, row.id, row.hash[:], receipt); err != nil {
			g.verified = ""
			return err
		}
	}
	count := marker.count + len(rows)
	root := hex.EncodeToString(chain[:])
	updated, err := tx.ExecContext(ctx,
		"UPDATE marker SET lsn=?,row_count=?,digest=? WHERE id=1 AND phase='READY' AND lsn=? AND row_count=? AND digest=?",
		receipt, count, root, marker.lsn, marker.count, marker.digest)
	if err != nil {
		g.verified = ""
		return err
	}
	if affected, err := updated.RowsAffected(); err != nil || affected != 1 {
		g.verified = ""
		return fmt.Errorf("batch marker update affected %d rows: %w", affected, err)
	}
	if err := tx.Commit(); err != nil {
		g.verified = ""
		return err
	}
	if stopAt == "after_sqlite" {
		return errors.New("injected interruption after SQLite batch commit")
	}
	g.verified, g.verifiedLSN, g.wantRows = root, receipt, count
	return nil
}

func TestResearchIncrementalSortBatchPublicationV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_BATCH_PUBLICATION_V1") != "1" {
		t.Skip("opt-in private batch publication contract")
	}
	ctx := context.Background()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	t.Run("normal_mixed_retry_and_asof", func(t *testing.T) {
		root := t.TempDir()
		g := createIncrementalSortGate(t, root)
		defer func() {
			if g != nil {
				g.close()
			}
		}()
		early := pinnedWrite("batch-early", second.Add(110*time.Millisecond))
		future := pinnedWrite("batch-future", second.Add(130*time.Millisecond))
		before := g.store.Snapshot(ctx)
		if err := g.appendBatch(ctx, []ResearchEventWrite{early, future}, ""); err != nil {
			t.Fatal("all-new batch", err)
		}
		after := g.store.Snapshot(ctx)
		if after.RuntimeVersion != before.RuntimeVersion+2 || after.EvidenceEpoch != before.EvidenceEpoch+2 ||
			!g.store.ingestMotion[before.RuntimeVersion+1].Equal(early.Event.AvailableAt) ||
			!g.store.ingestMotion[before.RuntimeVersion+2].Equal(future.Event.AvailableAt) {
			t.Fatal("batch lost per-event runtime version or ingestion motion")
		}
		marker, err := g.readMarker(ctx)
		if err != nil || marker.count != 5 {
			t.Fatal("batch marker did not advance by two", marker, err)
		}
		ids, err := g.search(ctx, second.Add(120*time.Millisecond))
		if err != nil || len(ids) != 3 {
			t.Fatal("batch as-of query includes future or misses past", ids, err)
		}
		seen := make(map[string]bool, len(ids))
		for _, id := range ids {
			seen[id] = true
		}
		if !seen["past100"] || !seen["at120"] || !seen["batch-early"] || seen["future125"] || seen["batch-future"] {
			t.Fatal("incorrect exact-LSN as-of set", ids)
		}
		if err := g.appendBatch(ctx, []ResearchEventWrite{early, future}, ""); err != nil {
			t.Fatal("exact duplicate-only retry", err)
		}
		if retryMarker, err := g.readMarker(ctx); err != nil || retryMarker != marker || g.store.Snapshot(ctx) != after {
			t.Fatal("duplicate-only retry changed durable state", retryMarker, err)
		}
		middle := pinnedWrite("batch-middle", second.Add(115*time.Millisecond))
		if err := g.appendBatch(ctx, []ResearchEventWrite{early, middle}, ""); err != nil {
			t.Fatal("mixed duplicate/new batch", err)
		}
		if current, err := g.readMarker(ctx); err != nil || current.count != 6 ||
			g.store.Snapshot(ctx).RuntimeVersion != after.RuntimeVersion+1 ||
			!g.store.ingestMotion[after.RuntimeVersion+1].Equal(middle.Event.AvailableAt) {
			t.Fatal("mixed batch lost one-event motion", current, err)
		}
		if err := g.close(); err != nil {
			t.Fatal("close batch gate", err)
		}
		g = nil
		g, err = openIncrementalSortGate(root)
		if err != nil {
			t.Fatal("reopen batch gate", err)
		}
		if _, ok := g.capture(ctx); !ok || g.wantRows != 6 {
			t.Fatal("reopen lost verified batch journal", ok, g.wantRows)
		}
		conflict := pinnedWrite(early.Event.ID, second.Add(210*time.Millisecond))
		if err := g.appendBatch(ctx, []ResearchEventWrite{conflict}, ""); err == nil {
			t.Fatal("conflicting duplicate was accepted")
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("uncertain conflict path stayed READY in memory")
		}
	})
	t.Run("interruption_after_db", func(t *testing.T) {
		root := t.TempDir()
		g := createIncrementalSortGate(t, root)
		writes := []ResearchEventWrite{
			pinnedWrite("db-first", second.Add(111*time.Millisecond)),
			pinnedWrite("db-second", second.Add(112*time.Millisecond)),
		}
		if err := g.appendBatch(ctx, writes, "after_db"); err == nil {
			t.Fatal("DB interruption did not fire")
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("DB-only batch remained READY")
		}
		if err := g.close(); err != nil {
			t.Fatal(err)
		}
		g, err := openIncrementalSortGate(root)
		if err != nil {
			t.Fatal(err)
		}
		defer g.close()
		if _, ok := g.capture(ctx); ok {
			t.Fatal("DB-only batch republished on reopen")
		}
		col, err := g.store.db.GetCollection(g.name)
		if err != nil {
			t.Fatal(err)
		}
		for _, write := range writes {
			if _, err := col.Get(ctx, write.Event.ID); err != nil {
				t.Fatal("committed batch row absent after reopen", write.Event.ID, err)
			}
		}
	})
	t.Run("interruption_after_sqlite", func(t *testing.T) {
		root := t.TempDir()
		g := createIncrementalSortGate(t, root)
		writes := []ResearchEventWrite{
			pinnedWrite("sql-first", second.Add(113*time.Millisecond)),
			pinnedWrite("sql-second", second.Add(114*time.Millisecond)),
		}
		if err := g.appendBatch(ctx, writes, "after_sqlite"); err == nil {
			t.Fatal("SQLite interruption did not fire")
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("in-memory gate accepted an interrupted SQLite batch")
		}
		if err := g.close(); err != nil {
			t.Fatal(err)
		}
		g, err := openIncrementalSortGate(root)
		if err != nil {
			t.Fatal(err)
		}
		defer g.close()
		if _, ok := g.capture(ctx); !ok || g.wantRows != 5 {
			t.Fatal("committed batch journal did not verify on reopen", ok, g.wantRows)
		}
	})
	t.Run("bypassed_sortable_write", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		bypass := pinnedWrite("batch-bypass", second.Add(116*time.Millisecond))
		if _, receipt, err := g.store.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{bypass}); err != nil || receipt == 0 {
			t.Fatal("bypass writer failed", receipt, err)
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("batch gate accepted a bypassed write")
		}
		candidate := pinnedWrite("after-bypass", second.Add(117*time.Millisecond))
		if err := g.appendBatch(ctx, []ResearchEventWrite{candidate}, ""); err == nil {
			t.Fatal("batch gate laundered a bypassed write")
		}
		col, err := g.store.db.GetCollection(g.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := col.Get(ctx, candidate.Event.ID); !errors.Is(err, libra.ErrRecordNotFound) {
			t.Fatal("rejected batch committed", err)
		}
	})
	t.Run("bypassed_unkeyed_write", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		unkeyed := pinnedWrite("batch-unkeyed", second.Add(118*time.Millisecond))
		if _, err := g.store.PutResearchEventBatch(ctx, []ResearchEventWrite{unkeyed}); err != nil {
			t.Fatal("unkeyed bypass writer failed", err)
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("batch gate accepted an unkeyed bypassed write")
		}
		candidate := pinnedWrite("after-unkeyed", second.Add(119*time.Millisecond))
		if err := g.appendBatch(ctx, []ResearchEventWrite{candidate}, ""); err == nil {
			t.Fatal("batch gate laundered an unkeyed bypassed write")
		}
		col, err := g.store.db.GetCollection(g.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := col.Get(ctx, candidate.Event.ID); !errors.Is(err, libra.ErrRecordNotFound) {
			t.Fatal("rejected batch committed", err)
		}
	})
	t.Run("unkeyed_interleaving_before_receipt", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		unkeyed := pinnedWrite("interleaved-unkeyed", second.Add(118*time.Millisecond))
		candidate := pinnedWrite("interleaved-candidate", second.Add(119*time.Millisecond))
		err := g.appendBatchWithHook(ctx, []ResearchEventWrite{candidate}, "", func() error {
			_, err := g.store.PutResearchEventBatch(ctx, []ResearchEventWrite{unkeyed})
			return err
		})
		if err == nil {
			t.Fatal("batch laundered an unkeyed same-Store write between precheck and receipt")
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("interleaved unkeyed write remained READY")
		}
	})
}

func TestResearchIncrementalSortBatchCostV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_BATCH_COST_V1") != "1" {
		t.Skip("opt-in private batch publication cost diagnostic")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	control := createIncrementalSortGate(t, t.TempDir())
	candidate := createIncrementalSortGate(t, t.TempDir())
	defer control.close()
	defer candidate.close()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	rowCount := 3
	makeGroup := func(n int) []ResearchEventWrite {
		group := make([]ResearchEventWrite, n)
		for i := range group {
			group[i] = pinnedWrite(fmt.Sprintf("group-%04d", rowCount+i),
				second.Add(200*time.Millisecond+time.Duration(rowCount+i)*time.Microsecond))
		}
		return group
	}
	advanceBoth := func(group []ResearchEventWrite, singles bool) (time.Duration, time.Duration) {
		t.Helper()
		advanceControl := func() time.Duration {
			start := time.Now()
			if singles {
				for _, write := range group {
					if err := control.append(ctx, write, ""); err != nil {
						t.Fatal("single control append", rowCount, err)
					}
				}
			} else if err := control.appendBatch(ctx, group, ""); err != nil {
				t.Fatal("control growth batch", rowCount, err)
			}
			return time.Since(start)
		}
		advanceCandidate := func() time.Duration {
			start := time.Now()
			if err := candidate.appendBatch(ctx, group, ""); err != nil {
				t.Fatal("candidate batch append", rowCount, err)
			}
			return time.Since(start)
		}
		var controlTime, candidateTime time.Duration
		if (rowCount/16)%2 == 0 {
			controlTime, candidateTime = advanceControl(), advanceCandidate()
		} else {
			candidateTime, controlTime = advanceCandidate(), advanceControl()
		}
		rowCount += len(group)
		return controlTime, candidateTime
	}
	for _, target := range []int{259, 1027} {
		for rowCount < target {
			advanceBoth(makeGroup(min(16, target-rowCount)), false)
		}
		var controlNS, candidateNS, captureNS []int64
		var controlTotal, candidateTotal time.Duration
		for i := 0; i < 8; i++ {
			controlTime, candidateTime := advanceBoth(makeGroup(16), true)
			controlNS = append(controlNS, controlTime.Nanoseconds())
			candidateNS = append(candidateNS, candidateTime.Nanoseconds())
			controlTotal += controlTime
			candidateTotal += candidateTime
		}
		for i := 0; i < 100; i++ {
			start := time.Now()
			_, ok := candidate.capture(ctx)
			elapsed := time.Since(start)
			if !ok || candidate.wantRows != rowCount {
				t.Fatal("batch candidate lost READY", rowCount, candidate.wantRows, ok)
			}
			captureNS = append(captureNS, elapsed.Nanoseconds())
		}
		t.Logf("start_rows=%d end_rows=%d singles_group_p50=%s singles_group_p99=%s batch_group_p50=%s batch_group_p99=%s total_singles=%s total_batch=%s total_ratio=%.3f capture_p99=%s",
			target, rowCount, pinnedPercentile(controlNS, .5), pinnedPercentile(controlNS, .99),
			pinnedPercentile(candidateNS, .5), pinnedPercentile(candidateNS, .99),
			controlTotal, candidateTotal, float64(candidateTotal)/float64(controlTotal), pinnedPercentile(captureNS, .99))
	}
}

func TestResearchSortJournalCoexistenceV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_JOURNAL_COEXIST_V1") != "1" {
		t.Skip("opt-in private Recall journal coexistence probe")
	}
	ctx := context.Background()
	g := createIncrementalSortGate(t, t.TempDir())
	defer g.close()
	beforeLSN, ok := g.capture(ctx)
	if !ok {
		t.Fatal("initial gate not READY")
	}
	beforeSnapshot := g.store.Snapshot(ctx)
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	entry := model.BayesianJournalEntry{ID: "frontier-001", TenantID: "tenant-a",
		SessionID: "private-lab", AsOf: second.Add(120 * time.Millisecond),
		QueryDigest: "private-journal-probe", Snapshot: beforeSnapshot}
	if err := g.store.PutBayesianJournal(ctx, entry); err != nil {
		t.Fatal("valid Recall frontier journal", err)
	}
	if saved, err := g.store.GetBayesianJournal(ctx, entry.TenantID, entry.ID); err != nil || saved.ID != entry.ID {
		t.Fatal("frontier journal was not durable", saved.ID, err)
	}
	afterLSN, err := g.store.db.LatestCommitLSN(ctx)
	if err != nil {
		t.Fatal(err)
	}
	afterSnapshot := g.store.Snapshot(ctx)
	_, ready := g.capture(ctx)
	t.Logf("journal before_lsn=%d after_lsn=%d runtime_snapshot_unchanged=%v gate_ready=%v",
		beforeLSN, afterLSN, beforeSnapshot == afterSnapshot, ready)
	if !ready {
		t.Error("valid frontier journal made EventFrame publication unavailable")
	}
	write := pinnedWrite("post-journal-event", second.Add(110*time.Millisecond))
	if err := g.appendBatch(ctx, []ResearchEventWrite{write}, ""); err != nil {
		t.Error("valid frontier journal blocked later authorized EventFrame append:", err)
		return
	}
	ids, err := g.search(ctx, second.Add(120*time.Millisecond))
	if err != nil || len(ids) != 3 {
		t.Error("post-journal exact-LSN as-of search failed", ids, err)
	}
}
