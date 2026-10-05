package observationlearners

import (
	"math"
	"reflect"
	"testing"
)

func TestSnapshotBankDetachedAndCoherent(t *testing.T) {
	b := newSnapshotBank()
	models := jointFixture(true, false)
	if err := b.publish(0, models); err != nil {
		t.Fatal(err)
	}
	s, err := b.snapshot()
	if err != nil {
		t.Fatal(err)
	}
	want, err := s.Forecast(511, 0)
	if err != nil {
		t.Fatal(err)
	}
	models[0].cells[partialIndex(511, 0)].weighted = .99
	if got, _ := s.Forecast(511, 0); got != want {
		t.Fatal("caller changed detached model")
	}
	for i := uint64(0); i < 32; i++ {
		law, _ := b.snapshot()
		if err := b.issue(i, law, 511, 0); err != nil {
			t.Fatal(err)
		}
		if err := b.filter.deliver(i, true); err != nil {
			t.Fatal(err)
		}
	}
	before := *b
	if err := b.publish(1, jointFixture(false, false)); err == nil {
		t.Fatal("changed measure accepted")
	}
	if !reflect.DeepEqual(*b, before) {
		t.Fatal("publication failure changed owner")
	}
	if err := b.publish(1, jointFixture(true, true)); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Forecast(511, 0); got != want {
		t.Fatal("publication changed old preview")
	}
	if err := b.issue(32, s, 511, 0); err == nil {
		t.Fatal("old preview issued")
	}
	law, _ := b.snapshot()
	// Independently marginalize all complete assignments for each partial view.
	for mask := uint16(0); mask < 512; mask++ {
		for value := mask; ; value = (value - 1) & mask {
			sum, n := 0., 0.
			for x := uint16(0); x < 512; x++ {
				if x&mask == value {
					p, err := law.Forecast(511, x)
					if err != nil {
						t.Fatal(err)
					}
					sum += p
					n++
				}
			}
			got, err := law.Forecast(mask, value)
			if err != nil || math.Abs(got-sum/n) > 1e-11 {
				t.Fatal("incoherent snapshot law", mask, value, err)
			}
			if value == 0 {
				break
			}
		}
	}
	if err := b.issue(32, law, 511, 0); err != nil {
		t.Fatal(err)
	}
	if err := b.filter.deliver(32, false); err != nil {
		t.Fatal(err)
	}
	before = *b
	if err := b.issue(33, law, 511, 0); err == nil {
		t.Fatal("pre-feedback preview issued")
	}
	if !reflect.DeepEqual(*b, before) {
		t.Fatal("stale preview changed owner")
	}
}

func TestSnapshotBankBadPublication(t *testing.T) {
	b := newSnapshotBank()
	before := *b
	if err := b.publish(0, [4]*ConditionalForest{}); err == nil {
		t.Fatal("nil models")
	}
	if !reflect.DeepEqual(*b, before) {
		t.Fatal("failed first publication committed")
	}
	if _, err := (snapshotLaw{}).Forecast(0, 0); err == nil {
		t.Fatal("zero law")
	}
	models := jointFixture(true, false)
	models[0].cells[0].weighted++
	if err := b.publish(0, models); err == nil {
		t.Fatal("incoherent model")
	}
	if !reflect.DeepEqual(*b, before) {
		t.Fatal("bad table changed owner")
	}
}
