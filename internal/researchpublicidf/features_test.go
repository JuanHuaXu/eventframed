package researchpublicidf

import (
	"context"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"math"
	"strings"
	"sync"
	"testing"
	"time"
)

func fixture(t *testing.T) (*Sources, time.Time, string) {
	t.Helper()
	clock := time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)
	epoch := strings.Repeat("a", 64)
	s, e := NewSources(context.Background(), []r.Source{{ID: "a", Title: "alpha rare", Body: "alpha alpha v2", AvailableAt: clock}, {ID: "b", Title: "alpha", Body: "common v3", AvailableAt: clock}}, clock, epoch)
	if e != nil {
		t.Fatal(e)
	}
	return s, clock, epoch
}

func TestIDFUnionAndAbsentTerms(t *testing.T) {
	s, _, epoch := fixture(t)
	if s.df["alpha"] != 2 || s.df["rare"] != 1 || s.n != 2 {
		t.Fatal("document union counted term occurrences")
	}
	v, e := s.Features(context.Background(), "alpha rare unseen", "a", epoch, 0)
	if e != nil {
		t.Fatal(e)
	}
	a, rare, absent := math.Log(1.2), math.Log(2), math.Log(6)
	for i, want := range []float64{(a + rare) / (a + rare + absent), a / (a + rare + absent), (a + rare) / (a + 2*rare + absent)} {
		if math.Abs(v[i]-want) > 1e-14 {
			t.Fatalf("feature %d: %g != %g", i, v[i], want)
		}
	}
	v, e = s.Features(context.Background(), "v2 v3", "a", epoch, 200)
	if e != nil || math.Abs(v[3]-.5) > 1e-14 || v[6] != 0 || v[7] != 0 {
		t.Fatalf("numeric/native: %v %v", v, e)
	}
}

func TestIDFSourceEpochFutureAndCancellation(t *testing.T) {
	s, clock, epoch := fixture(t)
	for _, tc := range []struct {
		q, id, epoch string
		rank         int
	}{{"alpha", "unknown", epoch, 0}, {"alpha", "a", strings.Repeat("b", 64), 0}, {"", "a", epoch, 0}, {"alpha", "a", epoch, 201}} {
		if _, e := s.Features(context.Background(), tc.q, tc.id, tc.epoch, tc.rank); e == nil {
			t.Fatal("invalid source context accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := s.Features(ctx, "alpha", "a", epoch, 0); e == nil {
		t.Fatal("cancel ignored")
	}
	if _, e := s.Features(nil, "alpha", "a", epoch, 0); e == nil {
		t.Fatal("nil context")
	}
	if _, e := (*Sources)(nil).Features(context.Background(), "alpha", "a", epoch, 0); e == nil {
		t.Fatal("nil source index")
	}
	if _, e := NewSources(context.Background(), []r.Source{{ID: "a", Title: "alpha", Body: "rare", AvailableAt: clock.Add(time.Second)}}, clock, epoch); e == nil {
		t.Fatal("future source allowed to affect df")
	}
	if _, e := NewSources(ctx, []r.Source{{ID: "a", Title: "alpha", Body: "rare", AvailableAt: clock}}, clock, epoch); e == nil {
		t.Fatal("cancelled build")
	}
}

func TestIDFImmutableConcurrentAndNativeInvariant(t *testing.T) {
	s, _, epoch := fixture(t)
	base, e := s.Features(context.Background(), "alpha rare", "a", epoch, 0)
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for n := 0; n < 64; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			v, e := s.Features(context.Background(), "alpha rare", "a", epoch, n)
			if e != nil || v != base {
				t.Errorf("native cue or concurrent state changed features: %v", e)
			}
		}(n)
	}
	wg.Wait()
	if base[0] != 1 || base[2] != 1 {
		t.Fatal("exact matches are not one")
	}
	if _, e := (r.Model{}).Score(r.Row{ID: "a", Features: base}); e != nil {
		t.Fatal(e)
	}
}
