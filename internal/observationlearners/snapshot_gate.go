package observationlearners

import (
	"errors"
	"math"
)

const snapshotGateBoundary = 12800. // 8 windows * 8 slots * 2 reserved views / .01

type snapshotGateTest struct {
	count    int
	active   [9]float64
	rejected bool
}

// Tests and first-crossing credits belong to (publication window, snapshot ID).
// Unrejected retained predictors get a separately budgeted new test window.
// Rejection persists for the same immutable ID; credits never follow slot reuse.
// Non-rejection is not a certificate that a snapshot is true.
type snapshotGate struct {
	epoch    int
	ids      [8]uint64
	boundary float64
	tests    [8]snapshotGateTest
	credits  [8][9]float64
}

func (g *snapshotGate) reset(epoch int, ids [8]uint64, boundary float64) error {
	if g == nil || epoch < 0 || epoch >= 8 || math.IsNaN(boundary) || math.IsInf(boundary, 0) || boundary <= 1 {
		return errors.New("invalid snapshot gate contract")
	}
	if g.boundary > 1 && (epoch != g.epoch+1 || boundary != g.boundary) {
		return errors.New("snapshot gate window or budget changed")
	}
	n := 0
	seen := map[uint64]bool{}
	for i, id := range ids {
		if id != 0 {
			if (id-1)%8 != uint64(i) || id > uint64(4*(epoch+1)) || int((id-1)/4) < epoch-1 {
				return errors.New("invalid gate generation")
			}
			if seen[id] {
				return errors.New("duplicate gate identity")
			}
			seen[id] = true
			n++
		}
	}
	if n != 4 && n != 8 {
		return errors.New("invalid gate population")
	}
	next := snapshotGate{epoch: epoch, ids: ids, boundary: boundary}
	for i, id := range ids {
		if id == 0 || id != g.ids[i] || !g.tests[i].rejected {
			continue
		}
		next.tests[i].rejected = true
		for j := range next.credits[i] {
			next.credits[i][j] = math.Inf(-1)
			if j == 0 || ids[j-1] != 0 && ids[j-1] == g.ids[j-1] {
				next.credits[i][j] = g.credits[i][j]
			}
		}
	}
	*g = next
	return nil
}

func (g *snapshotGate) route(ids [8]uint64, w [8]float64) ([9]float64, error) {
	var out [9]float64
	if g == nil || g.boundary <= 1 || ids != g.ids {
		return out, errors.New("stale snapshot gate")
	}
	total := 0.
	for i, v := range w {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || ids[i] == 0 && v != 0 {
			return out, errors.New("invalid snapshot gate weight")
		}
		total += v
	}
	if math.Abs(total-1) > 1e-12 {
		return out, errors.New("unnormalized snapshot gate weights")
	}
	eligible := func(j int) bool { return j == 0 || ids[j-1] != 0 && !g.tests[j-1].rejected }
	for i, mass := range w {
		if ids[i] == 0 {
			continue
		}
		if !g.tests[i].rejected {
			out[i+1] += mass
			continue
		}
		z := math.Inf(-1)
		for j, c := range g.credits[i] {
			if eligible(j) {
				if math.IsNaN(c) || math.IsInf(c, 1) {
					return out, errors.New("invalid snapshot rejection credit")
				}
				z = falsificationLogAdd(z, c)
			}
		}
		if math.IsInf(z, -1) {
			out[0] += mass
			continue
		}
		for j, c := range g.credits[i] {
			if eligible(j) {
				out[j] += mass * math.Exp(c-z)
			}
		}
	}
	total = 0
	for _, v := range out {
		total += v
	}
	if total <= 0 || math.IsNaN(total) {
		return out, errors.New("empty snapshot routed law")
	}
	for j := range out {
		out[j] /= total
	}
	return out, nil
}

func (g *snapshotGate) observe(epoch int, ids [8]uint64, raw [8]float64, y bool) error {
	if g == nil || g.boundary <= 1 || epoch != g.epoch || ids != g.ids {
		return errors.New("wrong snapshot gate owner")
	}
	n := 0
	for i, id := range ids {
		if id == 0 {
			if raw[i] != 0 {
				return errors.New("inactive gate emission")
			}
			continue
		}
		if math.IsNaN(raw[i]) || raw[i] <= 0 || raw[i] >= 1 || g.tests[i].count >= 32 {
			return errors.New("invalid gate emission or length")
		}
		n++
	}
	c := *g
	logp := func(p float64) float64 {
		if y {
			return math.Log(p)
		}
		return math.Log1p(-p)
	}
	for i, id := range ids {
		if id == 0 {
			continue
		}
		s := &c.tests[i]
		old := s.rejected
		s.count++
		dormant := math.Inf(-1)
		if s.count < 32 {
			dormant = math.Log(float64(32 - s.count))
		}
		var terms [9]float64
		for j := range terms {
			terms[j] = math.Inf(-1)
		}
		z := math.Inf(-1)
		for j := 0; j < 9; j++ {
			if j == i+1 || j > 0 && ids[j-1] == 0 {
				continue
			}
			q, prior := .5, .5
			if j > 0 {
				q = raw[j-1]
				prior = .5 / float64(n-1)
			}
			previous := s.active[j]
			if s.count == 1 {
				previous = math.Inf(-1)
			}
			s.active[j] = logp(q) - logp(raw[i]) + falsificationLogAdd(previous, 0)
			terms[j] = math.Log(prior) + falsificationLogAdd(s.active[j], dormant) - math.Log(32)
			z = falsificationLogAdd(z, terms[j])
		}
		if z >= math.Log(c.boundary) {
			s.rejected = true
		}
		if !old && s.rejected {
			for j := range terms {
				c.credits[i][j] = terms[j] - z
			}
		}
	}
	*g = c
	return nil
}

type snapshotRoutedLaw struct {
	base    snapshotLaw
	weights [9]float64
}

func (s snapshotRoutedLaw) cell(mask, values uint16) conditionalCell {
	i := partialIndex(mask, values)
	var mass float64
	for _, e := range s.base.banks {
		if e != nil {
			mass = e.models[0].cells[i].mass / e.roots[0]
			break
		}
	}
	weighted := s.weights[0] * .5 * mass
	for j, id := range s.base.ids {
		if id != 0 {
			e := s.base.banks[j/4]
			weighted += s.weights[j+1] * e.models[j%4].cells[i].weighted / e.roots[j%4]
		}
	}
	return conditionalCell{mass: mass, weighted: weighted}
}

func (s snapshotRoutedLaw) Forecast(mask, values uint16) (float64, error) {
	if mask >= 512 || values&^mask != 0 {
		return 0, errors.New("invalid routed observation")
	}
	total := 0.
	for j, w := range s.weights {
		if math.IsNaN(w) || math.IsInf(w, 0) || w < 0 || j > 0 && s.base.ids[j-1] == 0 && w != 0 {
			return 0, errors.New("invalid routed weight")
		}
		total += w
	}
	if math.Abs(total-1) > 1e-12 {
		return 0, errors.New("invalid routed total")
	}
	for j, id := range s.base.ids {
		if id != 0 && (s.base.banks[j/4] == nil || !s.base.banks[j/4].ready) {
			return 0, errors.New("missing routed snapshot")
		}
	}
	c := s.cell(mask, values)
	if c.mass <= 0 {
		return 0, errors.New("empty routed input measure")
	}
	return math.Max(0, math.Min(1, c.weighted/c.mass)), nil
}
