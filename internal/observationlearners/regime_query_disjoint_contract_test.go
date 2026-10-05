package observationlearners

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestRegimeDisjointContracts(t *testing.T) {
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
		if input.Phase != 0 || input.Index != 0 || (input.Case != 0 && input.Case != 5 && input.Case != 19 && input.Case != 20) {
			continue
		}
		before, _ := json.Marshal(input)
		r := runRegimeDisjoint(input)
		if r.Error != "" {
			t.Fatal(r.Error)
		}
		if !reflect.DeepEqual(r.Original, runRegimeOutcome(input)) {
			t.Fatal("original control drift")
		}
		for i, x := range r.Candidate.Decision.Probes {
			if x != input.Steps[137+i].X {
				t.Fatal("wrong probe window")
			}
		}
		if input.Schedule == 0 && (r.ExtraBatches != 0 || r.Original.Predictions != r.Candidate.Predictions) {
			t.Fatal("complete delivery changed")
		}
		for arm := 1; arm < 4; arm++ {
			if r.Candidate.Decision.Costs[arm] != r.Original.Decision.Costs[arm] {
				t.Fatal("paid cost changed")
			}
		}
		mutant := input
		mutant.Steps = append([]softV120Step(nil), input.Steps...)
		known := map[int]bool{}
		for _, j := range r.Original.Decision.Origins {
			known[j] = true
		}
		for j := range mutant.Steps {
			s := &mutant.Steps[j]
			s.Q = 1 - s.Q
			if !known[j] {
				s.Y = !s.Y
			}
			if j > 160 {
				s.X ^= 511
			}
		}
		old, err := decideRegimeOutcome(mutant)
		if err != nil {
			t.Fatal(err)
		}
		newDecision, err := disjointRegimeDecision(mutant, old)
		if err != nil || !reflect.DeepEqual(old, r.Original.Decision) || !reflect.DeepEqual(newDecision, r.Candidate.Decision) {
			t.Fatal("decision-time leak")
		}
		post := input
		post.Steps = append([]softV120Step(nil), input.Steps...)
		paid := map[int]bool{}
		for _, bundle := range []regimeOutcomeRecord{r.Original, r.Candidate} {
			for _, j := range bundle.Decision.Selected {
				if j >= 0 {
					paid[j] = true
				}
			}
		}
		for j := range post.Steps {
			s := &post.Steps[j]
			s.Q = 1 - s.Q
			if j >= 161 || ((s.Missing || j+s.Delay > 161) && !paid[j]) {
				s.Y = !s.Y
			}
		}
		if !reflect.DeepEqual(publishRegimeOutcome(post, r.Candidate.Decision), r.Candidate) {
			t.Fatal("publication-time leak")
		}
		if input.Case == 0 && input.Schedule == 1 {
			var got [2]regimeDisjointRecord
			var wg sync.WaitGroup
			for i := range got {
				wg.Add(1)
				go func(i int) { defer wg.Done(); got[i] = runRegimeDisjoint(input) }(i)
			}
			wg.Wait()
			for _, v := range got {
				if !reflect.DeepEqual(v, r) {
					t.Fatal("concurrent detached replay")
				}
			}
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("input mutated")
		}
		checked++
	}
	if checked != 8 {
		t.Fatal("fixture count", checked)
	}
	t.Log("eight fixtures: unchanged controls/marginals, costs, disjoint probes, both leakage boundaries and concurrent replay PASS")
}
