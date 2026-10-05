// Package researchlineage stores opt-in source mutation history for research.
// It is not a production store migration or a cross-database transaction.
package researchlineage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	_ "modernc.org/sqlite"
)

type EventKey struct{ Tenant, Event string }

type Ledger struct {
	mu        sync.Mutex
	db        *sql.DB
	release   func() error
	initial   model.Snapshot
	committed model.Snapshot
	closed    bool
}

// Create starts history only for a genuinely new sidecar. An absent old
// sidecar must never be interpreted as permission to recreate lost history.
func Create(path string, current model.Snapshot) (*Ledger, error) {
	return open(path, current, true)
}

// Open resumes only when the backend still matches the last durable checkpoint.
func Open(path string, current model.Snapshot) (*Ledger, error) {
	return open(path, current, false)
}

func open(path string, current model.Snapshot, create bool) (_ *Ledger, err error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("absolute research lineage path required")
	}
	if create {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return nil, err
		}
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	path = filepath.Join(parent, filepath.Base(path))
	if resolved, resolveErr := filepath.EvalSymlinks(path); resolveErr == nil {
		path = resolved
	} else if !os.IsNotExist(resolveErr) {
		return nil, resolveErr
	}
	release, err := acquireOwnership(path + ".lock")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = release()
		}
	}()
	flags := os.O_RDONLY
	if create {
		flags = os.O_CREATE | os.O_EXCL | os.O_RDWR
	}
	f, err := os.OpenFile(path, flags, 0600)
	if err != nil {
		return nil, err
	}
	info, statErr := f.Stat()
	closeErr := f.Close()
	if statErr != nil {
		return nil, statErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("research lineage file must be private and regular")
	}
	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	q.Add("_pragma", "locking_mode=EXCLUSIVE")
	q.Set("_journal_mode", "WAL")
	q.Set("_synchronous", "FULL")
	q.Set("_busy_timeout", "5000")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = db.Close()
		}
	}()
	db.SetMaxOpenConns(1)
	ctx := context.Background()
	if create {
		if _, err = db.ExecContext(ctx, `CREATE TABLE lineage_meta (
			id INTEGER PRIMARY KEY CHECK (id=1), initial BLOB NOT NULL, committed BLOB NOT NULL
		); CREATE TABLE lineage_touch (
			tenant TEXT NOT NULL, event TEXT NOT NULL, version INTEGER NOT NULL,
			PRIMARY KEY (tenant,event)
		); CREATE TABLE lineage_motion (
			version INTEGER PRIMARY KEY, available_at TEXT NOT NULL
		)`); err != nil {
			return nil, err
		}
		encoded, encodeErr := json.Marshal(current)
		if encodeErr != nil {
			return nil, encodeErr
		}
		if _, err = db.ExecContext(ctx, "INSERT INTO lineage_meta(id,initial,committed) VALUES(1,?,?)", encoded, encoded); err != nil {
			return nil, err
		}
	}
	var initialRaw, committedRaw []byte
	if err = db.QueryRowContext(ctx, "SELECT initial,committed FROM lineage_meta WHERE id=1").Scan(&initialRaw, &committedRaw); err != nil {
		return nil, fmt.Errorf("research lineage checkpoint unavailable: %w", err)
	}
	var initial, committed model.Snapshot
	if err = json.Unmarshal(initialRaw, &initial); err != nil {
		return nil, err
	}
	if err = json.Unmarshal(committedRaw, &committed); err != nil {
		return nil, err
	}
	if committed != current || initial.RuntimeVersion > committed.RuntimeVersion {
		return nil, errors.New("research lineage backend checkpoint mismatch")
	}
	// Older v2 sidecars gain only an empty motion table. Missing past rows keep
	// old as-of snapshots invalid; this migration cannot invent history.
	if _, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS lineage_motion (
		version INTEGER PRIMARY KEY, available_at TEXT NOT NULL
	)`); err != nil {
		return nil, err
	}
	return &Ledger{db: db, release: release, initial: initial, committed: committed}, nil
}

func (l *Ledger) Record(ctx context.Context, before, after model.Snapshot, touched []EventKey) error {
	return l.RecordWithMotion(ctx, before, after, touched, time.Time{})
}

// RecordWithMotion atomically checkpoints source touches and one certified
// ingestion time. A zero time denotes a general mutation, never an ingestion.
func (l *Ledger) RecordWithMotion(ctx context.Context, before, after model.Snapshot, touched []EventKey, ingestionAt time.Time) error {
	if ctx == nil {
		return errors.New("nil research lineage context")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed || l.committed != before || after.RuntimeVersion <= before.RuntimeVersion || after.RuntimeVersion > math.MaxInt64 {
		return errors.New("invalid research lineage transition")
	}
	for _, key := range touched {
		if key.Tenant == "" || key.Event == "" {
			return errors.New("invalid research lineage event key")
		}
	}
	if !ingestionAt.IsZero() {
		expected := before
		if before.RuntimeVersion == math.MaxInt64 || before.EvidenceEpoch == math.MaxUint64 || len(touched) != 1 {
			return errors.New("invalid research ingestion motion")
		}
		expected.RuntimeVersion++
		expected.EvidenceEpoch++
		if after != expected {
			return errors.New("ingestion motion does not match version transition")
		}
	}
	encoded, err := json.Marshal(after)
	if err != nil {
		return err
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, key := range touched {
		if _, err := tx.ExecContext(ctx, `INSERT INTO lineage_touch(tenant,event,version) VALUES(?,?,?)
			ON CONFLICT(tenant,event) DO UPDATE SET version=excluded.version`, key.Tenant, key.Event, after.RuntimeVersion); err != nil {
			return err
		}
	}
	if !ingestionAt.IsZero() {
		if _, err := tx.ExecContext(ctx, "INSERT INTO lineage_motion(version,available_at) VALUES(?,?)", after.RuntimeVersion, ingestionAt.UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	if after.RuntimeVersion >= 4096 {
		if _, err := tx.ExecContext(ctx, "DELETE FROM lineage_motion WHERE version<=?", after.RuntimeVersion-4096); err != nil {
			return err
		}
	}
	old, err := json.Marshal(before)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, "UPDATE lineage_meta SET committed=? WHERE id=1 AND committed=?", encoded, old)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil || changed != 1 {
		return errors.New("research lineage checkpoint moved")
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	l.committed = after
	return nil
}

// Motion returns a detached, bounded map for a publisher reopened at the
// exact sidecar checkpoint. Gaps are preserved so compatibility fails closed.
func (l *Ledger) Motion(ctx context.Context) (map[uint64]time.Time, error) {
	if ctx == nil {
		return nil, errors.New("nil research lineage context")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil, errors.New("research lineage closed")
	}
	rows, err := l.db.QueryContext(ctx, "SELECT version,available_at FROM lineage_motion")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	motion := make(map[uint64]time.Time)
	for rows.Next() {
		var version int64
		var encoded string
		if err := rows.Scan(&version, &encoded); err != nil {
			return nil, err
		}
		at, err := time.Parse(time.RFC3339Nano, encoded)
		if err != nil || at.IsZero() || version <= 0 || uint64(version) > l.committed.RuntimeVersion || l.committed.RuntimeVersion-uint64(version) >= 4096 {
			return nil, errors.New("invalid research ingestion motion row")
		}
		motion[uint64(version)] = at
	}
	return motion, rows.Err()
}

func (l *Ledger) Unchanged(ctx context.Context, from, target model.Snapshot, key EventKey) (bool, error) {
	if ctx == nil || key.Tenant == "" || key.Event == "" {
		return false, errors.New("invalid research lineage query")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed || target != l.committed || from.RuntimeVersion < l.initial.RuntimeVersion || from.RuntimeVersion > target.RuntimeVersion ||
		(from.RuntimeVersion == l.initial.RuntimeVersion && from != l.initial) {
		return false, errors.New("research lineage history unavailable")
	}
	var last int64
	err := l.db.QueryRowContext(ctx, "SELECT version FROM lineage_touch WHERE tenant=? AND event=?", key.Tenant, key.Event).Scan(&last)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return uint64(last) <= from.RuntimeVersion, nil
}

func (l *Ledger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	return errors.Join(l.db.Close(), l.release())
}
