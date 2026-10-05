# Priority-aware packing composition

A confirmed remaining failure: enabling packing's diversity selector undoes a
plain temporal permutation because that selector compares numeric scores again.
A two-record regression reproduces the high-score contradicted record winning
despite the non-contradicted record being placed first.

Research PackPriority now delegates to the existing packing implementation while
keeping priority-only values local to selection. For a mixed partition it adds3
to local copies of non-contradicted scores. With original scores in[0,1] and
diversity/priority penalties in[0,1], this preserves class priority during the
greedy diversity step. It restores complete original candidate records before
returning, so selection values cannot become forecast scores or confidence.
Unsupported penalties, malformed plans and incomplete identities are rejected.

Adaptive expansion is decided using the original candidates/scores FIRST. The
selection pass then disables only re-evaluation of that expansion decision,
not the requested diversity policy. Token limits, evidence occupancy and
cross-candidate correlation checks remain delegated to packing.Select over the
whole frontier. No per-class independent packing that could bypass cross-class
duplicate suppression is used. A compatible record over budget is still skipped.
All-contradicted/unknown-only cases preserve existing behavior.

## Checks

Race-enabled tests pass:

-Diversity permutation counterexample reproduced; adapter restores intended order
 and original output scores/records.
-Over-budget compatible candidate cannot bypass token budget.
-Same evidence group across classes is still suppressed; suppression count kept.
-Adaptive expansion matches original-score decision, without artificial gap input.
-All-contradicted fallback matches original packing behavior.
-Actual service applyRankDeltas followed by PlanPriority and PackPriority survives
 opposing0.25 corrections at2/50/200 with diversity enabled. Scores, proper laws
 and earlier modulation certainty remain unchanged.

```sh
go test -race ./internal/researchcalendar -run TestPriorityPacking -v
go test -race ./internal/service -run TestCalendarPriorityWithDiversityAfterDeltas -v
```

This remains test-composed, not an upgraded service.Recall implementation.
Caller must supply a plan for the exact current candidate ordering and snapshot;
the helper checks identities but does not establish snapshot freshness itself.
The separate plan's calibration status remains not_evaluated. No durable journal,
fresh public-data service replay or performance result is claimed. Adaptive mode
can invoke packing twice and its extra cost must be measured, not hidden.

Next integration must carry the typed priority/calibration sidecar through the
actual post-correction packing boundary and journal. Production remains untouched;
this closes a component-policy interaction, not any whole roadmap direction.
