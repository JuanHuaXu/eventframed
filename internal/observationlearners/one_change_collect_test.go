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

func TestOneChangeCollect(t *testing.T) {
	source, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_ONE_CHANGE_OUTPUT")
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
	results := make([][3]oneChangeRecord, 384)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				for k, clock := range []int{160, 192, 224} {
					results[j.index][k] = oneChangeSnapshot(j.input, clock)
				}
			}
		}()
	}
	count, seen := 0, 0
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
		seen++
		if r.Case != 0 && r.Case != 19 && r.Case != 20 {
			continue
		}
		if count >= 384 {
			decodeError = fmt.Errorf("extra records")
			break
		}
		jobs <- job{count, r}
		count++
		if count%64 == 0 {
			t.Logf("dispatched%d/384", count)
		}
	}
	close(jobs)
	wg.Wait()
	if decodeError != nil {
		t.Fatal(decodeError)
	}
	if count != 384 || seen != 2688 {
		t.Fatalf("counts%d/%d", count, seen)
	}
	hashes := map[string]string{}
	for _, p := range []string{"one_change_research_test.go", "one_change_collect_test.go", "segment_posterior.go", "segment_batch_research_test.go", "../../docs/experiments/mmm-one-change-protocol.md"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(struct {
		Version, InputSHA256               string
		Records, Workers, Cap, MinimumSide int
		NoChangePrior, GenericPrior        float64
		Hashes                             map[string]string
	}{"one-change-v1", fmt.Sprintf("%x", hash.Sum(nil)), 1152, 4, 64, 8, .9, .95, hashes}); err != nil {
		t.Fatal(err)
	}
	failed := 0
	for _, snapshots := range results {
		for _, r := range snapshots {
			if r.Error != "" {
				failed++
			}
			if err := enc.Encode(r); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if failed > 0 {
		t.Fatalf("retained%d failures", failed)
	}
	t.Log("1152 bounded one-change snapshots collected without oracle boundary")
}
