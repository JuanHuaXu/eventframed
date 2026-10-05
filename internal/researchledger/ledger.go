// Package researchledger provides an unconfigured durable research log. Payload
// semantics and evidence authentication belong to the future consumer, not SQL.
package researchledger

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type Key struct{ Tenant, Journal, Event, Contract string }
type Entry struct {
	Sequence int64
	Kind     string
	Key      Key
	Payload  json.RawMessage
}
type Ledger struct {
	db        *sql.DB
	release   func() error
	closeOnce sync.Once
	closeErr  error
}

func Open(path string) (*Ledger, error) {
	return openLockMode(path, false)
}

// OpenExclusiveResearch is restricted to a single-owner research database with
// no external readers. Durable WAL acknowledgment is unchanged; Open is control.
func OpenExclusiveResearch(path string) (*Ledger, error) {
	return openLockMode(path, true)
}

func openLockMode(path string, exclusive bool) (*Ledger, error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("absolute ledger path required")
	}
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return nil, e
	}
	// Resolve existing symlink aliases before choosing the sidecar lock.
	parent, e := filepath.EvalSymlinks(filepath.Dir(path))
	if e != nil {
		return nil, e
	}
	path = filepath.Join(parent, filepath.Base(path))
	if resolved, e := filepath.EvalSymlinks(path); e == nil {
		path = resolved
	} else if !os.IsNotExist(e) {
		return nil, e
	}
	release, e := acquireLedgerOwnership(path + ".lock")
	if e != nil {
		return nil, e
	}
	handedOff := false
	defer func() {
		if !handedOff {
			_ = release()
		}
	}()
	// Create with private permissions before opening; fixtures contain no secrets,
	// but a future caller must not expose evidence via a permissive initial file.
	f, e := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = f.Close(); e != nil {
		return nil, e
	}
	if e = os.Chmod(path, 0600); e != nil {
		return nil, e
	}
	u := url.URL{Scheme: "file", Path: path}
	if exclusive {
		// The pinned driver applies _pragma before shorthand journal/sync keys
		// on EVERY connection. A startup-only setting would miss replacements.
		q := u.Query()
		q.Add("_pragma", "locking_mode=EXCLUSIVE")
		q.Set("_journal_mode", "WAL")
		q.Set("_synchronous", "FULL")
		q.Set("_busy_timeout", "5000")
		u.RawQuery = q.Encode()
	}
	db, e := sql.Open("sqlite", u.String())
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(1)
	for _, q := range []string{"PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", "PRAGMA busy_timeout=5000", `CREATE TABLE IF NOT EXISTS research_log(sequence INTEGER PRIMARY KEY AUTOINCREMENT, identity TEXT NOT NULL, kind TEXT NOT NULL CHECK(kind IN ('admit','feedback')), payload BLOB NOT NULL, UNIQUE(identity,kind))`} {
		if _, e = db.Exec(q); e != nil {
			db.Close()
			return nil, e
		}
	}
	if _, e = db.Exec(`CREATE TABLE IF NOT EXISTS research_bootstrap(id INTEGER PRIMARY KEY CHECK(id=1), seal BLOB NOT NULL CHECK(length(seal)=32))`); e != nil {
		db.Close()
		return nil, e
	}
	if _, e = db.Exec(`CREATE TABLE IF NOT EXISTS research_bootstrap_motion(id INTEGER PRIMARY KEY CHECK(id=1), origin BLOB NOT NULL CHECK(length(origin) BETWEEN 1 AND 1024))`); e != nil {
		db.Close()
		return nil, e
	}
	handedOff = true
	return &Ledger{db: db, release: release}, nil
}

// Bootstrap returns the immutable research-epoch commitment, if one exists.
func (l *Ledger) Bootstrap(ctx context.Context) ([]byte, error) {
	var seal []byte
	err := l.db.QueryRowContext(ctx, "SELECT seal FROM research_bootstrap WHERE id=1").Scan(&seal)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(seal) != 32 {
		return nil, errors.New("invalid research bootstrap seal")
	}
	return seal, nil
}

