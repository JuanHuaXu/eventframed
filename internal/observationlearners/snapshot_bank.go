package observationlearners

import (
	"errors"
	"math"
)

// Owner retains at most two detached publications. Issued likelihood messages
// contain scalar probabilities, so late evidence does not retain retired tables.
// This component has no acceptance gate and is not a serving-policy replacement.
type snapshotBank struct {
	filter snapshotAdvice
	banks  [2]*observationExperts
}

func newSnapshotBank() *snapshotBank {
	f, _ := newSnapshotAdvice(.001)
	return &snapshotBank{filter: f}
}

func (b *snapshotBank) publish(version int, models [4]*ConditionalForest) error {
	if b == nil {
		return errors.New("nil snapshot bank")
	}
	c := *b
	if err := c.filter.publish(version); err != nil {
		return err
	}
	e, err := newObservationExperts(models)
	if err != nil {
		return err
	}
	// Common input measure permits one fixed mixture throughout acquisition.
	// A genuinely changed measure requires a different validated contract.
	if previous := c.banks[1-version%2]; previous != nil {
		for i, cell := range e.models[0].cells {
			if !jointClose(cell.mass/e.roots[0], previous.models[0].cells[i].mass/previous.roots[0]) {
				return errors.New("snapshot banks have different input laws")
			}
		}
	}
	c.banks[version%2] = e
	*b = c
	return nil
}

type snapshotLaw struct {
	banks   [2]*observationExperts
	ids     [8]uint64
	weights [8]float64
}

func (b *snapshotBank) snapshot() (snapshotLaw, error) {
	if b == nil || !b.filter.ready {
		return snapshotLaw{}, errors.New("nil snapshot bank")
	}
	w, err := b.filter.current.weights()
	if err != nil {
		return snapshotLaw{}, err
	}
	return snapshotLaw{banks: b.banks, ids: b.filter.current.ids, weights: w}, nil
}

func (s snapshotLaw) raw(mask, values uint16) ([8]float64, error) {
	var out [8]float64
	if mask >= 512 || values&^mask != 0 {
		return out, errors.New("invalid snapshot observation")
	}
	for i, id := range s.ids {
		if id == 0 {
			continue
		}
		e := s.banks[i/4]
		if e == nil || !e.ready {
			return out, errors.New("missing snapshot publication")
		}
		p, err := e.models[i%4].Forecast(mask, values)
		if err != nil {
			return out, err
		}
		out[i] = p
	}
	return out, nil
}

func (s snapshotLaw) Forecast(mask, values uint16) (float64, error) {
	raw, err := s.raw(mask, values)
	if err != nil {
		return 0, err
	}
	total, p := 0., 0.
	for i, w := range s.weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || s.ids[i] == 0 && w != 0 {
			return 0, errors.New("invalid snapshot weight")
		}
		total += w
		p += w * raw[i]
	}
	if math.Abs(total-1) > 1e-12 {
		return 0, errors.New("empty or unnormalized snapshot law")
	}
	return p / total, nil
}

func (b *snapshotBank) issue(origin uint64, s snapshotLaw, mask, values uint16) error {
	if b == nil {
		return errors.New("nil snapshot bank")
	}
	current, err := b.snapshot()
	if err != nil {
		return err
	}
	// A detached preview remains readable but cannot issue after feedback or
	// publication changes the current law, even if generation IDs still match.
	if current != s {
		return errors.New("stale snapshot preview")
	}
	raw, err := s.raw(mask, values)
	if err != nil {
		return err
	}
	return b.filter.issue(origin, s.ids, raw)
}
