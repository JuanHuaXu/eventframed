package observationlearners

import "testing"

func TestHazardJournalPointMass(t *testing.T) {
	a, err := newObservationExperts(jointFixture(true, false))
	if err != nil {
		t.Fatal(err)
	}
	b, err := newObservationExperts(jointFixture(true, true))
	if err != nil {
		t.Fatal(err)
	}
	g, ref := newHazardAdviceJournal(), newMarkovAdviceJournal()
	var point [hazardRateCount]float64
	point[1] = 1
	g.filter, err = newHazardAdvice(point)
	if err != nil {
		t.Fatal(err)
	}
	for i := uint64(0); i < 256; i++ {
		if err := g.setClock(i); err != nil {
			t.Fatal(err)
		}
		if err := ref.setClock(i); err != nil {
			t.Fatal(err)
		}
		if i%32 == 0 {
			e := a
			if i/32%2 == 1 {
				e = b
			}
			if err := g.publish(e); err != nil {
				t.Fatal(err)
			}
			if err := ref.publish(e); err != nil {
				t.Fatal(err)
			}
		}
		x, err := g.predict(i, &jointReader{x: uint16(i), epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		y, err := ref.predict(i, &jointReader{x: uint16(i), epoch: 1}, 1)
		if err != nil {
			t.Fatal(err)
		}
		compareJointTrace(t, x, y)
		if err := g.deliver(i, i%3 == 0); err != nil {
			t.Fatal(err)
		}
		if err := ref.deliver(i, i%3 == 0); err != nil {
			t.Fatal(err)
		}
		gw, rw := g.advice.weights(), ref.advice.weights()
		for j := range gw {
			adviceNear(t, gw[j], rw[j])
		}
	}
}
