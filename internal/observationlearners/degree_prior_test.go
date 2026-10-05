package observationlearners

import (
	"fmt"
	"math"
	"math/bits"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

func degreeVariance(mask uint16, mode int) float64 {
	if mask == 0 || mode == 2 {
		return 1
	}
	if mode == 1 {
		return 36. / 255
	}
	return 9. / []float64{1, 9, 36, 84, 126}[bits.OnesCount16(mask)]
}

// Dynamic elementary-symmetric products avoid building a512x512 Gram table.
// Signs depend only on Hamming distance; the kernel still uses all coordinates.
func degreeKernel(hamming, mode int) float64 {
	e := [5]float64{1}
	for i := 0; i < 9; i++ {
		z := 1.
		if i < hamming {
			z = -1
		}
		for d := 4; d >= 1; d-- {
			e[d] += z * e[d-1]
		}
	}
	value := 1.
	for d := 1; d <= 4; d++ {
		value += degreeVariance(uint16((1<<d)-1), mode) * e[d]
	}
	return value
}

// mode2 is a unit-prior audit control, never a newly tested candidate.
func fitDegree(samples []observation.Sample, mode int) (*spectralModel, error) {
	if len(samples) < 1 || len(samples) > 64 || mode < 0 || mode > 2 {
		return nil, fmt.Errorf("degree bounds")
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return nil, fmt.Errorf("degree input")
		}
	}
	spectralInit()
	var kernel [10]float64
	for h := range kernel {
		kernel[h] = degreeKernel(h, mode)
	}
	n := len(samples)
	var a, l [64][64]float64
	var y, z, alpha [64]float64
	for i, s := range samples {
		y[i] = -1
		if s.Outcome {
			y[i] = 1
		}
		for j := 0; j <= i; j++ {
			v := kernel[bits.OnesCount16(s.Bits^samples[j].Bits)]
			if i == j {
				v++
			}
			a[i][j], a[j][i] = v, v
			for k := 0; k < j; k++ {
				v -= l[i][k] * l[j][k]
			}
			if i == j {
				if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
					return nil, fmt.Errorf("degree Cholesky")
				}
				l[i][j] = math.Sqrt(v)
			} else {
				l[i][j] = v / l[j][j]
			}
		}
		v := y[i]
		for j := 0; j < i; j++ {
			v -= l[i][j] * z[j]
		}
		z[i] = v / l[i][i]
	}
	for i := n - 1; i >= 0; i-- {
		v := z[i]
		for j := i + 1; j < n; j++ {
			v -= l[j][i] * alpha[j]
		}
		alpha[i] = v / l[i][i]
	}
	m := new(spectralModel)
	for j := range spectralMasks {
		m.Features = append(m.Features, j)
	}
	for i, s := range samples {
		v := -y[i]
		for j := 0; j < n; j++ {
			v += a[i][j] * alpha[j]
		}
		m.Residual = math.Max(m.Residual, math.Abs(v))
		for j, mask := range spectralMasks {
			m.Coeff[j] += degreeVariance(mask, mode) * alpha[i] * spectralPhi[s.Bits][j]
		}
	}
	if math.IsNaN(m.Residual) || m.Residual > 1e-8 {
		return nil, fmt.Errorf("degree solve residual %g", m.Residual)
	}
	return m, nil
}

func fitDegreeExperiment(samples []observation.Sample, mode int) (*spectralModel, error) {
	if mode == 0 {
		return fitSpectral(samples, 0)
	}
	return fitDegree(samples, mode-1)
}

func TestDegreeContracts(t *testing.T) {
	spectralInit()
	for mode := 0; mode < 3; mode++ {
		want := 37.
		if mode == 2 {
			want = 256
		}
		if math.Abs(degreeKernel(0, mode)-want) > 1e-12 {
			t.Fatal("diagonal")
		}
		for h := 0; h <= 9; h++ {
			y := uint16((1 << h) - 1)
			sum := 0.
			for j, m := range spectralMasks {
				sum += degreeVariance(m, mode) * spectralPhi[0][j] * spectralPhi[y][j]
			}
			if math.Abs(sum-degreeKernel(h, mode)) > 1e-11 {
				t.Fatal("feature kernel", h, mode, sum)
			}
		}
		var weights [256]float64
		for j, m := range spectralMasks {
			weights[j] = degreeVariance(m, mode)
		}
		for x := 0; x < 512; x++ {
			for y := 0; y < 512; y++ {
				sum := 0.
				for j, w := range weights {
					sum += w * spectralPhi[x][j] * spectralPhi[y][j]
				}
				h := bits.OnesCount16(uint16(x ^ y))
				if math.Abs(sum-degreeKernel(h, mode)) > 1e-11 {
					t.Fatal("pair kernel", x, y, mode)
				}
			}
		}
	}
	for _, truth := range []bool{false, true} {
		s := make([]observation.Sample, 64)
		for i := range s {
			s[i] = observation.Sample{Bits: 511, Outcome: truth}
		}
		for mode := 0; mode < 3; mode++ {
			m, e := fitDegree(s, mode)
			if e != nil {
				t.Fatal(e)
			}
			p, e := m.predict(511)
			if e != nil || (p > .5) != truth {
				t.Fatal(p, e)
			}
		}
	}
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	a, e := fitSpectral(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	b, e := fitDegree(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	for x := 0; x < 512; x++ {
		p, e := a.predict(uint16(x))
		if e != nil {
			t.Fatal(e)
		}
		q, e := b.predict(uint16(x))
		if e != nil || math.Abs(p-q) > 1e-11 {
			t.Fatal("unit prior", p, q, e)
		}
	}
	for _, bad := range [][]observation.Sample{nil, make([]observation.Sample, 65), {{Bits: 512}}} {
		if _, e := fitDegree(bad, 0); e == nil {
			t.Fatal("invalid accepted")
		}
	}
}

func BenchmarkDegreeFit(b *testing.B) {
	spectralInit()
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	for mode := 0; mode < 2; mode++ {
		b.Run([]string{"degree", "scalar"}[mode], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, e := fitDegree(s, mode); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}

func BenchmarkDegreePredict(b *testing.B) {
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 19 % 512), Outcome: i%3 == 0}
	}
	for mode := 0; mode < 2; mode++ {
		m, e := fitDegree(s, mode)
		if e != nil {
			b.Fatal(e)
		}
		b.Run([]string{"degree", "scalar"}[mode], func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, e := m.predict(uint16(i % 512)); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
