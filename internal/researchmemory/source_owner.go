package researchmemory

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

// SourceOwner assigns learner IDs behind a persistent service-source identity.
// It is an opt-in research primitive, not an evidence authority or service plugin.
// The caller must hold the service's admission guard. No feedback or model-serving
// API is exposed until fresh feedback and retained-history authority are wired.
type SourceOwner struct {
	batchReads      bool
	reuseAdmissions bool
	mu              sync.Mutex
	d               *Durable
	log             *researchledger.Ledger
	seed            int64
}

// SourceAdmissionRequest has no caller-selected learner ID. Inputs must remain
// immutable until Admit returns; each journal/event may occur only once per batch.
type SourceAdmissionRequest struct {
	Features uint16
	Baseline float64
	At       time.Time
	Binding  ServiceBinding
}

func OpenSourceOwner(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64) (*SourceOwner, error) {
	return openSourceOwner(ctx, path, tenant, stream, epoch, seed, researchledger.Open)
}

func openSourceOwner(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64, open func(string) (*researchledger.Ledger, error)) (*SourceOwner, error) {
	log, err := open(path)
	if err != nil {
		return nil, err
	}
	// Validate complete canonical history before installing the opt-in index.
	// Neither replay nor index activation silently merges old duplicate evidence.
	a, err := replayLedger(ctx, log, tenant, stream, epoch, seed, true)
	if err != nil {
		log.Close()
		return nil, err
	}
	if err = log.EnableServiceIdentity(ctx); err != nil {
		log.Close()
		return nil, err
	}
	b, err := backgroundFromOwnedAdapter(a, 256)
	if err != nil {
		log.Close()
		return nil, err
	}
	d := &Durable{log: preparedBatchLedger{log}, worker: b, tenant: tenant, stream: stream, epoch: epoch}
	return &SourceOwner{d: d, log: log, seed: seed}, nil
}

func (o *SourceOwner) key(journal, event string) researchledger.ServiceIdentity {
	return researchledger.ServiceIdentity{Tenant: o.d.tenant, Stream: o.d.stream, Contract: RecordContract, Journal: journal, Event: event}
}

// lookup is called under o.mu. The index checks bounded storage identity; this
// layer additionally validates the actual canonical original and model contract.
func (o *SourceOwner) lookup(ctx context.Context, journal, event string) (RecordedPrediction, error) {
	if o.d.closed || o.d.stopped {
		return RecordedPrediction{}, errors.New("source owner stopped")
	}
	e, err := o.log.GetServiceAdmission(ctx, o.key(journal, event))
	if err != nil {
		return RecordedPrediction{}, err
	}
	return o.sourceOriginal(e, journal, event)
}

func (o *SourceOwner) sourceOriginal(e researchledger.Entry, journal, event string) (RecordedPrediction, error) {
	if e.Sequence <= 0 || e.Kind != "admit" || e.Key.Tenant != o.d.tenant || e.Key.Journal != o.d.stream || e.Key.Contract != RecordContract {
		return RecordedPrediction{}, errors.New("source original envelope mismatch")
	}
	var r RecordedPrediction
	if err := decodeCanonical(e.Payload, &r); err != nil {
		return RecordedPrediction{}, err
	}
	if err := r.Validate(); err != nil {
		return RecordedPrediction{}, err
	}
	if r.Binding == nil || r.Binding.Tenant != o.d.tenant || r.Binding.JournalID != journal || r.Binding.EventID != event || r.Seed != o.seed || r.Prediction.Epoch != o.d.epoch || e.Key.Event != strconv.FormatUint(r.Prediction.ID, 10) {
		return RecordedPrediction{}, errors.New("source original contract mismatch")
	}
	return r, nil
}

func (o *SourceOwner) Lookup(ctx context.Context, journal, event string) (RecordedPrediction, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.lookup(ctx, journal, event)
}

