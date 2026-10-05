# Research ranking integration

Status: optional in-process research hook, disabled by default. No production
deployment or public-task LLM runs are claimed by these checks.

The service calls the hook after backend ranking and resolution preference, but
before recall truncation and packing. The hook receives only bounded extracted
5W1H features and the pre-residual usefulness baseline, not text, IDs or labels.
The frontier is capped at 200. Each research probability contributes a correction
clipped to [-0.25, 0.25]; final ranking scores remain in [0, 1]. This is a declared
experimental conversion, not an empirically calibrated probability-to-rank map.
Retrieval scores, existing rank deltas and forecast laws are not rewritten.
ResearchRankDelta records the additional change. Packet confidence is cleared:
the old packing calibration is not valid for an experimental ordering.

Callbacks must be read-only, cancellation-aware and use a frozen model. The
service can retry stale snapshots, so callbacks must never fit or consume labels.
Frozen.Score creates no pending evidence records. Adapter fitting replaces model
instances, leaving already published models unchanged. Frozen models reject
queries before their latest admitted feedback or from another epoch.

These interfaces are not exposed through daemon flags or the plugin. A research
runner must explicitly configure a tenant and bind its frozen model epoch to the
experiment's declared snapshot. Feedback collection and publication remain
separate from the retrieval callback. This is not yet an integrated learned-law
serving path and must not be described as one.

## Checks on 2026-09-12

- Twenty-candidate direct boundary check: candidate 20 is promoted; all candidates
  are visible. This does not establish end-to-end top-ten answer improvement.
- Callback mutation of its copied baseline cannot alter the correction anchor.
- Wrong cardinality, NaN and out-of-range probabilities reject before mutation.
- Other tenants and cancelled calls do not invoke the callback.
- Actual Recall with an untrained frozen model preserves candidate output and
  receives all three nominated candidates before packing two.
- Published model predictions survive subsequent refits unchanged; repeated
  frozen reads create neither labels nor pending records.
- Targeted race tests for Frozen, ResearchRank and ResearchShadow passed.
- go vet for internal/researchmemory and internal/service passed.

Remaining: realistic candidate-volume end-to-end evaluation, trained-adapter
quality, epoch publication/retry coverage, correlated evidence, and load with
actual persistence and retrieval. Prepared public tasks remain unexecuted.
