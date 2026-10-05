package observationlearners

import (
	"encoding/json"
	"io"
	"math"
	"os"
	"reflect"
	"testing"
)

func TestRefitCadenceContracts(t *testing.T) {
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
		if err := d.Decode(&input); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || (input.Case != 0 && input.Case != 20) {
			continue
		}
		before, _ := json.Marshal(input)
		old, err := acquisitionTrainRun(input, 0, 256)
		if err != nil {
			t.Fatal(err)
		}
		for j, ps := range old.Predictions {
			for a, p := range ps {
				if math.Abs(p-input.Steps[j].P[[]int{0, 1, 2, 3, 12}[a]]) > 1e-12 {
					t.Fatal("32-clock control changed")
				}
			}
		}
		fast, err := acquisitionTrainRunCadence(input, 0, 256, 0, 8)
		if err != nil {
			t.Fatal(err)
		}
		if len(fast.Fits) != 32 || len(fast.Queries) != 0 {
			t.Fatal("wrong intervention")
		}
		if !reflect.DeepEqual(fast.Predictions[:8], old.Predictions[:8]) {
			t.Fatal("changed before first new fit")
		}
		for k, fit := range fast.Fits {
			if fit.Clock != k*8 {
				t.Fatal("fit clock")
			}
			var origins []int
			for j := -16; j < fit.Clock; j++ {
				if j < 0 || (!input.Steps[j].Missing && j+input.Steps[j].Delay <= fit.Clock) {
					origins = append(origins, j)
				}
			}
			for w, cap := range []int{64, 32} {
				if !reflect.DeepEqual(fit.Origins[w], origins[max(0, len(origins)-cap):]) {
					t.Fatal("fit origin mismatch")
				}
			}
		}
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
			other, err := acquisitionTrainRunCadence(changed, 0, clock+1, 0, 8)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(other.Predictions, fast.Predictions[:clock+1]) {
				t.Fatal("future evidence affected faster fit")
			}
			prefixes++
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("input mutated")
		}
		fixtures++
	}
	if fixtures != 4 || prefixes != 8 {
		t.Fatalf("fixtures%d prefixes%d", fixtures, prefixes)
	}
	t.Log("four paired control fixtures, 128 fast-fit bundles, eight poisoned prefixes PASS")
}
