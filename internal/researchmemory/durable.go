package researchmemory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

// Durable is an exclusively owned research stream, not a service plugin. IDs
// are caller-retained monotonically increasing prediction IDs, not event IDs.
// Unlabeled abandonment is a distinct durable terminal record, never a label.
type durableLog interface {
	Get(context.Context, researchledger.Key, string) (researchledger.Entry, error)
	Append(context.Context, researchledger.Key, string, json.RawMessage) (int64, bool, error)
	ReadAfter(context.Context, int64, int) ([]researchledger.Entry, error)
	Close() error
}

type sourceIndexedLog interface {
	GetServiceAdmission(context.Context, researchledger.ServiceIdentity) (researchledger.Entry, error)
}

type Durable struct {
	measurement     *admissionPhases
	mu              sync.Mutex
	log             durableLog
	worker          *Background
	tenant, stream  string
	epoch           uint64
	stopped, closed bool
}

// Tenant is fixed when this exclusively owned stream is opened.
func (d *Durable) Tenant() string {
	if d == nil {
		return ""
	}
	return d.tenant
}

// Epoch is fixed for the lifetime of this research stream.
func (d *Durable) Epoch() uint64 {
	if d == nil {
		return 0
	}
	return d.epoch
}

func OpenDurable(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64) (*Durable, error) {
	log, e := researchledger.Open(path)
	if e != nil {
		return nil, e
	}
	b, e := ResumeBackground(ctx, log, tenant, stream, epoch, seed, 256)
	if e != nil {
		log.Close()
		return nil, e
	}
	return &Durable{log: log, worker: b, tenant: tenant, stream: stream, epoch: epoch}, nil
}

// OpenBoundDurable starts or resumes a research epoch from a separately
// validated component transfer. The caller must rebuild the Adapter and seal
// from the same authenticated source evidence on every open. Neither the seal
// nor this wrapper grants service-side source authority.
func OpenBoundDurable(ctx context.Context, path string, bootstrap *SealedBoundBootstrap, validate func(context.Context, RecordedPrediction) error) (*Durable, error) {
	return openBoundDurable(ctx, path, bootstrap, false, nil, validate)
}

// OpenMotionBoundDurable requires the caller's fresh proof that the immutable
// stored origin remains as-of compatible with the current target. The origin
// is also HMAC-bound; this lower-level API itself grants no store authority.
func OpenMotionBoundDurable(ctx context.Context, path string, bootstrap *SealedBoundBootstrap, validateOrigin func(context.Context, model.Snapshot) error, validateOriginal func(context.Context, RecordedPrediction) error) (*Durable, error) {
	return openBoundDurable(ctx, path, bootstrap, true, validateOrigin, validateOriginal)
}

func openBoundDurable(ctx context.Context, path string, bootstrap *SealedBoundBootstrap, motion bool, validateOrigin func(context.Context, model.Snapshot) error, validate func(context.Context, RecordedPrediction) error) (*Durable, error) {
	if ctx == nil || bootstrap == nil || bootstrap.state == nil || validate == nil || (motion && validateOrigin == nil) {
		return nil, errors.New("invalid bound durable bootstrap")
	}
	state := bootstrap.state
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.used || state.motion != motion || (motion && (state.origin.RuntimeVersion == 0 || state.origin.ContractVersion == 0)) || state.adapter == nil || state.tenant == "" || state.stream == "" || state.epoch == 0 || len(state.seal) != 32 || state.adapter.epoch != state.epoch || state.adapter.seed != state.seed || state.adapter.next != 0 || len(state.adapter.pending) != 0 || state.adapter.total != len(state.adapter.samples) {
		return nil, errors.New("invalid bound durable bootstrap")
	}
	state.used = true
	log, err := researchledger.Open(path)
	if err != nil {
		return nil, err
	}
	if motion {
		var origin []byte
		origin, err = json.Marshal(state.origin)
		if err == nil {
			err = log.BindMotionBootstrap(ctx, state.seal, origin)
		}
		if err == nil {
			err = validateOrigin(ctx, state.origin)
		}
	} else {
		err = log.BindBootstrap(ctx, state.seal)
	}
	if err != nil {
		log.Close()
		return nil, err
	}
	if err = log.EnableServiceIdentity(ctx); err != nil {
		log.Close()
		return nil, err
	}
	a, err := replayLedgerOnAdapter(ctx, log, state.tenant, state.stream, state.adapter, true, true, validate)
	if err != nil {
		log.Close()
		return nil, err
	}
	b, err := backgroundFromOwnedAdapter(a, 256)
	if err != nil {
		log.Close()
		return nil, err
	}
	return &Durable{log: log, worker: b, tenant: state.tenant, stream: state.stream, epoch: state.epoch}, nil
}
func (d *Durable) key(id uint64) researchledger.Key {
	return researchledger.Key{Tenant: d.tenant, Journal: d.stream, Event: strconv.FormatUint(id, 10), Contract: RecordContract}
}

