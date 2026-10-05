package observationlearners

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

// Research-only dual ridge. No oracle metadata is accepted by this boundary.
var spectralOnce sync.Once
var spectralMasks []uint16
var spectralPhi [512][256]float64
var spectralGram [2][512][512]float64

func spectralInit() {
	spectralOnce.Do(func() {
		spectralMasks = append(spectralMasks, 0)
		for i := 0; i < 9; i++ {
			spectralMasks = append(spectralMasks, 1<<i)
		}
		for m := uint16(1); m < 512; m++ {
			d := bits.OnesCount16(m)
			if d >= 2 && d <= 4 {
				spectralMasks = append(spectralMasks, m)
			}
		}
		for x := 0; x < 512; x++ {
			for j, m := range spectralMasks {
				spectralPhi[x][j] = 1
				if (bits.OnesCount16(m)-bits.OnesCount16(uint16(x)&m))%2 != 0 {
					spectralPhi[x][j] = -1
				}
			}
		}
		for x := 0; x < 512; x++ {
			for y := 0; y <= x; y++ {
				a, b := 0., 0.
				for j := range spectralMasks {
					v := spectralPhi[x][j] * spectralPhi[y][j]
					b += v
					if j < 10 {
						a += v
					}
				}
				spectralGram[0][x][y], spectralGram[0][y][x] = a, a
				spectralGram[1][x][y], spectralGram[1][y][x] = b, b
			}
		}
	})
}

type spectralModel struct {
	Coeff    [256]float64
	Features []int
	Residual float64
}

