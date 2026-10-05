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

func TestFrozenPublicationGuards(t *testing.T) {
	input := softV120Record{Steps: make([]softV120Step, 256)}
	for i := range input.Steps {
		input.Steps[i].Missing = true
		input.Steps[i].Y = true
	}
	input.Initial[15].Outcome = true
	h, origins, err := frozenPublicationHistory(input, 63, -1, false)
	if err != nil || len(origins) != 16 || !h[15].Outcome {
		t.Fatal("noquery sentinel overwrote initial origin -1", err)
	}
	h, origins, err = frozenPublicationHistory(input, 64, 152, false)
	if err != nil || len(origins) != 17 || h[168].Outcome || !h[15].Outcome {
		t.Fatal("explicit answer or initial evidence", err)
	}
	for _, cap := range []int{0, 62, 65} {
		if _, _, err := frozenPublicationHistory(input, cap, -1, false); err == nil {
			t.Fatal("capacity", cap)
		}
	}
	for _, j := range []int{-2, 0, 151, 160} {
		if _, _, err := frozenPublicationHistory(input, 64, j, false); err == nil {
			t.Fatal("invalid query", j)
		}
	}
	input.Steps[152].Missing = false
	if _, _, err := frozenPublicationHistory(input, 64, 152, false); err == nil {
		t.Fatal("known query")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fitFrozenPublication(ctx, input, 64, -1, false); err != context.Canceled {
		t.Fatal("cancellation", err)
	}
}

func TestFrozenPublicationContracts(t *testing.T) {
	source, covPath := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_QUERY_COVERAGE_INPUT")
	if source == "" || covPath == "" {
		t.Skip("explicit artifacts required")
	}
	f, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := os.Open(covPath)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	d, cd := json.NewDecoder(f), json.NewDecoder(c)
	var head, chead struct{ Version string }
	if err := d.Decode(&head); err != nil || head.Version != "soft-learners-v120" {
		t.Fatal("source header", err)
	}
	if err := cd.Decode(&chead); err != nil || chead.Version != "query-coverage-v1" {
		t.Fatal("coverage header", err)
	}
	checked, checks := 0, 0
	maxError := 0.
	poisoned := false
	close := func(a, b [512]float64) {
		t.Helper()
		for x := range a {
			e := math.Abs(a[x] - b[x])
			maxError = math.Max(maxError, e)
			if e > 1e-10 {
				t.Fatalf("clock reuse: %.17g != %.17g", a[x], b[x])
			}
			checks++
		}
	}
	for {
		var input softV120Record
		var coverage coverageQueryRecord
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if err := cd.Decode(&coverage); err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || (input.Case != 0 && input.Case != 19 && input.Case != 20) {
			continue
		}
		if coverage.Phase != input.Phase || coverage.Case != input.Case || coverage.Index != input.Index || coverage.Schedule != input.Schedule {
			t.Fatal("artifact identity")
		}
		before := append([]softV120Step(nil), input.Steps...)
		mA, err := fitFrozenPublication(context.Background(), input, 63, -1, false)
		if err != nil {
			t.Fatal(err)
		}
		mB, err := fitFrozenPublication(context.Background(), input, 64, -1, false)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(mA.origins, coverage.Origins) {
			t.Fatal("base support")
		}
		if len(coverage.Pool) > 0 {
			close(mA.predictions, coverage.Base)
			for _, i := range []int{0, len(coverage.Pool) - 1} {
				j := coverage.Pool[i]
				for y := 0; y < 2; y++ {
					hA, oA, err := frozenPublicationHistory(input, 63, j, y == 1)
					if err != nil {
						t.Fatal(err)
					}
					hB, oB, err := frozenPublicationHistory(input, 64, j, y == 1)
					if err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(hA, hB) || !reflect.DeepEqual(oA, oB) {
						t.Fatal("paid A/B support")
					}
					m, err := fitFrozenPublication(context.Background(), input, 64, j, y == 1)
					if err != nil {
						t.Fatal(err)
					}
					close(m.predictions, coverage.Branches[i].Conditional[y])
				}
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
			m, err := fitFrozenPublication(context.Background(), bad, 64, -1, false)
			if err != nil || !reflect.DeepEqual(m, mB) {
				t.Fatal("fitter oracle/future leak", err)
			}
			poisoned = true
		}
		if !reflect.DeepEqual(before, input.Steps) {
			t.Fatal("input mutated")
		}
		result := runPublicationQuery(context.Background(), input, coverage)
		if result.Error != "" || result.Fits != 1+boolInt(len(coverage.Pool) == 0) {
			t.Fatal("worker", result.Error, result.Fits)
		}
		close(result.FrozenForecast, mA.predictions)
		close(result.RestoredForecast, mB.predictions)
		checked++
		if checked == 6 {
			break
		}
	}
	if checked != 6 || !poisoned {
		t.Fatal("fixture coverage", checked)
	}
	t.Logf("six fixtures; %d reused-forecast comparisons; max error %.3g; support, explicit answers, future/oracle isolation and ownership PASS", checks, maxError)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
