package researchregimelogsummary

import (
	"fmt"
	"reflect"
	"testing"
)

func exactTestPolicy() AdmissionPolicy {
	zero := 0.
	// A synthetic contract witness tests wiring, not actual float certification.
	return AdmissionPolicy{.01, .02, &zero, "unit-test-assumption-only"}
}

func TestBoundedIssueChecksSelectedLawAndIsTransactional(t *testing.T) {
	m, _ := New([]float64{.3, .8}, Config{.25, .1, 9})
	before := m.Snapshot()
	if _, _, err := m.IssueBounded(m.token, 0, 0, exactTestPolicy()); err == nil {
		t.Fatal("aggressive truncation admitted")
	}
	if !reflect.DeepEqual(before, m.Snapshot()) {
		t.Fatal("rejected issue published state")
	}
	if _, _, err := m.IssueBounded(m.token, 0, 0, AdmissionPolicy{MaximumTV: 1, MaximumBrierShift: 1}); err == nil {
		t.Fatal("missing numerical budget admitted")
	}
	n, _ := New([]float64{.3, .8}, Config{.25, .1, 0})
	old := n.token
	r, gate, err := n.IssueBounded(old, 0, 0, exactTestPolicy())
	if err != nil || gate.TotalTV != 0 || gate.BrierShift != 0 || r.Token != n.token {
		t.Fatalf("exact-support issue %v %+v", err, gate)
	}
	if _, _, err = n.IssueBounded(old, 0, 1, exactTestPolicy()); err == nil {
		t.Fatal("stale budget law accepted")
	}
	if len(n.rows) != 1 {
		t.Fatal("stale call appended a row")
	}
}

func TestBoundedPendingUsesBranchEnvelopeAndRequiresBudget(t *testing.T) {
	m, _ := New([]float64{.3, .8}, Config{.25, .1, 0})
	r, _ := m.Issue(0, 0)
	before := m.Snapshot()
	q, err := m.PendingCleanBounded(r.Token, r.Ordinal, 1, exactTestPolicy())
	if err != nil {
		t.Fatal(err)
	}
	for _, branch := range q.Branches {
		c := branch.Admission
		if c.Token != r.Token || c.ModelTV != 0 || len(branch.NextClean) != 2 {
			t.Fatal("wrong branch certificate")
		}
	}
	if !reflect.DeepEqual(before, m.Snapshot()) {
		t.Fatal("counterfactual query published state")
	}
	if _, err = m.PendingCleanBounded(r.Token, r.Ordinal, 1, AdmissionPolicy{MaximumTV: 1, MaximumBrierShift: 1}); err == nil {
		t.Fatal("query ignored missing numerical witness")
	}
}

func TestRejectedIssueDoesNotCorruptWarmPrefixCache(t *testing.T) {
	m, _ := New([]float64{.3, .8}, Config{.25, .1, 18})
	for i := 0; i < 319; i++ {
		if _, err := m.Issue(i%2, int64(i)); err != nil {
			t.Fatal(err)
		}
	}
	before := m.Snapshot()
	cache := append([]prefix(nil), m.cache...)
	if _, _, err := m.IssueBounded(m.token, 0, 319, exactTestPolicy()); err == nil {
		t.Fatal("unsafe warm state admitted")
	}
	if !reflect.DeepEqual(before, m.Snapshot()) || !reflect.DeepEqual(cache, m.cache) {
		t.Fatal("rejected trial mutated live prefix cache")
	}
}

func BenchmarkBoundedIssue(b *testing.B) {
	for _, members := range []int{2, 150, 200} {
		for _, bounded := range []bool{false, true} {
			b.Run(fmt.Sprintf("M%d/bounded=%t", members, bounded), func(b *testing.B) {
				m := fixture(b, members, 64)
				// Budget one benchmarks admission mechanics, not usable fidelity.
				// The zero numerical witness is only a synthetic wiring assumption.
				zero := 0.
				policy := AdmissionPolicy{1, 1, &zero, "benchmark-assumption-only"}
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					trial := *m
					if bounded {
						r, _, err := trial.IssueBounded(trial.token, 0, 65, policy)
						if err != nil {
							b.Fatal(err)
						}
						benchReceipt = r
					} else {
						r, err := trial.Issue(0, 65)
						if err != nil {
							b.Fatal(err)
						}
						benchReceipt = r
					}
				}
			})
		}
	}
}
