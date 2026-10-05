package researchmemory

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// BoundLabels reads the complete terminal evidence available by cutoff from
// this exclusively owned stream. The caller cannot select favorable IDs.
// Source validity still requires a separate target-snapshot validator.
func (d *Durable) BoundLabels(ctx context.Context, cutoff time.Time) ([]BoundLabel, error) {
	labels, _, err := d.BoundLabelsWithSequence(ctx, cutoff)
	return labels, err
}

// BoundLabelsWithSequence returns a quiescence token for a later guarded
// publication. It is invalidated by any later durable admission or terminal.
func (d *Durable) BoundLabelsWithSequence(ctx context.Context, cutoff time.Time) ([]BoundLabel, int64, error) {
	if ctx == nil || cutoff.IsZero() {
		return nil, 0, errors.New("invalid bound-label read")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.stopped {
		return nil, 0, errors.New("durable stream stopped")
	}
	predictions := make(map[uint64]RecordedPrediction)
	labels := make([]BoundLabel, 0)
	var after int64
	var nextID uint64
	var lastFeedback time.Time
	for {
		page, err := d.log.ReadAfter(ctx, after, 128)
		if err != nil {
			return nil, 0, err
		}
		if len(page) == 0 {
			return labels, after, nil
		}
		for _, entry := range page {
			if err := ctx.Err(); err != nil {
				return nil, 0, err
			}
			if entry.Sequence <= after || entry.Key.Tenant != d.tenant || entry.Key.Journal != d.stream || entry.Key.Contract != RecordContract {
				return nil, 0, errors.New("bound-label stream identity mismatch")
			}
			id, err := strconv.ParseUint(entry.Key.Event, 10, 64)
			if err != nil || id == 0 || entry.Key.Event != strconv.FormatUint(id, 10) {
				return nil, 0, errors.New("invalid bound-label event identity")
			}
			switch entry.Kind {
			case "admit":
				var prediction RecordedPrediction
				if err := decodeCanonical(entry.Payload, &prediction); err != nil {
					return nil, 0, err
				}
				if err := prediction.Validate(); err != nil {
					return nil, 0, err
				}
				if id != nextID+1 || prediction.Prediction.ID != id || prediction.Prediction.Epoch != d.epoch || prediction.Seed != d.worker.adapter.seed || (prediction.Binding != nil && prediction.Binding.Tenant != d.tenant) {
					return nil, 0, errors.New("bound-label admission order mismatch")
				}
				nextID = id
				predictions[id] = prediction
				if len(predictions) > 256 {
					return nil, 0, errors.New("bound-label pending history exceeds learner capacity")
				}
			case "feedback":
				prediction, admitted := predictions[id]
				if !admitted {
					return nil, 0, errors.New("terminal record without admission")
				}
				delete(predictions, id)
				var feedback RecordedFeedback
				if err := decodeCanonical(entry.Payload, &feedback); err == nil {
					if feedback.ID != id || feedback.Available.IsZero() || feedback.Available.Before(prediction.At) || feedback.Available.Before(lastFeedback) {
						return nil, 0, errors.New("invalid bound-label feedback")
					}
					lastFeedback = feedback.Available
					if !feedback.Available.After(cutoff) {
						if prediction.Binding == nil || prediction.Witness == nil || len(labels) == 256 {
							return nil, 0, errors.New("ineligible or oversized bound-label transfer")
						}
						labels = append(labels, BoundLabel{Prediction: prediction, Feedback: feedback})
					}
				} else {
					var discard RecordedDiscard
					if err := decodeCanonical(entry.Payload, &discard); err != nil || !discard.Discard || discard.ID != id || discard.Available.IsZero() || discard.Available.Before(prediction.At) {
						return nil, 0, errors.New("invalid bound-label discard")
					}
				}
			default:
				return nil, 0, fmt.Errorf("invalid bound-label kind %q", entry.Kind)
			}
			after = entry.Sequence
		}
	}
}

// WithUnchangedSequence excludes concurrent durable appends through work.
// Call it only after acquiring the store publication guard; work must not
// reenter this durable stream or the store mutator.
func (d *Durable) WithUnchangedSequence(ctx context.Context, sequence int64, work func() error) error {
	if ctx == nil || sequence < 0 || work == nil {
		return errors.New("invalid durable publication guard")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.stopped {
		return errors.New("durable stream stopped")
	}
	rows, err := d.log.ReadAfter(ctx, sequence, 1)
	if err != nil {
		return err
	}
	if len(rows) != 0 {
		return errors.New("durable stream changed before publication")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return work()
}
