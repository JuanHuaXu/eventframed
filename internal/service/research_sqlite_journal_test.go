package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

// This adapter exists only in tests. It deliberately does not supply
// cross-process ownership or a durable binding to the LibraVDB file identity.
type researchSQLiteJournalStore struct {
	store.EventStore
	guard interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	}
	db        *sql.DB
	closeOnce sync.Once
	closeErr  error
}

func researchSQLiteJournalSpan(ctx context.Context, name string, start time.Time) {
	if trace, ok := ctx.Value(recallProfileKey{}).(*recallProfileTrace); ok {
		trace.record(name, start, time.Now())
	}
}

func openResearchSQLiteJournalStore(base store.EventStore, path, owner string, create bool) (*researchSQLiteJournalStore, error) {
	guard, ok := base.(interface {
		WithResearchAsOfSnapshotWait(context.Context, model.Snapshot, time.Time, func() error) error
	})
	if !ok || !filepath.IsAbs(path) || owner == "" {
		return nil, errors.New("invalid research SQLite journal owner or path")
	}
	if create {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, err
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if err != nil {
			return nil, err
		}
		if err := f.Close(); err != nil {
			return nil, err
		}
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, fmt.Errorf("unsafe or missing research journal file: %w", err)
	}
	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	q.Set("mode", "rw")
	q.Set("_journal_mode", "WAL")
	q.Set("_synchronous", "FULL")
	q.Set("_busy_timeout", "5000")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	handedOff := false
	defer func() {
		if !handedOff {
			_ = db.Close()
		}
	}()
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if create {
		for _, statement := range []string{
			`CREATE TABLE journal_meta(id INTEGER PRIMARY KEY CHECK(id=1), owner TEXT NOT NULL, version INTEGER NOT NULL CHECK(version=1))`,
			`CREATE TABLE journals(tenant TEXT NOT NULL, id TEXT NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(tenant,id)) WITHOUT ROWID`,
		} {
			if _, err := db.Exec(statement); err != nil {
				return nil, err
			}
		}
		if _, err := db.Exec(`INSERT INTO journal_meta(id,owner,version) VALUES(1,?,1)`, owner); err != nil {
			return nil, err
		}
	}
	var recordedOwner string
	var version int
	if err := db.QueryRow(`SELECT owner,version FROM journal_meta WHERE id=1`).Scan(&recordedOwner, &version); err != nil || recordedOwner != owner || version != 1 {
		return nil, fmt.Errorf("research journal owner or schema mismatch: %w", err)
	}
	if _, err := db.Exec(`SELECT COUNT(*) FROM journals`); err != nil {
		return nil, err
	}
	var journalMode string
	var synchronous int
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		return nil, err
	}
	if err := db.QueryRow(`PRAGMA synchronous`).Scan(&synchronous); err != nil {
		return nil, err
	}
	if journalMode != "wal" || synchronous != 2 {
		return nil, fmt.Errorf("research journal lost WAL/FULL settings: mode=%s synchronous=%d", journalMode, synchronous)
	}
	handedOff = true
	return &researchSQLiteJournalStore{EventStore: base, guard: guard, db: db}, nil
}

func (s *researchSQLiteJournalStore) readRaw(ctx context.Context, tenant, id string) ([]byte, error) {
	var payload []byte
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM journals WHERE tenant=? AND id=?`, tenant, id).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, store.ErrJournalNotFound
	}
	return payload, err
}

func (s *researchSQLiteJournalStore) PutBayesianJournal(ctx context.Context, entry model.BayesianJournalEntry) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if entry.TenantID == "" || entry.ID == "" || entry.AsOf.IsZero() {
		return errors.New("invalid research journal entry")
	}
	encodeStart := time.Now()
	encoded, err := json.Marshal(entry)
	researchSQLiteJournalSpan(ctx, "sqlite_encode", encodeStart)
	if err != nil {
		return err
	}
	lookupStart := time.Now()
	prior, err := s.readRaw(ctx, entry.TenantID, entry.ID)
	researchSQLiteJournalSpan(ctx, "sqlite_lookup", lookupStart)
	if err == nil {
		if bytes.Equal(prior, encoded) {
			return nil
		}
		return store.ErrJournalConflict
	}
	if !errors.Is(err, store.ErrJournalNotFound) {
		return err
	}
	guardStart := time.Now()
	guardErr := s.guard.WithResearchAsOfSnapshotWait(ctx, entry.Snapshot, entry.AsOf, func() error {
		researchSQLiteJournalSpan(ctx, "sqlite_guard_wait", guardStart)
		insertStart := time.Now()
		result, insertErr := s.db.ExecContext(ctx, `INSERT INTO journals(tenant,id,payload) VALUES(?,?,?) ON CONFLICT(tenant,id) DO NOTHING`, entry.TenantID, entry.ID, encoded)
		researchSQLiteJournalSpan(ctx, "sqlite_insert", insertStart)
		if insertErr != nil {
			return insertErr
		}
		inserted, insertErr := result.RowsAffected()
		if insertErr != nil {
			return insertErr
		}
		if inserted == 1 {
			return nil
		}
		current, readErr := s.readRaw(ctx, entry.TenantID, entry.ID)
		if readErr != nil {
			return readErr
		}
		if !bytes.Equal(current, encoded) {
			return store.ErrJournalConflict
		}
		return nil
	})
	researchSQLiteJournalSpan(ctx, "sqlite_guard_total", guardStart)
	return guardErr
}

func (s *researchSQLiteJournalStore) GetBayesianJournal(ctx context.Context, tenant, id string) (model.BayesianJournalEntry, error) {
	encoded, err := s.readRaw(ctx, tenant, id)
	if err != nil {
		return model.BayesianJournalEntry{}, err
	}
	var entry model.BayesianJournalEntry
	if err := json.Unmarshal(encoded, &entry); err != nil {
		return model.BayesianJournalEntry{}, err
	}
	if entry.TenantID != tenant || entry.ID != id {
		return model.BayesianJournalEntry{}, errors.New("research journal identity mismatch")
	}
	return entry, nil
}

func (s *researchSQLiteJournalStore) count(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM journals`).Scan(&n)
	return n, err
}

