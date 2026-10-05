package researchbatch

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	_ "modernc.org/sqlite"
)

var errInjectedCrash = errors.New("injected crash after durable phase")

type batchIntent struct {
	Before, After model.Snapshot
	At            time.Time
	Writes        []libravdbstore.ResearchEventWrite
	New           []int
	Duplicate     []bool
}

type batchPrototype struct {
	backend    *libravdbstore.Store
	db         *sql.DB
	snapshot   model.Snapshot
	motion     map[uint64]time.Time
	quarantine bool
}

func openBatchPrototype(root string, create bool) (_ *batchPrototype, err error) {
	backend, err := libravdbstore.Open(libravdbstore.Config{
		Path: filepath.Join(root, "events.libravdb"), Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = backend.Close()
		}
	}()
	if create {
		virgin := model.Snapshot{
			PolicyVersion: 1, ContractVersion: model.ContractVersion, GraphVersion: 1,
			PosteriorVersion: 1, ResidualVersion: 1, AbstractionVersion: 1, AgencyVersion: 1,
		}
		if backend.Snapshot(context.Background()) != virgin {
			return nil, errors.New("cannot initialize sidecar over existing backend history")
		}
	}
	path := filepath.Join(root, "batch-intent.sqlite")
	if create {
		f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if e != nil {
			return nil, e
		}
		if e = f.Close(); e != nil {
			return nil, e
		}
	} else if _, e := os.Stat(path); e != nil {
		return nil, e
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = db.Close()
		}
	}()
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{"PRAGMA journal_mode=WAL", "PRAGMA synchronous=FULL", "PRAGMA busy_timeout=5000"} {
		if _, err = db.Exec(pragma); err != nil {
			return nil, err
		}
	}
	p := &batchPrototype{backend: backend, db: db}
	if create {
		if _, err = db.Exec(`CREATE TABLE meta(id INTEGER PRIMARY KEY CHECK(id=1), snapshot BLOB NOT NULL);
			CREATE TABLE intent(id INTEGER PRIMARY KEY CHECK(id=1), payload BLOB NOT NULL);
			CREATE TABLE motion(version INTEGER PRIMARY KEY, available_at TEXT NOT NULL);
			CREATE TABLE accepted(tenant TEXT NOT NULL, event TEXT NOT NULL, digest TEXT NOT NULL,
				version INTEGER NOT NULL, payload BLOB NOT NULL, PRIMARY KEY(tenant,event))`); err != nil {
			return nil, err
		}
		encoded, e := json.Marshal(backend.Snapshot(context.Background()))
		if e != nil {
			return nil, e
		}
		if _, err = db.Exec("INSERT INTO meta(id,snapshot) VALUES(1,?)", encoded); err != nil {
			return nil, err
		}
	}
	if err = p.recover(context.Background()); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *batchPrototype) close() error {
	return errors.Join(p.db.Close(), p.backend.Close())
}

func readSnapshot(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) (model.Snapshot, error) {
	var encoded []byte
	if err := q.QueryRowContext(ctx, "SELECT snapshot FROM meta WHERE id=1").Scan(&encoded); err != nil {
		return model.Snapshot{}, err
	}
	var snapshot model.Snapshot
	if err := json.Unmarshal(encoded, &snapshot); err != nil {
		return model.Snapshot{}, err
	}
	return snapshot, nil
}

func (p *batchPrototype) readIntent(ctx context.Context) (*batchIntent, error) {
	var encoded []byte
	if err := p.db.QueryRowContext(ctx, "SELECT payload FROM intent WHERE id=1").Scan(&encoded); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	var intent batchIntent
	if err := json.Unmarshal(encoded, &intent); err != nil {
		return nil, err
	}
	return &intent, nil
}

func (p *batchPrototype) loadPublished(ctx context.Context) error {
	snapshot, err := readSnapshot(ctx, p.db)
	if err != nil {
		return err
	}
	if snapshot != p.backend.Snapshot(ctx) {
		p.quarantine = true
		return errors.New("backend/sidecar checkpoint mismatch")
	}
	rows, err := p.db.QueryContext(ctx, "SELECT version,available_at FROM motion")
	if err != nil {
		return err
	}
	defer rows.Close()
	motion := make(map[uint64]time.Time)
	for rows.Next() {
		var version uint64
		var encoded string
		if err := rows.Scan(&version, &encoded); err != nil {
			return err
		}
		at, err := time.Parse(time.RFC3339Nano, encoded)
		if err != nil || at.IsZero() || version == 0 || version > snapshot.RuntimeVersion {
			return errors.New("invalid durable motion")
		}
		motion[version] = at
	}
	if err := rows.Err(); err != nil {
		return err
	}
	p.snapshot, p.motion = snapshot, motion
	return nil
}

