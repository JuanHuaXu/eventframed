package researchmemory

import (
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"math"
	"time"
)

const RecordContract = "research-memory-original-forecast-v1"

// ServiceBinding records provenance/dependencies; it does not certify that they
// remain valid. A service bridge must verify them at admission and feedback.
type ServiceBinding struct {
	Tenant, JournalID, EventID string
	Snapshot                   model.Snapshot
}

func (b ServiceBinding) validate() error {
	if b.Tenant == "" || b.JournalID == "" || b.EventID == "" || len(b.Tenant) > 4096 || len(b.JournalID) > 4096 || len(b.EventID) > 4096 || b.Snapshot.RuntimeVersion == 0 || b.Snapshot.ContractVersion == 0 {
		return errors.New("invalid service binding")
	}
	return nil
}

// RecordedPrediction contains the actual pre-outcome experts. Validation checks
// shape and contract, not provenance: only an owned journal may supply records.
type RecordedPrediction struct {
	Contract     string
	Seed         int64
	Prediction   Prediction
	At           time.Time
	Outer, Inner [4]float64
	Ready        bool
	Binding      *ServiceBinding `json:",omitempty"`
	Witness      *SourceWitness  `json:",omitempty"`
}

func (a *Adapter) Record(id uint64) (RecordedPrediction, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	p, ok := a.pending[id]
	if !ok {
		return RecordedPrediction{}, errors.New("unknown prediction")
	}
	return RecordedPrediction{Contract: RecordContract, Seed: a.seed, Prediction: p.Prediction, At: p.At, Outer: p.Outer, Inner: p.Inner, Ready: p.Ready}, nil
}

// Record exports the immutable pre-outcome journal, not a new forecast. Call
// before Feedback consumes its pending identity; the worker's seed is immutable.
func (b *Background) Record(id uint64) (RecordedPrediction, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	p, ok := b.pending[id]
	if b.closed || !ok {
		return RecordedPrediction{}, errors.New("unknown background prediction")
	}
	return RecordedPrediction{Contract: RecordContract, Seed: b.adapter.seed, Prediction: p.Prediction, At: p.At, Outer: p.Outer, Inner: p.Inner, Ready: p.Ready}, nil
}

func (r RecordedPrediction) Validate() error {
	if r.Binding != nil {
		if e := r.Binding.validate(); e != nil {
			return e
		}
	}
	if r.Witness != nil {
		if r.Binding == nil {
			return errors.New("unbound source witness")
		}
		if e := r.Witness.Validate(); e != nil {
			return e
		}
	}
	if r.Contract != RecordContract || r.Prediction.ID == 0 || r.Prediction.Features >= 512 || r.At.IsZero() {
		return errors.New("invalid prediction record")
	}
	values := []float64{r.Prediction.Probability}
	values = append(values, r.Outer[:]...)
	values = append(values, r.Inner[:]...)
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
			return errors.New("invalid recorded probability")
		}
	}
	if !r.Ready && r.Prediction.Probability != r.Outer[0] {
		return errors.New("invalid cold forecast")
	}
	return nil
}

// RestorePrediction reconstructs an ordered journal on a replay-owned Adapter.
// It deliberately does not call Predict: current weights may differ from those
// that emitted this record. Complete admission order is required, including any
// later-abandoned predictions. External storage/consumer owns replay identity.
func (a *Adapter) RestorePrediction(r RecordedPrediction) error {
	if e := r.Validate(); e != nil {
		return e
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if r.Seed != a.seed || r.Prediction.Epoch != a.epoch || a.next == ^uint64(0) || r.Prediction.ID != a.next+1 || len(a.pending) >= 256 {
		return errors.New("record contract/order mismatch")
	}
	a.pending[r.Prediction.ID] = pending{Prediction: r.Prediction, At: r.At, Outer: r.Outer, Inner: r.Inner, Ready: r.Ready}
	a.next = r.Prediction.ID
	return nil
}