func (s *researchSQLiteJournalStore) closeJournal() error {
	s.closeOnce.Do(func() { s.closeErr = s.db.Close() })
	return s.closeErr
}

func TestResearchSQLiteJournalContractV20(t *testing.T) {
	for _, motion := range []string{"future", "backfill", "policy"} {
		t.Run(motion, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			base, err := researchpublicationstore.New(memorystore.New())
			if err != nil {
				t.Fatal(err)
			}
			defer base.Close()
			path := filepath.Join(t.TempDir(), "journal.sqlite")
			if _, err := openResearchSQLiteJournalStore(base, path, "fixture-owner", false); err == nil {
				t.Fatal("missing journal reopened")
			}
			sidecar, err := openResearchSQLiteJournalStore(base, path, "fixture-owner", true)
			if err != nil {
				t.Fatal(err)
			}
			asOf := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			entry := model.BayesianJournalEntry{ID: "j1", TenantID: "tenant-a", SessionID: "session-a", AsOf: asOf, Snapshot: base.Snapshot(ctx)}
			switch motion {
			case "future", "backfill":
				at := asOf.Add(time.Hour)
				if motion == "backfill" {
					at = asOf.Add(-time.Hour)
				}
				event := testutil.Event("motion-event", "public test fixture", at)
				if _, err := base.Put(ctx, event, make([]float32, 8), "digest"); err != nil {
					t.Fatal(err)
				}
			case "policy":
				if _, err := base.BindBayesianPolicy(ctx, "changed-policy"); err != nil {
					t.Fatal(err)
				}
			}
			err = sidecar.PutBayesianJournal(ctx, entry)
			if motion != "future" {
				if err == nil {
					t.Fatal("invalid motion accepted")
				}
				if n, countErr := sidecar.count(ctx); countErr != nil || n != 0 {
					t.Fatalf("rejected journal committed: count=%d err=%v", n, countErr)
				}
				_ = sidecar.closeJournal()
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := sidecar.closeJournal(); err != nil {
				t.Fatal(err)
			}
			if _, err := openResearchSQLiteJournalStore(base, path, "wrong-owner", false); err == nil {
				t.Fatal("wrong owner reopened journal")
			}
			reopened, err := openResearchSQLiteJournalStore(base, path, "fixture-owner", false)
			if err != nil {
				t.Fatal(err)
			}
			defer reopened.closeJournal()
			got, err := reopened.GetBayesianJournal(ctx, entry.TenantID, entry.ID)
			if err != nil || got.SessionID != entry.SessionID {
				t.Fatalf("reopen lost acknowledged journal: %+v err=%v", got, err)
			}
			var group sync.WaitGroup
			retries := make(chan error, 16)
			for range 16 {
				group.Add(1)
				go func() {
					defer group.Done()
					retries <- reopened.PutBayesianJournal(ctx, entry)
				}()
			}
			group.Wait()
			close(retries)
			for retryErr := range retries {
				if retryErr != nil {
					t.Fatal(retryErr)
				}
			}
			entry.SessionID = "changed"
			if err := reopened.PutBayesianJournal(ctx, entry); !errors.Is(err, store.ErrJournalConflict) {
				t.Fatalf("changed journal retry: got %v", err)
			}
		})
	}
}
