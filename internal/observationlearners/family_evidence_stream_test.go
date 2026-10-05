package observationlearners

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observation"
)

type familyEvidenceClock struct {
	Clock                                            int
	Origins                                          [2][]int
	GenericMass, GenericLog, BooleanLog, LogEvidence [2]float64
}
type familyEvidenceStream struct {
	Predictions [][2]float64
	Fits        []familyEvidenceClock
}

func familyEvidenceRun(input softV120Record, stride, stop int) (familyEvidenceStream, error) {
	var out familyEvidenceStream
	if len(input.Steps) != 256 || (stride != 8 && stride != 32) || stop < 0 || stop > 256 {
		return out, fmt.Errorf("invalid family stream")
	}
	var tables [2][512]float64
	for t := 0; t < stop; t++ {
		if t%stride == 0 {
			var available []int
			for j := -16; j < t; j++ {
				if j < 0 || (!input.Steps[j].Missing && j+input.Steps[j].Delay <= t) {
					available = append(available, j)
				}
			}
			fit := familyEvidenceClock{Clock: t}
			for w, cap := range []int{64, 32} {
				origins := available[max(0, len(available)-cap):]
				fit.Origins[w] = append([]int(nil), origins...)
				samples := make([]observation.Sample, len(origins))
				for i, j := range origins {
					if j < 0 {
						samples[i] = input.Initial[j+16]
					} else {
						samples[i] = observation.Sample{Bits: input.Steps[j].X, Outcome: input.Steps[j].Y}
					}
				}
				m, err := fitFamilyEvidence(samples, .95)
				if err != nil {
					return out, err
				}
				tables[w] = m.predictions
				fit.GenericMass[w], fit.GenericLog[w], fit.BooleanLog[w], fit.LogEvidence[w] = m.genericMass, m.genericLog, m.booleanLog, m.logEvidence
			}
			out.Fits = append(out.Fits, fit)
		}
		x := input.Steps[t].X
		if x >= 512 {
			return out, fmt.Errorf("invalid query bits")
		}
		out.Predictions = append(out.Predictions, [2]float64{tables[0][x], tables[1][x]})
	}
	return out, nil
}

func TestFamilyEvidenceStreamContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit consumed input required")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Version != "soft-learners-v120" {
		t.Fatal("wrong input")
	}
	fixtures, prefixes := 0, 0
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || (input.Case != 0 && input.Case != 20) {
			continue
		}
		before, _ := json.Marshal(input)
		for _, stride := range []int{32, 8} {
			result, err := familyEvidenceRun(input, stride, 256)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Fits) != 256/stride {
				t.Fatal("fit count")
			}
			for i, fit := range result.Fits {
				if fit.Clock != i*stride {
					t.Fatal("fit timing")
				}
				if stride == 32 && !reflect.DeepEqual(fit.Origins, input.Fits[i].Origins) {
					t.Fatal("original origins differ")
				}
				for w, origins := range fit.Origins {
					for _, j := range origins {
						if j >= 0 && (j >= fit.Clock || input.Steps[j].Missing || j+input.Steps[j].Delay > fit.Clock) {
							t.Fatal("unavailable evidence")
						}
					}
					logZ := segmentLogAdd(math.Log(.95)+fit.GenericLog[w], math.Log1p(-.95)+fit.BooleanLog[w])
					if math.Abs(logZ-fit.LogEvidence[w]) > 1e-12 || math.Abs(math.Exp(math.Log(.95)+fit.GenericLog[w]-logZ)-fit.GenericMass[w]) > 1e-12 {
						t.Fatal("evidence linkage")
					}
				}
			}
			for _, ps := range result.Predictions {
				for _, p := range ps {
					if math.IsNaN(p) || p <= 0 || p >= 1 {
						t.Fatal("probability")
					}
				}
			}
			if input.Case == 20 && input.Schedule == 1 {
				for _, clock := range []int{33, 161} {
					changed := input
					changed.Steps = append([]softV120Step(nil), input.Steps...)
					for j := range changed.Steps {
						s := &changed.Steps[j]
						s.Q = 1 - s.Q
						if j >= clock || s.Missing || j+s.Delay > clock {
							s.Y = !s.Y
						}
					}
					other, err := familyEvidenceRun(changed, stride, clock+1)
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(other.Predictions, result.Predictions[:clock+1]) || !reflect.DeepEqual(other.Fits, result.Fits[:len(other.Fits)]) {
						t.Fatal("future evidence leaked")
					}
					prefixes++
				}
			}
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("input mutation")
		}
		fixtures++
	}
	if fixtures != 4 || prefixes != 4 {
		t.Fatalf("coverage%d/%d", fixtures, prefixes)
	}
	t.Log("four stream fixtures, two cadences/windows, exact32-clock origins, four poisoned prefixes PASS")
}
