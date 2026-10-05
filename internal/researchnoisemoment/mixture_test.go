package researchnoisemoment

import (
	"math"
	"testing"
)

func configTest() Config {
	return Config{Family: "rich", Prior: "moment", Strength: 2, Hazard: 1. / 16, Shared: true}
}

func TestMixtureReceiptIdentityAndFencing(t *testing.T) {
	m, e := NewMixture([]float64{.3, .9}, 1, 128, configTest())
	if e != nil {
		t.Fatal(e)
	}
	a, _ := m.Issue(0, 1)
	b, _ := m.Issue(1, 2)
	bad := a
	bad.children[1] = b.children[1]
	if _, e = m.Resolve(bad, true, 3); e == nil || m.Pending() != 2 || m.poison {
		t.Fatal("cross-child identity preflight")
	}
	q := a.Forecast()
	a.q = math.NaN()
	a.member = 1
	a.ordinal = 64
	a.at = 99
	r, e := m.Resolve(a, true, 3)
	if e != nil || r.Forecast != q || r.Member != 0 || r.TrialOrdinal != 1 || r.IssuedAt != 1 {
		t.Fatal("untrusted receipt fields", r, e)
	}
	if _, e = m.Resolve(a, false, 4); e == nil {
		t.Fatal("replay")
	}
	// A valid ticket can still encounter numerical failure after an earlier child
	// commits. The whole bundle must become unusable, not serve a mixed state.
	m.children[1].prior[Atoms*MaxFamilies] = math.NaN()
	if _, e = m.Resolve(b, true, 4); e == nil || !m.poison {
		t.Fatal("partial resolve fence")
	}
	if _, e = m.Predict(0); e == nil {
		t.Fatal("poison forecast")
	}
	if _, e = m.Weights(); e == nil {
		t.Fatal("poison weights")
	}
	if e = m.BeginEpoch(2, 5); e != nil || m.Pending() != 0 {
		t.Fatal("epoch repair", e)
	}
	if _, e = m.Resolve(b, true, 5); e == nil {
		t.Fatal("old epoch")
	}
}

func TestMixtureCapsCancellationAndPartialIssue(t *testing.T) {
	m, _ := NewMixture([]float64{.3, .9}, 1, 2, configTest())
	other, _ := NewMixture([]float64{.3, .9}, 1, 2, configTest())
	a, _ := m.Issue(0, 1)
	b, _ := m.Issue(1, 2)
	if _, e := other.Resolve(a, true, 2); e == nil {
		t.Fatal("owner")
	}
	if _, e := m.Resolve(a, true, 1); e == nil {
		t.Fatal("clock")
	}
	if _, e := m.Issue(0, 3); e == nil || m.Pending() != 2 {
		t.Fatal("cap")
	}
	if e := m.Cancel(a, 3); e != nil {
		t.Fatal(e)
	}
	if _, e := m.Resolve(a, true, 3); e == nil {
		t.Fatal("cancel")
	}
	if e := m.Cancel(b, 4); e != nil {
		t.Fatal(e)
	}
	if e := m.BeginEpoch(1, 4); e == nil {
		t.Fatal("epoch monotonicity")
	}
	m.children[1].issued[0] = MaxTrials
	if _, e := m.Issue(0, 5); e == nil || !m.poison {
		t.Fatal("partial issue fence")
	}
	if e := m.Cancel(b, 6); e == nil {
		t.Fatal("poison cancel")
	}
	if e := m.BeginEpoch(2, 6); e != nil {
		t.Fatal(e)
	}
	for j := 0; j < MaxTrials; j++ {
		if _, e := m.Issue(0, int64(7+j)); e != nil {
			t.Fatal(e)
		}
		m.capForTest(128)
	}
	if _, e := m.Issue(0, 72); e == nil {
		t.Fatal("history cap")
	}
}

// Change only the pending cap, never outcomes or model parameters, to exercise
// the independent per-member history cap after the two-slot cap check above.
func (m *Mixture) capForTest(n int) {
	for _, c := range m.children {
		c.cap = n
	}
}

func TestNoiseDomainAndEvidence(t *testing.T) {
	for _, eta := range []float64{-.1, .5, 1, math.NaN(), math.Inf(1)} {
		c := configTest()
		c.Noise = eta
		if _, e := NewMemoV49([]float64{.3, .9}, 1, 128, c); e == nil {
			t.Fatal("noise domain")
		}
	}
	c := configTest()
	c.Shared = false
	m, e := NewMemoV49([]float64{.3, .9}, 1, 128, c)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = m.LogEvidence(); e == nil {
		t.Fatal("private evidence mislabeled shared")
	}
	if _, e = NewMixture([]float64{.3, .9}, 1, 128, c); e == nil {
		t.Fatal("private mixture")
	}
	c.Shared = true
	c.Noise = .1
	if _, e = NewMixture([]float64{.3, .9}, 1, 128, c); e == nil {
		t.Fatal("implicit grid")
	}
}
