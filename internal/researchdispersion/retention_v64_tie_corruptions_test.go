package researchdispersion

import (
	"encoding/json"
	"testing"
)

func TestRetentionV64TieCorruptions(t *testing.T) {
	p := makePairedV60(2026105407, 0, 14, 0)
	a, e := runRetentionV64(p, "random", "fixed150")
	if e != nil {
		t.Fatal(e)
	}
	scoreWindowV37(p.World, &a.windowArmV37)
	if e = independentRetentionV64TieAudit(p, a); e != nil {
		t.Fatal(e)
	}
	clone := func() armPairedV60 {
		b, _ := json.Marshal(a)
		var out armPairedV60
		if e = json.Unmarshal(b, &out); e != nil {
			t.Fatal(e)
		}
		return out
	}
	for _, field := range []string{"issue", "observed", "receipt", "request", "choice", "snapshot", "metric", "pending", "cost"} {
		x := clone()
		switch field {
		case "issue":
			x.Issued[0] += .1
		case "observed":
			x.ObservedIssued[0] += .1
		case "receipt":
			x.Receipts[0].ArrivedAt++
		case "request":
			x.AuditOutcomes[0].Forecast += .1
		case "choice":
			x.Decisions[0].Members[0] = x.Decisions[0].Members[1]
		case "snapshot":
			x.Snapshots[0].Forecast[0] += .1
		case "metric":
			x.Recovery++
		case "pending":
			x.Snapshots[0].Pending++
		case "cost":
			x.Breakdown.AccountedNS++
		}
		if e = independentRetentionV64TieAudit(p, x); e == nil {
			t.Fatal("corruption accepted", field)
		}
	}
	q := p
	q.Second = append([]bool(nil), p.Second...)
	for k := 1800; k < 2400; k++ {
		q.Second[k] = !q.Second[k]
	}
	z, e := runRetentionV64(q, "random", "fixed150")
	if e != nil {
		t.Fatal(e)
	}
	changed := false
	for k := range a.Issued {
		if k <= 1800 && a.Issued[k] != z.Issued[k] {
			t.Fatal("future leaked", k)
		}
		if k > 1800 && a.Issued[k] != z.Issued[k] {
			changed = true
		}
	}
	if !changed {
		t.Fatal("vacuous future fork")
	}
	t.Log("9 corrupted fields rejected; future fork differs only after visible evidence")
}
