package observationlearners

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestRegimeCounterfactualContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit source")
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
	checked, changed := 0, 0
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || (input.Case != 0 && input.Case != 19 && input.Case != 20) {
			continue
		}
		before, _ := json.Marshal(input)
		r := runRegimeCounterfactual(input)
		if r.Error != "" {
			t.Fatal(r.Error)
		}
		if !reflect.DeepEqual(input, regimeCounterfactualInput(regimeCounterfactualInput(input))) {
			t.Fatal("involution")
		}
		for i, j := range r.Original.Choices {
			b, a := i/4, i%4
			old, new := r.Original.Bundles[b], r.Flipped.Bundles[b]
			flip := j >= 0 && !old.Redundant[a]
			if !flip && (old.Predictions[a] != new.Predictions[a] || old.LogEvidence[a] != new.LogEvidence[a]) {
				t.Fatal("unchanged control moved")
			}
			single := input
			single.Steps = append([]softV120Step(nil), input.Steps...)
			if flip {
				single.Steps[j].Y = !single.Steps[j].Y
			}
			for k := range single.Steps {
				s := &single.Steps[k]
				s.Q = 1 - s.Q
				if k >= 161 || ((s.Missing || k+s.Delay > 161) && k != j) {
					s.Y = !s.Y
				}
			}
			expected := publishRegimeOutcome(single, regimeOutcomeDecision{Clock: 160, Selected: [4]int{j, j, j, j}})
			if expected.Error != "" || expected.Predictions[0] != new.Predictions[a] || expected.LogEvidence[0] != new.LogEvidence[a] {
				t.Fatal("cross-branch or oracle leak")
			}
			if flip && old.Predictions[a] != new.Predictions[a] {
				changed++
			}
		}
		if input.Case == 0 && input.Schedule == 1 {
			var got [2]regimeCounterfactualRecord
			var wg sync.WaitGroup
			for i := range got {
				wg.Add(1)
				go func(i int) { defer wg.Done(); got[i] = runRegimeCounterfactual(input) }(i)
			}
			wg.Wait()
			for _, x := range got {
				if !reflect.DeepEqual(x, r) {
					t.Fatal("concurrent replay")
				}
			}
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("ownership")
		}
		checked++
	}
	if checked != 6 || changed == 0 {
		t.Fatal("fixtures or non-noop", checked, changed)
	}
	t.Logf("six fixtures; %d changed branches; isolated-label equivalence, unchanged controls, oracle leakage and race replay PASS", changed)
}