func (d *Durable) Admit(ctx context.Context, id uint64, features uint16, baseline float64, at time.Time) (Prediction, bool, error) {
	return d.admit(ctx, id, features, baseline, at, nil, nil)
}

func (d *Durable) AdmitBound(ctx context.Context, id uint64, features uint16, baseline float64, at time.Time, binding ServiceBinding) (Prediction, bool, error) {
	if e := binding.validate(); e != nil {
		return Prediction{}, false, e
	}
	if binding.Tenant != d.tenant {
		return Prediction{}, false, errors.New("binding tenant mismatch")
	}
	return d.admit(ctx, id, features, baseline, at, &binding, nil)
}

// AdmitBoundWithWitness persists an optional research source commitment beside
// the original forecast. The caller must have validated the source while
// holding the service admission guard; this method cannot verify the MAC key.
func (d *Durable) AdmitBoundWithWitness(ctx context.Context, id uint64, features uint16, baseline float64, at time.Time, binding ServiceBinding, witness SourceWitness) (Prediction, bool, error) {
	if err := binding.validate(); err != nil {
		return Prediction{}, false, err
	}
	if err := witness.Validate(); err != nil {
		return Prediction{}, false, err
	}
	if binding.Tenant != d.tenant || witness.FeatureContract != FeatureContract {
		return Prediction{}, false, errors.New("source witness contract mismatch")
	}
	return d.admit(ctx, id, features, baseline, at, &binding, &witness)
}

