package observationlearners

import (
	"fmt"
	"math"
	"reflect"

	"github.com/JuanHuaXu/eventframed/internal/transfergenerator"
)

// Teacher reconstruction is evaluation-only. No fitter accepts this object.
type populationQueryOracle struct {
	inputMap [512]uint16
	truth    func(uint16, int) (float64, error)
}

func populationOracle(input softV120Record) (populationQueryOracle, error) {
	var o populationQueryOracle
	if input.Phase < 0 || input.Phase > 1 || input.Case < 0 || input.Case >= 21 || input.Index < 0 || input.Index >= 32 {
		return o, fmt.Errorf("invalid identity")
	}
	if input.Case < 9 {
		if input.Teacher == nil {
			return o, fmt.Errorf("missing teacher")
		}
		spec := input.Teacher.Spec
		if spec.Phase != input.Phase || spec.Index != input.Index || int(spec.Family) != input.Case/3 || int(spec.Mode) != input.Case%3 {
			return o, fmt.Errorf("teacher identity mismatch")
		}
		_, teacher, err := transfergenerator.Generate(spec)
		if err != nil {
			return o, err
		}
		if !reflect.DeepEqual(teacher.Describe(), *input.Teacher) {
			return o, fmt.Errorf("teacher reconstruction mismatch")
		}
		o.truth = teacher.Truth
		for i := range o.inputMap {
			o.inputMap[i] = uint16(i)
		}
	} else {
		c := input.Case - 9
		expected, err := softV118Generate(input.Phase, input.Case, input.Index, softV120TransferBase, softV120BooleanBase)
		if err != nil || expected.Rules != input.Rules {
			return o, fmt.Errorf("Boolean rule identity mismatch")
		}
		for i := range o.inputMap {
			o.inputMap[i] = stackV93Input(uint16(i), replicationV102Cases[c])
		}
		o.truth = func(x uint16, t int) (float64, error) {
			name, rule := replicationV102Cases[c], input.Rules[0]
			if c >= 10 {
				name = "majority3"
				if c == 11 {
					name = "parity4"
				}
				if t >= 128 {
					rule = input.Rules[1]
					if c == 10 {
						name = "parity4"
					} else {
						name = "majority3"
					}
				}
			}
			return stackV93Truth(x, rule, name), nil
		}
	}
	return o, nil
}

type populationQueryKernel struct{ A, B, C [512]float64 }

func makePopulationKernel(o populationQueryOracle) (populationQueryKernel, error) {
	var k populationQueryKernel
	for i := 0; i < 31; i++ {
		a := math.Pow(.99, float64(i))
		b := .5 * (1 - a)
		for _, x := range o.inputMap {
			if x >= 512 {
				return k, fmt.Errorf("invalid input map")
			}
			q, err := o.truth(x, 161+i)
			if err != nil {
				return k, err
			}
			if math.IsNaN(q) || q < 0 || q > 1 {
				return k, fmt.Errorf("invalid probability")
			}
			w := 1. / (31 * 512)
			k.A[x] += w * a * a
			k.B[x] += w * 2 * a * (b - q)
			k.C[x] += w * (b*b + q*(1-2*b))
		}
	}
	return k, nil
}
func (k populationQueryKernel) risk(p [512]float64) (float64, error) {
	sum := 0.
	for x, v := range p {
		if math.IsNaN(v) || v < 0 || v > 1 {
			return 0, fmt.Errorf("invalid prediction")
		}
		sum += k.A[x]*v*v + k.B[x]*v + k.C[x]
	}
	if math.IsNaN(sum) || sum < -1e-12 || sum > 1+1e-12 {
		return 0, fmt.Errorf("invalid risk")
	}
	return sum, nil
}
