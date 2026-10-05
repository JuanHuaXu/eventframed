package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

type incrementalSortGate struct{ *sortPublicationGate }

func incrementalRowHash(id string, metadata map[string]interface{}, vector []float32) ([32]byte, error) {
	event, err := decodeStoredEvent(metadata, true)
	if err != nil {
		return [32]byte{}, err
	}
	if event.ID != id || event.TenantID != "tenant-a" {
		return [32]byte{}, fmt.Errorf("invalid durable EventFrame identity %q", id)
	}
	key, ok := metadata["available_at_sort"].(string)
	if !ok || key != researchAvailabilitySortKey(event.AvailableAt) {
		return [32]byte{}, fmt.Errorf("invalid durable sort key for %q", id)
	}
	payload, ok := metadata["event_json"].(string)
	if !ok {
		return [32]byte{}, fmt.Errorf("missing durable payload for %q", id)
	}
	encoded, err := json.Marshal(struct {
		ID, Payload, Key string
		Vector           []float32
	}{id, payload, key, vector})
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(encoded), nil
}

func incrementalChainNext(prior, row [32]byte, lsn, seq uint64) [32]byte {
	var input [80]byte
	copy(input[:32], prior[:])
	copy(input[32:64], row[:])
	binary.BigEndian.PutUint64(input[64:72], lsn)
	binary.BigEndian.PutUint64(input[72:80], seq)
	return sha256.Sum256(input[:])
}

func createIncrementalSortGate(t *testing.T, root string) *incrementalSortGate {
	t.Helper()
	ctx := context.Background()
	g, err := openSortPublicationGate(root, true)
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.store.db.CreateCollection(ctx, g.name, libra.WithDimension(4),
		libra.WithMetric(libra.CosineDistance), libra.WithHNSW(16, 200, 100),
		libra.WithMemoryMapping(true),
		libra.WithMetadataSchema(libra.MetadataSchema{
			"available_at": libra.StringField, "available_at_sort": libra.StringField,
		}))
	if err != nil {
		g.close()
		t.Fatal(err)
	}
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	writes := []ResearchEventWrite{
		pinnedWrite("past100", second.Add(100*time.Millisecond)),
		pinnedWrite("at120", second.Add(120*time.Millisecond)),
		pinnedWrite("future125", second.Add(125*time.Millisecond)),
	}
	if _, lsn, err := g.store.PutResearchEventBatchSortableReceipt(ctx, writes); err != nil || lsn == 0 {
		g.close()
		t.Fatal("genesis rows", lsn, err)
	}
	if err := g.publish(ctx); err != nil {
		g.close()
		t.Fatal("genesis publication", err)
	}
	gate := &incrementalSortGate{g}
	if err := gate.initJournal(ctx); err != nil {
		g.close()
		t.Fatal("genesis journal", err)
	}
	return gate
}

