package observationlearners

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestQueryBurstLearnerContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit consumed artifact required")
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
		natural, err := acquisitionTrainRun(input, 0, 256)
		if err != nil {
			t.Fatal(err)
		}
		for mode := 2; mode <= 3; mode++ {
			zero, err := acquisitionTrainRunMode(input, 0, 256, mode)
			if err != nil || !reflect.DeepEqual(zero, natural) {
				t.Fatal("natural path changed", err)
			}
			for policy := 1; policy <= 3; policy++ {
				result, err := acquisitionTrainRunMode(input, policy, 256, mode)
				if err != nil {
					t.Fatal(err)
				}
				if input.Schedule == 0 && (len(result.Queries) != 0 || !reflect.DeepEqual(result.Predictions, natural.Predictions)) {
					t.Fatal("complete feedback changed")
				}
				if input.Schedule == 1 && len(result.Queries) != 31 {
					t.Fatalf("unequal budget case%d mode%d policy%d: %d", input.Case, mode, policy, len(result.Queries))
				}
				seen := map[int]bool{}
				for _, a := range result.Arrivals {
					s := input.Steps[a.Origin]
					if seen[a.Origin] {
						t.Fatal("duplicate arrival")
					}
					seen[a.Origin] = true
					paid := false
					for _, q := range result.Queries {
						paid = paid || (q.Origin == a.Origin && q.Reveal <= a.Clock)
					}
					if a.Paid != paid || (!paid && (s.Missing || a.Origin+s.Delay > a.Clock)) {
						t.Fatal("unavailable evidence")
					}
					if a.Mixer != (a.Origin >= a.Clock-32) {
						t.Fatal("wrong mixer expiry treatment")
					}
					p := result.Predictions[a.Origin][4]
					if !s.Y {
						p = 1 - p
					}
					if a.Surprise != (p < .2) {
						t.Fatal("wrong surprise anchor")
					}
				}
				seen = map[int]bool{}
				for _, q := range result.Queries {
					if seen[q.Origin] || q.Reveal != q.Clock+1 || q.Origin >= q.Clock || q.Origin < max(0, q.Clock-32) {
						t.Fatal("invalid query")
					}
					seen[q.Origin] = true
				}
				for _, fit := range result.Fits {
					for _, origins := range fit.Origins {
						for _, j := range origins {
							if j < 0 {
								continue
							}
							s := input.Steps[j]
							paid := false
							for _, q := range result.Queries {
								paid = paid || (q.Origin == j && q.Reveal <= fit.Clock)
							}
							if j >= fit.Clock || (!paid && (s.Missing || j+s.Delay > fit.Clock)) {
								t.Fatal("unrevealed fit label")
							}
						}
					}
				}
				if input.Schedule == 1 && input.Case == 20 {
					for _, clock := range []int{32, 160} {
						changed := input
						changed.Steps = append([]softV120Step(nil), input.Steps...)
						for j := range changed.Steps {
							s := &changed.Steps[j]
							s.Q = 1 - s.Q
							arrived := false
							for _, a := range result.Arrivals {
								arrived = arrived || (a.Origin == j && a.Clock <= clock)
							}
							if !arrived {
								s.Y = !s.Y
							}
						}
						other, err := acquisitionTrainRunMode(changed, policy, clock+1, mode)
						if err != nil {
							t.Fatal(err)
						}
						var queries []acquisitionTrainQuery
						for _, q := range result.Queries {
							if q.Clock <= clock {
								queries = append(queries, q)
							}
						}
						if !reflect.DeepEqual(other.Predictions, result.Predictions[:clock+1]) || !reflect.DeepEqual(other.Queries, queries) {
							t.Fatal("future evidence changed prefix")
						}
						prefixes++
					}
				}
			}
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("input mutated")
		}
		fixtures++
	}
	if fixtures != 4 || prefixes != 12 {
		t.Fatalf("fixtures%d prefixes%d", fixtures, prefixes)
	}
	t.Log("four fixtures, two timings, three selectors, 12 poisoned prefixes passed")
}
