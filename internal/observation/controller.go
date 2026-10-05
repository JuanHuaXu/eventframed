// Package observation is an isolated finite observation-allocation prototype.
// It is not wired into serving and does not publish beliefs or causal claims.
package observation

import (
	"errors"
	"math"
	"math/bits"
	"math/rand"
)

const Dimensions = 9
const Universe = 1 << Dimensions

type Sample struct {
	Bits    uint16
	Outcome bool
}
type count struct{ n, yes uint32 }

// Model is immutable after fitting. Repeated inspection must never add support.
type Model struct{ table [Universe][Universe]count }

func Fit(samples []Sample) (*Model, error) {
	if len(samples) == 0 || len(samples) > 8192 {
		return nil, errors.New("fitting size outside 1..8192")
	}
	var full [Universe]count
	for _, s := range samples {
		if s.Bits >= Universe {
			return nil, errors.New("invalid fitting bits")
		}
		c := &full[s.Bits]
		c.n++
		if s.Outcome {
			c.yes++
		}
	}
	m := new(Model)
	for mask := 0; mask < Universe; mask++ {
		for value, c := range full {
			p := &m.table[mask][value&mask]
			p.n += c.n
			p.yes += c.yes
		}
	}
	return m, nil
}

func (m *Model) Support() uint32 { return m.table[0][0].n }
func (m *Model) predict(mask, value uint16) float64 {
	c := m.table[mask][value&mask]
	return float64(c.yes+1) / float64(c.n+2)
}

type View struct {
	Scope int `json:"scope"`
	Depth int `json:"depth"`
}

func (v View) Mask() uint16 { return uint16((1<<(v.Depth+1))-1) << (3 * v.Scope) }

type Reader interface {
	Epoch() uint64
	Read(View) (mask, values uint16, err error)
}

type Step struct {
	View        View    `json:"view"`
	Attempted   uint16  `json:"attempted"`
	Observed    uint16  `json:"observed"`
	Values      uint16  `json:"values"`
	Probability float64 `json:"probability"`
}
type Result struct {
	Probability float64 `json:"probability"`
	Cost        int     `json:"cost"`
	Observed    int     `json:"observed"`
	Trace       []Step  `json:"trace"`
	Stop        string  `json:"stop"`
}

var Policies = []string{"fixed", "scope", "depth", "breadth", "depth_first", "random", "mmm", "exhaustive"}

func entropy(p float64) float64 { return -p*math.Log(p) - (1-p)*math.Log1p(-p) }

func (m *Model) value(mask, value, added uint16) float64 {
	parent := m.table[mask][value&mask]
	if parent.n == 0 {
		return 0
	}
	expected := 0.0
	// At most eight possible observations for a three-coordinate view.
	for sub := added; ; sub = (sub - 1) & added {
		c := m.table[mask|added][(value&mask)|sub]
		if c.n > 0 {
			expected += float64(c.n) / float64(parent.n) * entropy(m.predict(mask|added, value|sub))
		}
		if sub == 0 {
			break
		}
	}
	return math.Max(0, entropy(m.predict(mask, value))-expected)
}

func Run(m *Model, reader Reader, epoch uint64, policy string, seed int64) (Result, error) {
	valid := false
	for _, p := range Policies {
		if p == policy {
			valid = true
		}
	}
	if m == nil || reader == nil || !valid {
		return Result{}, errors.New("invalid model, reader, or policy")
	}
	budget := 6
	if policy == "exhaustive" {
		budget = 9
	}
	rng := rand.New(rand.NewSource(seed))
	var mask, values, attempted uint16
	result := Result{Probability: m.predict(0, 0)}
	for result.Cost < budget {
		if reader.Epoch() != epoch {
			return Result{}, errors.New("stale observation snapshot")
		}
		if result.Cost > 0 && policy != "exhaustive" && (result.Probability <= .1 || result.Probability >= .9) {
			result.Stop = "confidence"
			break
		}
		var candidates []View
		for scope := 0; scope < 3; scope++ {
			for depth := 0; depth < 3; depth++ {
				v := View{scope, depth}
				cost := bits.OnesCount16(v.Mask() &^ attempted)
				if cost == 0 || cost > budget-result.Cost {
					continue
				}
				if result.Cost == 0 && v != (View{0, 0}) {
					continue
				}
				if policy == "fixed" && v != (View{0, 0}) {
					continue
				}
				if policy == "scope" && depth != 0 {
					continue
				}
				if policy == "depth" && scope != 0 {
					continue
				}
				candidates = append(candidates, v)
			}
		}
		if len(candidates) == 0 {
			result.Stop = "no_affordable_view"
			break
		}
		chosen := candidates[0]
		if policy == "random" {
			chosen = candidates[rng.Intn(len(candidates))]
		}
		if policy == "breadth" || policy == "scope" {
			for _, v := range candidates {
				if v.Depth < chosen.Depth || (v.Depth == chosen.Depth && v.Scope < chosen.Scope) {
					chosen = v
				}
			}
		}
		if policy == "mmm" {
			best := -1.0
			for _, v := range candidates {
				added := v.Mask() &^ attempted
				score := m.value(mask, values, added) / float64(bits.OnesCount16(added))
				if score > best+1e-12 {
					best = score
					chosen = v
				}
			}
		}
		newMask, newValues, err := reader.Read(chosen)
		if err != nil {
			return Result{}, err
		}
		if reader.Epoch() != epoch {
			return Result{}, errors.New("snapshot changed during observation")
		}
		if newMask&^chosen.Mask() != 0 || newValues&^newMask != 0 {
			return Result{}, errors.New("reader returned undeclared coordinates")
		}
		if (newValues^values)&(newMask&mask) != 0 {
			return Result{}, errors.New("snapshot changed observed values")
		}
		result.Cost += bits.OnesCount16(chosen.Mask() &^ attempted)
		attempted |= chosen.Mask()
		mask |= newMask
		values |= newValues
		result.Probability = m.predict(mask, values)
		result.Trace = append(result.Trace, Step{chosen, attempted, mask, values, result.Probability})
	}
	if result.Stop == "" {
		result.Stop = "budget"
	}
	result.Observed = bits.OnesCount16(mask)
	return result, nil
}
