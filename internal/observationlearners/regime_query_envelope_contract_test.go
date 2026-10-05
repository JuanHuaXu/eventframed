package observationlearners

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestRegimeEnvelopeContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit source required")
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
	checked := 0
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
		r := runRegimeEnvelope(input)
		if r.Error != "" {
			t.Fatal(r.Error)
		}
		old := runRegimeOutcome(input)
		if old.Error != "" {
			t.Fatal(old.Error)
		}
		for arm, selected := range old.Decision.Selected {
			found := false
			for i, choice := range r.Choices {
				if selected == choice {
					b, a := r.Bundles[i/4], i%4
					if b.Predictions[a] != old.Predictions[arm] || b.AtPublication[a] != old.AtPublication[arm] || b.LogEvidence[a] != old.LogEvidence[arm] || !reflect.DeepEqual(b.Origins[a], old.Origins[arm]) {
						t.Fatal("old policy mismatch")
					}
					found = true
				}
			}
			if !found {
				t.Fatal("missing choice")
			}
		}
		mutant := input
		mutant.Steps = append([]softV120Step(nil), input.Steps...)
		for j := range mutant.Steps {
			mutant.Steps[j].Y = !mutant.Steps[j].Y
			mutant.Steps[j].Q = 1 - mutant.Steps[j].Q
			mutant.Steps[j].X ^= 511
		}
		choices, err := regimeEnvelopeChoices(mutant)
		if err != nil || !reflect.DeepEqual(choices, r.Choices) {
			t.Fatal("choice enumeration reads content")
		}
		for i, selected := range r.Choices {
			post := input
			post.Steps = append([]softV120Step(nil), input.Steps...)
			for j := range post.Steps {
				s := &post.Steps[j]
				s.Q = 1 - s.Q
				if j >= 161 || ((s.Missing || j+s.Delay > 161) && j != selected) {
					s.Y = !s.Y
				}
			}
			d := regimeOutcomeDecision{Clock: 160, Selected: [4]int{selected, selected, selected, selected}}
			got := publishRegimeOutcome(post, d)
			want := r.Bundles[i/4]
			a := i % 4
			if got.Error != "" || got.Predictions[0] != want.Predictions[a] || got.LogEvidence[0] != want.LogEvidence[a] {
				t.Fatal("unrevealed outcome leak")
			}
		}
		if input.Case == 0 && input.Schedule == 1 {
			var got [2]regimeEnvelopeRecord
			var wg sync.WaitGroup
			for i := range got {
				wg.Add(1)
				go func(i int) { defer wg.Done(); got[i] = runRegimeEnvelope(input) }(i)
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
	if checked != 6 {
		t.Fatal("fixture count", checked)
	}
	if r := runRegimeEnvelope(softV120Record{}); r.Error == "" {
		t.Fatal("invalid length accepted")
	}
	t.Log("six fixtures: all candidate branches, unchanged policy forecasts, per-branch leakage, ownership and concurrent replay PASS")
}
