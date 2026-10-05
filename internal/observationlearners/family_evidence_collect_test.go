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

type familyEvidenceCollected struct {
	Phase, Case, Index, Schedule int
	Results                      [2]familyEvidenceStream
	Error                        string `json:",omitempty"`
}

func TestFamilyEvidenceCollect(t *testing.T) {
	source, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_FAMILY_EVIDENCE_OUTPUT")
	if source == "" || output == "" {
		t.Skip("explicit research paths required")
	}
	f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	input, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	hash := sha256.New()
	d := json.NewDecoder(io.TeeReader(input, hash))
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Version != "soft-learners-v120" {
		t.Fatal("source version")
	}
	type job struct {
		index int
		input softV120Record
	}
	jobs := make(chan job, 8)
	results := make([]familyEvidenceCollected, 2688)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := familyEvidenceCollected{Phase: j.input.Phase, Case: j.input.Case, Index: j.input.Index, Schedule: j.input.Schedule}
				for i, stride := range []int{32, 8} {
					out, err := familyEvidenceRun(j.input, stride, 256)
					r.Results[i] = out
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
		if count == len(results) {
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
		t.Fatalf("records%d", count)
	}
	hashes := map[string]string{}
	for _, p := range []string{"family_evidence_collect_test.go", "family_evidence_stream_test.go", "family_evidence_test.go", "family.go", "family_prior.go", "subset.go", "boolean_specialist.go", "segment_posterior.go", "../../docs/experiments/mmm-family-evidence-protocol.md"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(struct {
		Version, InputSHA256 string
		Records, Workers     int
		Hashes               map[string]string
	}{"family-evidence-stream-v1", fmt.Sprintf("%x", hash.Sum(nil)), count, 4, hashes}); err != nil {
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
	t.Log("all2688 four-variant records collected; independent scoring required")
}
