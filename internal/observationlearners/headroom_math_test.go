package observationlearners

import (
	"errors"
	"math"
	"testing"
)

type headroomPoint struct {
	Q       float64
	Experts [4]float64
	Bank    float64
}
type headroomMetric struct{ Generic, Short, Bank, Hull, Pair, Alpha float64 }

// This function consumes simulator truth solely for retrospective evaluation.
// Its output is deliberately not a forecast constructor or admission API.
func scoreHeadroom(points []headroomPoint) (headroomMetric, error) {
	var r headroomMetric
	if len(points) == 0 {
		return r, errors.New("empty headroom block")
	}
	numerator, denominator := 0., 0.
	for _, p := range points {
		for _, v := range []float64{p.Q, p.Bank, p.Experts[0], p.Experts[1], p.Experts[2], p.Experts[3]} {
			if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return r, errors.New("invalid headroom probability")
			}
		}
		d := p.Experts[2] - p.Experts[0]
		numerator += d * (p.Q - p.Experts[0])
		denominator += d * d
	}
	if denominator > 0 {
		r.Alpha = math.Max(0, math.Min(1, numerator/denominator))
	}
	for _, p := range points {
		lo, hi := p.Experts[0], p.Experts[0]
		for _, x := range p.Experts {
			lo = math.Min(lo, x)
			hi = math.Max(hi, x)
		}
		loss := func(x float64) float64 { return ((x-p.Q)*(x-p.Q) + p.Q*(1-p.Q)) / float64(len(points)) }
		r.Generic += loss(p.Experts[0])
		r.Short += loss(p.Experts[2])
		r.Bank += loss(p.Bank)
		r.Hull += loss(math.Max(lo, math.Min(hi, p.Q)))
		r.Pair += loss(p.Experts[0] + r.Alpha*(p.Experts[2]-p.Experts[0]))
	}
	if r.Hull > r.Pair+1e-12 || r.Pair > r.Generic+1e-12 || r.Pair > r.Short+1e-12 {
		return r, errors.New("headroom ordering")
	}
	return r, nil
}

func TestHeadroomMath(t *testing.T) {
	for _, c := range []struct {
		q    float64
		want float64
	}{{.5, .5}, {.1, 0}, {.9, 1}} {
		r, e := scoreHeadroom([]headroomPoint{{Q: c.q, Experts: [4]float64{.2, .2, .8, .8}, Bank: .4}})
		if e != nil || math.Abs(r.Alpha-c.want) > 1e-12 {
			t.Fatal("alpha", r, e)
		}
	}
	r, e := scoreHeadroom([]headroomPoint{{Q: .8, Experts: [4]float64{.3, .3, .3, .3}, Bank: .3}})
	if e != nil || r.Alpha != 0 || r.Pair != r.Generic || r.Hull != r.Short {
		t.Fatal("tie", r, e)
	}
	if _, e = scoreHeadroom(nil); e == nil {
		t.Fatal("empty")
	}
	for _, q := range []float64{math.NaN(), math.Inf(1), -1, 2} {
		if _, e = scoreHeadroom([]headroomPoint{{Q: q}}); e == nil {
			t.Fatal("invalid")
		}
	}
	// One common weight cannot use a different oracle optimum on every event.
	r, e = scoreHeadroom([]headroomPoint{{Q: .1, Experts: [4]float64{.2, .2, .8, .8}}, {Q: .9, Experts: [4]float64{.2, .2, .8, .8}}})
	if e != nil || r.Hull >= r.Pair || math.Abs(r.Alpha-.5) > 1e-12 {
		t.Fatal("block distinction", r, e)
	}
}
