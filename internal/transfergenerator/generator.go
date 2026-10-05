// Package transfergenerator provides independent synthetic research teachers.
// It imports no forecasting implementation or archived Boolean truth helper.
package transfergenerator

import (
	"errors"
	"math"
	"math/rand"
)

const ValidationSeedBase int64 = 2176111600

type Family int

const (
	Additive Family = iota
	Hierarchy
	LocalTable
)

type Mode int

const (
	Stationary Mode = iota
	Abrupt
	Gradual
)

type Spec struct {
	SeedBase     int64
	Family       Family
	Mode         Mode
	Phase, Index int
}

// Observation is the only packet intended for future learner adapters. Teacher
// identity, relevant coordinates, target probability and change time are absent.
type Observation struct {
	X       uint16
	Y       bool
	Delay   uint8
	Missing bool
}
type Evidence struct {
	Initial [16]Observation
	Frames  [256]Observation
}

type Parameters struct {
	Coordinates   [6]uint8
	Coefficients  [5]float64
	Bias          float64
	Branches      [7]uint8
	Probabilities [16]float64
}

// Description is explicitly oracle-side. Returned arrays are detached values.
type Description struct {
	Spec   Spec
	Seeds  [5]int64
	Change int
	Models [2]Parameters
}

type Teacher struct {
	description Description
	ready       bool
}

func makeParameters(r *rand.Rand, f Family) Parameters {
	var p Parameters
	permutation := r.Perm(6)
	for i, v := range permutation {
		p.Coordinates[i] = uint8(v)
	}
	switch f {
	case Additive:
		p.Bias = -.4 + .8*float64(r.Intn(2))
		for i := range p.Coefficients {
			p.Coefficients[i] = (.4 + .3*float64(i)) * (2*float64(r.Intn(2)) - 1)
		}
	case Hierarchy:
		p.Branches = [7]uint8{p.Coordinates[0], p.Coordinates[1], p.Coordinates[2], p.Coordinates[3], p.Coordinates[4], p.Coordinates[4], p.Coordinates[5]}
		for i := 0; i < 8; i++ {
			p.Probabilities[i] = .1 + .8*r.Float64()
		}
	case LocalTable:
		for i := range p.Probabilities {
			p.Probabilities[i] = .1 + .8*r.Float64()
		}
	}
	return p
}

func probability(p Parameters, f Family, x uint16) float64 {
	switch f {
	case Additive:
		z := p.Bias
		for j, a := range p.Coefficients {
			z += a * (2*float64((x>>p.Coordinates[j])&1) - 1)
		}
		return .1 + .8/(1+math.Exp(-z))
	case Hierarchy:
		node := 0
		for node < 7 {
			node = 2*node + 1 + int((x>>p.Branches[node])&1)
		}
		return p.Probabilities[node-7]
	default:
		key := 0
		for j := 3; j >= 0; j-- {
			key = 2*key + int((x>>p.Coordinates[j])&1)
		}
		return p.Probabilities[key]
	}
}

func (t Teacher) Describe() Description { return t.description }

func (t Teacher) Truth(x uint16, step int) (float64, error) {
	if !t.ready || x >= 512 || step < -16 || step >= 256 {
		return 0, errors.New("invalid teacher query")
	}
	d := t.description
	w := 0.
	if d.Spec.Mode != Stationary && step >= d.Change {
		w = 1
		if d.Spec.Mode == Gradual {
			w = math.Min(1, float64(step-d.Change+1)/32)
		}
	}
	a := probability(d.Models[0], d.Spec.Family, x)
	b := probability(d.Models[1], d.Spec.Family, x)
	return (1-w)*a + w*b, nil
}

func (t Teacher) Conditional(mask, values uint16, step int) (float64, error) {
	if mask >= 512 || values&^mask != 0 {
		return 0, errors.New("invalid conditional query")
	}
	sum, n := 0., 0.
	for x := uint16(0); x < 512; x++ {
		if x&mask == values {
			q, err := t.Truth(x, step)
			if err != nil {
				return 0, err
			}
			sum += q
			n++
		}
	}
	return sum / n, nil
}

func Generate(spec Spec) (Evidence, Teacher, error) {
	var data Evidence
	if spec.SeedBase <= 0 || spec.SeedBase > 1<<62 || spec.Family < Additive || spec.Family > LocalTable || spec.Mode < Stationary || spec.Mode > Gradual || spec.Phase < 0 || spec.Phase > 1 || spec.Index < 0 || spec.Index >= 32 {
		return data, Teacher{}, errors.New("invalid transfer generator spec")
	}
	seed := spec.SeedBase + int64(spec.Phase*1000000+int(spec.Family)*100000+int(spec.Mode)*10000+spec.Index*10)
	d := Description{Spec: spec}
	for i := range d.Seeds {
		d.Seeds[i] = seed + int64(i)
	}
	r := rand.New(rand.NewSource(d.Seeds[0]))
	for i := range d.Models {
		d.Models[i] = makeParameters(r, spec.Family)
	}
	d.Change = 97 + 2*r.Intn(32)
	teacher := Teacher{description: d, ready: true}
	xs, ys := rand.New(rand.NewSource(d.Seeds[1])), rand.New(rand.NewSource(d.Seeds[2]))
	delays, misses := rand.New(rand.NewSource(d.Seeds[3])), rand.New(rand.NewSource(d.Seeds[4]))
	for step := -16; step < 256; step++ {
		x := uint16(xs.Intn(512))
		q, _ := teacher.Truth(x, step)
		row := Observation{X: x, Y: ys.Float64() < q}
		if step < 0 {
			data.Initial[step+16] = row
		} else {
			row.Delay = uint8(delays.Intn(32))
			row.Missing = misses.Float64() < .2
			data.Frames[step] = row
		}
	}
	return data, teacher, nil
}
