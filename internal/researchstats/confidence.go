// Package researchstats provides research-only inference over prespecified
// paired stream means. It does not validate experimental independence or
// preregistration, and is not connected to production or experiment verdicts.
package researchstats

import (
	"errors"
	"math"
	"strings"
)

// Counts are capped where every integer is exactly representable in float64.
const maxUnits uint64 = 1 << 53

type meanState struct {
	n               uint64
	sum, correction float64
}

func (s *meanState) add(x float64) {
	// Neumaier summation preserves small paired gains amid cancellation.
	t := s.sum + x
	if math.Abs(s.sum) >= math.Abs(x) {
		s.correction += (s.sum - t) + x
	} else {
		s.correction += (x - t) + s.sum
	}
	s.sum = t
	s.n++
}

func (s meanState) mean() float64 {
	if s.n == 0 {
		return 0
	}
	return math.Max(-1, math.Min(1, (s.sum+s.correction)/float64(s.n)))
}

// ConfidenceSequence owns a frozen family of two-sided comparisons, sharing
// one familywise alpha equally. Each comparison has its own stream count.
// Use NewConfidenceSequence; the zero value is invalid. A single goroutine
// must own it, and it must not be copied, reset, or restarted to recycle alpha.
// Merely bounding observations is insufficient: each comparison must satisfy
// E[X_n | F_{n-1}] = mu for one constant mu, with X_n in [-1,1].
type ConfidenceSequence struct {
	alpha       float64
	comparisons []string
	states      map[string]*meanState
	logBudget   float64
}

// Interval is centered at the observed stream mean with an unclipped Radius.
// Lower and Upper are clipped to [-1,1]. At N=0 they are [-1,1], Radius=2,
// and Mean=0 is only a placeholder, not an estimate. Intervals need not nest.
type Interval struct {
	N                          uint64
	Mean, Radius, Lower, Upper float64
}

// NewConfidenceSequence freezes alpha in (0,1) and the complete, nonempty
// comparison family. Names must be unique and nonblank. Input names are copied.
// Freezing in code is not evidence of freezing before inspecting outcomes.
func NewConfidenceSequence(alpha float64, comparisons []string) (*ConfidenceSequence, error) {
	if !finite(alpha) || alpha <= 0 || alpha >= 1 || len(comparisons) == 0 {
		return nil, errors.New("invalid alpha or empty comparison family")
	}
	s := &ConfidenceSequence{
		alpha: alpha, comparisons: append([]string(nil), comparisons...),
		states:    make(map[string]*meanState, len(comparisons)),
		logBudget: math.Log(2) + math.Log(float64(len(comparisons))) - math.Log(alpha),
	}
	for _, name := range s.comparisons {
		if strings.TrimSpace(name) == "" {
			return nil, errors.New("blank comparison name")
		}
		if _, exists := s.states[name]; exists {
			return nil, errors.New("duplicate comparison name")
		}
		s.states[name] = &meanState{}
	}
	return s, nil
}

// Configuration returns copies of the frozen family specification.
func (s *ConfidenceSequence) Configuration() (float64, []string, error) {
	if s == nil || len(s.states) == 0 {
		return 0, nil, errors.New("uninitialized confidence sequence")
	}
	return s.alpha, append([]string(nil), s.comparisons...), nil
}

func (s *ConfidenceSequence) state(comparison string) (*meanState, error) {
	if s == nil || len(s.states) == 0 {
		return nil, errors.New("uninitialized confidence sequence")
	}
	v, ok := s.states[comparison]
	if !ok {
		return nil, errors.New("comparison outside frozen family")
	}
	return v, nil
}

// AddStream adds ONE completed paired stream mean, not one clock or forecast.
// Positive values mean control loss minus candidate loss (candidate benefit).
// If fits are the independent units, supply one prespecified cluster mean per
// fit instead. The caller must prevent duplicates and outcome-selected ordering,
// exclusion, windows, or weighting. Invalid input leaves all state unchanged.
func (s *ConfidenceSequence) AddStream(comparison string, pairedMean float64) error {
	v, err := s.state(comparison)
	if err != nil {
		return err
	}
	if !finite(pairedMean) || pairedMean < -1 || pairedMean > 1 {
		return errors.New("paired stream mean must be finite and in [-1,1]")
	}
	if v.n >= maxUnits {
		return errors.New("stream count exceeds exact float64 integer range")
	}
	v.add(pairedMean)
	return nil
}

// Anytime returns the union-bound Hoeffding CS at the current stream count.
// With m frozen comparisons, delta_n=alpha/(m*n*(n+1)) and radius
// sqrt(2*log(2/delta_n)/n). The union over all comparisons and all n>=1
// fails with probability at most alpha under the constant conditional mean
// assumption. Dependence between comparisons is allowed. See the contract's proof.
func (s *ConfidenceSequence) Anytime(comparison string) (Interval, error) {
	return s.interval(comparison, true)
}

// FixedSample returns a different bound: sqrt(2*log(2*m/alpha)/n).
// Its familywise validity requires sample counts fixed independently of the
// evaluated outcomes. Repeated peeking or selecting N from losses invalidates
// that claim; use Anytime for optional stopping. This call spends no extra
// alpha, and the two methods are not two separate inferential opportunities.
func (s *ConfidenceSequence) FixedSample(comparison string) (Interval, error) {
	return s.interval(comparison, false)
}

func (s *ConfidenceSequence) interval(comparison string, anytime bool) (Interval, error) {
	v, err := s.state(comparison)
	if err != nil {
		return Interval{}, err
	}
	if v.n == 0 {
		return Interval{Radius: 2, Lower: -1, Upper: 1}, nil
	}
	n := float64(v.n)
	logTerm := s.logBudget
	if anytime {
		// Work in logs so tiny alpha and large n cannot under/overflow delta_n.
		logTerm += math.Log(n) + math.Log1p(n)
	}
	r := math.Sqrt(2 * logTerm / n)
	mean := v.mean()
	return Interval{N: v.n, Mean: mean, Radius: r,
		Lower: math.Max(-1, mean-r), Upper: math.Min(1, mean+r)}, nil
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
