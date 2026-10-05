package researchdispersion

import (
	"math"
	"reflect"
	"testing"
)

func TestInvalidNormalizationIsAtomicV34(t *testing.T) {
	m, err := New(bases(11), "adaptive")
	if err != nil {
		t.Fatal(err)
	}
	// Inject impossible internal state; public inputs cannot produce it.
	m.logs[0] = math.Inf(1)
	before := *m
	before.n = append([]uint16(nil), m.n...)
	before.success = append([]uint16(nil), m.success...)
	if err := m.Observe(0, 1, true); err == nil || !reflect.DeepEqual(before, *m) {
		t.Fatal("invalid normalization mutated state")
	}
}
