package observationlearners

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestAcquisitionBoundaryContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit consumed data required")
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
	checked, boundaries := 0, 0
	for {
		var input softV120Record
		if err := d.Decode(&input); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Case != 20 || input.Index != 0 {
			continue
		}
		for policy := 0; policy < 4; policy++ {
			old, err := acquisitionTrainRun(input, policy, 256)
			if err != nil {
				t.Fatal(err)
			}
			aligned, err := acquisitionTrainRunTiming(input, policy, 256, true)
			if err != nil {
				t.Fatal(err)
			}
			if len(old.Queries) != len(aligned.Queries) {
				t.Fatal("cost changed")
			}
			if (policy == 0 || input.Schedule == 0) && !reflect.DeepEqual(old, aligned) {
				t.Fatal("no-acquisition path changed")
			}
			for _, q := range aligned.Queries {
				end := q.Clock
				if q.Clock%32 == 31 {
					end++
					boundaries++
					found := false
					for _, fit := range aligned.Fits {
						if fit.Clock == q.Reveal {
							for _, j := range fit.Origins[0] {
								found = found || j == q.Origin
							}
						}
					}
					if !found {
						t.Fatal("boundary label missed next fit")
					}
				}
				if q.Reveal != q.Clock+1 || q.Origin < end-8 || q.Origin >= end {
					t.Fatal("invalid query")
				}
			}
			for _, clock := range []int{31, 32, 159, 160} {
				changed := input
				changed.Steps = append([]softV120Step(nil), input.Steps...)
				paid := map[int]bool{}
				var prefix []acquisitionTrainQuery
				for _, q := range aligned.Queries {
					if q.Reveal <= clock {
						paid[q.Origin] = true
					}
					if q.Clock <= clock {
						prefix = append(prefix, q)
					}
				}
				for j := range changed.Steps {
					s := &changed.Steps[j]
					s.Q = 1 - s.Q
					if j >= clock || ((s.Missing || j+s.Delay > clock) && !paid[j]) {
						s.Y = !s.Y
					}
				}
				other, err := acquisitionTrainRunTiming(changed, policy, clock+1, true)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(other.Predictions, aligned.Predictions[:clock+1]) || !reflect.DeepEqual(other.Queries, prefix) {
					t.Fatal("future label affected aligned prefix")
				}
			}
		}
		checked++
	}
	if checked != 2 || boundaries != 21 {
		t.Fatalf("coverage checked%d boundary%d", checked, boundaries)
	}
	t.Log("two schedules, four policies, 21 shifted acquisitions and 32 as-of prefixes PASS")
}