// Admit resolves all source identities before staging anything. A changed
// snapshot, feature or baseline under an existing key conflicts, not new evidence.
// The SQLite unique index reserves each new source in the original's transaction.
// Request memory is O(batch), not O(history). Startup scans the complete history
// and may construct an on-disk index; this is not a bounded-startup-time claim.
func (o *SourceOwner) Admit(ctx context.Context, requests []SourceAdmissionRequest) ([]AdmissionResult, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	var phaseStart time.Time
	if o.d.measurement != nil {
		phaseStart = time.Now()
	}
	if o.d.closed || o.d.stopped {
		return nil, errors.New("source owner stopped")
	}
	if len(requests) == 0 || len(requests) > 256 {
		return nil, errors.New("invalid source batch size")
	}
	var originals []sourceResolution
	if o.batchReads {
		refs := make([]SourceReference, len(requests))
		for i, r := range requests {
			if err := r.Binding.validate(); err != nil {
				return nil, err
			}
			if r.Binding.Tenant != o.d.tenant {
				return nil, errors.New("source tenant mismatch")
			}
			refs[i] = SourceReference{r.Binding.JournalID, r.Binding.EventID}
		}
		var err error
		originals, err = o.batchOriginals(ctx, refs)
		if err != nil {
			return nil, err
		}
	}
	o.d.worker.mu.Lock()
	next := o.d.worker.next
	o.d.worker.mu.Unlock()
	seen := make(map[researchledger.ServiceIdentity]bool, len(requests))
	resolved := make([]AdmissionRequest, len(requests))
	for i, request := range requests {
		if err := request.Binding.validate(); err != nil {
			return nil, err
		}
		if request.Binding.Tenant != o.d.tenant {
			return nil, errors.New("source tenant mismatch")
		}
		key := o.key(request.Binding.JournalID, request.Binding.EventID)
		if seen[key] {
			return nil, errors.New("duplicate source in batch")
		}
		seen[key] = true
		var r RecordedPrediction
		var err error
		if o.batchReads {
			if originals[i].found {
				r = originals[i].record
			} else {
				err = sql.ErrNoRows
			}
		} else {
			r, err = o.lookup(ctx, key.Journal, key.Event)
		}
		var id uint64
		switch {
		case err == nil:
			if r.Prediction.Features != request.Features || r.Outer[0] != request.Baseline || !r.At.Equal(request.At) || !reflect.DeepEqual(*r.Binding, request.Binding) {
				return nil, errors.New("conflicting source retry")
			}
			id = r.Prediction.ID
		case errors.Is(err, sql.ErrNoRows):
			if next == ^uint64(0) {
				return nil, errors.New("source ID exhausted")
			}
			next++
			id = next
		default:
			return nil, err
		}
		binding := request.Binding
		resolved[i] = AdmissionRequest{ID: id, Features: request.Features, Baseline: request.Baseline, At: request.At, Binding: &binding}
	}
	if o.d.measurement != nil {
		o.d.measurement.SourceNS += time.Since(phaseStart).Nanoseconds()
	}
	if o.reuseAdmissions {
		if !o.batchReads || len(originals) != len(resolved) {
			return nil, errors.New("admission resolution unavailable")
		}
		return o.d.admitBatchResolved(ctx, resolved, true, originals)
	}
	return o.d.AdmitBatchWithSnapshotReads(ctx, resolved)
}

// Discard ends an unlabeled observation without turning it into negative evidence.
// Its source identity remains reserved, including after reopen.
func (o *SourceOwner) Discard(ctx context.Context, journal, event string, available time.Time) (bool, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	r, err := o.lookup(ctx, journal, event)
	if err != nil {
		return false, err
	}
	return o.d.Discard(ctx, r.Prediction.ID, available)
}

type SourceDiscardRequest struct {
	JournalID, EventID string
	Available          time.Time
}

// DiscardBatch resolves and validates the entire source list before the existing
// atomic terminal operation. Unlike a loop of Discard calls, a late unknown or
// conflicting member cannot leave earlier members partially discarded. The same
// private owner lock excludes admissions/close until the durable result is known.
func (o *SourceOwner) DiscardBatch(ctx context.Context, requests []SourceDiscardRequest) ([]bool, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.d.closed || o.d.stopped {
		return nil, errors.New("source owner stopped")
	}
	if len(requests) == 0 || len(requests) > 256 {
		return nil, errors.New("invalid source discard batch size")
	}
	var originals []sourceResolution
	if o.batchReads {
		refs := make([]SourceReference, len(requests))
		for i, r := range requests {
			refs[i] = SourceReference{r.JournalID, r.EventID}
		}
		var err error
		originals, err = o.batchOriginals(ctx, refs)
		if err != nil {
			return nil, err
		}
	}
	resolved := make([]DiscardRequest, len(requests))
	seen := make(map[researchledger.ServiceIdentity]bool, len(requests))
	for i, request := range requests {
		key := o.key(request.JournalID, request.EventID)
		if seen[key] {
			return nil, errors.New("duplicate source discard")
		}
		seen[key] = true
		var r RecordedPrediction
		var err error
		if o.batchReads {
			if originals[i].found {
				r = originals[i].record
			} else {
				err = sql.ErrNoRows
			}
		} else {
			r, err = o.lookup(ctx, request.JournalID, request.EventID)
		}
		if err != nil {
			return nil, err
		}
		resolved[i] = DiscardRequest{ID: r.Prediction.ID, Available: request.Available}
	}
	return o.d.DiscardBatchWithSnapshotReads(ctx, resolved)
}

func (o *SourceOwner) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.d.Close()
}
