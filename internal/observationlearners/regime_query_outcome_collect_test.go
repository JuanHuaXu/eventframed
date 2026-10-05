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

func TestRegimeOutcomeCollect(t *testing.T) {
	source, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_REGIME_OUTCOME_OUTPUT")
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
		i     int
		input softV120Record
	}
	jobs := make(chan job, 8)
	results := make([]regimeOutcomeRecord, 2688)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.i] = runRegimeOutcome(j.input)
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
		if count >= 2688 {
			decodeError = fmt.Errorf("extra records")
			break
		}
		jobs <- job{count, r}
		count++
		if count%336 == 0 {
			t.Logf("dispatched%d/2688", count)
		}
	}
	close(jobs)
	wg.Wait()
	if decodeError != nil {
		t.Fatal(decodeError)
	}
	if count != 2688 {
		t.Fatal("record count", count)
	}
	hashes := map[string]string{}
	for _, p := range []string{"regime_query_outcome_test.go", "regime_query_outcome_contract_test.go", "regime_query_outcome_collect_test.go", "regime_query_tape_test.go", "regime_query_research_test.go", "regime_query_batch_test.go", "segment_posterior.go", "segment_batch_research_test.go", "../../docs/experiments/mmm-regime-query-outcome-protocol.md"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(struct {
		Version, InputSHA256                                                             string
		Records, Workers, DecisionClock, RevealClock, Forecasts, BaseCap, PublicationCap int
		Hazard, GenericPrior                                                             float64
		Hashes                                                                           map[string]string
	}{"regime-query-outcome-v1", fmt.Sprintf("%x", hash.Sum(nil)), 2688, 4, 160, 161, 31, 63, 64, .01, .95, hashes}); err != nil {
		t.Fatal(err)
	}
	failed, batches, fits := 0, 0, 0
	for _, r := range results {
		if r.Error != "" {
			failed++
		}
		if len(r.Decision.Pool) > 0 {
			batches++
		}
		fits += r.ActualFits
		if err := enc.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
	if failed > 0 {
		t.Fatalf("retained%d failures", failed)
	}
	t.Logf("2688 records retained; %d shared query batches and %d actual publication fits", batches, fits)
}