func (g *incrementalSortGate) initJournal(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	marker, err := g.readMarker(ctx)
	if err != nil {
		return err
	}
	if marker.phase != "READY" || marker.digest != g.verified || marker.lsn != g.verifiedLSN {
		return errors.New("genesis marker is not verified")
	}
	col, err := g.store.db.GetCollection(g.name)
	if err != nil {
		return err
	}
	records, err := col.ListAll(ctx)
	if err != nil {
		return err
	}
	if len(records) != marker.count {
		return fmt.Errorf("genesis row count mismatch: %d != %d", len(records), marker.count)
	}
	if _, err := g.sidecar.ExecContext(ctx, `CREATE TABLE journal(
		seq INTEGER PRIMARY KEY, id TEXT NOT NULL UNIQUE,
		row_hash BLOB NOT NULL, lsn INTEGER NOT NULL)`); err != nil {
		return err
	}
	// The genesis order is fixed by ID; later entries follow receipt order.
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	tx, err := g.sidecar.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var chain [32]byte
	for i, record := range records {
		row, err := incrementalRowHash(record.ID, record.Metadata, record.Vector)
		if err != nil {
			return err
		}
		seq := uint64(i + 1)
		chain = incrementalChainNext(chain, row, marker.lsn, seq)
		if _, err := tx.ExecContext(ctx, "INSERT INTO journal(seq,id,row_hash,lsn) VALUES(?,?,?,?)",
			seq, record.ID, row[:], marker.lsn); err != nil {
			return err
		}
	}
	root := hex.EncodeToString(chain[:])
	result, err := tx.ExecContext(ctx, "UPDATE marker SET digest=? WHERE id=1 AND phase='READY' AND lsn=? AND row_count=? AND digest=?",
		root, marker.lsn, marker.count, marker.digest)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected != 1 {
		return fmt.Errorf("genesis marker update affected %d rows", affected)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	g.verified = root
	return nil
}

func openIncrementalSortGate(root string) (*incrementalSortGate, error) {
	g, err := openSortPublicationGate(root, false)
	if err != nil {
		return nil, err
	}
	gate := &incrementalSortGate{g}
	marker, err := gate.readMarker(context.Background())
	if err != nil {
		g.close()
		return nil, err
	}
	lsn, err := g.store.db.LatestCommitLSN(context.Background())
	if err != nil {
		g.close()
		return nil, err
	}
	if marker.phase != "READY" || marker.lsn != lsn || marker.count < 3 {
		g.verified, g.verifiedLSN = "", 0
		return gate, nil
	}
	rootHash, count, err := gate.verifyJournal(context.Background())
	if err == nil && count == marker.count && rootHash == marker.digest {
		g.wantRows = count
		g.verified = rootHash
		g.verifiedLSN = lsn
	} else {
		g.verified, g.verifiedLSN = "", 0
	}
	return gate, nil
}

func (g *incrementalSortGate) verifyJournal(ctx context.Context) (string, int, error) {
	col, err := g.store.db.GetCollection(g.name)
	if err != nil {
		return "", 0, err
	}
	rows, err := g.sidecar.QueryContext(ctx, "SELECT seq,id,row_hash,lsn FROM journal ORDER BY seq")
	if err != nil {
		return "", 0, err
	}
	defer rows.Close()
	var chain [32]byte
	count := 0
	for rows.Next() {
		var seq, lsn int64
		var id string
		var storedHash []byte
		if err := rows.Scan(&seq, &id, &storedHash, &lsn); err != nil {
			return "", 0, err
		}
		count++
		if seq != int64(count) || lsn <= 0 || len(storedHash) != sha256.Size {
			return "", 0, fmt.Errorf("invalid journal entry %d", count)
		}
		record, err := col.Get(ctx, id)
		if err != nil {
			return "", 0, err
		}
		actual, err := incrementalRowHash(record.ID, record.Metadata, record.Vector)
		if err != nil {
			return "", 0, err
		}
		if !equalHash(storedHash, actual) {
			return "", 0, fmt.Errorf("journal row hash mismatch for %q", id)
		}
		chain = incrementalChainNext(chain, actual, uint64(lsn), uint64(seq))
	}
	if err := rows.Err(); err != nil {
		return "", 0, err
	}
	all, err := col.ListAll(ctx)
	if err != nil {
		return "", 0, err
	}
	if len(all) != count {
		return "", 0, fmt.Errorf("journal/collection row count mismatch: %d != %d", count, len(all))
	}
	return hex.EncodeToString(chain[:]), count, nil
}

func equalHash(stored []byte, actual [32]byte) bool {
	if len(stored) != len(actual) {
		return false
	}
	for i := range stored {
		if stored[i] != actual[i] {
			return false
		}
	}
	return true
}

func (g *incrementalSortGate) append(ctx context.Context, write ResearchEventWrite, stopAt string) error {
	return g.appendWithHook(ctx, write, stopAt, nil)
}

func (g *incrementalSortGate) appendWithHook(ctx context.Context, write ResearchEventWrite, stopAt string, afterPrecheck func() error) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	marker, err := g.readMarker(ctx)
	if err != nil {
		return err
	}
	if g.verified == "" || marker.phase != "READY" || marker.digest != g.verified ||
		marker.lsn != g.verifiedLSN || marker.count != g.wantRows {
		return errors.New("incremental gate is not current")
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
		return fmt.Errorf("LibraVDB moved outside the publication gate: %d != %d", before, marker.lsn)
	}
	if afterPrecheck != nil {
		if err := afterPrecheck(); err != nil {
			return err
		}
	}
	results, receipt, err := g.store.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{write})
	if err != nil {
		g.verified = ""
		return err
	}
	if receipt == 0 {
		if len(results) != 1 || !results[0].Duplicate {
			g.verified = ""
			return errors.New("zero receipt without an exact duplicate")
		}
		g.store.writeMu.RLock()
		afterSnapshot := g.store.snapshot
		after, err := g.store.db.LatestCommitLSN(ctx)
		g.store.writeMu.RUnlock()
		if err != nil || after != before || afterSnapshot != beforeSnapshot {
			g.verified = ""
			return fmt.Errorf("duplicate-only append moved LibraVDB state: LSN %d -> %d: %v", before, after, err)
		}
		return nil
	}
	g.store.writeMu.RLock()
	afterSnapshot := g.store.snapshot
	latest, err := g.store.db.LatestCommitLSN(ctx)
	g.store.writeMu.RUnlock()
	if err != nil {
		g.verified = ""
		return err
	}
	if latest != receipt || len(results) != 1 || results[0].Duplicate {
		g.verified = ""
		return fmt.Errorf("receipt is not the exact latest commit: %d != %d", receipt, latest)
	}
	expectedSnapshot := beforeSnapshot
	expectedSnapshot.RuntimeVersion++
	expectedSnapshot.EvidenceEpoch++
	if afterSnapshot != expectedSnapshot {
		g.verified = ""
		return fmt.Errorf("single append snapshot moved outside journaled event: before=%+v after=%+v",
			beforeSnapshot, afterSnapshot)
	}
	col, err := g.store.db.GetCollection(g.name)
	if err != nil {
		g.verified = ""
		return err
	}
	record, err := col.Get(ctx, write.Event.ID)
	if err != nil {
		g.verified = ""
		return err
	}
	row, err := incrementalRowHash(record.ID, record.Metadata, record.Vector)
	if err != nil {
		g.verified = ""
		return err
	}
	if stopAt == "after_db" {
		return errors.New("injected interruption after LibraVDB commit")
	}
	prior, err := hex.DecodeString(marker.digest)
	if err != nil {
		g.verified = ""
		return err
	}
	if len(prior) != sha256.Size {
		g.verified = ""
		return fmt.Errorf("invalid prior chain length: %d", len(prior))
	}
	var priorHash [32]byte
	copy(priorHash[:], prior)
	seq := uint64(marker.count + 1)
	root := incrementalChainNext(priorHash, row, receipt, seq)
	rootText := hex.EncodeToString(root[:])
	tx, err := g.sidecar.BeginTx(ctx, nil)
	if err != nil {
		g.verified = ""
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "INSERT INTO journal(seq,id,row_hash,lsn) VALUES(?,?,?,?)",
		seq, write.Event.ID, row[:], receipt); err != nil {
		g.verified = ""
		return err
	}
	result, err := tx.ExecContext(ctx, "UPDATE marker SET lsn=?,row_count=?,digest=? WHERE id=1 AND phase='READY' AND lsn=? AND row_count=? AND digest=?",
		receipt, seq, rootText, marker.lsn, marker.count, marker.digest)
	if err != nil {
		g.verified = ""
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		g.verified = ""
		return err
	} else if affected != 1 {
		g.verified = ""
		return fmt.Errorf("incremental marker update affected %d rows", affected)
	}
	if err := tx.Commit(); err != nil {
		g.verified = ""
		return err
	}
	if stopAt == "after_sqlite" {
		return errors.New("injected interruption after SQLite commit")
	}
	g.verified, g.verifiedLSN, g.wantRows = rootText, receipt, int(seq)
	return nil
}

