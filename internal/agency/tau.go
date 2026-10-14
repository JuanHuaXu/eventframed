package agency

import (
	"errors"
	"math"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

const MaxTauPoints = 32
const MaxTauHorizonMS int64 = 7 * 24 * 60 * 60 * 1000

// PlanTau maximizes a conservative expected incremental utility over a finite
// candidate set. The TV allowance is supplied by the forecast contract; this
// arithmetic does not establish the allowance's empirical coverage.
func PlanTau(f model.TauForecast, p model.TauTurnPolicy, now time.Time) (model.TauDecision, error) {
	if err := ValidateTau(f, p, now); err != nil {
		return model.TauDecision{}, err
	}
	best := model.TauDecision{Reason: "no admissible positive-utility turn"}
	var candidates []int64
	if p.CandidateDelaysMS != nil {
		candidates = *p.CandidateDelaysMS
	}
	if p.CandidateDelaysMS == nil {
		for _, point := range f.Points {
			if point.Mark == p.TargetKind {
				candidates = append(candidates, point.DelayMS-p.LeadMS-p.SlackMS)
			}
		}
	}
	for _, offset := range candidates {
		s := f.Origin.Add(time.Duration(offset) * time.Millisecond)
		latest := s.Add(time.Duration(p.SlackMS) * time.Millisecond)
		if s.Before(now.Add(time.Duration(p.MinDelayMS)*time.Millisecond)) || latest.After(f.ValidUntil) || tauQuietBlocked(s, latest, p) {
			continue
		}
		// Intersect the useful windows of every allowed actual dispatch time.
		lo := s.Sub(f.Origin).Milliseconds() + p.LeadMS + p.SlackMS
		hi := s.Sub(f.Origin).Milliseconds() + p.LeadMS + p.WindowMS
		mass := 0.0
		for _, candidate := range f.Points {
			if candidate.Mark == p.TargetKind && candidate.DelayMS >= lo && candidate.DelayMS <= hi {
				mass += candidate.Probability
			}
		}
		lower := math.Max(0, math.Min(1, mass)-f.TVRadius)
		u := (p.Benefit+p.FalseCost)*lower - p.FalseCost - p.TurnCost - 1e-12
		if lower < p.MinProbability || u <= p.Margin {
			continue
		}
		if !best.Schedule || u > best.LowerUtility || u == best.LowerUtility && s.Before(best.DeliverAt) {
			best = model.TauDecision{Schedule: true, Reason: "positive conservative turn utility", DeliverAt: s, LatestDispatch: latest, Probability: mass, LowerProbability: lower, LowerUtility: u}
		}
	}
	return best, nil
}

func ValidateTau(f model.TauForecast, p model.TauTurnPolicy, now time.Time) error {
	if !boundedString(f.ID, 256) || f.Origin.IsZero() || f.IssuedAt.Before(f.Origin) || f.IssuedAt.After(now) || !f.ValidUntil.After(now) || f.HorizonMS < 1 || f.HorizonMS > MaxTauHorizonMS || f.ValidUntil.After(f.Origin.Add(time.Duration(f.HorizonMS)*time.Millisecond)) {
		return errors.New("invalid tau forecast identity, as-of time or horizon")
	}
	if len(f.Points) > MaxTauPoints || !probability(f.NoEvent) || !probability(f.TVRadius) {
		return errors.New("invalid tau support, no-event mass or TV radius")
	}
	mass := f.NoEvent
	last := int64(0)
	seen := map[TauAtomKey]bool{}
	for _, point := range f.Points {
		if !boundedString(point.Mark, 256) || point.DelayMS < 1 || point.DelayMS > f.HorizonMS || point.DelayMS < last || !probability(point.Probability) {
			return errors.New("invalid or unsorted tau atom")
		}
		key := TauAtomKey{point.Mark, point.DelayMS}
		if seen[key] {
			return errors.New("duplicate tau atom")
		}
		seen[key] = true
		last = point.DelayMS
		mass += point.Probability
	}
	if math.Abs(mass-1) > 1e-12 {
		return errors.New("tau law including no-event mass must sum to one")
	}
	if !boundedString(p.TargetKind, 256) || !probability(p.Benefit) || !probability(p.FalseCost) || !probability(p.TurnCost) || !probability(p.MinProbability) || !probability(p.Margin) {
		return errors.New("invalid tau utility policy")
	}
	if p.Margin == 0 {
		return errors.New("tau utility margin must be positive")
	}
	var supplied []int64
	if p.CandidateDelaysMS != nil {
		supplied = *p.CandidateDelaysMS
	}
	if len(supplied) > MaxTauPoints {
		return errors.New("too many tau candidates")
	}
	lastCandidate := int64(-1)
	for _, s := range supplied {
		if s < 0 || s > f.HorizonMS || s <= lastCandidate {
			return errors.New("tau candidates must be ordered distinct horizon offsets")
		}
		lastCandidate = s
	}
	if (p.QuietStartUTC == nil) != (p.QuietEndUTC == nil) {
		return errors.New("tau quiet hours require both endpoints")
	}
	if p.QuietStartUTC != nil && (*p.QuietStartUTC < 0 || *p.QuietStartUTC > 23 || *p.QuietEndUTC < 0 || *p.QuietEndUTC > 23 || *p.QuietStartUTC == *p.QuietEndUTC) {
		return errors.New("invalid tau quiet hours")
	}
	if p.LeadMS < 0 || p.LeadMS > MaxTauHorizonMS || p.SlackMS < 2000 || p.SlackMS > MaxTauHorizonMS || p.WindowMS < p.SlackMS || p.WindowMS > MaxTauHorizonMS || p.MinDelayMS < 1000 || p.MinDelayMS > MaxTauHorizonMS || p.CooldownMS < p.SlackMS || p.CooldownMS > MaxTauHorizonMS {
		return errors.New("invalid tau lead, dispatch slack, window, delay or cooldown")
	}
	return nil
}

func probability(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= 1 }

type TauAtomKey struct {
	mark  string
	delay int64
}

func tauQuietBlocked(start, end time.Time, p model.TauTurnPolicy) bool {
	if p.QuietStartUTC == nil {
		return false
	}
	day := time.Date(start.UTC().Year(), start.UTC().Month(), start.UTC().Day(), 0, 0, 0, 0, time.UTC).Add(-24 * time.Hour)
	for !day.After(end) {
		lo := day.Add(time.Duration(*p.QuietStartUTC) * time.Hour)
		hi := day.Add(time.Duration(*p.QuietEndUTC) * time.Hour)
		if !hi.After(lo) {
			hi = hi.Add(24 * time.Hour)
		}
		if start.Before(hi) && !end.Before(lo) {
			return true
		}
		day = day.Add(24 * time.Hour)
	}
	return false
}

// FitTau is a Dirichlet(1,...,1) histogram working forecast. It preserves a
// no-event atom and uses complete evidence windows, not unobserved futures.
func FitTau(id string, origin, now time.Time, horizon int64, bins []int64, episodes []model.TauEpisode, events map[string]model.Event, target string) (model.TauForecast, error) {
	if len(episodes) < 1 || len(episodes) > 15 || len(bins) < 1 || len(bins) > MaxTauPoints || horizon < 1 || horizon > MaxTauHorizonMS {
		return model.TauForecast{}, errors.New("bounded closed episodes and timing bins required")
	}
	counts := make([]float64, len(bins)+1)
	last := int64(0)
	for i, b := range bins {
		if b <= last || b > horizon {
			return model.TauForecast{}, errors.New("timing bin edges must increase inside horizon")
		}
		last = b
		counts[i] = 1
	}
	if last != horizon {
		return model.TauForecast{}, errors.New("last timing bin must close the full horizon")
	}
	counts[len(bins)] = 1
	used := map[string]bool{}
	type interval struct {
		session string
		lo, hi  time.Time
	}
	intervals := []interval{}
	for _, ep := range episodes {
		a, aok := events[ep.AnchorID]
		o, ook := events[ep.OutcomeID]
		if !aok || !ook || ep.AnchorID == ep.OutcomeID || used[ep.AnchorID] || used[ep.OutcomeID] || a.AvailableAt.After(now) || o.AvailableAt.After(now) || a.SessionID != o.SessionID || !o.OccurredAt.After(a.OccurredAt) {
			return model.TauForecast{}, errors.New("episodes require distinct available matching-session evidence")
		}
		used[ep.AnchorID] = true
		used[ep.OutcomeID] = true
		if len(intervals) > 0 {
			prior := intervals[len(intervals)-1]
			if a.SessionID < prior.session || a.SessionID == prior.session && !a.OccurredAt.After(prior.hi) {
				return model.TauForecast{}, errors.New("tau episodes must be session/time ordered and disjoint")
			}
		}
		intervals = append(intervals, interval{a.SessionID, a.OccurredAt, o.OccurredAt})
		delay := o.OccurredAt.Sub(a.OccurredAt)
		if ep.NoEvent {
			if o.Kind != "tau_no_event" || delay < time.Duration(horizon)*time.Millisecond {
				return model.TauForecast{}, errors.New("no-event evidence must explicitly close the horizon")
			}
			counts[len(bins)]++
			continue
		}
		if o.Kind != target || delay > time.Duration(horizon)*time.Millisecond {
			return model.TauForecast{}, errors.New("target outcome must be in horizon")
		}
		for i, b := range bins {
			if delay <= time.Duration(b)*time.Millisecond {
				counts[i]++
				break
			}
		}
	}
	total := float64(len(episodes) + len(counts))
	f := model.TauForecast{ID: id, Origin: origin, IssuedAt: now, ValidUntil: origin.Add(time.Duration(horizon) * time.Millisecond), HorizonMS: horizon, NoEvent: counts[len(bins)] / total}
	for i, b := range bins {
		f.Points = append(f.Points, model.TauPoint{Mark: target, DelayMS: b, Probability: counts[i] / total})
	}
	return f, nil
}