func sameWrite(a, b libravdbstore.ResearchEventWrite) bool {
	left, leftErr := json.Marshal(a.Event)
	right, rightErr := json.Marshal(b.Event)
	return leftErr == nil && rightErr == nil && a.Digest == b.Digest && bytes.Equal(left, right) && reflect.DeepEqual(a.Vector, b.Vector)
}

func (p *batchPrototype) verifyWrite(ctx context.Context, w libravdbstore.ResearchEventWrite) error {
	got, err := p.backend.GetEventsWithVectors(ctx, w.Event.TenantID, []string{w.Event.ID}, w.Event.AvailableAt)
	if err != nil {
		return fmt.Errorf("accepted event readback %s: %w", w.Event.ID, err)
	}
	if len(got) != 1 {
		return fmt.Errorf("accepted event readback count mismatch %s", w.Event.ID)
	}
	expected := w.Event
	expected.Embedding = append([]float32(nil), w.Vector...)
	expected.EmbeddingModel = "research:d4"
	left, leftErr := json.Marshal(expected)
	right, rightErr := json.Marshal(got[0])
	if leftErr != nil || rightErr != nil || !bytes.Equal(left, right) {
		return fmt.Errorf("accepted event readback mismatch %s", w.Event.ID)
	}
	return nil
}

func (p *batchPrototype) prepare(ctx context.Context, writes []libravdbstore.ResearchEventWrite) (*batchIntent, error) {
	if p.quarantine || len(writes) == 0 || len(writes) > 16 || writes[0].Event.TenantID == "" || writes[0].Event.AvailableAt.IsZero() {
		return nil, errors.New("invalid batch preparation")
	}
	if intent, err := p.readIntent(ctx); err != nil || intent != nil {
		return nil, errors.New("unresolved prior intent")
	}
	before, err := readSnapshot(ctx, p.db)
	if err != nil || before != p.backend.Snapshot(ctx) || before != p.snapshot {
		return nil, errors.New("unowned batch snapshot")
	}
	intent := &batchIntent{Before: before, At: writes[0].Event.AvailableAt, Writes: writes,
		Duplicate: make([]bool, len(writes))}
	seen := make(map[string]libravdbstore.ResearchEventWrite)
	for i, w := range writes {
		if w.Event.TenantID != writes[0].Event.TenantID || !w.Event.AvailableAt.Equal(intent.At) || w.Digest == "" ||
			w.Event.ID == "" || len(w.Vector) != 4 || w.Event.Composition != nil {
			return nil, errors.New("mixed or invalid batch input")
		}
		if err := w.Event.Validate(4); err != nil {
			return nil, err
		}
		for _, v := range w.Vector {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return nil, errors.New("nonfinite batch vector")
			}
		}
		key := w.Event.TenantID + "\x00" + w.Event.ID
		if prior, ok := seen[key]; ok {
			if !sameWrite(prior, w) {
				return nil, store.ErrIdempotencyConflict
			}
			intent.Duplicate[i] = true
			continue
		}
		seen[key] = w
		var encoded []byte
		err := p.db.QueryRowContext(ctx, "SELECT payload FROM accepted WHERE tenant=? AND event=?", w.Event.TenantID, w.Event.ID).Scan(&encoded)
		if err == nil {
			var prior libravdbstore.ResearchEventWrite
			if err := json.Unmarshal(encoded, &prior); err != nil {
				return nil, err
			}
			if !sameWrite(prior, w) {
				return nil, store.ErrIdempotencyConflict
			}
			if err := p.verifyWrite(ctx, w); err != nil {
				return nil, err
			}
			intent.Duplicate[i] = true
		} else if errors.Is(err, sql.ErrNoRows) {
			intent.New = append(intent.New, i)
		} else {
			return nil, err
		}
	}
	if uint64(len(intent.New)) > math.MaxUint64-before.RuntimeVersion ||
		uint64(len(intent.New)) > math.MaxUint64-before.EvidenceEpoch {
		return nil, errors.New("batch version overflow")
	}
	intent.After = before
	intent.After.RuntimeVersion += uint64(len(intent.New))
	intent.After.EvidenceEpoch += uint64(len(intent.New))
	if len(intent.New) == 0 {
		return intent, nil
	}
	encoded, err := json.Marshal(intent)
	if err != nil {
		return nil, err
	}
	if _, err := p.db.ExecContext(ctx, "INSERT INTO intent(id,payload) VALUES(1,?)", encoded); err != nil {
		return nil, err
	}
	return intent, nil
}

