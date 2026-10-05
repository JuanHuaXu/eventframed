package researchindex

import (
	"math"
	"sync"
	"testing"
)

func TestBuildPairTableCollisionAndEncoding(t *testing.T) {
	c, e := NewBuildPairTable(1)
	if e != nil {
		t.Fatal(e)
	}
	if _, ok := c.Lookup(0, 0); ok {
		t.Fatal("empty slot hit")
	}
	for _, bits := range []uint32{0, 0x80000000, 0x7fc00001, 0x3f800000} {
		c.Store(0, 0, math.Float32frombits(bits))
		v, ok := c.Lookup(0, 0)
		if !ok || math.Float32bits(v) != bits {
			t.Fatal("bits changed")
		}
	}
	c.Store(1, 2, 3)
	if _, ok := c.Lookup(0, 0); ok {
		t.Fatal("collision false hit")
	}
	if _, ok := c.Lookup(2, 1); ok {
		t.Fatal("directed keys aliased")
	}
	for _, ids := range [][2]uint32{{65536, 1}, {1, 65536}, {65535, 65535}} {
		c.Store(ids[0], ids[1], 7)
		if _, ok := c.Lookup(ids[0], ids[1]); ok {
			t.Fatal("unsupported IDs aliased")
		}
	}
	for _, size := range []int{0, 3, 65537} {
		if _, e := NewBuildPairTable(size); e == nil {
			t.Fatal("bad size")
		}
	}
}
func TestBuildPairTableConcurrentNoFalseHits(t *testing.T) {
	c, _ := NewBuildPairTable(4)
	var wg sync.WaitGroup
	f := func(a, b uint32) float32 { return math.Float32frombits(a*15485863 + b) }
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				a, b := uint32(i%17), uint32(w)
				c.Store(a, b, f(a, b))
				if v, ok := c.Lookup(a, b); ok && math.Float32bits(v) != math.Float32bits(f(a, b)) {
					t.Error("torn or false hit")
				}
			}
		}(w)
	}
	wg.Wait()
	s := c.Stats()
	if s.Hits+s.Misses != 8000 || s.Entries > 4 {
		t.Fatal(s)
	}
}
func TestBuildPairTableNoPerLookupAllocation(t *testing.T) {
	c, _ := NewBuildPairTable(65536)
	if n := testing.AllocsPerRun(1000, func() { c.Store(1, 2, 3); c.Lookup(1, 2); c.Lookup(2, 3) }); n != 0 {
		t.Fatal("hot allocation", n)
	}
}