func TestResearchIncrementalSortPublicationV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_INCREMENTAL_V1") != "1" {
		t.Skip("opt-in private incremental sort-key publication test")
	}
	ctx := context.Background()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	newGate := func(t *testing.T) (*incrementalSortGate, string) {
		t.Helper()
		root := t.TempDir()
		return createIncrementalSortGate(t, root), root
	}
	reopen := func(t *testing.T, g *incrementalSortGate, root string) *incrementalSortGate {
		t.Helper()
		if err := g.close(); err != nil {
			t.Fatal(err)
		}
		fresh, err := openIncrementalSortGate(root)
		if err != nil {
			t.Fatal(err)
		}
		return fresh
	}
	t.Run("normal", func(t *testing.T) {
		g, root := newGate(t)
		defer func() { g.close() }()
		middle := pinnedWrite("middle110", second.Add(110*time.Millisecond))
		future := pinnedWrite("future130", second.Add(130*time.Millisecond))
		for _, write := range []ResearchEventWrite{middle, future} {
			if err := g.append(ctx, write, ""); err != nil {
				t.Fatal("authorized append", err)
			}
		}
		marker, err := g.readMarker(ctx)
		if err != nil || marker.count != 5 {
			t.Fatal("marker did not advance for two rows", marker, err)
		}
		if err := g.append(ctx, middle, ""); err != nil {
			t.Fatal("exact duplicate", err)
		}
		if after, err := g.readMarker(ctx); err != nil || after != marker {
			t.Fatal("duplicate changed marker", after, err)
		}
		ids, err := g.search(ctx, second.Add(120*time.Millisecond))
		if err != nil || len(ids) != 3 {
			t.Fatal("incorrect incremental as-of set", ids, err)
		}
		seen := map[string]bool{}
		for _, id := range ids {
			seen[id] = true
		}
		if !seen["past100"] || !seen["middle110"] || !seen["at120"] || seen["future125"] || seen["future130"] {
			t.Fatal("future leakage or past omission", ids)
		}
		g = reopen(t, g, root)
		if _, ok := g.capture(ctx); !ok {
			t.Fatal("reopened journal was not verified")
		}
		ids, err = g.search(ctx, second.Add(120*time.Millisecond))
		if err != nil || len(ids) != 3 {
			t.Fatal("reopened as-of query changed", ids, err)
		}
		seen = map[string]bool{}
		for _, id := range ids {
			seen[id] = true
		}
		if !seen["past100"] || !seen["middle110"] || !seen["at120"] || seen["future125"] || seen["future130"] {
			t.Fatal("reopened as-of set leaked or omitted rows", ids)
		}
	})
	for _, cut := range []string{"after_db", "after_sqlite"} {
		t.Run(cut, func(t *testing.T) {
			g, root := newGate(t)
			defer func() { g.close() }()
			write := pinnedWrite("interrupted-"+cut, second.Add(111*time.Millisecond))
			if err := g.append(ctx, write, cut); err == nil {
				t.Fatal("injection did not interrupt")
			}
			if _, ok := g.capture(ctx); ok {
				t.Fatal("current gate accepted interrupted publication")
			}
			g = reopen(t, g, root)
			_, ok := g.capture(ctx)
			if cut == "after_db" && ok || cut == "after_sqlite" && !ok {
				t.Fatal("reopened interrupted publication had wrong READY state", cut, ok)
			}
		})
	}
	t.Run("journal_tamper", func(t *testing.T) {
		g, root := newGate(t)
		defer func() { g.close() }()
		if _, err := g.sidecar.ExecContext(ctx, "UPDATE journal SET row_hash=? WHERE seq=1", make([]byte, 32)); err != nil {
			t.Fatal(err)
		}
		g = reopen(t, g, root)
		if _, ok := g.capture(ctx); ok {
			t.Fatal("tampered journal was verified")
		}
	})
	t.Run("legacy_write", func(t *testing.T) {
		g, _ := newGate(t)
		defer g.close()
		legacy := pinnedWrite("unkeyed", second.Add(112*time.Millisecond))
		if _, err := g.store.PutResearchEventBatch(ctx, []ResearchEventWrite{legacy}); err != nil {
			t.Fatal(err)
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("legacy write did not stale READY")
		}
		candidate := pinnedWrite("would-launder", second.Add(113*time.Millisecond))
		if err := g.append(ctx, candidate, ""); err == nil {
			t.Fatal("unpublished legacy write was laundered")
		}
		col, err := g.store.db.GetCollection(g.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := col.Get(ctx, candidate.Event.ID); !errors.Is(err, libra.ErrRecordNotFound) {
			t.Fatal("rejected append committed", err)
		}
	})
	t.Run("bypassed_sortable_write", func(t *testing.T) {
		g, root := newGate(t)
		defer func() { g.close() }()
		bypass := pinnedWrite("sort-bypass", second.Add(114*time.Millisecond))
		if _, receipt, err := g.store.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{bypass}); err != nil || receipt == 0 {
			t.Fatal("direct sortable bypass", receipt, err)
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("direct sortable write did not stale READY")
		}
		candidate := pinnedWrite("after-bypass", second.Add(115*time.Millisecond))
		if err := g.append(ctx, candidate, ""); err == nil {
			t.Fatal("gate silently adopted a bypassed sortable write")
		}
		g = reopen(t, g, root)
		if _, ok := g.capture(ctx); ok {
			t.Fatal("reopen accepted an unjournaled sortable write")
		}
	})
	t.Run("unkeyed_interleaving_before_single_receipt", func(t *testing.T) {
		g, _ := newGate(t)
		defer g.close()
		unkeyed := pinnedWrite("single-interleaved-unkeyed", second.Add(116*time.Millisecond))
		candidate := pinnedWrite("single-interleaved-candidate", second.Add(117*time.Millisecond))
		err := g.appendWithHook(ctx, candidate, "", func() error {
			_, err := g.store.PutResearchEventBatch(ctx, []ResearchEventWrite{unkeyed})
			return err
		})
		if err == nil {
			t.Fatal("single append laundered an unkeyed same-Store write between precheck and receipt")
		}
		if _, ok := g.capture(ctx); ok {
			t.Fatal("interleaved unkeyed write remained READY")
		}
	})
}

