package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestRegimeDisjointCollect(t *testing.T) {
	source, baseline, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_REGIME_OUTCOME_BASELINE"), os.Getenv("EVENTFRAME_REGIME_DISJOINT_OUTPUT")
	if source == "" || baseline == "" || output == "" {
		t.Skip("explicit paths required")
	}
	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	f, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	bf, err := os.Open(baseline)
	if err != nil {
		t.Fatal(err)
	}
	defer bf.Close()
	sh, bh := sha256.New(), sha256.New()
	d := json.NewDecoder(io.TeeReader(f, sh))
	bd := json.NewDecoder(io.TeeReader(bf, bh))
	var sourceHeader softV120Artifact
	var baselineHeader struct{ Version, InputSHA256 string }
	if err := d.Decode(&sourceHeader); err != nil {
		t.Fatal(err)
	}
	if err := bd.Decode(&baselineHeader); err != nil {
		t.Fatal(err)
	}
	if sourceHeader.Version != "soft-learners-v120" || baselineHeader.Version != "regime-query-outcome-v1" {
		t.Fatal("source versions")
	}
	type job struct {
		i        int
		input    softV120Record
		baseline regimeOutcomeRecord
	}
	jobs := make(chan job, 8)
	results := make([]regimeDisjointRecord, 2688)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := runRegimeDisjoint(j.input)
				if !reflect.DeepEqual(r.Original, j.baseline) {
					r.Error = "original artifact control mismatch"
				}
				results[j.i] = r
			}
		}()
	}
	count := 0
	var decodeError error
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			decodeError = err
			break
		}
		var old regimeOutcomeRecord
		if err := bd.Decode(&old); err != nil {
			decodeError = err
			break
		}
		if count >= 2688 {
			decodeError = fmt.Errorf("extra source")
			break
		}
		jobs <- job{count, input, old}
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
		t.Fatal("count", count)
	}
	var trailing any
	if err := bd.Decode(&trailing); err != io.EOF {
		t.Fatal("baseline trailing data", err)
	}
	inputHash := fmt.Sprintf("%x", sh.Sum(nil))
	if baselineHeader.InputSHA256 != inputHash {
		t.Fatal("baseline source mismatch")
	}
	hashes := map[string]string{}
	for _, p := range []string{"regime_query_disjoint_test.go", "regime_query_disjoint_contract_test.go", "regime_query_disjoint_collect_test.go", "regime_query_outcome_test.go", "regime_query_tape_test.go", "regime_query_batch_test.go", "regime_query_research_test.go", "segment_posterior.go", "segment_batch_research_test.go", "../../docs/experiments/mmm-regime-query-disjoint-protocol.md"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	enc := json.NewEncoder(out)
	if err := enc.Encode(struct {
		Version, InputSHA256, BaselineSHA256     string
		Records, Workers, ProbeStart, ProbeCount int
		Hashes                                   map[string]string
	}{"regime-query-disjoint-v1", inputHash, fmt.Sprintf("%x", bh.Sum(nil)), 2688, 4, 137, 8, hashes}); err != nil {
		t.Fatal(err)
	}
	failed, batches, oldFits, newFits := 0, 0, 0, 0
	for _, r := range results {
		if r.Error != "" {
			failed++
		}
		batches += r.ExtraBatches
		oldFits += r.Original.ActualFits
		newFits += r.Candidate.ActualFits
		if err := enc.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := out.Sync(); err != nil {
		t.Fatal(err)
	}
	if failed > 0 {
		t.Fatalf("retained%d failures", failed)
	}
	t.Logf("2688 records and exact original controls; %d extra batches; old/new publication fits%d/%d", batches, oldFits, newFits)
}
