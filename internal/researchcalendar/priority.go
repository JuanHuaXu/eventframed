package researchcalendar

import (
	"context"
	"errors"
	"math"

	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

type TemporalDecision struct {
	EventID      string
	State        string // unknown, compatible, or contradicted; never a truth certificate
	SourceField  string
	CalendarDate string
}

type PriorityPlan struct {
	Order             []int
	Decisions         []TemporalDecision // original input order
	AllContradicted   bool
	CalibrationStatus string
	Method            string
}

// PlanPriority returns a permutation only. Caller must apply it AFTER all score
// corrections and BEFORE packing, without another numeric sort. It never changes
// input scores, proper laws, or the certainty used to modulate earlier deltas.
func PlanPriority(ctx context.Context, req retrieval.RankRequest) (PriorityPlan, error) {
	if err := ctx.Err(); err != nil {
		return PriorityPlan{}, err
	}
	b, ok := ctx.Value(bindingKey{}).(binding)
	if !ok || b.tenant == "" || b.tenant != req.UserID || len(b.original) > 4096 || frame.QueryText(Focus(b.original)) != req.QueryText {
		return PriorityPlan{}, errors.New("calendar priority binding mismatch")
	}
	n := len(req.Candidates)
	if n == 0 || n > 200 || req.K1 != n || req.K2 != n {
		return PriorityPlan{}, errors.New("calendar priority requires complete bounded frontier")
	}
	p := PriorityPlan{CalibrationStatus: "not_evaluated", Method: "research/calendar-priority-v1"}
	rs := rules(b.original)
	good, bad := make([]int, 0, n), make([]int, 0, n)
	seen := make(map[string]bool, n)
	for i, c := range req.Candidates {
		if err := ctx.Err(); err != nil {
			return PriorityPlan{}, err
		}
		if c.ID == "" || seen[c.ID] || len(c.Text) > 8192 || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) || c.Score < 0 || c.Score > 1 {
			return PriorityPlan{}, errors.New("invalid calendar priority candidate")
		}
		seen[c.ID] = true
		d := TemporalDecision{EventID: c.ID, State: "unknown", SourceField: "what"}
		v, fieldOK := what(c.Text)
		date, dateOK := Date(v)
		if fieldOK && dateOK && len(rs) > 0 {
			d.CalendarDate = date.Format("2006-01-02")
			d.State = "compatible"
			for _, r := range rs {
				if r.kind == "before" && date.Year() >= r.year || r.kind == "after" && date.Year() <= r.year || r.kind == "on" && !date.Equal(r.date) || r.kind == "exclude" && date.Equal(r.date) {
					d.State = "contradicted"
					break
				}
			}
		}
		p.Decisions = append(p.Decisions, d)
		if d.State == "contradicted" {
			bad = append(bad, i)
		} else {
			good = append(good, i)
		}
	}
	p.AllContradicted = len(bad) == n
	if p.AllContradicted {
		for i := 0; i < n; i++ {
			p.Order = append(p.Order, i)
		}
	} else {
		p.Order = append(good, bad...)
	}
	if err := ctx.Err(); err != nil {
		return PriorityPlan{}, err
	}
	return p, nil
}
