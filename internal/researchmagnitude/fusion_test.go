package researchmagnitude

import (
	"context"
	"math"
	"reflect"
	"testing"
)

func TestEndpointsGapsAndOwnership(t *testing.T) {
	x := []float64{.4, .5, 1}
	for _, w := range []float64{0, .25, .5, .75, 1} {
		out, err := Scores(context.Background(), x, w)
		if err != nil {
			t.Fatal(err)
		}
		for i, v := range out {
			want := (1-w)/float64(i+1) + w*x[i]
			if math.Abs(v-want) > 1e-15 {
				t.Fatal("formula", i, v, want)
			}
		}
		out[0] = 0
		if !reflect.DeepEqual(x, []float64{.4, .5, 1}) {
			t.Fatal("borrowed input")
		}
	}
	a, _ := Scores(context.Background(), []float64{.8, .9}, .75)
	b, _ := Scores(context.Background(), []float64{.1, .9}, .75)
	if !(a[0] > a[1] && b[1] > b[0]) {
		t.Fatal("discarded lexical gap")
	}
}

func TestInvalidAndCancelledReject(t *testing.T) {
	for _, w := range []float64{-1, 2, math.NaN(), math.Inf(1)} {
		if v, e := Scores(context.Background(), []float64{.1}, w); e == nil || v != nil {
			t.Fatal("bad weight")
		}
	}
	for _, x := range [][]float64{{math.NaN()}, {math.Inf(1)}, {-.1}, {1.1}, make([]float64, 201)} {
		if v, e := Scores(context.Background(), x, .5); e == nil || v != nil {
			t.Fatal("bad coverage")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if v, e := Scores(ctx, nil, .5); e == nil || v != nil {
		t.Fatal("cancelled empty call")
	}
	if v, e := Scores(context.Background(), make([]float64, 200), .5); e != nil || len(v) != 200 {
		t.Fatal("valid cap")
	}
}

func BenchmarkMagnitude200(b *testing.B) {
	x := make([]float64, 200)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, e := Scores(context.Background(), x, .5); e != nil {
			b.Fatal(e)
		}
	}
}
