package observationlearners

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestRegimeOutcomeSelectors(t *testing.T) {
	winner, err := chooseRegimeMaximum([]int{9, 3, 1}, []float64{.5, .5, .1})
	if err != nil || winner != 3 {
		t.Fatal("tie rule")
	}
	for _, bad := range [][]float64{nil, {math.NaN()}, {math.Inf(1)}} {
		if _, err := chooseRegimeMaximum([]int{1}, bad); err == nil {
			t.Fatal("invalid score")
		}
	}
	if _, err := decideRegimeOutcome(softV120Record{}); err == nil {
		t.Fatal("invalid tape")
	}
}

func TestRegimeOutcomeContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit consumed source required")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	var header softV120Artifact
	if err := decoder.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Version != "soft-learners-v120" {
		t.Fatal("source version")
	}
	checked, paidEffects := 0, 0
	for {
		var input softV120Record
		err := decoder.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || (input.Case != 0 && input.Case != 5 && input.Case != 8 && input.Case != 12 && input.Case != 19 && input.Case != 20) {
			continue
		}
		before, _ := json.Marshal(input)
		d, err := decideRegimeOutcome(input)
		if err != nil {
			t.Fatal(err)
		}
		view, pool, probes, err := regimeQueryTapeView(input, 160)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(pool, d.Pool) || !reflect.DeepEqual(probes, d.Probes) {
			t.Fatal("view interface drift")
		}
		if len(pool) > 0 {
			old, err := newRegimeQueryState(context.Background(), -16, 160, view)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(old.base.origins, d.Origins) {
				t.Fatal("factored view support")
			}
		}
		mutant := input
		mutant.Steps = append([]softV120Step(nil), input.Steps...)
		selectedSupport := map[int]bool{}
		for _, j := range d.Origins {
			selectedSupport[j] = true
		}
		for j := range mutant.Steps {
			v := &mutant.Steps[j]
			v.Q = 1 - v.Q
			if !selectedSupport[j] {
				v.Y = !v.Y
			}
			if j > 160 {
				v.X ^= 511
			}
		}
		otherDecision, err := decideRegimeOutcome(mutant)
		if err != nil || !reflect.DeepEqual(d, otherDecision) {
			t.Fatal("decision sees hidden outcome/Q/future X")
		}
		r := publishRegimeOutcome(input, d)
		if r.Error != "" {
			t.Fatal(r.Error)
		}
		if r.Decision.Selected[0] != -1 || r.Decision.Costs[0] != 0 {
			t.Fatal("no-query cost")
		}
		for arm, j := range d.Selected {
			if arm > 0 {
				want := 0
				if len(d.Pool) > 0 {
					want = 1
				}
				if d.Costs[arm] != want {
					t.Fatal("unequal unit costs")
				}
			}
			for i, o := range r.Origins[arm] {
				if i > 0 && o <= r.Origins[arm][i-1] {
					t.Fatal("origin order")
				}
				if o >= 0 {
					v := input.Steps[o]
					if o >= 161 || (!(!v.Missing && o+v.Delay <= 161) && o != j) {
						t.Fatal("unavailable publication evidence")
					}
				}
			}
			if j >= 0 {
				included := false
				for _, o := range r.Origins[arm] {
					included = included || o == j
				}
				if !included {
					t.Fatal("paid label not fitted")
				}
				v := input.Steps[j]
				redundant := !v.Missing && j+v.Delay <= 161
				if redundant != r.Redundant[arm] {
					t.Fatal("redundancy accounting")
				}
				if redundant && !reflect.DeepEqual(r.Origins[arm], r.Origins[0]) {
					t.Fatal("redundant purchase changed support")
				}
			}
			if input.Schedule == 0 && (j != -1 || r.Predictions[arm] != r.Predictions[0] || !reflect.DeepEqual(r.Origins[arm], r.Origins[0])) {
				t.Fatal("complete delivery drift")
			}
		}
		post := input
		post.Steps = append([]softV120Step(nil), input.Steps...)
		paid := map[int]bool{}
		for _, j := range d.Selected {
			if j >= 0 {
				paid[j] = true
			}
		}
		for j := range post.Steps {
			v := &post.Steps[j]
			v.Q = 1 - v.Q
			if j >= 161 || ((v.Missing || j+v.Delay > 161) && !paid[j]) {
				v.Y = !v.Y
			}
		}
		if other := publishRegimeOutcome(post, d); !reflect.DeepEqual(r, other) {
			t.Fatal("publication sees unrevealed outcomes/Q")
		}
		if other := runRegimeOutcome(input); !reflect.DeepEqual(r, other) {
			t.Fatal("detached replay drift")
		}
		if input.Schedule == 1 {
			for arm := 1; arm < 4; arm++ {
				if !r.Redundant[arm] && d.Selected[arm] >= 0 {
					changed := input
					changed.Steps = append([]softV120Step(nil), input.Steps...)
					j := d.Selected[arm]
					changed.Steps[j].Y = !changed.Steps[j].Y
					counter := publishRegimeOutcome(changed, d)
					if counter.Error != "" {
						t.Fatal(counter.Error)
					}
					if counter.Predictions[arm] != r.Predictions[arm] {
						paidEffects++
					}
					break
				}
			}
		}
		if input.Case == 0 && input.Schedule == 1 {
			var got [2]regimeOutcomeRecord
			var wg sync.WaitGroup
			for i := range got {
				wg.Add(1)
				go func(i int) { defer wg.Done(); got[i] = runRegimeOutcome(input) }(i)
			}
			wg.Wait()
			for _, v := range got {
				if !reflect.DeepEqual(v, r) {
					t.Fatal("concurrent replay drift")
				}
			}
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("source mutated")
		}
		checked++
	}
	if checked != 12 || paidEffects == 0 {
		t.Fatalf("coverage%d, paid forecast effects%d", checked, paidEffects)
	}
	t.Logf("12 fixtures: decision/reveal separation, as-of poisoning, cost/redundancy, detached replay and concurrency PASS; %d paid-label flips changed forecasts", paidEffects)
}
