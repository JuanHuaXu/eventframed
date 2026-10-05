package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
)

type queryBurstCollected struct {
	Phase, Case, Index, Schedule int
	Results                      [7]acquisitionTrainResult
	Error                        string `json:",omitempty"`
}

// Keep failed records and all forecasts. Quality scoring is a separate pass.
func TestQueryBurstCollect(t *testing.T) {
	inputPath, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_QUERY_BURST_OUTPUT")
	if inputPath == "" || output == "" {
		t.Skip("explicit research paths required")
	}
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	input, err := os.Open(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	h := sha256.New()
	d := json.NewDecoder(io.TeeReader(input, h))
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Version != "soft-learners-v120" {
		t.Fatal("wrong input")
	}
	type job struct {
		index int
		input softV120Record
	}
	jobs := make(chan job, 8)
	results := make([]queryBurstCollected, 2688)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := queryBurstCollected{Phase: j.input.Phase, Case: j.input.Case, Index: j.input.Index, Schedule: j.input.Schedule}
				for arm := 0; arm < 7; arm++ {
					mode, policy := 2, arm
					if arm >= 4 {
						mode, policy = 3, arm-3
					}
					out, err := acquisitionTrainRunMode(j.input, policy, 256, mode)
					r.Results[arm] = out
					if err != nil {
						r.Error = err.Error()
						break
					}
				}
				results[j.index] = r
			}
		}()
	}
	count := 0
	var decodeError error
	for {
		var r softV120Record
		err := d.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			decodeError = err
			break
		}
		if count >= len(results) {
			decodeError = fmt.Errorf("too many records")
			break
		}
		jobs <- job{count, r}
		count++
		if count%128 == 0 {
			t.Logf("dispatched%d/2688", count)
		}
	}
	close(jobs)
	wg.Wait()
	if decodeError != nil {
		t.Fatal(decodeError)
	}
	if count != 2688 {
		t.Fatalf("record count%d", count)
	}
	hashes := map[string]string{}
	for _, path := range []string{"query_burst_collect_test.go", "query_burst_schedule_test.go", "acquisition_train_research_test.go", "query_burst_integration_test.go", "subset.go", "boolean_specialist.go", "markov_advice.go", "aged_advice.go", "../../docs/experiments/mmm-query-burst-protocol.md"} {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		hashes[path] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	enc := json.NewEncoder(file)
	if err := enc.Encode(struct {
		Version, InputSHA256 string
		Hashes               map[string]string
		Arms                 []string
		Workers              int
	}{"query-burst-consumed-v1", fmt.Sprintf("%x", h.Sum(nil)), hashes, []string{"natural", "periodic-random", "periodic-entropy", "periodic-disagreement", "burst-random", "burst-entropy", "burst-disagreement"}, 4}); err != nil {
		t.Fatal(err)
	}
	failed := 0
	for _, r := range results {
		if r.Error != "" {
			failed++
		}
		if err := enc.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := file.Sync(); err != nil {
		t.Fatal(err)
	}
	if failed > 0 {
		t.Fatalf("retained%d failed records", failed)
	}
	t.Log("collected2688 records; budgets and quality require independent scoring")
}
