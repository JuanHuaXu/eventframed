# Separate constraint priority: composition prototype

The margin adapter's large score gaps changed packet rank certainty even when
all evidence and forecasts were unchanged. A new research PlanPriority function
returns a permutation and typed date decisions instead of replacement scores.

Ordering contract: apply all numeric corrections first; compute the temporal
partition from original query and actual what-field content; apply its stable
permutation immediately before packing. No later numeric sort. Unknown evidence
stays non-contradicted, not certified correct. All-contradicted input preserves
original order with an explicit flag. Decision states are unknown, compatible
with the recognized date predicates, or contradicted by those predicates.
They are not factual-truth or source-authentication certificates.

The plan reports CalibrationStatus=not_evaluated. It never emits a synthetic
confidence percentage. Earlier elastic modulation retains its original numeric
boundary input. Reordering changes the served selection, so calibration for the
new packet still requires evidence; preserving individual forecast laws does
not supply it automatically.

## Test Evidence

Race-enabled tests pass at2/50/200 candidates with the same allowed opposing
0.25 correction fixtures used against the ordinal adapter:

-Actual service applyRankDeltas first applies corrections and sorts numerically.
-PlanPriority then returns a separate permutation; actual packing.Select packs1.
-The temporal winner is preserved at every size.
-Every numeric score and proper law is unchanged by priority projection.
-The earlier modulation certainty equals the original score-boundary value.
-Request binding, complete-frontier requirement, input immutability and
 unsupported/all-contradicted behavior are checked by the focused unit test.

```sh
go test -race ./internal/researchcalendar ./internal/service -run 'TestPriorityPlanDoesNotEncodeConfidence|TestCalendarPriorityAfterDeltas' -v
```

This is a TEST-COMPOSED path through real delta/packing functions, not a call to
an upgraded service.Recall implementation. The production service has no such
post-delta hook or calibration-status packet field yet. No learned correction
frequency, loaded performance, durable explanation storage, broad snapshot
integration or fresh accuracy validation is claimed. The test records are
injected perturbations, not newly learned outcomes.

## Integration Requirement

Prefer testing this separate priority channel over further inflating scores.
An actual service integration must execute it after all numeric/resolution
corrections and account explicitly for diversity packing; carry snapshot-bound
method/decision explanations and calibration availability in its output/journal;
and verify request isolation, stale-retry behavior and proper-law preservation.
It must not relabel a post-pack wrapper or an unconsumed sidecar as integration.
Production and frozen earlier implementations remain unchanged. Whole roadmap
goals are still open.
