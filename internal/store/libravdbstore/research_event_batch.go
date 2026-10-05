package libravdbstore

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	libra "github.com/xDarkicex/libravdb/libravdb"
)

// ResearchEventWrite is a pre-embedded event, not an OpenClaw capture contract.
// Callers must not mutate its event or vector while the call is in flight.
type ResearchEventWrite struct {
	Event  model.Event
	Vector []float32
	Digest string
}

// PutResearchEventBatch is an unwired, same-tenant transaction experiment.
// Results share the final snapshot; no intermediate snapshot is published.
// A commit error is an unknown outcome, not a promise of rollback. Before
// continuing after such an error, close/reopen and reconcile using event IDs.
// This deliberately does not implement request cancellation or a batching queue.
func (s *Store) PutResearchEventBatch(ctx context.Context, writes []ResearchEventWrite) ([]store.PutResult, error) {
	return s.putResearchEventBatch(ctx, writes, false, nil)
}

// PutResearchEventBatchSortable is a research-only candidate for a collection
// that declares available_at_sort as a string metadata field. The sort key is
// committed with the event and runtime state, never supplied by the caller.
func (s *Store) PutResearchEventBatchSortable(ctx context.Context, writes []ResearchEventWrite) ([]store.PutResult, error) {
	return s.putResearchEventBatch(ctx, writes, true, nil)
}

// PutResearchEventBatchSortableReceipt is an unwired single-process candidate.
// A zero LSN means every input was an exact duplicate and no commit occurred.
// A commit error can have an unknown outcome and requires reconciliation.
func (s *Store) PutResearchEventBatchSortableReceipt(ctx context.Context, writes []ResearchEventWrite) ([]store.PutResult, uint64, error) {
	var lsn uint64
	results, err := s.putResearchEventBatch(ctx, writes, true, &lsn)
	return results, lsn, err
}

func researchAvailabilitySortKey(at time.Time) string {
	return at.UTC().Format("2006-01-02T15:04:05.000000000Z")
}

func researchSortableColumnBindable(ctx context.Context, s *Store, name string) error {
	cat := s.db.Catalog()
	if cat == nil {
		return fmt.Errorf("sortable collection %q has no SQL catalog", name)
	}
	// EventFrame collection names are lowercase ASCII; LibraVDB hashes SQL
	// identifiers with FNV-1a after ASCII case folding.
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	table, err := cat.GetTable(h.Sum64())
	if err != nil {
		return fmt.Errorf("sortable collection %q is absent from SQL catalog: %w", name, err)
	}
	var idType, keyType uint16
	var hasID, hasKey bool
	for _, column := range cat.AllColumns(table) {
		switch column.Name {
		case "id":
			idType, hasID = column.Type, true
		case "available_at_sort":
			keyType, hasKey = column.Type, true
		}
	}
	if !hasID || !hasKey || keyType != idType {
		return fmt.Errorf("sortable collection %q requires a SQL text available_at_sort field", name)
	}
	lsn, err := s.db.LatestCommitLSN(ctx)
	if err != nil {
		return err
	}
	_, err = s.db.QueryWithParams(ctx,
		"SELECT id FROM "+name+" AS OF LSN $snapshot_lsn WHERE available_at_sort <= $available_by_sort LIMIT 0",
		libra.QueryParams{"snapshot_lsn": int64(lsn),
			"available_by_sort": researchAvailabilitySortKey(time.Unix(0, 0))})
	return err
}

