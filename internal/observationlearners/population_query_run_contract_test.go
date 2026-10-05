package observationlearners

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestPopulationQueryBranch(t *testing.T) {
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
	count := 0
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
		r := runPopulationQuery(input)
		if r.Error != "" {
			t.Fatal(r.Error)
		}
		old := runRegimeEnvelope(input)
		if old.Error != "" {
			t.Fatal(old.Error)
		}
		for i, b := range r.Branches {
			y := 0
			if b.ActualY {
				y = 1
			}
			want := old.Bundles[i/4]
			a := i % 4
			if b.Predictions[y] != want.Predictions[a] || b.LogEvidence[y] != want.LogEvidence[a] || !reflect.DeepEqual(b.Origins, want.Origins[a]) {
				t.Fatal("old branch mismatch")
			}
			mutant := input
			mutant.Steps = append([]softV120Step(nil), input.Steps...)
			mutant.Case = 99
			mutant.Teacher = nil
			mutant.Rules = [2]uint16{}
			for j := range mutant.Steps {
				s := &mutant.Steps[j]
				s.Q = -1
				if j >= 161 || s.Missing || j+s.Delay > 161 {
					s.Y = !s.Y
				}
			}
			m, _, _, err := fitPopulationQueryBranch(mutant, b.Origin, true)
			if err != nil {
				t.Fatal(err)
			}
			for j, p := range b.AtPublication[1] {
				if p != m.predictions[input.Steps[161+j].X] {
					t.Fatal("oracle/unavailable data influenced fit")
				}
			}
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("ownership")
		}
		count++
	}
	if count != 6 {
		t.Fatal("fixture count", count)
	}
	t.Log("six fixtures: old branch equality, explicit answer override, oracle/unavailable-label isolation and ownership PASS")
}
