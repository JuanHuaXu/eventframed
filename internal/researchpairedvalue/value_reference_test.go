package researchpairedvalue_test

import (
	"math"
	"testing"

	ref "github.com/JuanHuaXu/eventframed/internal/researchpairedref"
	p "github.com/JuanHuaXu/eventframed/internal/researchpairedvalue"
	valueRef "github.com/JuanHuaXu/eventframed/internal/researchpairedvalueref"
)

type valueEvent struct {
	kind, member, ordinal int
	y                     bool
}

// Rebuild from revealed history, not the implementation's hypothetical rows.
func valueReference(t *testing.T, base []float64, cfg p.Config, history []valueEvent, extra *valueEvent) *ref.Reference {
	t.Helper()
	r, e := ref.New(base, cfg)
	if e != nil {
		t.Fatal(e)
	}
	apply := func(a valueEvent) {
		var e error
		switch a.kind {
		case 0:
			e = r.Issue(a.member)
		case 1:
			e = r.Observe(a.member, a.ordinal, 1, a.y)
		case 2:
			e = r.Audit(a.member, a.ordinal)
		case 3:
			e = r.Observe(a.member, a.ordinal, 2, a.y)
		}
		if e != nil {
			t.Fatal(e)
		}
	}
	for _, a := range history {
		apply(a)
	}
	if extra != nil {
		apply(valueEvent{kind: 2, member: extra.member, ordinal: extra.ordinal})
		apply(*extra)
	}
	return r
}

func TestPredictiveValueIndependentJointHistories(t *testing.T) {
	base := []float64{.25, .47, .78, .925}
	weights := []float64{.1, .2, .3, .4}
	positive, cases := 0, 0
	for _, hazard := range []float64{0, 1. / 16, .3} {
		for pattern := 0; pattern < 3; pattern++ {
			cfg := p.Config{Strength: 2, Hazard: hazard}
			m, e := p.New(base, 1, 256, cfg)
			if e != nil {
				t.Fatal(e)
			}
			history := []valueEvent{}
			tickets := make([]p.Ticket, 24)
			for k := range tickets {
				tickets[k], e = m.Issue(k%4, int64(k))
				if e != nil {
					t.Fatal(e)
				}
				history = append(history, valueEvent{kind: 0, member: k % 4})
			}
			at := int64(24)
			for step := range tickets {
				k := (step*7 + 23) % len(tickets)
				if k == 5 { // canceled and unrevealed, not an observed negative
					if e = m.Cancel(tickets[k], at); e != nil {
						t.Fatal(e)
					}
				} else if k != 9 { // one still-pending first measurement
					y := (k+pattern)%5 < 2
					if _, e = m.Resolve(tickets[k], y, at); e != nil {
						t.Fatal(e)
					}
					history = append(history, valueEvent{kind: 1, member: k % 4, ordinal: k/4 + 1, y: y})
				}
				at++
			}
			for _, k := range []int{19, 22} {
				second, e := m.RequestAudit(tickets[k], at)
				if e != nil {
					t.Fatal(e)
				}
				history = append(history, valueEvent{kind: 2, member: k % 4, ordinal: k/4 + 1})
				at++
				if _, e = m.Resolve(second, k%2 == 0, at); e != nil {
					t.Fatal(e)
				}
				history = append(history, valueEvent{kind: 3, member: k % 4, ordinal: k/4 + 1, y: k%2 == 0})
				at++
			}
			indices := []int{0, 1, 2, 3, 12, 13, 14, 15}
			origins := make([]p.Ticket, len(indices))
			for j, k := range indices {
				origins[j] = tickets[k]
			}
			// At most one batch slot per member is a deliberate computational cap.
			for begin := 0; begin < len(origins); begin += len(base) {
				options, e := m.PredictionValues(origins[begin:begin+len(base)], weights)
				if e != nil {
					t.Fatal(e)
				}
				before := valueReference(t, base, cfg, history, nil)
				independent, e := valueRef.New(base, cfg)
				if e != nil {
					t.Fatal(e)
				}
				for _, event := range history {
					var e error
					switch event.kind {
					case 0:
						e = independent.Issue(event.member)
					case 1:
						e = independent.Observe(event.member, event.ordinal, 1, event.y)
					case 2:
						e = independent.Audit(event.member, event.ordinal)
					case 3:
						e = independent.Observe(event.member, event.ordinal, 2, event.y)
					}
					if e != nil {
						t.Fatal(e)
					}
				}
				for j, o := range options {
					k := indices[begin+j]
					a, e := before.Options(k%4, k/4+1)
					if e != nil {
						t.Fatal(e)
					}
					near(t, o.Observed, a.Observed)
					currentRisk, expectedRisk, movement := 0., 0., 0.
					tower := make([]float64, len(base))
					for branch := 0; branch < 2; branch++ {
						probability := a.Observed
						if branch == 1 {
							probability = 1 - probability
						}
						if probability == 0 {
							continue
						}
						r := valueReference(t, base, cfg, history, &valueEvent{kind: 3, member: k % 4, ordinal: k/4 + 1, y: branch == 0})
						for i := range base {
							q, _, e := before.Predict(i)
							if e != nil {
								t.Fatal(e)
							}
							conditional, _, e := r.Predict(i)
							if e != nil {
								t.Fatal(e)
							}
							tower[i] += probability * conditional
							expectedRisk += probability * weights[i] * conditional * (1 - conditional)
							movement += probability * weights[i] * math.Pow(conditional-q, 2)
							if branch == 0 {
								currentRisk += weights[i] * q * (1 - q)
							}
						}
					}
					near(t, o.Value, movement)
					near(t, o.Value, currentRisk-expectedRisk)
					probability, value, e := independent.PredictionValue(k%4, k/4+1, weights)
					if e != nil {
						t.Fatal(e)
					}
					near(t, probability, o.Observed)
					near(t, value, movement)
					for i := range base {
						q, _, _ := before.Predict(i)
						near(t, tower[i], q)
						served, _, _ := m.Predict(i)
						near(t, served, q)
					}
					if o.Value > 1e-8 {
						positive++
					}
					cases++
				}
			}
		}
	}
	if positive == 0 || cases != 72 {
		t.Fatalf("vacuous value audit: cases=%d positive=%d", cases, positive)
	}
	t.Logf("72 independent joint-history origins; %d positive values", positive)
}
