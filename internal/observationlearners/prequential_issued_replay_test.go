package observationlearners

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"testing"
)

type issuedReplayResult struct {
	Phase, Case, Index, Schedule, Checks, FitLabels int
	MaxError                                        float64
	PoisonChecked                                   bool
	Error                                           string `json:",omitempty"`
}

func replayIssued(input softV120Record) issuedReplayResult {
	r := issuedReplayResult{Phase: input.Phase, Case: input.Case, Index: input.Index, Schedule: input.Schedule}
	h := make([]segmentPacket, 144)
	for i := range h {
		j := i - 16
		if j < 0 {
			h[i] = segmentPacket{Bits: input.Initial[i].Bits, Outcome: input.Initial[i].Outcome, Arrives: 0}
		} else {
			s := input.Steps[j]
			arrival := j + s.Delay
			if s.Missing {
				arrival = -1
			}
			h[i] = segmentPacket{Bits: s.X, Outcome: s.Y, Arrives: arrival}
		}
	}
	m, err := fitSegmentPosteriorBatch(context.Background(), -16, 128, h, 64, .01, .95)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.FitLabels = len(m.origins)
	for j := 128; j < 160; j++ {
		s := input.Steps[j]
		if s.X >= 512 {
			r.Error = "invalid query input"
			return r
		}
		delta := math.Abs(m.predictions[s.X] - s.P[10])
		if math.IsNaN(delta) || delta > 1e-10 {
			r.Error = "issued forecast mismatch"
			return r
		}
		r.MaxError = math.Max(r.MaxError, delta)
		r.Checks++
	}
	if input.Phase == 0 && input.Index == 0 && (input.Case == 0 || input.Case == 5 || input.Case == 8 || input.Case == 12 || input.Case == 19 || input.Case == 20) {
		for i := range h {
			if h[i].Arrives < 0 || h[i].Arrives > 128 {
				h[i].Outcome = !h[i].Outcome
			}
		}
		other, err := fitSegmentPosteriorBatch(context.Background(), -16, 128, h, 64, .01, .95)
		if err != nil {
			r.Error = err.Error()
			return r
		}
		if m.predictions != other.predictions {
			r.Error = "excluded label influenced fit"
			return r
		}
		r.PoisonChecked = true
	}
	return r
}
func TestPrequentialIssuedReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit source")
	}
	full := os.Getenv("EVENTFRAME_PREQUENTIAL_FULL") == "1"
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	hash := sha256.New()
	d := json.NewDecoder(io.TeeReader(f, hash))
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Version != "soft-learners-v120" {
		t.Fatal("version")
	}
	type job struct {
		i     int
		input softV120Record
	}
	jobs := make(chan job, 8)
	results := make([]issuedReplayResult, 2688)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				results[j.i] = replayIssued(j.input)
			}
		}()
	}
	count, sourceCount := 0, 0
	var decodeErr error
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			decodeErr = err
			break
		}
		sourceCount++
		fixture := input.Phase == 0 && input.Index == 0 && (input.Case == 0 || input.Case == 5 || input.Case == 8 || input.Case == 12 || input.Case == 19 || input.Case == 20)
		if !full && !fixture {
			continue
		}
		if count >= len(results) {
			decodeErr = fmt.Errorf("extra source")
			break
		}
		jobs <- job{count, input}
		count++
	}
	close(jobs)
	wg.Wait()
	if decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if sourceCount != 2688 || (full && count != 2688) || (!full && count != 12) {
		t.Fatal("counts", sourceCount, count)
	}
	results = results[:count]
	inputHash := fmt.Sprintf("%x", hash.Sum(nil))
	if inputHash != "5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f" {
		t.Fatal("source hash")
	}
	checks, poison, failed := 0, 0, 0
	maxError := 0.
	seen := map[[4]int]bool{}
	for _, r := range results {
		key := [4]int{r.Phase, r.Case, r.Index, r.Schedule}
		if seen[key] {
			t.Fatal("duplicate")
		}
		seen[key] = true
		if r.Error != "" {
			failed++
		}
		checks += r.Checks
		if r.PoisonChecked {
			poison++
		}
		maxError = math.Max(maxError, r.MaxError)
	}
	hashes := map[string]string{}
	for _, p := range []string{"prequential_issued_replay_test.go", "soft_learners_v120_test.go", "segment_posterior.go", "segment_batch_research_test.go", "../../docs/experiments/mmm-prequential-projection-protocol.md"} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		hashes[p] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	if output := os.Getenv("EVENTFRAME_PREQUENTIAL_REPLAY_OUTPUT"); output != "" {
		out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer out.Close()
		if err := json.NewEncoder(out).Encode(struct {
			Version, InputSHA256 string
			Full                 bool
			Checks, Poison       int
			MaxError             float64
			Hashes               map[string]string
			Records              []issuedReplayResult
		}{"prequential-issued-v1", inputHash, full, checks, poison, maxError, hashes, results}); err != nil {
			t.Fatal(err)
		}
		if err := out.Sync(); err != nil {
			t.Fatal(err)
		}
	}
	if failed > 0 || checks != count*32 || poison != 12 {
		t.Fatal("failed checks", failed, checks, poison)
	}
	t.Logf("%d issued forecasts; %d poisoned fits; maximum discrepancy %.3g", checks, poison, maxError)
}
