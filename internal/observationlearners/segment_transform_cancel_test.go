package observationlearners

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

// A deterministic context cancels at a chosen checkpoint, avoiding timing-based
// claims about which phase was interrupted. Fits access it on one goroutine.
type checkpointContext struct {
	context.Context
	checks, limit int
}

func (c *checkpointContext) Err() error {
	c.checks++
	if c.checks >= c.limit {
		return context.Canceled
	}
	return nil
}

func TestTransformCancellation(t *testing.T) {
	h := transformHistory(272)
	before := append([]segmentPacket(nil), h...)
	expected, e := fitSegmentPosteriorTransform(-16, 256, h, 64, .01, .95)
	if e != nil {
		t.Fatal(e)
	}
	active := &checkpointContext{Context: context.Background(), limit: 100000}
	got, e := fitSegmentPosteriorContext(active, -16, 256, h, 64, .01, .95)
	if e != nil || !reflect.DeepEqual(got, expected) {
		t.Fatal("uncancelled parity", e)
	}
	total := active.checks
	for _, limit := range []int{1, 2, 3, 10, 100, total - 2, total - 1, total} {
		ctx := &checkpointContext{Context: context.Background(), limit: limit}
		m, e := fitSegmentPosteriorContext(ctx, -16, 256, h, 64, .01, .95)
		if !errors.Is(e, context.Canceled) || m != nil {
			t.Fatal("partial model escaped", limit, e)
		}
	}
	if !reflect.DeepEqual(h, before) {
		t.Fatal("history mutation")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	m, e := fitSegmentPosteriorContext(ctx, -16, 256, h, 64, .01, .95)
	if !errors.Is(e, context.DeadlineExceeded) || m != nil {
		t.Fatal("deadline not honored", e)
	}
	recovered, e := fitSegmentPosteriorContext(context.Background(), -16, 256, h, 64, .01, .95)
	if e != nil || !reflect.DeepEqual(recovered, expected) {
		t.Fatal("cancel poisoned next fit", e)
	}
	t.Logf("checked %d checkpoints, eight deterministic cutoffs, actual deadline and recovery", total)
}

func BenchmarkTransformContextPair(b *testing.B) {
	h := transformHistory(272)
	for _, name := range []string{"without_context", "active_deadline"} {
		b.Run(name, func(b *testing.B) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var m *segmentPosterior
				var e error
				if name == "without_context" {
					m, e = fitSegmentPosteriorTransform(-16, 256, h, 64, .01, .95)
				} else {
					m, e = fitSegmentPosteriorContext(ctx, -16, 256, h, 64, .01, .95)
				}
				if e != nil {
					b.Fatal(e)
				}
				segmentBenchmarkModel = m
			}
		})
	}
}