func TestResearchIncrementalSortPublicationCostV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_INCREMENTAL_COST_V1") != "1" {
		t.Skip("opt-in private incremental publication cost test")
	}
	ctx := context.Background()
	g := createIncrementalSortGate(t, t.TempDir())
	defer g.close()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	var appendNS, captureNS []int64
	for i := 0; i < 116; i++ {
		write := pinnedWrite(fmt.Sprintf("cost-%03d", i), second.Add(time.Duration(i+200)*time.Millisecond))
		start := time.Now()
		if err := g.append(ctx, write, ""); err != nil {
			t.Fatal(err)
		}
		if i >= 16 {
			appendNS = append(appendNS, time.Since(start).Nanoseconds())
		}
	}
	for i := 0; i < 100; i++ {
		start := time.Now()
		if _, ok := g.capture(ctx); !ok {
			t.Fatal("cost gate denied READY")
		}
		captureNS = append(captureNS, time.Since(start).Nanoseconds())
	}
	t.Logf("append p50=%s p99=%s capture p50=%s p99=%s",
		pinnedPercentile(appendNS, .5), pinnedPercentile(appendNS, .99),
		pinnedPercentile(captureNS, .5), pinnedPercentile(captureNS, .99))
	if pinnedPercentile(appendNS, .99) > 100*time.Millisecond || pinnedPercentile(captureNS, .99) > time.Millisecond {
		t.Error("incremental publication exceeded frozen isolated ceilings")
	}
}

