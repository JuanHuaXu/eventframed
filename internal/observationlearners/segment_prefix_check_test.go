package observationlearners

import (
	"context"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"reflect"
	"testing"
	"unsafe"
)

func TestPrefixLikelihoodParity(t *testing.T) {
	ctx := context.Background()
	for _, n := range []int{0, 1, 16, 63} {
		for _, mass := range []float64{0, .95, 1} {
			prefix := ridgeTestSamples(n)
			p, e := prepareSegmentPrefix(ctx, prefix, mass)
			if e != nil {
				t.Fatal(e)
			}
			for _, next := range []observation.Sample{{Bits: 0, Outcome: false}, {Bits: 511, Outcome: true}, {Bits: 17, Outcome: true}} {
				samples := append(append([]observation.Sample(nil), prefix...), next)
				got, e := buildPrefixLikelihoods(ctx, samples, mass, p, nil)
				if e != nil {
					t.Fatal(e)
				}
				want, e := buildSegmentLikelihoodsTable(ctx, samples, mass)
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(got, want) {
					t.Fatal("prefix mismatch", n, mass, next)
				}
			}
			if n > 0 {
				changed := append(append([]observation.Sample(nil), prefix...), observation.Sample{})
				changed[0].Outcome = !changed[0].Outcome
				if _, e := buildPrefixLikelihoods(ctx, changed, mass, p, nil); e == nil {
					t.Fatal("changed prefix accepted")
				}
			}
		}
	}
	t.Logf("36 complete likelihood/tail parity fixtures; bytes per retained interval %d", unsafe.Sizeof(prefixInterval{}))
}

func BenchmarkPrefixLikelihood(b *testing.B) {
	samples := ridgeTestSamples(64)
	p, e := prepareSegmentPrefix(context.Background(), samples[:63], .95)
	if e != nil {
		b.Fatal(e)
	}
	for _, name := range []string{"full", "prepared_append", "prepare63"} {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var out *segmentLikelihoods
				var e error
				switch name {
				case "full":
					out, e = buildSegmentLikelihoodsTable(context.Background(), samples, .95)
				case "prepared_append":
					out, e = buildPrefixLikelihoods(context.Background(), samples, .95, p, nil)
				case "prepare63":
					_, e = prepareSegmentPrefix(context.Background(), samples[:63], .95)
				}
				if e != nil {
					b.Fatal(e)
				}
				segmentBenchmarkLikelihoods = out
			}
		})
	}
}