func fitSpectral(samples []observation.Sample, mode int) (*spectralModel, error) {
	if len(samples) < 1 || len(samples) > 64 || mode < 0 || mode > 2 {
		return nil, fmt.Errorf("spectral bounds")
	}
	for _, s := range samples {
		if s.Bits >= 512 {
			return nil, fmt.Errorf("spectral input")
		}
	}
	spectralInit()
	m := new(spectralModel)
	for j := 0; j < 10; j++ {
		m.Features = append(m.Features, j)
	}
	n := len(samples)
	if mode == 2 {
		for j := 10; j < 256; j++ {
			m.Features = append(m.Features, j)
		}
	}
	if mode == 1 {
		type entry struct {
			j         int
			magnitude float64
		}
		var selected []entry
		threshold := math.Sqrt(2 * math.Log(2*246/.1) / float64(n))
		for j := 10; j < 256; j++ {
			sum := 0.
			for _, s := range samples {
				y := -1.
				if s.Outcome {
					y = 1
				}
				sum += y * spectralPhi[s.Bits][j]
			}
			a := math.Abs(sum / float64(n))
			if a >= threshold {
				selected = append(selected, entry{j, a})
			}
		}
		// At most246 elements; deterministic insertion sort preserves mask ties.
		for i := 1; i < len(selected); i++ {
			for k := i; k > 0 && selected[k].magnitude > selected[k-1].magnitude; k-- {
				selected[k], selected[k-1] = selected[k-1], selected[k]
			}
		}
		if len(selected) > 16 {
			selected = selected[:16]
		}
		for _, e := range selected {
			m.Features = append(m.Features, e.j)
		}
	}
	var gram, l [64][64]float64
	var target, z, alpha [64]float64
	for i, s := range samples {
		target[i] = -1
		if s.Outcome {
			target[i] = 1
		}
		for j := 0; j <= i; j++ {
			g := 0.
			if mode == 1 {
				for _, f := range m.Features {
					g += spectralPhi[s.Bits][f] * spectralPhi[samples[j].Bits][f]
				}
			} else {
				g = spectralGram[mode/2][s.Bits][samples[j].Bits]
			}
			if i == j {
				g++
			}
			gram[i][j], gram[j][i] = g, g
			v := g
			for k := 0; k < j; k++ {
				v -= l[i][k] * l[j][k]
			}
			if i == j {
				if v <= 0 || math.IsNaN(v) || math.IsInf(v, 0) {
					return nil, fmt.Errorf("spectral Cholesky")
				}
				l[i][j] = math.Sqrt(v)
			} else {
				l[i][j] = v / l[j][j]
			}
		}
		v := target[i]
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
	for i, s := range samples {
		v := -target[i]
		for j := 0; j < n; j++ {
			v += gram[i][j] * alpha[j]
		}
		m.Residual = math.Max(m.Residual, math.Abs(v))
		for _, f := range m.Features {
			m.Coeff[f] += alpha[i] * spectralPhi[s.Bits][f]
		}
	}
	if math.IsNaN(m.Residual) || m.Residual > 1e-8 {
		return nil, fmt.Errorf("spectral solve residual %g", m.Residual)
	}
	return m, nil
}
func (m *spectralModel) predict(x uint16) (float64, error) {
	if m == nil || len(m.Features) == 0 || x >= 512 {
		return 0, fmt.Errorf("spectral prediction")
	}
	v := 0.
	for _, j := range m.Features {
		v += m.Coeff[j] * spectralPhi[x][j]
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("spectral nonfinite")
	}
	return math.Max(1e-12, math.Min(1-1e-12, (1+v)/2)), nil
}

func TestSpectralContracts(t *testing.T) {
	spectralInit()
	if len(spectralMasks) != 256 {
		t.Fatal(len(spectralMasks))
	}
	// Exhaust the six active bits. Every monomial on this support is orthogonal;
	// nuisance bits are fixed, so test the known ridge prediction via its Gram.
	for _, mask := range []uint16{1, 3, 7, 15} {
		samples := make([]observation.Sample, 64)
		for x := range samples {
			samples[x] = observation.Sample{Bits: uint16(x), Outcome: bits.OnesCount16(uint16(x)&mask)%2 == 0}
		}
		before := append([]observation.Sample(nil), samples...)
		for mode := 0; mode < 3; mode++ {
			m, e := fitSpectral(samples, mode)
			if e != nil {
				t.Fatal(e)
			}
			for x, s := range samples {
				p, e := m.predict(uint16(x))
				if e != nil {
					t.Fatal(e)
				}
				if mode == 0 && mask == 1 {
					y := -1.
					if s.Outcome {
						y = 1
					}
					if math.Abs(p-(1+y*64./65)/2) > 1e-12 {
						t.Fatalf("analytic main effect: %g", p)
					}
				}
				if mode == 0 && mask != 1 {
					if math.Abs(p-.5) > 1e-10 {
						t.Fatalf("linear parity: %g", p)
					}
				} else if (p > .5) != s.Outcome {
					t.Fatalf("mode%d mask%d", mode, mask)
				}
			}
			if !reflect.DeepEqual(samples, before) {
				t.Fatal("mutated input")
			}
		}
	}
	for _, truth := range []bool{false, true} {
		s := make([]observation.Sample, 64)
		for i := range s {
			s[i] = observation.Sample{Bits: 511, Outcome: truth}
		}
		for mode := 0; mode < 3; mode++ {
			m, e := fitSpectral(s, mode)
			if e != nil {
				t.Fatal(e)
			}
			p, e := m.predict(511)
			if e != nil || (p > .5) != truth {
				t.Fatal(p, e)
			}
		}
	}
	for _, s := range [][]observation.Sample{nil, make([]observation.Sample, 65), {{Bits: 512}}} {
		if _, e := fitSpectral(s, 0); e == nil {
			t.Fatal("accepted bad input")
		}
	}
}

type spectralRecord struct {
	Phase, Case, Index, Schedule int
	P                            [256][6]float64
	MaxResidual                  float64
	FeatureTotal                 [6]int
	Fits                         int
}

func runSpectral(r softV120Record) (spectralRecord, error) {
	out := spectralRecord{Phase: r.Phase, Case: r.Case, Index: r.Index, Schedule: r.Schedule}
	if len(r.Steps) != 256 || len(r.Fits) != 8 {
		return out, fmt.Errorf("source shape")
	}
	var models [6]*spectralModel
	for clock, s := range r.Steps {
		if clock%32 == 0 {
			var available []int
			for i := -16; i < clock; i++ {
				if i < 0 || (!r.Steps[i].Missing && i+r.Steps[i].Delay <= clock) {
					available = append(available, i)
				}
			}
			for w, cap := range []int{64, 32} {
				origins := available
				if len(origins) > cap {
					origins = origins[len(origins)-cap:]
				}
				if !reflect.DeepEqual(origins, r.Fits[clock/32].Origins[w]) {
					return out, fmt.Errorf("as-of mismatch")
				}
				samples := make([]observation.Sample, len(origins))
				for j, i := range origins {
					if i < 0 {
						samples[j] = r.Initial[i+16]
					} else {
						samples[j] = observation.Sample{Bits: r.Steps[i].X, Outcome: r.Steps[i].Y}
					}
				}
				for mode := 0; mode < 3; mode++ {
					m, e := fitSpectral(samples, mode)
					if e != nil {
						return out, e
					}
					k := 2*mode + w
					models[k] = m
					out.MaxResidual = math.Max(out.MaxResidual, m.Residual)
					out.FeatureTotal[k] += len(m.Features)
					out.Fits++
				}
			}
		}
		for k, m := range models {
			p, e := m.predict(s.X)
			if e != nil {
				return out, e
			}
			out.P[clock][k] = p
		}
	}
	return out, nil
}

func TestSpectralRun(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SPECTRAL_SOURCE")
	if path == "" {
		t.Skip("explicit research source required")
	}
	dst := os.Getenv("EVENTFRAME_SPECTRAL_OUTPUT")
	if dst == "" {
		t.Fatal("output required")
	}
	in, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer in.Close()
	out, e := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	dec, enc := json.NewDecoder(in), json.NewEncoder(out)
	var header softV120Artifact
	if e = dec.Decode(&header); e != nil || header.Version != "soft-learners-v120" {
		t.Fatalf("source header: %v", e)
	}
	count := 0
	for {
		var r softV120Record
		e = dec.Decode(&r)
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		v, e := runSpectral(r)
		if e != nil {
			t.Fatalf("record%d: %v", count, e)
		}
		if e = enc.Encode(v); e != nil {
			t.Fatal(e)
		}
		count++
	}
	if count != 2688 {
		t.Fatalf("records%d", count)
	}
	if e = out.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Logf("records=%d fits=%d", count, count*48)
}

func BenchmarkSpectralFit(b *testing.B) {
	spectralInit()
	s := make([]observation.Sample, 64)
	for i := range s {
		s[i] = observation.Sample{Bits: uint16(i * 7 % 512), Outcome: i%3 == 0}
	}
	for mode := 0; mode < 3; mode++ {
		b.Run(fmt.Sprint(mode), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, e := fitSpectral(s, mode); e != nil {
					b.Fatal(e)
				}
			}
		})
	}
}