func TestResearchIncrementalSortStartupScaleV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_STARTUP_SCALE_V1") != "1" {
		t.Skip("opt-in private incremental publication startup diagnostic")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	root := t.TempDir()
	g := createIncrementalSortGate(t, root)
	defer func() {
		if g != nil {
			g.close()
		}
	}()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	rowCount := 3
	for _, target := range []int{35, 259, 1027} {
		for rowCount < target {
			write := pinnedWrite(fmt.Sprintf("startup-%04d", rowCount),
				second.Add(200*time.Millisecond+time.Duration(rowCount)*time.Microsecond))
			if err := g.append(ctx, write, ""); err != nil {
				t.Fatal("grow private startup fixture", rowCount, err)
			}
			rowCount++
		}
		if err := g.close(); err != nil {
			t.Fatal("close startup fixture", err)
		}
		g = nil
		var controls, candidates []int64
		runControl := func() time.Duration {
			t.Helper()
			start := time.Now()
			base, err := openSortPublicationGate(root, false)
			elapsed := time.Since(start)
			if err != nil {
				t.Fatal("control reopen", err)
			}
			if err := base.close(); err != nil {
				t.Fatal("control close", err)
			}
			return elapsed
		}
		runCandidate := func() time.Duration {
			t.Helper()
			start := time.Now()
			verified, err := openIncrementalSortGate(root)
			elapsed := time.Since(start)
			if err != nil {
				t.Fatal("verified reopen", err)
			}
			if _, ok := verified.capture(ctx); !ok || verified.wantRows != target {
				verified.close()
				t.Fatal("verified startup rejected or lost rows", target, verified.wantRows, ok)
			}
			if err := verified.close(); err != nil {
				t.Fatal("verified close", err)
			}
			return elapsed
		}
		for pair := 0; pair < 5; pair++ {
			var control, candidate time.Duration
			first := "control"
			if pair%2 == 0 {
				control, candidate = runControl(), runCandidate()
			} else {
				first = "candidate"
				candidate, control = runCandidate(), runControl()
			}
			controls = append(controls, control.Nanoseconds())
			candidates = append(candidates, candidate.Nanoseconds())
			t.Logf("rows=%d pair=%d first=%s control=%s verified=%s ratio=%.3f",
				target, pair, first, control, candidate, float64(candidate)/float64(control))
		}
		t.Logf("rows=%d control_p50=%s control_max=%s verified_p50=%s verified_max=%s p50_ratio=%.3f",
			target, pinnedPercentile(controls, .5), pinnedPercentile(controls, 1),
			pinnedPercentile(candidates, .5), pinnedPercentile(candidates, 1),
			float64(pinnedPercentile(candidates, .5))/float64(pinnedPercentile(controls, .5)))
		var err error
		g, err = openIncrementalSortGate(root)
		if err != nil {
			t.Fatal("resume owner after diagnostic reopens", err)
		}
	}
}

