package libravdbstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	libra "github.com/xDarkicex/libravdb/libravdb"
)

type sortPublishedView struct {
	lsn uint64
	at  time.Time
}

// sortPublishedReader reads only a journal-certified LSN. It is valid only
// under the declared one-owner, gate-only writer contract.
type sortPublishedReader struct {
	gate    *incrementalSortGate
	current atomic.Pointer[sortPublishedView]
}

func newSortPublishedReader(ctx context.Context, gate *incrementalSortGate) (*sortPublishedReader, error) {
	r := &sortPublishedReader{gate: gate}
	if err := r.publish(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *sortPublishedReader) publish(ctx context.Context) error {
	lsn, ok := r.gate.capture(ctx)
	if !ok {
		r.current.Store(nil)
		return errors.New("no fully published journal view")
	}
	r.current.Store(&sortPublishedView{lsn: lsn, at: time.Now()})
	return nil
}

func (r *sortPublishedReader) poison() { r.current.Store(nil) }

func (r *sortPublishedReader) search(ctx context.Context, at time.Time) ([]string, time.Duration, error) {
	view := r.current.Load()
	if view == nil {
		return nil, 0, errors.New("no certified published view")
	}
	age := time.Since(view.at)
	if age < 0 || age >= 250*time.Millisecond {
		return nil, age, fmt.Errorf("certified view age %s exceeds 250 ms", age)
	}
	lease, err := r.gate.store.db.SnapshotAtLSN(ctx, view.lsn)
	if err != nil {
		return nil, age, err
	}
	defer lease.Close()
	query := "SELECT id, embedding <-> $query_vec AS distance FROM " + r.gate.name +
		" AS OF LSN $snapshot_lsn d WHERE available_at_sort <= $available_by_sort ORDER BY distance LIMIT 10"
	rows, err := r.gate.store.db.QueryWithParams(ctx, query, libra.QueryParams{
		"snapshot_lsn": int64(view.lsn), "query_vec": []float32{1, 0, 0, 0},
		"available_by_sort": researchAvailabilitySortKey(at),
	})
	if err != nil {
		return nil, age, err
	}
	ids := make([]string, len(rows.Results))
	for i, row := range rows.Results {
		ids[i] = row.ID
	}
	age = time.Since(view.at)
	if age >= 250*time.Millisecond || r.current.Load() == nil {
		return nil, age, errors.New("published view expired or was poisoned during read")
	}
	return ids, age, nil
}

func TestResearchSortPublishedViewV1(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_SORT_PUBLISHED_VIEW_V1") != "1" {
		t.Skip("opt-in private published-view lifecycle contract")
	}
	ctx := context.Background()
	second := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	asOf := second.Add(120 * time.Millisecond)
	t.Run("old_view_during_db_only_commit_and_poison", func(t *testing.T) {
		root := t.TempDir()
		g := createIncrementalSortGate(t, root)
		defer func() {
			if g != nil {
				g.close()
			}
		}()
		r, err := newSortPublishedReader(ctx, g)
		if err != nil {
			t.Fatal(err)
		}
		old := r.current.Load().lsn
		newWrite := pinnedWrite("pending-early", second.Add(110*time.Millisecond))
		if err := g.appendBatch(ctx, []ResearchEventWrite{newWrite}, "after_db"); err == nil {
			t.Fatal("DB-only interruption did not fire")
		}
		ids, age, err := r.search(ctx, asOf)
		if err != nil || age >= 250*time.Millisecond || len(ids) != 2 {
			t.Fatal("old certified LSN was not readable during pending publication", ids, age, err)
		}
		for _, id := range ids {
			if strings.HasPrefix(id, "pending-") || strings.HasPrefix(id, "future") {
				t.Fatal("unpublished or future event leaked through old LSN", ids)
			}
		}
		if current := r.current.Load(); current == nil || current.lsn != old {
			t.Fatal("DB-only commit advanced published pointer", current)
		}
		r.poison()
		if _, _, err := r.search(ctx, asOf); err == nil {
			t.Fatal("poisoned view remained readable")
		}
		if err := g.close(); err != nil {
			t.Fatal(err)
		}
		g = nil
		g, err = openIncrementalSortGate(root)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := newSortPublishedReader(ctx, g); err == nil {
			t.Fatal("DB-only interrupted journal published on reopen")
		}
	})
	t.Run("expiry", func(t *testing.T) {
		g := createIncrementalSortGate(t, t.TempDir())
		defer g.close()
		r, err := newSortPublishedReader(ctx, g)
		if err != nil {
			t.Fatal(err)
		}
		view := r.current.Load()
		r.current.Store(&sortPublishedView{lsn: view.lsn, at: time.Now().Add(-251 * time.Millisecond)})
		if _, _, err := r.search(ctx, asOf); err == nil {
			t.Fatal("expired published view was accepted")
		}
	})
	t.Run("complete_batch_and_reopen", func(t *testing.T) {
		root := t.TempDir()
		g := createIncrementalSortGate(t, root)
		defer func() {
			if g != nil {
				g.close()
			}
		}()
		r, err := newSortPublishedReader(ctx, g)
		if err != nil {
			t.Fatal(err)
		}
		newWrite := pinnedWrite("published-early", second.Add(110*time.Millisecond))
		if err := g.appendBatch(ctx, []ResearchEventWrite{newWrite}, ""); err != nil {
			t.Fatal(err)
		}
		if err := r.publish(ctx); err != nil {
			t.Fatal(err)
		}
		ids, _, err := r.search(ctx, asOf)
		if err != nil || len(ids) != 3 {
			t.Fatal("new published view unreadable", ids, err)
		}
		if err := g.close(); err != nil {
			t.Fatal(err)
		}
		g = nil
		g, err = openIncrementalSortGate(root)
		if err != nil {
			t.Fatal(err)
		}
		r, err = newSortPublishedReader(ctx, g)
		if err != nil {
			t.Fatal("fully committed batch failed re-verification", err)
		}
		ids, _, err = r.search(ctx, asOf)
		if err != nil || len(ids) != 3 {
			t.Fatal("reopened published view unreadable", ids, err)
		}
	})
}
