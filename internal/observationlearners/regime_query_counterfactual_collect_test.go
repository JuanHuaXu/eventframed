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

func TestRegimeCounterfactualCollect(t *testing.T) {
	source, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_REGIME_COUNTERFACTUAL_OUTPUT")
	if source == "" || output == "" {
		t.Skip("explicit paths required")
	}
	f, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	sh := sha256.New()
	d := json.NewDecoder(io.TeeReader(f, sh))
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
	results := make([]regimeCounterfactualRecord, 2688)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.i] = runRegimeCounterfactual(j.input)
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
		if count >= 2688 {
			decodeError = fmt.Errorf("extra source")
			break
		}
		jobs <- job{count, input}
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
	hash := fmt.Sprintf("%x", sh.Sum(nil))
	if hash != "5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f" {
		t.Fatal("source hash")
	}
	hashes := map[string]string{}
	for _, p := range []string{"regime_query_counterfactual_test.go", "regime_query_counterfactual_contract_test.go", "regime_query_counterfactual_collect_test.go", "regime_query_envelope_test.go", "regime_query_outcome_test.go", "segment_posterior.go", "segment_batch_research_test.go", "soft_learners_v118_test.go", "../transfergenerator/generator.go", "../../docs/experiments/mmm-regime-query-counterfactual-protocol.md"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	enc := json.NewEncoder(out)
	if err := enc.Encode(struct {
		Version, InputSHA256 string
		Records, Workers     int
		Hashes               map[string]string
	}{"regime-query-counterfactual-v1", hash, 2688, 4, hashes}); err != nil {
		t.Fatal(err)
	}
	fits, failed := 0, 0
	for _, r := range results {
		fits += r.Original.ActualFits + r.Flipped.ActualFits
		if r.Error != "" {
			failed++
		}
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
	t.Logf("2688 original/flipped pairs; %d fits", fits)
}
