package researchmemory

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func BenchmarkMotionBoundRebuild64(b *testing.B) {
	labels, origin, cutoff := boundTransferFixture(64, true, false)
	target := origin
	target.RuntimeVersion++
	key := []byte("0123456789abcdef0123456789abcdef")
	retain := func(_ context.Context, _ model.Snapshot, label BoundLabel) (bool, error) {
		return label.Prediction.Binding.EventID == "source-b", nil
	}
	ctx := context.Background()
	b.ResetTimer()
	for range b.N {
		if _, _, err := RebuildMotionSealedBoundLabels(ctx, target, origin, "tenant", "motion-bench", 2, 42, cutoff, labels, retain, "motion-key", key); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMotionBoundOpen64(b *testing.B) {
	if b.N > 128 {
		b.Skip("bounded SQLite research benchmark: use -benchtime=100x")
	}
	labels, origin, cutoff := boundTransferFixture(64, true, false)
	target := origin
	target.RuntimeVersion++
	key := []byte("0123456789abcdef0123456789abcdef")
	retain := func(_ context.Context, _ model.Snapshot, label BoundLabel) (bool, error) {
		return label.Prediction.Binding.EventID == "source-b", nil
	}
	ctx := context.Background()
	path := filepath.Join(b.TempDir(), "motion-bench.sqlite")
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		bootstrap, _, err := RebuildMotionSealedBoundLabels(ctx, target, origin, "tenant", "motion-bench", 2, 42, cutoff, labels, retain, "motion-key", key)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		worker, err := OpenMotionBoundDurable(ctx, path, bootstrap, func(context.Context, model.Snapshot) error { return nil }, func(context.Context, RecordedPrediction) error { return nil })
		b.StopTimer()
		if err != nil {
			b.Fatal(err)
		}
		if err := worker.Close(); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
	}
}
