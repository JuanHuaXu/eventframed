package researchmemory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// Discard durably consumes an unresolved prediction without updating the model
// or its evidence clock. Available dates abandonment, not an observed outcome.
func (d *Durable) Discard(ctx context.Context, id uint64, available time.Time) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed || d.stopped {
		return false, errors.New("durable stream stopped")
	}
	old, e := d.log.Get(ctx, d.key(id), "feedback")
	if e == nil {
		var r RecordedDiscard
		if e = decodeCanonical(old.Payload, &r); e != nil {
			return false, errors.New("prediction already labeled")
		}
		if !r.Discard || r.ID != id || !r.Available.Equal(available) {
			return false, errors.New("conflicting discard retry")
		}
		return true, nil
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return false, e
	}
	p, e := d.worker.Record(id)
	if e != nil || available.IsZero() || available.Before(p.At) {
		return false, errors.New("invalid discard")
	}
	raw, e := json.Marshal(RecordedDiscard{id, true, available})
	if e != nil {
		return false, e
	}
	if _, _, e = d.log.Append(ctx, d.key(id), "feedback", raw); e != nil {
		d.stopped = true
		return false, e
	}
	d.worker.Discard(id)
	return false, nil
}
