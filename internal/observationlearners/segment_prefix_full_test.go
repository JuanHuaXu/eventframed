package observationlearners

import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"reflect"
	"sync"
	"testing"
)

// This helper uses reference eligibility only to prepare a fixture. The
// candidate independently recomputes and checks eligibility on every full fit.
func prefixForModel(t testing.TB, h []segmentPacket, m *segmentPosterior, mass float64) *segmentPrefix {
	t.Helper()
	samples := make([]observation.Sample, len(m.origins)-1)
	for i := range samples {
		p := h[m.origins[i]-m.left]
		samples[i] = observation.Sample{Bits: p.Bits, Outcome: p.Outcome}
	}
	p, e := prepareSegmentPrefix(context.Background(), samples, mass)
	if e != nil {
		t.Fatal(e)
	}
	return p
}

func TestPrefixFullChronology(t *testing.T) {
	count := 0
	for _, n := range []int{32, 64, 272} {
		for _, cap := range []int{16, 64} {
			for _, mass := range []float64{0, .95, 1} {
				for _, hazard := range []float64{.01, .5} {
					h := transformHistory(n)
					a, e := fitSegmentPosteriorTable(context.Background(), -16, n-16, h, cap, hazard, mass)
					if e != nil {
						t.Fatal(e)
					}
					p := prefixForModel(t, h, a, mass)
					b, e := fitSegmentPosteriorPrefix(context.Background(), p, -16, n-16, h, cap, hazard, mass)
					if e != nil {
						t.Fatal(e)
					}
					if !reflect.DeepEqual(a, b) {
						t.Fatal("full posterior mismatch", n, cap, mass, hazard)
					}
					count++
				}
			}
		}
	}
	t.Logf("%d full posterior pairs including chronology and hazard", count)
}

func TestPrefixOwnershipAndCancellation(t *testing.T) {
	ctx := context.Background()
	h := transformHistory(272)
	ref, e := fitSegmentPosteriorTable(ctx, -16, 256, h, 64, .01, .95)
	if e != nil {
		t.Fatal(e)
	}
	p := prefixForModel(t, h, ref, .95)
	saved := *p
	saved.samples = append([]observation.Sample(nil), p.samples...)
	saved.states = append([]prefixInterval(nil), p.states...)
	var models [4]*segmentPosterior
	var errs [4]error
	var wg sync.WaitGroup
	for i := range models {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			models[i], errs[i] = fitSegmentPosteriorPrefix(ctx, p, -16, 256, h, 64, .01, .95)
		}(i)
	}
	wg.Wait()
	for i, m := range models {
		if errs[i] != nil || !reflect.DeepEqual(m, ref) {
			t.Fatal("shared prefix", i, errs[i])
		}
	}
	cancelled := &checkpointContext{Context: ctx, limit: 10}
	if m, e := fitSegmentPosteriorPrefix(cancelled, p, -16, 256, h, 64, .01, .95); !errors.Is(e, context.Canceled) || m != nil {
		t.Fatal("cancelled output", e)
	}
	if !reflect.DeepEqual(*p, saved) {
		t.Fatal("prefix mutated")
	}
	changed := append([]segmentPacket(nil), h...)
	changed[ref.origins[0]+16].Outcome = !changed[ref.origins[0]+16].Outcome
	if _, e := fitSegmentPosteriorPrefix(ctx, p, -16, 256, changed, 64, .01, .95); e == nil {
		t.Fatal("changed old label reused")
	}
	changed = append([]segmentPacket(nil), h...)
	changed[ref.origins[0]+16].Arrives = 257
	if _, e := fitSegmentPosteriorPrefix(ctx, p, -16, 256, changed, 64, .01, .95); e == nil {
		t.Fatal("changed eligible set reused")
	}
	// Equivalent values at changed origins remain mathematically reusable only
	// because full origin/hazard chronology is recomputed, never cached here.
	recovered, e := fitSegmentPosteriorPrefix(ctx, p, -16, 256, h, 64, .01, .95)
	if e != nil || !reflect.DeepEqual(recovered, ref) {
		t.Fatal("recovery", e)
	}
}

func BenchmarkPrefixFull(b *testing.B) {
	h := transformHistory(272)
	ref, e := fitSegmentPosteriorTable(context.Background(), -16, 256, h, 64, .01, .95)
	if e != nil {
		b.Fatal(e)
	}
	p := prefixForModel(b, h, ref, .95)
	for _, name := range []string{"full", "prepared"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var m *segmentPosterior
				var e error
				if name == "full" {
					m, e = fitSegmentPosteriorTable(context.Background(), -16, 256, h, 64, .01, .95)
				} else {
					m, e = fitSegmentPosteriorPrefix(context.Background(), p, -16, 256, h, 64, .01, .95)
				}
				if e != nil {
					b.Fatal(e)
				}
				segmentBenchmarkModel = m
			}
		})
	}
}
