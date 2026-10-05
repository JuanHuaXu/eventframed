package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"testing"
)

type acquisitionTrainCollected struct {
	Phase, Case, Index, Schedule int
	Results                      [4]acquisitionTrainResult
	Brier, Accuracy, LogLoss     [4][5][2]float64
	Error                        string `json:",omitempty"`
}

func acquisitionTrainCollect(input softV120Record) acquisitionTrainCollected {
	return acquisitionTrainCollectTiming(input, false)
}

func acquisitionTrainCollectTiming(input softV120Record, aligned bool) acquisitionTrainCollected {
	out := acquisitionTrainCollected{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule}
	for policy := 0; policy < 4; policy++ {
		r, err := acquisitionTrainRunTiming(input, policy, 256, aligned)
		if err != nil {
			out.Error = err.Error()
			return out
		}
		out.Results[policy] = r
		if policy > 1 && len(r.Queries) != len(out.Results[1].Queries) {
			out.Error = "query budget mismatch"
			return out
		}
		if input.Schedule == 0 && len(r.Queries) != 0 {
			out.Error = "paid complete evidence"
			return out
		}
		for t, ps := range r.Predictions {
			for arm, p := range ps {
				if math.IsNaN(p) || p <= 0 || p >= 1 {
					out.Error = "invalid probability"
					return out
				}
				if policy == 0 && math.Abs(p-input.Steps[t].P[[]int{0, 1, 2, 3, 12}[arm]]) > 1e-12 {
					out.Error = "natural control mismatch"
					return out
				}
				if input.Schedule == 0 && p != out.Results[0].Predictions[t][arm] {
					out.Error = "immediate mismatch"
					return out
				}
				q := input.Steps[t].Q
				loss := (p-q)*(p-q) + q*(1-q)
				accuracy := 1 - q
				if p >= .5 {
					accuracy = q
				}
				logLoss := -q*math.Log(p) - (1-q)*math.Log1p(-p)
				out.Brier[policy][arm][0] += loss / 256
				out.Accuracy[policy][arm][0] += accuracy / 256
				out.LogLoss[policy][arm][0] += logLoss / 256
				if t >= 192 {
					out.Brier[policy][arm][1] += loss / 64
					out.Accuracy[policy][arm][1] += accuracy / 64
					out.LogLoss[policy][arm][1] += logLoss / 64
				}
			}
		}
	}
	return out
}

func TestAcquisitionTrainCollect(t *testing.T) {
	inputPath, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_OUTPUT")
	aligned := os.Getenv("EVENTFRAME_ACQUISITION_ALIGNED") == "1"
	if inputPath == "" || output == "" {
		t.Skip("explicit research paths required")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("output must not exist")
	}
	input, err := os.Open(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	hash := sha256.New()
	decoder := json.NewDecoder(io.TeeReader(input, hash))
	var header softV120Artifact
	if err := decoder.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Version != "soft-learners-v120" {
		t.Fatal("wrong version")
	}
	type job struct {
		index int
		input softV120Record
	}
	jobs := make(chan job, 8)
	results := make([]acquisitionTrainCollected, 2688)
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.index] = acquisitionTrainCollectTiming(j.input, aligned)
			}
		}()
	}
	count := 0
	var decodeError error
	for {
		var r softV120Record
		err := decoder.Decode(&r)
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
			t.Logf("dispatched %d/2688", count)
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
	for _, path := range []string{"acquisition_train_collect_test.go", "acquisition_train_research_test.go", "acquisition_boundary_test.go", "subset.go", "boolean_specialist.go", "markov_advice.go", "aged_advice.go", "../../docs/experiments/mmm-acquisition-train-protocol.md", "../../docs/experiments/mmm-acquisition-boundary-protocol.md"} {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		hashes[path] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(file)
	version := "acquisition-train-consumed-v1"
	if aligned {
		version = "acquisition-train-boundary-v1"
	}
	err = encoder.Encode(map[string]any{"version": version, "inputSHA256": fmt.Sprintf("%x", hash.Sum(nil)), "hashes": hashes, "workers": 4, "records": count})
	failures := 0
	for _, r := range results {
		if r.Error != "" {
			failures++
		}
		if err == nil {
			err = encoder.Encode(r)
		}
	}
	closeErr := file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	if failures > 0 {
		t.Fatalf("retained%d failed records", failures)
	}
}
