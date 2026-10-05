package researchindex

import (
	"errors"
	"math"
	"sync/atomic"
)

type pairCounters struct {
	hits, misses atomic.Uint64
	padding      [48]byte
}

// BuildPairTable is a build-local, direct-mapped cache. A single atomic word
// publishes both the full directed key and exact float32 bits: collisions may
// evict, but cannot produce torn key/value hits. Inputs outside its 16-bit ID
// encoding fall back to the original metric. Never share across vector epochs.
type BuildPairTable struct {
	slots    []atomic.Uint64
	counters [64]pairCounters
}

func NewBuildPairTable(size int) (*BuildPairTable, error) {
	if size < 1 || size > 65536 || size&(size-1) != 0 {
		return nil, errors.New("pair table requires bounded power-of-two size")
	}
	return &BuildPairTable{slots: make([]atomic.Uint64, size)}, nil
}
func pairTag(a, b uint32) (uint32, bool) {
	if a > 65535 || b > 65535 {
		return 0, false
	}
	key := a<<16 | b
	if key == math.MaxUint32 {
		return 0, false
	}
	return key + 1, true // zero remains an unoccupied slot, including for pair(0,0).
}
func pairSlot(tag uint32, mask int) int {
	tag ^= tag >> 16
	tag *= 0x7feb352d
	tag ^= tag >> 15
	return int(tag) & mask
}
func (t *BuildPairTable) Lookup(a, b uint32) (float32, bool) {
	tag, valid := pairTag(a, b)
	slot := pairSlot(tag, len(t.slots)-1)
	counter := &t.counters[slot&63]
	if valid {
		word := t.slots[slot].Load()
		if uint32(word>>32) == tag {
			counter.hits.Add(1)
			return math.Float32frombits(uint32(word)), true
		}
	}
	counter.misses.Add(1)
	return 0, false
}
func (t *BuildPairTable) Store(a, b uint32, value float32) {
	tag, valid := pairTag(a, b)
	if !valid {
		return
	}
	t.slots[pairSlot(tag, len(t.slots)-1)].Store(uint64(tag)<<32 | uint64(math.Float32bits(value)))
}
func (t *BuildPairTable) Stats() BuildPairStats {
	var s BuildPairStats
	for i := range t.counters {
		s.Hits += t.counters[i].hits.Load()
		s.Misses += t.counters[i].misses.Load()
	}
	for i := range t.slots {
		if t.slots[i].Load() != 0 {
			s.Entries++
		}
	}
	return s
}
