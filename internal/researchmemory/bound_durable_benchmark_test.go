package researchmemory

import (
	"context"
	"testing"
)

func BenchmarkBoundDurableEpochHandoff(b *testing.B) {
	ctx := context.Background()
	labels, target, cutoff := boundTransferFixture(64, true, false)
	key := make([]byte, 32)
	key[0] = 9
	build := func() *SealedBoundBootstrap {
		bootstrap, retained, err := RebuildSealedBoundLabels(ctx, target, "tenant", "bench-stream", 2, 42, cutoff, labels, retainSourceB, "bench-key", key)
		if err != nil || retained != 32 {
			b.Fatalf("rebuild retained=%d err=%v", retained, err)
		}
		return bootstrap
	}
	b.Run("sealed_rebuild_64", func(b *testing.B) {
		for range b.N {
			_ = build()
		}
	})
	b.Run("bound_reopen_empty", func(b *testing.B) {
		path := b.TempDir() + "/bound.sqlite"
		validate := func(context.Context, RecordedPrediction) error { return nil }
		first, err := OpenBoundDurable(ctx, path, build(), validate)
		if err != nil {
			b.Fatal(err)
		}
		if err = first.Close(); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for range b.N {
			worker, err := OpenBoundDurable(ctx, path, build(), validate)
			if err != nil {
				b.Fatal(err)
			}
			if err = worker.Close(); err != nil {
				b.Fatal(err)
			}
		}
	})
}