func (p *batchPrototype) verifyAccepted(ctx context.Context, intent *batchIntent) error {
	for _, i := range intent.New {
		if err := p.verifyWrite(ctx, intent.Writes[i]); err != nil {
			return err
		}
	}
	return nil
}

func (p *batchPrototype) finalize(ctx context.Context, intent *batchIntent) error {
	if p.backend.Snapshot(ctx) != intent.After || len(intent.New) == 0 {
		return errors.New("invalid finalization snapshot")
	}
	if err := p.verifyAccepted(ctx, intent); err != nil {
		return err
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	before, err := readSnapshot(ctx, tx)
	if err != nil || before != intent.Before {
		return errors.New("sidecar checkpoint moved")
	}
	for j, i := range intent.New {
		w := intent.Writes[i]
		version := intent.Before.RuntimeVersion + uint64(j) + 1
		payload, err := json.Marshal(w)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO motion(version,available_at) VALUES(?,?)", version, intent.At.UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO accepted(tenant,event,digest,version,payload) VALUES(?,?,?,?,?)",
			w.Event.TenantID, w.Event.ID, w.Digest, version, payload); err != nil {
			return err
		}
	}
	after, err := json.Marshal(intent.After)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE meta SET snapshot=? WHERE id=1", after); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM intent WHERE id=1"); err != nil {
		return err
	}
	return tx.Commit()
}

func (p *batchPrototype) recover(ctx context.Context) error {
	intent, err := p.readIntent(ctx)
	if err != nil {
		return err
	}
	if intent != nil {
		current := p.backend.Snapshot(ctx)
		switch current {
		case intent.Before:
			for _, i := range intent.New {
				w := intent.Writes[i]
				got, err := p.backend.GetEvents(ctx, w.Event.TenantID, []string{w.Event.ID}, intent.At.Add(time.Second))
				if err == nil && len(got) > 0 {
					return errors.New("uncommitted intent has visible event")
				}
				if err != nil && !errors.Is(err, store.ErrEventNotFound) {
					return err
				}
			}
			if _, err := p.db.ExecContext(ctx, "DELETE FROM intent WHERE id=1"); err != nil {
				return err
			}
		case intent.After:
			if err := p.finalize(ctx, intent); err != nil {
				return err
			}
		default:
			return errors.New("unreconcilable batch snapshot")
		}
	}
	return p.loadPublished(ctx)
}

func (p *batchPrototype) apply(ctx context.Context, writes []libravdbstore.ResearchEventWrite, crash string) ([]store.PutResult, error) {
	intent, err := p.prepare(ctx, writes)
	if err != nil {
		return nil, err
	}
	if len(intent.New) == 0 {
		results, err := p.backend.PutResearchEventBatch(ctx, writes)
		if err != nil {
			p.quarantine = true
			return nil, err
		}
		if len(results) != len(writes) || p.backend.Snapshot(ctx) != intent.Before {
			p.quarantine = true
			return nil, errors.New("duplicate-only backend confirmation mismatch")
		}
		for _, result := range results {
			if !result.Duplicate || result.Snapshot != intent.Before {
				p.quarantine = true
				return nil, errors.New("duplicate-only backend identity mismatch")
			}
		}
		return results, nil
	}
	if crash == "intent" {
		return nil, errInjectedCrash
	}
	results, err := p.backend.PutResearchEventBatch(ctx, writes)
	if err != nil {
		p.quarantine = true
		return nil, err
	}
	if len(results) != len(writes) {
		p.quarantine = true
		return nil, errors.New("short raw batch result")
	}
	for i, r := range results {
		if r.Duplicate != intent.Duplicate[i] || r.Snapshot != intent.After {
			p.quarantine = true
			return nil, errors.New("raw batch identity mismatch")
		}
	}
	if crash == "backend" {
		return nil, errInjectedCrash
	}
	if err := p.finalize(ctx, intent); err != nil {
		p.quarantine = true
		return nil, err
	}
	if crash == "sidecar" {
		return nil, errInjectedCrash
	}
	if err := p.loadPublished(ctx); err != nil {
		return nil, err
	}
	return results, nil
}

func (p *batchPrototype) asOf(ctx context.Context, captured model.Snapshot, at time.Time) bool {
	if p.quarantine || p.snapshot != p.backend.Snapshot(ctx) {
		return false
	}
	intent, err := p.readIntent(ctx)
	return err == nil && intent == nil && store.ResearchSnapshotCompatible(captured, p.snapshot, at, p.motion)
}
