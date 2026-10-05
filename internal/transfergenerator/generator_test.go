package transfergenerator

import (
	"math"
	"reflect"
	"testing"
)

func reference(p Parameters, f Family, x uint16) float64 {
	bit := func(c uint8) int { return int(x>>c) & 1 }
	if f == Additive {
		z := p.Bias
		for j, a := range p.Coefficients {
			if bit(p.Coordinates[j]) == 0 {
				z -= a
			} else {
				z += a
			}
		}
		return .5 + .4*math.Tanh(z/2)
	}
	if f == Hierarchy {
		a := bit(p.Branches[0])
		b := bit(p.Branches[1+a])
		c := bit(p.Branches[3+2*a+b])
		return p.Probabilities[4*a+2*b+c]
	}
	key := 0
	for j := 0; j < 4; j++ {
		key += bit(p.Coordinates[j]) * (1 << j)
	}
	return p.Probabilities[key]
}

func TestIndependentTeachers(t *testing.T) {
	for phase := 0; phase < 2; phase++ {
		for family := Additive; family <= LocalTable; family++ {
			for mode := Stationary; mode <= Gradual; mode++ {
				for index := 0; index < 32; index++ {
					spec := Spec{SeedBase: ValidationSeedBase, Family: family, Mode: mode, Phase: phase, Index: index}
					data, teacher, err := Generate(spec)
					if err != nil {
						t.Fatal(err)
					}
					again, other, err := Generate(spec)
					if err != nil || data != again || teacher != other {
						t.Fatal("non-deterministic generator")
					}
					d := teacher.Describe()
					if d.Change < 97 || d.Change > 159 || d.Change%2 != 1 || d.Change%32 == 0 {
						t.Fatal("change clock")
					}
					for _, step := range []int{-16, d.Change - 1, d.Change, d.Change + 31, 255} {
						weight := 0.
						if mode == Abrupt && step >= d.Change {
							weight = 1
						}
						if mode == Gradual && step >= d.Change {
							weight = float64(step-d.Change+1) / 32
							if weight > 1 {
								weight = 1
							}
						}
						for x := uint16(0); x < 512; x++ {
							a, b := reference(d.Models[0], family, x), reference(d.Models[1], family, x)
							want := (1-weight)*a + weight*b
							got, err := teacher.Truth(x, step)
							if err != nil || math.Abs(want-got) > 1e-14 || got < .1 || got > .9 {
								t.Fatal("teacher formula", spec, x, step, got, want, err)
							}
							flipped, _ := teacher.Truth(x^448, step)
							if got != flipped {
								t.Fatal("irrelevant coordinates used")
							}
						}
					}
					for _, row := range data.Initial {
						if row.X >= 512 || row.Delay != 0 || row.Missing {
							t.Fatal("initial sample contract")
						}
					}
					for _, row := range data.Frames {
						if row.X >= 512 || row.Delay > 31 {
							t.Fatal("frame contract")
						}
					}
					d.Models[0].Bias = 999
					if teacher.Describe().Models[0].Bias == 999 {
						t.Fatal("description alias")
					}
				}
			}
		}
	}
}

func TestConditionalEnumeration(t *testing.T) {
	for family := Additive; family <= LocalTable; family++ {
		_, teacher, err := Generate(Spec{SeedBase: ValidationSeedBase, Family: family, Mode: Gradual, Phase: 1, Index: 0})
		if err != nil {
			t.Fatal(err)
		}
		d := teacher.Describe()
		step := d.Change + 7
		var full [512]float64
		for x := range full {
			full[x] = .75*reference(d.Models[0], family, uint16(x)) + .25*reference(d.Models[1], family, uint16(x))
		}
		for mask := uint16(0); mask < 512; mask++ {
			for value := mask; ; value = (value - 1) & mask {
				sum, n := 0., 0.
				for x, q := range full {
					if uint16(x)&mask == value {
						sum += q
						n++
					}
				}
				got, err := teacher.Conditional(mask, value, step)
				if err != nil || math.Abs(got-sum/n) > 1e-13 {
					t.Fatal("conditional average")
				}
				if value == 0 {
					break
				}
			}
		}
	}
}

func TestLearnerBoundaryAndValidation(t *testing.T) {
	typ := reflect.TypeOf(Observation{})
	want := []string{"X", "Y", "Delay", "Missing"}
	if typ.NumField() != len(want) {
		t.Fatal("learner packet has extra metadata")
	}
	for i, k := range want {
		if typ.Field(i).Name != k {
			t.Fatal("teacher metadata in learner packet")
		}
	}
	for _, spec := range []Spec{{}, {SeedBase: 1, Family: 3}, {SeedBase: 1, Mode: 3}, {SeedBase: 1, Phase: 2}, {SeedBase: 1, Index: 32}, {SeedBase: 1, Index: -1}, {SeedBase: 1<<63 - 1}} {
		if _, _, err := Generate(spec); err == nil {
			t.Fatal("invalid spec accepted", spec)
		}
	}
	_, teacher, _ := Generate(Spec{SeedBase: ValidationSeedBase})
	for _, query := range [][2]int{{512, 0}, {0, -17}, {0, 256}} {
		if _, err := teacher.Truth(uint16(query[0]), query[1]); err == nil {
			t.Fatal("invalid truth query")
		}
	}
	if _, err := teacher.Conditional(0, 1, 0); err == nil {
		t.Fatal("hidden value query")
	}
	if _, err := (Teacher{}).Truth(0, 0); err == nil {
		t.Fatal("uninitialized teacher")
	}
}

func TestValidationSeedSeparation(t *testing.T) {
	used := map[int64]bool{}
	for phase := 0; phase < 2; phase++ {
		for f := 0; f < 3; f++ {
			for m := 0; m < 3; m++ {
				for i := 0; i < 32; i++ {
					for role := 0; role < 5; role++ {
						seed := (ValidationSeedBase + int64(phase*1000000+f*100000+m*10000+i*10+role)) % 2147483647
						if used[seed] {
							t.Fatal("within-cohort seed collision")
						}
						used[seed] = true
					}
				}
			}
		}
	}
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300, 2042119400, 2046119500, 2050119600, 2054119700, 2062119900, 2066110000, 2078110100, 2082110200, 2090110300, 2100110400, 2110110600, 2120110800, 2130110900, 2140111100, 2144111200, 2150111300, 2160111500} {
		for phase := 0; phase < 2; phase++ {
			for c := 0; c < 12; c++ {
				for i := 0; i < 128; i++ {
					for role := 0; role < 5; role++ {
						if used[(base+int64(phase*1000000+c*10000+i*10+role))%2147483647] {
							t.Fatal("archived learner seed", base)
						}
					}
				}
			}
		}
	}
	for _, base := range []int64{3070119900, 3090110000, 3110110100, 3130110200} {
		for mode := int64(0); mode < 2; mode++ {
			for family := int64(0); family < 4096; family++ {
				for test := int64(0); test < 64; test++ {
					if used[(base+mode*10000000+family*1000+test)%2147483647] {
						t.Fatal("archived null seed", base)
					}
				}
			}
		}
	}
}