func (s *Store) putResearchEventBatch(ctx context.Context, writes []ResearchEventWrite, sortable bool, receiptLSN *uint64) ([]store.PutResult, error) {
	if len(writes) == 0 || len(writes) > 16 {
		return nil, fmt.Errorf("research batch requires 1..16 events")
	}
	for _, w := range writes {
		if w.Event.TenantID != writes[0].Event.TenantID || w.Digest == "" || w.Event.Composition != nil {
			return nil, fmt.Errorf("research batch requires one tenant, digests, and ordinary events")
		}
		if err := w.Event.Validate(s.config.Dimension); err != nil {
			return nil, err
		}
		if sortable {
			year := w.Event.AvailableAt.UTC().Year()
			if year < 0 || year > 9999 {
				return nil, fmt.Errorf("sortable availability requires a four-digit UTC year")
			}
		}
		if len(w.Vector) != s.config.Dimension {
			return nil, fmt.Errorf("invalid vector dimension")
		}
		for _, v := range w.Vector {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return nil, fmt.Errorf("nonfinite vector")
			}
		}
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	s.eventMu.Lock()
	defer s.eventMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	collection, err := s.collection(ctx, writes[0].Event.TenantID)
	if err != nil {
		return nil, err
	}
	if sortable {
		if err := researchSortableColumnBindable(ctx, s, collectionName(writes[0].Event.TenantID, s.config.EmbeddingModel)); err != nil {
			return nil, fmt.Errorf("sortable event batch requires SQL-bound available_at_sort: %w", err)
		}
	}
	results := make([]store.PutResult, len(writes))
	seen := make(map[string]ResearchEventWrite, len(writes))
	indices := make([]int, 0, len(writes))
	metadata := make([]map[string]interface{}, 0, len(writes))
	for i, w := range writes {
		if prior, ok := seen[w.Event.ID]; ok {
			if prior.Digest != w.Digest || sortable && !prior.Event.AvailableAt.Equal(w.Event.AvailableAt) {
				return nil, store.ErrIdempotencyConflict
			}
			results[i].Duplicate = true
			continue
		}
		seen[w.Event.ID] = w
		record, err := collection.Get(ctx, w.Event.ID)
		if err == nil {
			if digest, _ := record.Metadata["content_digest"].(string); digest != w.Digest {
				return nil, store.ErrIdempotencyConflict
			}
			if sortable {
				stored, decodeErr := decodeStoredEvent(record.Metadata, true)
				if decodeErr != nil || stored.ID != w.Event.ID || stored.TenantID != w.Event.TenantID ||
					!stored.AvailableAt.Equal(w.Event.AvailableAt) {
					return nil, store.ErrIdempotencyConflict
				}
				if key, _ := record.Metadata["available_at_sort"].(string); key != researchAvailabilitySortKey(stored.AvailableAt) {
					return nil, store.ErrIdempotencyConflict
				}
			}
			results[i].Duplicate = true
			continue
		}
		if !errors.Is(err, libra.ErrRecordNotFound) {
			return nil, err
		}
		payload, err := encodeStoredEvent(w.Event)
		if err != nil {
			return nil, err
		}
		indices = append(indices, i)
		entry := map[string]interface{}{
			"event_json": string(payload), "corpus_text": w.Event.FrameText(),
			"raw_content": w.Event.Content, "content_digest": w.Digest,
			"session_id": w.Event.SessionID, "kind": w.Event.Kind,
			"available_at": w.Event.AvailableAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"),
			"priority":     w.Event.Priority,
		}
		if sortable {
			entry["available_at_sort"] = researchAvailabilitySortKey(w.Event.AvailableAt)
		}
		metadata = append(metadata, entry)
	}
	next := s.snapshot
	if uint64(len(indices)) > math.MaxUint64-next.RuntimeVersion || uint64(len(indices)) > math.MaxUint64-next.EvidenceEpoch {
		return nil, fmt.Errorf("snapshot counter overflow")
	}
	next.RuntimeVersion += uint64(len(indices))
	next.EvidenceEpoch += uint64(len(indices))
	if len(indices) > 0 {
		state, err := s.stateMetadata(next)
		if err != nil {
			return nil, err
		}
		key := collectionName(writes[0].Event.TenantID, s.config.EmbeddingModel)
		commit := func(tx libra.Tx) error {
			for j, i := range indices {
				w := writes[i]
				if err := tx.Insert(ctx, key, w.Event.ID, w.Vector, metadata[j]); err != nil {
					return err
				}
			}
			return tx.Upsert(ctx, systemCollection, "runtime", nil, state)
		}
		if receiptLSN == nil {
			err = s.db.WithTx(ctx, commit)
		} else {
			var receipt libra.CommitReceipt
			receipt, err = s.db.WithTxReceipt(ctx, func(tx libra.ReceiptTx) error { return commit(tx) })
			if err == nil {
				if receipt.CommitLSN == 0 {
					return nil, fmt.Errorf("research batch committed without an exact receipt (reconcile before reuse)")
				}
				*receiptLSN = receipt.CommitLSN
			}
		}
		if err != nil {
			return nil, fmt.Errorf("research batch commit (reconcile before reuse): %w", err)
		}
		for j, i := range indices {
			recordIngestMotion(s.ingestMotion, s.snapshot.RuntimeVersion+uint64(j)+1, writes[i].Event.AvailableAt)
		}
		s.snapshot = next
	}
	for i := range results {
		results[i].Snapshot = next
	}
	return results, nil
}
