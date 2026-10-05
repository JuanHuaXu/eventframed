package researchcalendar

import (
	"context"
	"reflect"
	"testing"
)

func TestPriorityPlanDoesNotEncodeConfidence(t *testing.T) {
	ctx, r := fixture("Which event before 2010?")
	before := append(r.Candidates[:0:0], r.Candidates...)
	p, err := PlanPriority(ctx, r)
	if err != nil || !reflect.DeepEqual(p.Order, []int{1, 0}) || p.CalibrationStatus != "not_evaluated" {
		t.Fatal(p, err)
	}
	if !reflect.DeepEqual(before, r.Candidates) {
		t.Fatal("input scores or data changed")
	}
	if p.Decisions[0].State != "contradicted" || p.Decisions[1].State != "compatible" {
		t.Fatal(p.Decisions)
	}
	for _, q := range []string{"Which event earlier than 2010?", "Which event before 2000?"} {
		ctx, r = fixture(q)
		p, err = PlanPriority(ctx, r)
		if err != nil || !reflect.DeepEqual(p.Order, []int{0, 1}) {
			t.Fatal(p, err)
		}
	}
	if !p.AllContradicted {
		t.Fatal("missing contradiction status")
	}
	if _, err = PlanPriority(context.Background(), r); err == nil {
		t.Fatal("missing binding accepted")
	}
	r.K2 = 1
	if _, err = PlanPriority(ctx, r); err == nil {
		t.Fatal("partial frontier accepted")
	}
}
