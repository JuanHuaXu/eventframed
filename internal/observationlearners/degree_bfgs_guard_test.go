package observationlearners

import (
	"fmt"
	"math"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type degreeGuardStats struct {
	Old          learnedDegreeStats
	BFGS         degreeBFGSStats
	SelectedBFGS bool
	Evaluations  int
}

// This is an optimization safeguard, not a gate on future predictive quality.
// Both objectives must refer to fits on exactly the same eligible samples.
func degreePreferBFGS(old, next float64) (bool, error) {
	if math.IsNaN(old) || math.IsInf(old, 0) || math.IsNaN(next) || math.IsInf(next, 0) {
		return false, fmt.Errorf("finite incumbent objectives required")
	}
	return next <= old, nil
}

func fitGuardedDegreeBFGS(s []observation.Sample, noise bool) (*spectralModel, degreeGuardStats, error) {
	var st degreeGuardStats
	old, a, e := fitLearnedDegree(s, noise)
	st.Old = a
	if e != nil {
		return nil, st, e
	}
	next, b, e := fitDegreeBFGS(s, noise)
	st.BFGS = b
	st.Evaluations = a.Evaluations + b.Evaluations
	if e != nil {
		return nil, st, e
	}
	st.SelectedBFGS, e = degreePreferBFGS(a.Final, b.Final)
	if e != nil {
		return nil, st, e
	}
	if st.SelectedBFGS {
		return next, st, nil
	}
	return old, st, nil
}

func TestDegreeBFGSGuard(t *testing.T) {
	for _, c := range []struct {
		a, b float64
		want bool
	}{{10, 9, true}, {9, 10, false}, {9, 9, true}, {-10, -11, true}} {
		got, e := degreePreferBFGS(c.a, c.b)
		if e != nil || got != c.want {
			t.Fatal(c, got, e)
		}
	}
	if _, e := degreePreferBFGS(0, math.NaN()); e == nil {
		t.Fatal("nonfinite accepted")
	}
	s := make([]observation.Sample, 32)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	for _, noise := range []bool{false, true} {
		m, st, e := fitGuardedDegreeBFGS(s, noise)
		if e != nil {
			t.Fatal(e)
		}
		if st.Evaluations != st.Old.Evaluations+st.BFGS.Evaluations || st.Evaluations > 1490 {
			t.Fatal("unaccounted work")
		}
		var expected *spectralModel
		if st.SelectedBFGS {
			expected, _, e = fitDegreeBFGS(s, noise)
			if st.BFGS.Final > st.Old.Final {
				t.Fatal("objective regression")
			}
		} else {
			expected, _, e = fitLearnedDegree(s, noise)
		}
		if e != nil {
			t.Fatal(e)
		}
		for x := 0; x < 512; x++ {
			p, e := m.predict(uint16(x))
			if e != nil {
				t.Fatal(e)
			}
			q, e := expected.predict(uint16(x))
			if e != nil || p != q {
				t.Fatal("selected model mismatch", x, p, q, e)
			}
		}
	}
}

func BenchmarkDegreeBFGSGuard(b *testing.B) {
	spectralInit()
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	for _, noise := range []bool{false, true} {
		b.Run(fmt.Sprint(noise), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, _, e := fitGuardedDegreeBFGS(s, noise); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
