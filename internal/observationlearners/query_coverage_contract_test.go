package observationlearners

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"os"
	"reflect"
	"testing"
)

func TestCoverageQueryWeights(t *testing.T) {
	input := softV120Record{Steps: make([]softV120Step, 256)}
	for i := range input.Steps {
		input.Steps[i].X = 7
	}
	input.Steps[160].X = 9
	w, err := coverageQueryWeights(input)
	if err != nil || w[0][7] != 160 || w[0][9] != 1 || w[1][7] != 31 || w[1][9] != 1 {
		t.Fatal("weights", w, err)
	}
	input.Steps[161].X = 999
	if again, e := coverageQueryWeights(input); e != nil || again != w {
		t.Fatal("future input leak")
	}
	input.Steps[160].X = 512
	if _, err := coverageQueryWeights(input); err == nil {
		t.Fatal("invalid input accepted")
	}
	if _, err := coverageQueryWeights(softV120Record{}); err == nil {
		t.Fatal("invalid length")
	}
	var base [512]float64
	b := coverageQueryBranch{Mass: [2]float64{.5, .5}}
	base[7] = .5
	base[9] = .5
	b.Conditional[0][7] = .1
	b.Conditional[1][7] = .9
	b.Conditional[0][9] = .3
	b.Conditional[1][9] = .7
	gain, err := coverageQueryGain(base, b, w[0])
	if err != nil || math.Abs(gain-(160*.16+.04)/161) > 1e-14 {
		t.Fatal("duplicate weighting", gain, err)
	}
	if _, err := coverageQueryGain(base, b, [512]uint16{}); err == nil {
		t.Fatal("empty weights")
	}
	b.Mass = [2]float64{}
	if _, err := coverageQueryGain(base, b, w[0]); err == nil {
		t.Fatal("invalid mass")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if r := runCoverageQuery(ctx, input); r.Error == "" {
		t.Fatal("cancellation ignored")
	}
	if _, err := coverageQueryValue(context.Background(), nil, 0); err == nil {
		t.Fatal("nil state")
	}
}

func TestCoverageQueryContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit consumed source required")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header softV120Artifact
	if err := d.Decode(&header); err != nil || header.Version != "soft-learners-v120" {
		t.Fatal("source header", err)
	}
	checked, conditionalChecks := 0, 0
	maxError := 0.
	poisoned := false
	close := func(a, b float64) {
		t.Helper()
		e := math.Abs(a - b)
		maxError = math.Max(maxError, e)
		if e > 1e-10 {
			t.Fatalf("difference %.17g %.17g", a, b)
		}
	}
	for {
		var input softV120Record
		if err := d.Decode(&input); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || (input.Case != 0 && input.Case != 19 && input.Case != 20) {
			continue
		}
		before := append([]softV120Step(nil), input.Steps...)
		got := runCoverageQuery(context.Background(), input)
		if got.Error != "" {
			t.Fatal(got.Error)
		}
		old, err := decideRegimeOutcome(input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got.Pool, old.Pool) || !reflect.DeepEqual(got.Origins, old.Origins) || !reflect.DeepEqual(before, input.Steps) {
			t.Fatal("ownership or support")
		}
		if len(got.Pool) == 0 {
			if got.Fits != 0 || got.Selected != [2]int{-1, -1} {
				t.Fatal("empty pool")
			}
		} else {
			if got.Fits != 1+2*len(got.Pool) {
				t.Fatal("fit accounting")
			}
			close(got.BaseLogEvidence, old.BaseLogEvidence)
			for i, b := range got.Branches {
				v := old.Values[i]
				if b.Origin != v.Origin {
					t.Fatal("origin order")
				}
				for y := 0; y < 2; y++ {
					close(b.Mass[y], v.Mass[y])
					close(b.LogEvidence[y], v.LogEvidence[y])
					for j, x := range old.Probes {
						close(b.Conditional[y][x], v.Conditional[y][j])
						conditionalChecks++
					}
				}
				var probeWeights [512]uint16
				for _, x := range old.Probes {
					probeWeights[x]++
				}
				gain, err := coverageQueryGain(got.Base, b, probeWeights)
				if err != nil {
					t.Fatal(err)
				}
				close(gain, v.Gain)
				for x, p := range got.Base {
					close(p, b.Mass[0]*b.Conditional[0][x]+b.Mass[1]*b.Conditional[1][x])
					conditionalChecks++
				}
			}
			if !poisoned {
				bad := input
				bad.Steps = append([]softV120Step(nil), input.Steps...)
				bad.Teacher = nil
				bad.Case = 99
				bad.Phase = 9
				bad.Index = 999
				for j := range bad.Steps {
					bad.Steps[j].Q = -1
					if j >= 160 || bad.Steps[j].Missing || j+bad.Steps[j].Delay > 160 {
						bad.Steps[j].Y = !bad.Steps[j].Y
						if j < 160 {
							bad.Steps[j].Missing = true
							bad.Steps[j].Delay = 0
						}
					}
					if j > 160 {
						bad.Steps[j].X = 999
					}
				}
				again := runCoverageQuery(context.Background(), bad)
				again.Phase = got.Phase
				again.Case = got.Case
				again.Index = got.Index
				if !reflect.DeepEqual(again, got) {
					t.Fatal("oracle/future label/input leak")
				}
				poisoned = true
			}
		}
		checked++
	}
	if checked != 6 || !poisoned {
		t.Fatal("coverage", checked, poisoned)
	}
	t.Logf("%d fixtures; %d conditional identities; old conditionals, weights, support, poisoning and ownership PASS; max error %.3g", checked, conditionalChecks, maxError)
}