func TestResearchIncrementalSortAppendScaleV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_APPEND_SCALE_V1") != "1" {
		t.Skip("opt-in private incremental publication append scale diagnostic")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	control := createIncrementalSortGate(t, t.TempDir())
	candidate := createIncrementalSortGate(t, t.TempDir())
	defer control.close()
	defer candidate.close()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	rowCount := 3
	writePair := func(record bool, controlNS, candidateNS *[]int64) {
		t.Helper()
		write := pinnedWrite(fmt.Sprintf("append-scale-%04d", rowCount),
			second.Add(200*time.Millisecond+time.Duration(rowCount)*time.Microsecond))
		writeControl := func() {
			start := time.Now()
			outcomes, receipt, err := control.store.PutResearchEventBatchSortableReceipt(ctx, []ResearchEventWrite{write})
			elapsed := time.Since(start)
			if err != nil || receipt == 0 || len(outcomes) != 1 || outcomes[0].Duplicate {
				t.Fatal("raw sortable receipt write", rowCount, receipt, outcomes, err)
			}
			if record {
				*controlNS = append(*controlNS, elapsed.Nanoseconds())
			}
		}
		writeCandidate := func() {
			start := time.Now()
			err := candidate.append(ctx, write, "")
			elapsed := time.Since(start)
			if err != nil {
				t.Fatal("incremental published write", rowCount, err)
			}
			if record {
				*candidateNS = append(*candidateNS, elapsed.Nanoseconds())
			}
		}
		if rowCount%2 == 0 {
			writeControl()
			writeCandidate()
		} else {
			writeCandidate()
			writeControl()
		}
		rowCount++
	}
	for _, target := range []int{35, 259, 1027} {
		for rowCount < target {
			writePair(false, nil, nil)
		}
		var controlNS, candidateNS, captureNS []int64
		for i := 0; i < 100; i++ {
			writePair(true, &controlNS, &candidateNS)
		}
		for i := 0; i < 100; i++ {
			start := time.Now()
			_, ok := candidate.capture(ctx)
			elapsed := time.Since(start)
			if !ok || candidate.wantRows != rowCount {
				t.Fatal("published gate lost READY or row count", target, rowCount, candidate.wantRows, ok)
			}
			captureNS = append(captureNS, elapsed.Nanoseconds())
		}
		controlP50, controlP99 := pinnedPercentile(controlNS, .5), pinnedPercentile(controlNS, .99)
		candidateP50, candidateP99 := pinnedPercentile(candidateNS, .5), pinnedPercentile(candidateNS, .99)
		t.Logf("start_rows=%d end_rows=%d control_p50=%s control_p99=%s control_max=%s candidate_p50=%s candidate_p99=%s candidate_max=%s p50_ratio=%.3f p99_ratio=%.3f capture_p50=%s capture_p99=%s",
			target, rowCount, controlP50, controlP99, pinnedPercentile(controlNS, 1),
			candidateP50, candidateP99, pinnedPercentile(candidateNS, 1),
			float64(candidateP50)/float64(controlP50), float64(candidateP99)/float64(controlP99),
			pinnedPercentile(captureNS, .5), pinnedPercentile(captureNS, .99))
	}
}
