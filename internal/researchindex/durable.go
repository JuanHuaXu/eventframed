package researchindex

import (
	"context"
	"errors"
	"math"
)

var ErrRecoveryRequired = errors.New("research generation requires durable recovery")

// PersistGeneration must synchronously commit the supplied records and revision
// atomically. Nil means durable success, not queued acknowledgement. Any error
// is conservatively unknown. The callback must not mutate or retain mutations.
type PersistGeneration func(context.Context, uint64, []Mutation) error

// DurableGenerations is a research coordinator, not a storage engine. It blocks
// new views during a transaction so an uncertain outcome cannot be presented as
// current state. Existing immutable views remain valid historical snapshots.
type DurableGenerations struct {
	core    *Generations
	gate    chan struct{}
	persist PersistGeneration
	failed  bool // protected by gate
}

// RestoreDurable requires an authoritative, coherent durable snapshot from an
// exclusive recovery session. It does not authenticate or enumerate that state.
func RestoreDurable(dimension, capacity int, revision uint64, records []Mutation, persist PersistGeneration) (*DurableGenerations, error) {
	if persist == nil {
		return nil, errors.New("missing durable callback")
	}
	core, err := NewGenerations(dimension, capacity)
	if err != nil {
		return nil, err
	}
	if revision == 0 && len(records) > 0 {
		return nil, errors.New("nonempty state with zero revision")
	}
	base := make(map[string]record, len(records))
	for _, r := range records {
		if r.ID == "" || r.Delete || len(r.Vector) != dimension {
			return nil, errors.New("invalid recovered record")
		}
		if _, ok := base[r.ID]; ok {
			return nil, errors.New("duplicate recovered ID")
		}
		var norm float64
		for _, v := range r.Vector {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				return nil, errors.New("invalid recovered vector")
			}
			norm += float64(v) * float64(v)
		}
		if norm == 0 {
			return nil, errors.New("zero recovered vector")
		}
		base[r.ID] = record{vector: append([]float32(nil), r.Vector...), revision: revision}
	}
	core.current.Store(&generation{base: &baseGeneration{records: base}, delta: map[string]record{}, revision: revision})
	d := &DurableGenerations{core: core, gate: make(chan struct{}, 1), persist: persist}
	d.gate <- struct{}{}
	return d, nil
}

func (d *DurableGenerations) lock(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-d.gate:
	}
	if err := ctx.Err(); err != nil {
		d.gate <- struct{}{}
		return err
	}
	return nil
}

func (d *DurableGenerations) View(ctx context.Context) (View, error) {
	if err := d.lock(ctx); err != nil {
		return View{}, err
	}
	defer func() { d.gate <- struct{}{} }()
	if d.failed {
		return View{}, ErrRecoveryRequired
	}
	return d.core.View(), nil
}

func (d *DurableGenerations) Apply(ctx context.Context, mutations []Mutation) error {
	if err := d.lock(ctx); err != nil {
		return err
	}
	defer func() { d.gate <- struct{}{} }()
	if d.failed {
		return ErrRecoveryRequired
	}
	p, err := d.core.Prepare(ctx, mutations)
	if err != nil {
		return err
	}
	settled := false
	defer func() {
		if !settled {
			// Discarding an unpublished candidate is not claiming storage rollback.
			// The coordinator stays quarantined until replaced from durable state.
			d.failed = true
			_ = p.Abort()
		}
	}()
	owned := make([]Mutation, 0, len(mutations))
	for _, r := range mutations {
		prepared := p.next.delta[r.ID]
		owned = append(owned, Mutation{ID: r.ID, Vector: append([]float32(nil), prepared.vector...), Delete: prepared.deleted})
	}
	if err := d.persist(ctx, p.next.revision, owned); err != nil {
		return errors.Join(ErrRecoveryRequired, err)
	}
	if err := p.Commit(); err != nil {
		return errors.Join(ErrRecoveryRequired, err)
	}
	settled = true
	return nil
}
