package observationlearners

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"slices"
)

// This builder deliberately excludes labels newly eligible after decision160.
// The explicit answer is the only unobserved label admitted to a paid branch.
func frozenPublicationHistory(input softV120Record, baseCap, selected int, answer bool) ([]segmentPacket, []int, error) {
	if len(input.Steps) != 256 || (baseCap != 63 && baseCap != 64) || selected < -1 || selected >= 160 || (selected >= 0 && selected < 152) {
		return nil, nil, fmt.Errorf("invalid frozen publication request")
	}
	if selected >= 0 {
		s := input.Steps[selected]
		if !s.Missing && selected+s.Delay <= 160 {
			return nil, nil, fmt.Errorf("query already known")
		}
	}
	var origins []int
	for j := -16; j < 160; j++ {
		if j < 0 || (!input.Steps[j].Missing && j+input.Steps[j].Delay <= 160) {
			origins = append(origins, j)
		}
	}
	origins = origins[max(0, len(origins)-baseCap):]
	if selected >= 0 {
		origins = append(origins, selected)
		slices.Sort(origins)
		origins = origins[max(0, len(origins)-64):]
	}
	h := make([]segmentPacket, 177)
	for i := range h {
		j := i - 16
		h[i].Arrives = -1
		if j < 0 {
			h[i].Bits = input.Initial[i].Bits
		} else {
			h[i].Bits = input.Steps[j].X
		}
	}
	for _, j := range origins {
		h[j+16].Arrives = 161
		if j < 0 {
			h[j+16].Outcome = input.Initial[j+16].Outcome
		} else if j == selected {
			h[j+16].Outcome = answer
		} else {
			h[j+16].Outcome = input.Steps[j].Y
		}
	}
	return h, origins, nil
}

func fitFrozenPublication(ctx context.Context, input softV120Record, baseCap, selected int, answer bool) (*segmentPosterior, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	h, _, err := frozenPublicationHistory(input, baseCap, selected, answer)
	if err != nil {
		return nil, err
	}
	return fitSegmentPosteriorBatch(ctx, -16, 161, h, 64, .01, .95)
}

type publicationRisk struct{ Population, Sample float64 }
type publicationQueryBranch struct {
	Origin  int
	Origins []int
	Q       float64
	ActualY bool
	Risk    [2]publicationRisk
}
type publicationQueryRecord struct {
	Phase, Case, Index, Schedule int
	Baseline                     [2]publicationRisk
	BaselineOrigins              [2][]int
	FrozenForecast               [512]float64
	RestoredForecast             [512]float64
	Branches                     []publicationQueryBranch
	Fits                         int
	Error                        string `json:",omitempty"`
}

func frozenPublicationRisk(input softV120Record, k populationQueryKernel, p [512]float64) (publicationRisk, error) {
	r := publicationRisk{}
	if len(input.Steps) != 256 {
		return r, fmt.Errorf("invalid evaluation tape")
	}
	var err error
	r.Population, err = k.risk(p)
	if err != nil {
		return r, err
	}
	for i := 0; i < 31; i++ {
		s := input.Steps[161+i]
		if s.X >= 512 || math.IsNaN(s.Q) || s.Q < 0 || s.Q > 1 {
			return r, fmt.Errorf("invalid evaluation truth")
		}
		forecast := .5 + math.Pow(.99, float64(i))*(p[s.X]-.5)
		r.Sample += ((forecast-s.Q)*(forecast-s.Q) + s.Q*(1-s.Q)) / 31
	}
	return r, nil
}

func runPublicationQuery(ctx context.Context, input softV120Record, coverage coverageQueryRecord) publicationQueryRecord {
	r := publicationQueryRecord{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule}
	if coverage.Error != "" || coverage.Phase != input.Phase || coverage.Case != input.Case || coverage.Index != input.Index || coverage.Schedule != input.Schedule {
		r.Error = "coverage identity"
		return r
	}
	_, pool, _, err := regimeQueryTapeView(input, 160)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	if !reflect.DeepEqual(pool, coverage.Pool) || len(pool) != len(coverage.Branches) {
		r.Error = "coverage pool"
		return r
	}
	_, r.BaselineOrigins[0], err = frozenPublicationHistory(input, 63, -1, false)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	if !reflect.DeepEqual(r.BaselineOrigins[0], coverage.Origins) {
		r.Error = "coverage support"
		return r
	}
	baseA := coverage.Base
	if len(pool) == 0 {
		m, e := fitFrozenPublication(ctx, input, 63, -1, false)
		if e != nil {
			r.Error = e.Error()
			return r
		}
		r.Fits++
		baseA = m.predictions
	}
	baseB, err := fitFrozenPublication(ctx, input, 64, -1, false)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.Fits++
	r.BaselineOrigins[1] = baseB.origins
	r.FrozenForecast = baseA
	r.RestoredForecast = baseB.predictions
	o, err := populationOracle(input)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	k, err := makePopulationKernel(o)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	for a, p := range [2][512]float64{baseA, baseB.predictions} {
		r.Baseline[a], err = frozenPublicationRisk(input, k, p)
		if err != nil {
			r.Error = err.Error()
			return r
		}
	}
	for i, j := range pool {
		b := coverage.Branches[i]
		if b.Origin != j {
			r.Error = "coverage branch order"
			return r
		}
		hA, origins, e := frozenPublicationHistory(input, 63, j, false)
		if e != nil {
			r.Error = e.Error()
			return r
		}
		hB, originsB, e := frozenPublicationHistory(input, 64, j, false)
		if e != nil {
			r.Error = e.Error()
			return r
		}
		if !reflect.DeepEqual(origins, originsB) || !reflect.DeepEqual(hA, hB) {
			r.Error = "paid support differs between A and B"
			return r
		}
		branch := publicationQueryBranch{Origin: j, Origins: origins, Q: input.Steps[j].Q, ActualY: input.Steps[j].Y}
		if math.IsNaN(branch.Q) || branch.Q < 0 || branch.Q > 1 {
			r.Error = "invalid query truth"
			return r
		}
		for y := 0; y < 2; y++ {
			branch.Risk[y], err = frozenPublicationRisk(input, k, b.Conditional[y])
			if err != nil {
				r.Error = err.Error()
				return r
			}
		}
		r.Branches = append(r.Branches, branch)
	}
	return r
}