// BindBootstrap commits a caller-authenticated model identity before the first
// epoch record. Exact reopen is allowed; a changed identity is never adopted.
func (l *Ledger) BindBootstrap(ctx context.Context, seal []byte) error {
	if len(seal) != 32 {
		return errors.New("invalid research bootstrap seal")
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var old []byte
	err = tx.QueryRowContext(ctx, "SELECT seal FROM research_bootstrap WHERE id=1").Scan(&old)
	if err == nil {
		if !bytes.Equal(old, seal) {
			return errors.New("research bootstrap seal mismatch")
		}
		return tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var count int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM research_log").Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return errors.New("cannot bind bootstrap to nonempty log")
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO research_bootstrap(id,seal) VALUES(1,?)", seal); err != nil {
		return err
	}
	return tx.Commit()
}

// MotionBootstrap reads the origin bound to a motion-mode seal. A legacy seal
// without origin metadata cannot be promoted by guessing its old snapshot.
func (l *Ledger) MotionBootstrap(ctx context.Context) ([]byte, []byte, error) {
	seal, err := l.Bootstrap(ctx)
	if err != nil || seal == nil {
		return seal, nil, err
	}
	var origin []byte
	if err = l.db.QueryRowContext(ctx, "SELECT origin FROM research_bootstrap_motion WHERE id=1").Scan(&origin); errors.Is(err, sql.ErrNoRows) {
		return nil, nil, errors.New("legacy bootstrap lacks motion origin")
	} else if err != nil {
		return nil, nil, err
	}
	if len(origin) == 0 || len(origin) > 1024 || !json.Valid(origin) {
		return nil, nil, errors.New("invalid research motion origin")
	}
	return seal, origin, nil
}

// BindMotionBootstrap atomically binds a caller-authenticated origin and seal
// before the first log entry. Exact reopen is allowed; legacy logs reject.
func (l *Ledger) BindMotionBootstrap(ctx context.Context, seal, origin []byte) error {
	if len(seal) != 32 || len(origin) == 0 || len(origin) > 1024 || !json.Valid(origin) {
		return errors.New("invalid research motion bootstrap")
	}
	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var oldSeal []byte
	err = tx.QueryRowContext(ctx, "SELECT seal FROM research_bootstrap WHERE id=1").Scan(&oldSeal)
	if err == nil {
		if !bytes.Equal(oldSeal, seal) {
			return errors.New("research motion seal mismatch")
		}
		var oldOrigin []byte
		if err = tx.QueryRowContext(ctx, "SELECT origin FROM research_bootstrap_motion WHERE id=1").Scan(&oldOrigin); errors.Is(err, sql.ErrNoRows) {
			return errors.New("legacy bootstrap lacks motion origin")
		} else if err != nil {
			return err
		}
		if !bytes.Equal(oldOrigin, origin) {
			return errors.New("research motion origin mismatch")
		}
		return tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var count int
	if err = tx.QueryRowContext(ctx, "SELECT (SELECT COUNT(*) FROM research_log)+(SELECT COUNT(*) FROM research_bootstrap_motion)").Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return errors.New("cannot bind motion bootstrap to nonempty log")
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO research_bootstrap(id,seal) VALUES(1,?)", seal); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO research_bootstrap_motion(id,origin) VALUES(1,?)", origin); err != nil {
		return err
	}
	return tx.Commit()
}
func (l *Ledger) Close() error {
	l.closeOnce.Do(func() { l.closeErr = errors.Join(l.db.Close(), l.release()) })
	return l.closeErr
}

// Get supports exact retry lookup without loading the lifetime log into RAM.
func (l *Ledger) Get(ctx context.Context, key Key, kind string) (Entry, error) {
	identity, e := json.Marshal(key)
	if e != nil {
		return Entry{}, e
	}
	r := Entry{Key: key, Kind: kind}
	e = l.db.QueryRowContext(ctx, "SELECT sequence,payload FROM research_log WHERE identity=? AND kind=?", string(identity), kind).Scan(&r.Sequence, &r.Payload)
	return r, e
}

// Append acknowledges only after COMMIT. Exact retries return the existing
// sequence; changed bytes under the same identity are conflicts, not corrections.
// Feedback requires an admitted identity. This does NOT validate usefulness,
// temporal legality or original forecasts; the consumer must do that first.
func (l *Ledger) Append(ctx context.Context, key Key, kind string, payload json.RawMessage) (int64, bool, error) {
	for _, s := range []string{key.Tenant, key.Journal, key.Event, key.Contract} {
		if len(s) == 0 || len(s) > 4096 {
			return 0, false, errors.New("invalid ledger identity")
		}
	}
	if (kind != "admit" && kind != "feedback") || len(payload) > 1<<20 || !json.Valid(payload) {
		return 0, false, errors.New("invalid ledger record")
	}
	identity, _ := json.Marshal(key)
	tx, e := l.db.BeginTx(ctx, nil)
	if e != nil {
		return 0, false, e
	}
	defer tx.Rollback()
	var seq int64
	var old []byte
	e = tx.QueryRowContext(ctx, "SELECT sequence,payload FROM research_log WHERE identity=? AND kind=?", string(identity), kind).Scan(&seq, &old)
	if e == nil {
		if !bytes.Equal(old, payload) {
			return 0, false, errors.New("conflicting ledger retry")
		}
		return seq, true, tx.Commit()
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return 0, false, e
	}
	if kind == "feedback" {
		var n int
		if e = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM research_log WHERE identity=? AND kind='admit'", string(identity)).Scan(&n); e != nil {
			return 0, false, e
		}
		if n != 1 {
			return 0, false, errors.New("feedback without admission")
		}
	}
	r, e := tx.ExecContext(ctx, "INSERT INTO research_log(identity,kind,payload) VALUES(?,?,?)", string(identity), kind, []byte(payload))
	if e != nil {
		return 0, false, e
	}
	seq, e = r.LastInsertId()
	if e != nil {
		return 0, false, e
	}
	return seq, false, tx.Commit()
}

// ReadAfter bounds replay memory; durable identities are not evicted. There is
// deliberately no checkpoint/deletion API until its replay contract is tested.
func (l *Ledger) ReadAfter(ctx context.Context, after int64, limit int) ([]Entry, error) {
	if after < 0 || limit < 1 || limit > 256 {
		return nil, errors.New("invalid replay page")
	}
	rows, e := l.db.QueryContext(ctx, "SELECT sequence,identity,kind,payload FROM research_log WHERE sequence>? ORDER BY sequence LIMIT ?", after, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var r Entry
		var identity string
		if e = rows.Scan(&r.Sequence, &identity, &r.Kind, &r.Payload); e != nil {
			return nil, e
		}
		if e = json.Unmarshal([]byte(identity), &r.Key); e != nil {
			return nil, e
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
