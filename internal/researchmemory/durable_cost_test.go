package researchmemory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type durableCostResult struct {
	Labels, Trial             int
	Durable                   bool
	AdmitNS, FeedbackNS       []int64
	TotalNS, RestartNS, Bytes int64
}

func durableCostArm(t *testing.T, n, trial int, durable bool) durableCostResult {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "cost.sqlite")
	r := durableCostResult{Labels: n, Trial: trial, Durable: durable}
	var d *Durable
	var b *Background
	var e error
	if durable {
		d, e = OpenDurable(ctx, path, "fixture", "cost", 1, 42)
		if e == nil {
			b = d.worker
		}
	} else {
		b, e = NewBackground(1, 42, 256)
	}
	if e != nil {
		t.Fatal(e)
	}
	started := time.Now()
	for i := 1; i <= n; i++ {
		at := time.Now()
		var p Prediction
		if durable {
			p, _, e = d.Admit(ctx, uint64(i), uint16(i%512), .6, durableFixtureTime)
		} else {
			p, e = b.Predict(uint16(i%512), .6, 1, durableFixtureTime)
		}
		r.AdmitNS = append(r.AdmitNS, time.Since(at).Nanoseconds())
		if e != nil {
			t.Fatal(e)
		}
		at = time.Now()
		if durable {
			_, e = d.Feedback(ctx, p.ID, i%3 != 0, durableFixtureTime)
		} else {
			e = b.Feedback(p.ID, i%3 != 0, 1, durableFixtureTime)
		}
		r.FeedbackNS = append(r.FeedbackNS, time.Since(at).Nanoseconds())
		if e != nil {
			t.Fatal(e)
		}
		if i%64 == 0 {
			wait, cancel := context.WithTimeout(ctx, 5*time.Second)
			e = b.WaitProcessed(wait, uint64(i))
			cancel()
			if e != nil {
				t.Fatal(e)
			}
		}
	}
	r.TotalNS = time.Since(started).Nanoseconds()
	if count, failed, pending, queued := b.Counts(); count != uint64(n) || failed != 0 || pending != 0 || queued != 0 {
		t.Fatal(count, failed, pending, queued)
	}
	original := b.Snapshot()
	if durable {
		if e = d.Close(); e != nil {
			t.Fatal(e)
		}
		stat, e := os.Stat(path)
		if e != nil {
			t.Fatal(e)
		}
		r.Bytes = stat.Size()
		start := time.Now()
		reopened, e := OpenDurable(ctx, path, "fixture", "cost", 1, 42)
		r.RestartNS = time.Since(start).Nanoseconds()
		if e != nil {
			t.Fatal(e)
		}
		defer reopened.Close()
		for x := uint16(0); x < 512; x++ {
			want, e := original.Score(x, .6, 1, durableFixtureTime)
			got, f := reopened.worker.Snapshot().Score(x, .6, 1, durableFixtureTime)
			if e != nil || f != nil || want != got {
				t.Fatal("restart differs", x, want, got, e, f)
			}
		}
	} else {
		b.Close()
	}
	return r
}

func TestDurableCostStudy(t *testing.T) {
	path := os.Getenv("EVENTFRAME_DURABLE_COST_ARTIFACT")
	if path == "" {
		t.Skip("opt-in cost measurement")
	}
	f, e := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	for _, p := range []string{"internal/researchmemory/durable_cost_test.go", "internal/researchmemory/durable.go", "internal/researchmemory/background.go", "internal/researchmemory/completion.go", "internal/researchmemory/adapter.go", "internal/researchmemory/record.go", "internal/researchmemory/replay.go", "internal/researchledger/ledger.go", "docs/experiments/mmm-durable-cost-v10-protocol.md"} {
		b, e := os.ReadFile("../../" + p)
		if e != nil {
			t.Fatal(e)
		}
		sources[p] = string(b)
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	enc := json.NewEncoder(f)
	if e = enc.Encode(map[string]any{"Sources": sources, "Hashes": hashes, "expected_arms": 12}); e != nil {
		t.Fatal(e)
	}
	for _, n := range []int{256, 1024} {
		for trial := 0; trial < 3; trial++ {
			for order := 0; order < 2; order++ {
				r := durableCostArm(t, n, trial, (trial+order)%2 == 1)
				if e = enc.Encode(r); e != nil {
					t.Fatal(e)
				}
				if e = f.Sync(); e != nil {
					t.Fatal(e)
				}
				t.Logf("n%d trial%d durable%t total_ms%.3f restart_ms%.3f bytes%d", n, trial, r.Durable, float64(r.TotalNS)/1e6, float64(r.RestartNS)/1e6, r.Bytes)
			}
		}
	}
}
