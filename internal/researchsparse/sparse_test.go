package researchsparse

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type golden struct {
	Weights []float64
	Rows    []struct {
		Query       string
		Event       model.Event
		Features    [Dimension]float64
		Probability float64
	}
}

func fixture(t testing.TB) golden {
	t.Helper()
	b, e := os.ReadFile("../../research/public-task-pilot/sparse-go-golden.json")
	if e != nil {
		t.Fatal(e)
	}
	var g golden
	if e = json.Unmarshal(b, &g); e != nil {
		t.Fatal(e)
	}
	return g
}
func TestPythonParity(t *testing.T) {
	g := fixture(t)
	f, e := New(g.Weights)
	if e != nil {
		t.Fatal(e)
	}
	for i, r := range g.Rows {
		x, e := Extract(r.Query, r.Event)
		if e != nil {
			t.Fatal(e)
		}
		for k, v := range x.Values() {
			if math.Abs(v-r.Features[k]) > 1e-14 {
				t.Fatalf("feature mismatch row%d index%d", i, k)
			}
		}
		p, e := f.Score(x)
		if e != nil || math.Abs(p-r.Probability) > 1e-12 {
			t.Fatalf("probability mismatch row%d %v", i, e)
		}
		r.Event.Content = "irrelevant external content"
		r.Event.ID = "different-id"
		r.Event.Attributes = map[string]string{"answer": "not a feature"}
		x2, e := Extract(r.Query, r.Event)
		if e != nil || x != x2 {
			t.Fatal("metadata changed features")
		}
	}
	x, _ := Extract(g.Rows[0].Query, g.Rows[0].Event)
	before, _ := f.Score(x)
	g.Weights[0] = 999
	after, _ := f.Score(x)
	if before != after {
		t.Fatal("caller mutated model")
	}
}
func TestValidation(t *testing.T) {
	for _, q := range []string{"", "a the", strings.Repeat("x", 1025), "non-ascii é"} {
		if _, e := Extract(q, model.Event{}); e == nil {
			t.Fatal("invalid query accepted")
		}
	}
	if _, e := Extract("query", model.Event{What: model.Field{Value: strings.Repeat("a", 2049)}}); e == nil {
		t.Fatal("oversized field")
	}
	if _, e := New(nil); e == nil {
		t.Fatal("invalid weights")
	}
	w := make([]float64, Dimension)
	w[1] = math.NaN()
	if _, e := New(w); e == nil {
		t.Fatal("NaN weights")
	}
	x, _ := Extract("query", model.Event{})
	if _, e := (Frozen{}).Score(x); e == nil {
		t.Fatal("uninitialized model")
	}
	w[1] = 0
	f, _ := New(w)
	if _, e := f.Score(Features{}); e == nil {
		t.Fatal("uninitialized features")
	}
}

var benchScore float64

func BenchmarkFrozenScore(b *testing.B) {
	g := fixture(b)
	f, _ := New(g.Weights)
	x, _ := Extract(g.Rows[0].Query, g.Rows[0].Event)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchScore, _ = f.Score(x)
	}
}
func BenchmarkExtractScoreFrontier200(b *testing.B) {
	g := fixture(b)
	f, _ := New(g.Weights)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 200; j++ {
			r := g.Rows[j%len(g.Rows)]
			x, e := Extract(r.Query, r.Event)
			if e != nil {
				b.Fatal(e)
			}
			benchScore, e = f.Score(x)
			if e != nil {
				b.Fatal(e)
			}
		}
	}
}