// AdmitSourceBoundWithWitness allocates a learner ID behind the persistent
// service-source index. The caller must hold its source-validation guard and
// supply a witness constructed from the actual committed event.
func (d *Durable) AdmitSourceBoundWithWitness(ctx context.Context, features uint16, baseline float64, at time.Time, binding ServiceBinding, witness SourceWitness) (Prediction, bool, error) {
	if err := binding.validate(); err != nil {
		return Prediction{}, false, err
	}
	if err := witness.Validate(); err != nil {
		return Prediction{}, false, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped || d.closed || binding.Tenant != d.tenant || witness.FeatureContract != FeatureContract {
		return Prediction{}, false, errors.New("bound source stream stopped or mismatched")
	}
	indexed, ok := d.log.(sourceIndexedLog)
	if !ok {
		return Prediction{}, false, errors.New("bound source index unavailable")
	}
	key := researchledger.ServiceIdentity{Tenant: d.tenant, Stream: d.stream, Contract: RecordContract, Journal: binding.JournalID, Event: binding.EventID}
	entry, err := indexed.GetServiceAdmission(ctx, key)
	if err == nil {
		var old RecordedPrediction
		if err = decodeCanonical(entry.Payload, &old); err != nil {
			return Prediction{}, false, err
		}
		if err = old.Validate(); err != nil {
			return Prediction{}, false, err
		}
		if entry.Key.Event != strconv.FormatUint(old.Prediction.ID, 10) || old.Seed != d.worker.adapter.seed || old.Prediction.Epoch != d.epoch || old.Prediction.Features != features || old.Outer[0] != baseline || !old.At.Equal(at) || !reflect.DeepEqual(old.Binding, &binding) || !reflect.DeepEqual(old.Witness, &witness) {
			return Prediction{}, false, errors.New("conflicting bound source retry")
		}
		return old.Prediction, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Prediction{}, false, err
	}
	d.worker.mu.Lock()
	next := d.worker.next
	d.worker.mu.Unlock()
	if next == ^uint64(0) {
		return Prediction{}, false, errors.New("bound source IDs exhausted")
	}
	return d.admitLocked(ctx, next+1, features, baseline, at, &binding, &witness)
}

// AdmissionBySource resolves the immutable original for later guarded
// feedback. Lookup alone never authorizes a label.
func (d *Durable) AdmissionBySource(ctx context.Context, journal, event string) (RecordedPrediction, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped || d.closed {
		return RecordedPrediction{}, errors.New("bound source stream stopped")
	}
	indexed, ok := d.log.(sourceIndexedLog)
	if !ok {
		return RecordedPrediction{}, errors.New("bound source index unavailable")
	}
	entry, err := indexed.GetServiceAdmission(ctx, researchledger.ServiceIdentity{Tenant: d.tenant, Stream: d.stream, Contract: RecordContract, Journal: journal, Event: event})
	if err != nil {
		return RecordedPrediction{}, err
	}
	var record RecordedPrediction
	if err = decodeCanonical(entry.Payload, &record); err != nil {
		return RecordedPrediction{}, err
	}
	if err = record.Validate(); err != nil {
		return RecordedPrediction{}, err
	}
	if record.Binding == nil || record.Witness == nil || record.Binding.Tenant != d.tenant || record.Binding.JournalID != journal || record.Binding.EventID != event || record.Prediction.Epoch != d.epoch || record.Seed != d.worker.adapter.seed || entry.Key.Event != strconv.FormatUint(record.Prediction.ID, 10) {
		return RecordedPrediction{}, errors.New("bound source original mismatch")
	}
	return record, nil
}

func (d *Durable) admit(ctx context.Context, id uint64, features uint16, baseline float64, at time.Time, binding *ServiceBinding, witness *SourceWitness) (Prediction, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.admitLocked(ctx, id, features, baseline, at, binding, witness)
}

func (d *Durable) admitLocked(ctx context.Context, id uint64, features uint16, baseline float64, at time.Time, binding *ServiceBinding, witness *SourceWitness) (Prediction, bool, error) {
	if d.stopped || d.closed {
		return Prediction{}, false, errors.New("durable stream stopped")
	}
	old, e := d.log.Get(ctx, d.key(id), "admit")
	if e == nil {
		var r RecordedPrediction
		if e = decodeCanonical(old.Payload, &r); e != nil {
			return Prediction{}, false, e
		}
		if r.Prediction.ID != id || r.Prediction.Features != features || r.Outer[0] != baseline || !r.At.Equal(at) || !reflect.DeepEqual(r.Binding, binding) || !reflect.DeepEqual(r.Witness, witness) {
			return Prediction{}, false, errors.New("conflicting admission retry")
		}
		return r.Prediction, true, nil
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return Prediction{}, false, e
	}
	d.worker.mu.Lock()
	next := d.worker.next
	d.worker.mu.Unlock()
	if next == ^uint64(0) || id != next+1 {
		return Prediction{}, false, errors.New("admission ID out of order")
	}
	p, e := d.worker.Predict(features, baseline, d.epoch, at)
	if e != nil {
		return Prediction{}, false, e
	}
	r, e := d.worker.Record(id)
	if e != nil {
		d.stopped = true
		return Prediction{}, false, e
	}
	r.Binding = binding
	r.Witness = witness
	raw, e := json.Marshal(r)
	if e != nil {
		d.stopped = true
		return Prediction{}, false, e
	}
	if _, _, e = d.log.Append(ctx, d.key(id), "admit", raw); e != nil {
		d.stopped = true
		return Prediction{}, false, e
	}
	return p, false, nil
}

// Admission returns the persisted original record, including any service
// binding. Reading it is not a dependency certificate or feedback authorization.
func (d *Durable) Admission(ctx context.Context, id uint64) (RecordedPrediction, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.stopped {
		return RecordedPrediction{}, errors.New("durable stream stopped")
	}
	entry, e := d.log.Get(ctx, d.key(id), "admit")
	if e != nil {
		return RecordedPrediction{}, e
	}
	var r RecordedPrediction
	if e = decodeCanonical(entry.Payload, &r); e != nil {
		return RecordedPrediction{}, e
	}
	if e = r.Validate(); e != nil {
		return RecordedPrediction{}, e
	}
	if r.Prediction.ID != id || (r.Binding != nil && r.Binding.Tenant != d.tenant) {
		return RecordedPrediction{}, errors.New("stored binding mismatch")
	}
	return r, nil
}

func (d *Durable) Feedback(ctx context.Context, id uint64, useful bool, available time.Time) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped || d.closed {
		return false, errors.New("durable stream stopped")
	}
	record := RecordedFeedback{id, useful, available}
	raw, e := json.Marshal(record)
	if e != nil {
		return false, e
	}
	old, e := d.log.Get(ctx, d.key(id), "feedback")
	if e == nil {
		var previous RecordedFeedback
		if e = decodeCanonical(old.Payload, &previous); e != nil {
			return false, e
		}
		if previous.ID != id || previous.Useful != useful || !previous.Available.Equal(available) {
			return false, errors.New("conflicting feedback retry")
		}
		return true, nil
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return false, e
	}
	// This wrapper exclusively owns the worker. Fitting cannot change pending
	// identity or availability order, so validation survives the durable write.
	d.worker.mu.Lock()
	p, ok := d.worker.pending[id]
	last := d.worker.lastAvailable
	d.worker.mu.Unlock()
	if !ok || available.IsZero() || available.Before(p.At) || available.Before(last) {
		return false, errors.New("invalid durable feedback")
	}
	if _, _, e = d.log.Append(ctx, d.key(id), "feedback", raw); e != nil {
		d.stopped = true
		return false, e
	}
	if e = d.worker.Feedback(id, useful, d.epoch, available); e != nil {
		d.stopped = true
		return false, e
	}
	return false, nil
}

func (d *Durable) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return nil
	}
	d.closed = true
	d.worker.Close()
	return d.log.Close()
}
