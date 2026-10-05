package observation

import (
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

// EventReader exposes supplied local/episode/process views of an as-of snapshot.
// No automatic scope discovery or natural-language extraction is claimed.
type EventReader struct {
	Frames  [3][2]model.Event // previous, current within each supplied scope
	Tenant  string
	AsOf    time.Time
	Version uint64
}

func (r *EventReader) Epoch() uint64 { return r.Version }
func (r *EventReader) Read(v View) (uint16, uint16, error) {
	if v.Scope < 0 || v.Scope >= 3 || v.Depth < 0 || v.Depth >= 3 || r.AsOf.IsZero() || r.Tenant == "" {
		return 0, 0, errors.New("invalid observation view")
	}
	var mask, values uint16
	if r.Frames[v.Scope][0].OccurredAt.After(r.Frames[v.Scope][1].OccurredAt) {
		return 0, 0, errors.New("previous frame occurs after current frame")
	}
	// This finite fixture model has distinct coordinate sources, not duplicated
	// projections of the same observation. A general overlap model is future work.
	seen := make(map[string]bool, 6)
	for _, scope := range r.Frames {
		for _, event := range scope {
			if event.ID == "" || seen[event.ID] {
				return 0, 0, errors.New("missing or duplicate coordinate source")
			}
			seen[event.ID] = true
		}
	}
	for d := 0; d <= v.Depth; d++ {
		index := 1
		if d == 2 {
			index = 0
		}
		e := r.Frames[v.Scope][index]
		if e.TenantID != r.Tenant {
			return 0, 0, errors.New("cross-tenant observation")
		}
		if err := e.Validate(0); err != nil {
			return 0, 0, err
		}
		if e.AvailableAt.After(r.AsOf) {
			continue
		}
		field := e.What
		if d == 1 {
			field = e.How
		}
		// Inferred explanations, raw text and absent values cannot become evidence.
		if field.Source != model.SourceObserved || (field.Value != "0" && field.Value != "1") {
			continue
		}
		bit := uint16(1) << (3*v.Scope + d)
		mask |= bit
		if field.Value == "1" {
			values |= bit
		}
	}
	return mask, values, nil
}
