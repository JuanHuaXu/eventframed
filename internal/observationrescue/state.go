// Package observationrescue integrates existing Anti-Pigeon revision primitives
// with the isolated MMM controller. It grants no production publication rights.
package observationrescue

import (
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/observation"
)

const Window = 64
const AuditCapacity = 128
const MinAudits = 32
const RefitEvery = 16

var Arms = []string{"frozen", "audit_learning", "cp_fallback", "ap_fallback", "cp_audit", "ap_audit"}

func ChangePolicy() bayes.ChangePolicy {
	return bayes.ChangePolicy{Hazard: .05, Threshold: .30, MaxRun: 64,
		FastRate: .25, SlowRate: .025, DriftThreshold: .30, DriftPersistence: 12,
		MinSamples: 20, CUSUMSlack: .10, CUSUMThreshold: 8, CooldownSamples: 20}
}
func GroupPolicy() bayes.GroupPolicy {
	return bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8,
		MaxMembers: 64, EquivalenceMargin: .15, EquivalenceThreshold: .80,
		MaxUncertainBorrowing: .10, SharedEvidenceWeight: .5}
}

// Monitor compares two full-stream reliability contexts. Calling the revision
// primitive is not proof of an externally certified target-law diameter.
type Monitor struct {
	posterior model.BayesianPosterior
	reference [Window]bool
	live      [Window]bool
	n         int
}

func (m *Monitor) Observe(reference, live, eligible bool) (model.BayesianRevision, bool) {
	m.reference[m.n%Window], m.live[m.n%Window] = reference, live
	m.n++
	var cp bool
	m.posterior, cp = bayes.ApplyOutcomeAuthorized(m.posterior, live, 1, ChangePolicy(), eligible)
	group := model.BayesianPosterior{PosteriorKey: "ap:research-attention-reliability",
		MemberEvidence: make(map[string]model.BayesianMemberEvidence)}
	for i := 0; i < min(m.n, Window); i++ {
		bayes.UpdateMemberEvidence(&group, "reference", m.reference[i], 1)
		bayes.UpdateMemberEvidence(&group, "live", m.live[i], 1)
	}
	if m.n < Window {
		return model.BayesianRevision{Action: model.BayesianRevisionRetain}, false
	}
	return bayes.AssessRevision(group, []string{"reference", "live"}, "live", cp, eligible, GroupPolicy()), cp
}

type Forecast struct {
	Probability float64
	Inspection  observation.Result
	Epoch       uint64
	Support     uint32
}
type Feedback struct {
	Sequence           int
	Epoch              uint64
	ReferenceCorrect   bool
	LiveCorrect        bool
	ValidationEligible bool
	Audit              *observation.Sample
}
type Transition struct {
	Action      model.BayesianRevisionAction `json:"action"`
	ChangePoint bool                         `json:"cp"`
	Invalidated bool                         `json:"invalidated"`
	Refitted    bool                         `json:"refitted"`
}

// State is owned by one stream worker. Base and replacement models are immutable.
// Prediction precedes feedback; duplicate feedback cannot add audit support.
type State struct {
	arm         string
	base        *observation.Model
	working     *observation.Model
	monitor     Monitor
	epoch       uint64
	next        int
	pending     bool
	revoked     bool
	audits      []observation.Sample
	auditCount  int
	FitCount    int
	FitDuration time.Duration
}

func New(base *observation.Model, arm string) (*State, error) {
	valid := false
	for _, name := range Arms {
		if arm == name {
			valid = true
		}
	}
	if base == nil || !valid {
		return nil, errors.New("invalid observation rescue configuration")
	}
	return &State{base: base, arm: arm, epoch: 1}, nil
}

func (s *State) Predict(reader observation.Reader, sequence int) (Forecast, error) {
	if s.pending || sequence != s.next {
		return Forecast{}, errors.New("prediction order violation")
	}
	m := s.base
	policy := "mmm"
	if s.working != nil {
		m = s.working
	} else if s.revoked {
		policy = "breadth"
	}
	inspection, err := observation.Run(m, reader, reader.Epoch(), policy, int64(sequence))
	if err != nil {
		return Forecast{}, err
	}
	p := inspection.Probability
	if s.revoked && s.working == nil {
		p = .5
	}
	s.pending = true
	return Forecast{p, inspection, s.epoch, m.Support()}, nil
}

func (s *State) Observe(f Feedback) (Transition, error) {
	if !s.pending || f.Sequence != s.next || f.Epoch != s.epoch {
		return Transition{}, errors.New("missing, duplicate, or stale feedback")
	}
	if f.Audit != nil && f.Audit.Bits >= observation.Universe {
		return Transition{}, errors.New("invalid complete audit")
	}
	revision, cp := s.monitor.Observe(f.ReferenceCorrect, f.LiveCorrect, f.ValidationEligible)
	transition := Transition{Action: revision.Action, ChangePoint: cp}
	trigger := false
	switch s.arm {
	case "cp_fallback", "cp_audit":
		trigger = cp
	case "ap_fallback", "ap_audit":
		trigger = revision.Action != model.BayesianRevisionRetain
	}
	// Revocation is latched for this original dependency. Repeated evidence
	// against it cannot repeatedly erase a newly learning, unshared replacement.
	if trigger && !s.revoked {
		s.revoked = true
		s.working = nil
		s.audits = nil
		s.auditCount = 0
		s.epoch++
		transition.Invalidated = true
	}
	learn := s.arm == "audit_learning" || (s.revoked && (s.arm == "cp_audit" || s.arm == "ap_audit"))
	if learn && f.Audit != nil {
		s.audits = append(s.audits, *f.Audit)
		if len(s.audits) > AuditCapacity {
			s.audits = s.audits[1:]
		}
		s.auditCount++
		if s.auditCount >= MinAudits && (s.auditCount-MinAudits)%RefitEvery == 0 {
			start := time.Now()
			next, err := observation.Fit(s.audits)
			if err != nil {
				return Transition{}, err
			}
			s.FitDuration += time.Since(start)
			s.working = next
			s.FitCount++
			s.epoch++
			transition.Refitted = true
		}
	}
	s.pending = false
	s.next++
	return transition, nil
}
