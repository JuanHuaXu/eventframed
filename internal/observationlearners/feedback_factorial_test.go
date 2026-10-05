package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"
)

func feedbackIntervention(input softV120Record, mode int) (softV120Record, error) {
	if mode < 0 || mode > 3 {
		return softV120Record{}, fmt.Errorf("invalid feedback mode")
	}
	out := input
	out.Steps = append([]softV120Step(nil), input.Steps...)
	for i := range out.Steps {
		if mode < 2 {
			out.Steps[i].Delay = 0
		}
		if mode == 0 || mode == 2 {
			out.Steps[i].Missing = false
		}
	}
	return out, nil
}

func TestFeedbackFactorialContracts(t *testing.T) {
	path := os.Getenv("EVENTFRAME_FEEDBACK_FACTORIAL_INPUT")
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
	var pair [2]softV120Record
	for {
		var r softV120Record
		if err := d.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		if r.Phase == 0 && r.Case == 20 && r.Index == 0 {
			pair[r.Schedule] = r
		}
	}
	before, _ := json.Marshal(pair[1])
	for mode := 0; mode < 4; mode++ {
		data, err := feedbackIntervention(pair[1], mode)
		if err != nil {
			t.Fatal(err)
		}
		result, err := acquisitionTrainRun(data, 0, 256)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Queries) != 0 {
			t.Fatal("unexpected paid query")
		}
		if mode == 0 || mode == 3 {
			schedule := mode / 3
			for j, ps := range result.Predictions {
				for a, k := range []int{0, 1, 2, 3, 12} {
					if math.Abs(ps[a]-pair[schedule].Steps[j].P[k]) > 1e-12 {
						t.Fatal("control mismatch")
					}
				}
			}
		}
		for _, clock := range []int{32, 160} {
			changed := data
			changed.Steps = append([]softV120Step(nil), data.Steps...)
			for j := range changed.Steps {
				s := &changed.Steps[j]
				s.Q = 1 - s.Q
				if j >= clock || s.Missing || j+s.Delay > clock {
					s.Y = !s.Y
				}
			}
			other, err := acquisitionTrainRun(changed, 0, clock+1)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(other.Predictions, result.Predictions[:clock+1]) {
				t.Fatal("future leakage")
			}
		}
	}
	after, _ := json.Marshal(pair[1])
	if string(before) != string(after) {
		t.Fatal("intervention mutated input")
	}
	t.Log("four modes, two original controls, eight as-of prefixes PASS")
}

type feedbackFactorialRecord struct {
	Phase, Case, Index int
	Results            [2]acquisitionTrainResult
	Error              string `json:",omitempty"`
}

func TestFeedbackFactorialCollect(t *testing.T) {
	inputPath, output := os.Getenv("EVENTFRAME_FEEDBACK_FACTORIAL_INPUT"), os.Getenv("EVENTFRAME_FEEDBACK_FACTORIAL_OUTPUT")
	if inputPath == "" || output == "" {
		t.Skip("explicit paths required")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("output already exists")
	}
	file, err := os.Open(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	hash := sha256.New()
	decoder := json.NewDecoder(io.TeeReader(file, hash))
	var h softV120Artifact
	if err := decoder.Decode(&h); err != nil {
		t.Fatal(err)
	}
	if h.Version != "soft-learners-v120" {
		t.Fatal("version")
	}
	type job struct {
		index int
		input softV120Record
	}
	jobs := make(chan job, 8)
	records := make([]feedbackFactorialRecord, 1344)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := feedbackFactorialRecord{Phase: j.input.Phase, Case: j.input.Case, Index: j.input.Index}
				for mode := 1; mode <= 2; mode++ {
					data, err := feedbackIntervention(j.input, mode)
					if err == nil {
						r.Results[mode-1], err = acquisitionTrainRun(data, 0, 256)
					}
					if err != nil {
						r.Error = err.Error()
						break
					}
				}
				records[j.index] = r
			}
		}()
	}
	n := 0
	var failure error
	for {
		var r softV120Record
		err := decoder.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			failure = err
			break
		}
		if r.Schedule != 1 {
			continue
		}
		if n >= 1344 {
			failure = fmt.Errorf("too many records")
			break
		}
		jobs <- job{n, r}
		n++
		if n%128 == 0 {
			t.Logf("dispatched %d/1344", n)
		}
	}
	close(jobs)
	wg.Wait()
	if failure != nil {
		t.Fatal(failure)
	}
	if n != 1344 {
		t.Fatal("missing records")
	}
	hashes := map[string]string{}
	for _, p := range []string{"feedback_factorial_test.go", "acquisition_train_research_test.go", "subset.go", "boolean_specialist.go", "markov_advice.go", "aged_advice.go", "../../docs/experiments/mmm-feedback-factorial-protocol.md"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	enc := json.NewEncoder(out)
	err = enc.Encode(map[string]any{"version": "feedback-factorial-v1", "inputSHA256": fmt.Sprintf("%x", hash.Sum(nil)), "hashes": hashes, "records": n, "modes": []int{1, 2}, "workers": 4})
	failures := 0
	for _, r := range records {
		if r.Error != "" {
			failures++
		}
		if err == nil {
			err = enc.Encode(r)
		}
	}
	closed := out.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closed != nil {
		t.Fatal(closed)
	}
	if failures > 0 {
		t.Fatalf("retained %d failures", failures)
	}
}
