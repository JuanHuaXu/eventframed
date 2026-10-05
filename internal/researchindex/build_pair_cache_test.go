package researchindex

import (
	"math"
	"sync"
	"testing"
)

func TestBuildPairCachePreservesBitsDirectionAndCapacity(t *testing.T) {
	f := func(a, b uint32) float32 { return math.Float32frombits(a ^ (b << 3)) }
	c, err := NewBuildPairCache(4)
	if err != nil {
		t.Fatal(err)
	}
	for round := 0; round < 5; round++ {
		for a := uint32(0); a < 10; a++ {
			for b := uint32(0); b < 10; b++ {
				if math.Float32bits(c.Distance(a, b, func() float32 { return f(a, b) })) != math.Float32bits(f(a, b)) {
					t.Fatal("metric changed")
				}
			}
		}
	}
	s := c.Stats()
	if s.Entries != 4 || s.Hits == 0 || s.Hits+s.Misses != 500 {
		t.Fatal(s)
	}
	zero, _ := NewBuildPairCache(0)
	zero.Distance(1, 2, func() float32 { return f(1, 2) })
	zero.Distance(1, 2, func() float32 { return f(1, 2) })
	if s := zero.Stats(); s.Hits != 0 || s.Misses != 2 || s.Entries != 0 {
		t.Fatal(s)
	}
	for _, limit := range []int{-1, 65537} {
		if _, err := NewBuildPairCache(limit); err == nil {
			t.Fatal("unbounded cache accepted")
		}
	}
}

func TestBuildPairCacheConcurrentAndBuildIsolation(t *testing.T) {
	f := func(a, b uint32) float32 { return float32(a*100 + b) }
	c, _ := NewBuildPairCache(64)
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				a, b := uint32(i%10), uint32(i%7)
				if c.Distance(a, b, func() float32 { return f(a, b) }) != f(a, b) {
					t.Error("wrong distance")
				}
			}
		}()
	}
	wg.Wait()
	s := c.Stats()
	if s.Entries > 64 || s.Hits+s.Misses != 8000 {
		t.Fatal(s)
	}
	other, _ := NewBuildPairCache(64)
	if other.Distance(1, 2, func() float32 { return f(1, 2) + 1 }) == c.Distance(1, 2, func() float32 { return f(1, 2) }) {
		t.Fatal("separate builds shared state")
	}
}
