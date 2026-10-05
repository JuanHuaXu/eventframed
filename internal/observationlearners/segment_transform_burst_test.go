package observationlearners

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"
)

type transformJobTiming struct {
	ID           int     `json:"id"`
	QueueMS      float64 `json:"queue_ms"`
	ComputeMS    float64 `json:"compute_ms"`
	CompletionMS float64 `json:"completion_ms"`
}
type transformBurst struct {
	Trial     int                  `json:"trial"`
	Workers   int                  `json:"workers"`
	Transform bool                 `json:"transform"`
	Jobs      []transformJobTiming `json:"jobs"`
}

func TestTransformBurstExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_TRANSFORM_BURST_OUT")
	if path == "" {
		t.Skip("opt-in measurement")
	}
	old := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(old)
	history := transformHistory(272)
	ref, e := fitSegmentPosterior(-16, 256, history, 64, .01, .95)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := fitSegmentPosteriorTransform(-16, 256, history, 64, .01, .95); e != nil {
		t.Fatal(e)
	}
	var bursts []transformBurst
	for trial := 0; trial < 3; trial++ {
		for _, workers := range []int{1, 2, 4} {
			for arm := 0; arm < 2; arm++ {
				fast := (arm+trial)%2 == 1
				jobs := make(chan int, 12)
				timings := make([]transformJobTiming, 12)
				errs := make([]error, 12)
				models := make([]*segmentPosterior, 12)
				startGate := make(chan struct{})
				var wg sync.WaitGroup
				var submitted time.Time
				for w := 0; w < workers; w++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						<-startGate
						for id := range jobs {
							start := time.Now()
							if fast {
								models[id], errs[id] = fitSegmentPosteriorTransform(-16, 256, history, 64, .01, .95)
							} else {
								models[id], errs[id] = fitSegmentPosterior(-16, 256, history, 64, .01, .95)
							}
							end := time.Now()
							timings[id] = transformJobTiming{id, float64(start.Sub(submitted)) / 1e6, float64(end.Sub(start)) / 1e6, float64(end.Sub(submitted)) / 1e6}
						}
					}()
				}
				for id := 0; id < 12; id++ {
					jobs <- id
				}
				close(jobs)
				submitted = time.Now()
				close(startGate)
				wg.Wait()
				for id, m := range models {
					if errs[id] != nil || m == nil {
						t.Fatal("fit failed", id, errs[id])
					}
					for x, p := range ref.predictions {
						if d := math.Abs(p - m.predictions[x]); math.IsNaN(d) || d > 1e-12 {
							t.Fatal("parity", id, x)
						}
					}
				}
				bursts = append(bursts, transformBurst{trial, workers, fast, timings})
			}
		}
	}
	hashes := map[string]string{}
	for _, file := range []string{"segment_posterior.go", "segment_transform_research_test.go", "segment_transform_fit_test.go", "segment_transform_fullpair_test.go", "segment_transform_burst_test.go", "../../research/segment-transform-burst-protocol.md"} {
		data, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(data)
		hashes[file] = hex.EncodeToString(h[:])
	}
	artifact := struct {
		Go      string            `json:"go"`
		Procs   int               `json:"gomaxprocs"`
		Sources map[string]string `json:"sources"`
		Bursts  []transformBurst  `json:"bursts"`
	}{runtime.Version(), 4, hashes, bursts}
	data, e := json.MarshalIndent(artifact, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	_, writeErr := f.Write(append(data, '\n'))
	closeErr := f.Close()
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	t.Logf("recorded %d bursts / %d fits", len(bursts), len(bursts)*12)
}
