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

func TestPublicationQueryCollect(t *testing.T) {
	source, coverage, output := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT"), os.Getenv("EVENTFRAME_QUERY_COVERAGE_INPUT"), os.Getenv("EVENTFRAME_QUERY_PUBLICATION_OUTPUT")
	if source == "" || coverage == "" || output == "" {
		t.Skip("explicit paths required")
	}
	f, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	c, err := os.Open(coverage)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	sh, ch := sha256.New(), sha256.New()
	d, cd := json.NewDecoder(io.TeeReader(f, sh)), json.NewDecoder(io.TeeReader(c, ch))
	var header softV120Artifact
	if err := d.Decode(&header); err != nil || header.Version != "soft-learners-v120" {
		t.Fatal("source header", err)
	}
	var parent struct {
		Version, InputSHA256 string
		Records              int
	}
	if err := cd.Decode(&parent); err != nil || parent.Version != "query-coverage-v1" || parent.Records != 2688 {
		t.Fatal("coverage header", err)
	}
	hashes := map[string]string{}
	for _, p := range []string{"query_publication_test.go", "query_publication_contract_test.go", "query_publication_collect_test.go", "query_coverage_test.go", "regime_query_tape_test.go", "regime_query_research_test.go", "segment_posterior.go", "segment_batch_research_test.go", "population_query_kernel_test.go", "soft_learners_v120_test.go", "../../docs/experiments/mmm-query-publication-protocol.md", "../../go.mod"} {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	type job struct {
		index    int
		input    softV120Record
		coverage coverageQueryRecord
	}
	jobs := make(chan job, 8)
	results := make([]publicationQueryRecord, 2688)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.index] = runPublicationQuery(context.Background(), j.input, j.coverage)
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
		var cov coverageQueryRecord
		if err := cd.Decode(&cov); err != nil {
			decodeError = err
			break
		}
		if count >= len(results) {
			decodeError = fmt.Errorf("extra source")
			break
		}
		jobs <- job{count, input, cov}
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
	var extra json.RawMessage
	if err := cd.Decode(&extra); err != io.EOF {
		t.Fatal("coverage trailing record", err)
	}
	if count != 2688 {
		t.Fatal("count", count)
	}
	hash := fmt.Sprintf("%x", sh.Sum(nil))
	if hash != "5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f" || parent.InputSHA256 != hash {
		t.Fatal("source hash")
	}
	for p, expected := range hashes {
		b, e := os.ReadFile(p)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != expected {
			t.Fatal("source changed during collection", p, e)
		}
	}
	enc := json.NewEncoder(out)
	if err := enc.Encode(struct {
		Version, InputSHA256, CoverageSHA256 string
		Records, Workers                     int
		Hashes                               map[string]string
	}{"query-publication-v1", hash, fmt.Sprintf("%x", ch.Sum(nil)), 2688, 4, hashes}); err != nil {
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
	if failed != 0 {
		t.Fatalf("retained%d failed records", failed)
	}
	if fits != 4032 {
		t.Fatal("unexpected fits", fits)
	}
	t.Logf("2688 publication records; %d fits; source immutability PASS", fits)
}
