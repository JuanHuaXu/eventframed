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

type refitCadenceCollected struct {
	Phase, Case, Index, Schedule int
	Result                       acquisitionTrainResult
	Error                        string `json:",omitempty"`
}

func TestRefitCadenceCollect(t *testing.T) {
	inputPath, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_REFIT_CADENCE_OUTPUT")
	if inputPath == "" || output == "" {
		t.Skip("explicit research paths required")
	}
	f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
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
		t.Fatal("wrong source")
	}
	type job struct {
		i     int
		input softV120Record
	}
	jobs := make(chan job, 8)
	results := make([]refitCadenceCollected, 2688)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := refitCadenceCollected{Phase: j.input.Phase, Case: j.input.Case, Index: j.input.Index, Schedule: j.input.Schedule}
				out, err := acquisitionTrainRunCadence(j.input, 0, 256, 0, 8)
				r.Result = out
				if err != nil {
					r.Error = err.Error()
				}
				results[j.i] = r
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
	for _, path := range []string{"refit_cadence_collect_test.go", "refit_cadence_test.go", "acquisition_train_research_test.go", "subset.go", "boolean_specialist.go", "markov_advice.go", "aged_advice.go", "../../docs/experiments/mmm-refit-cadence-protocol.md"} {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		hashes[path] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(struct {
		Version, InputSHA256     string
		Stride, Records, Workers int
		Hashes                   map[string]string
	}{"refit-cadence-v1", fmt.Sprintf("%x", h.Sum(nil)), 8, count, 4, hashes}); err != nil {
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
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if failed > 0 {
		t.Fatalf("retained%d errors", failed)
	}
	t.Log("collected2688 faster-cadence trajectories; independent scoring required")
}
