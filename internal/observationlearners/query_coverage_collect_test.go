package observationlearners

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
)

func TestCoverageQueryCollect(t *testing.T) {
	source, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_QUERY_COVERAGE_OUTPUT")
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
	if err := d.Decode(&header); err != nil || header.Version != "soft-learners-v120" {
		t.Fatal("source header", err)
	}
	type job struct {
		index int
		input softV120Record
	}
	jobs := make(chan job, 8)
	results := make([]coverageQueryRecord, 2688)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.index] = runCoverageQuery(context.Background(), j.input)
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
	for _, p := range []string{"query_coverage_test.go", "query_coverage_contract_test.go", "query_coverage_collect_test.go", "regime_query_tape_test.go", "regime_query_research_test.go", "regime_query_outcome_test.go", "segment_posterior.go", "segment_batch_research_test.go", "soft_learners_v120_test.go", "../../docs/experiments/mmm-query-coverage-protocol.md", "../../go.mod"} {
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
	}{"query-coverage-v1", hash, 2688, 4, hashes}); err != nil {
		t.Fatal(err)
	}
	fits, failed := 0, 0
	for _, r := range results {
		fits += r.Fits
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
		t.Fatalf("retained%d failed records", failed)
	}
	t.Logf("2688 full-response records; %d fits", fits)
}
