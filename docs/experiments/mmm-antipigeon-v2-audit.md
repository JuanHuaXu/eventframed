# MMM + Anti-Pigeon v2 audit

## Confirmed and repaired

- Initial evaluation seed allocation reused design reference seeds as confirmation
  live seeds. Preserved that run as superseded and reran with disjoint split,
  scenario, stream and role ranges. A regression enumerates every seed and the
  fitting seed. No outcome-dependent parameter change accompanied the repair.

## Checks

- Unit coverage: divergent versus common revision, evidence eligibility, pending
  prediction, duplicate/stale feedback, next-step epoch publication, bounded audit
  buffer, and single original-dependency invalidation. Initial race tests passed.
- Every authoritative forecast's score, epoch transition, audit-only refit, and
  coordinate cap recomputed from preserved traces. All aggregate metrics and
  comparisons reproduced. Every stream and all six arms replayed deterministically
  with equality excluding measured fitting wall time. Source/protocol hashes match.
- Frozen fitting uses only earlier generator samples. Foreground observations
  precede outcomes; audits are nominated before either, observed after prediction,
  and can affect only subsequent predictions. No generator formula, shift time,
  or scenario label enters the controller/monitor.
- Existing Anti-Pigeon/changepoint primitives are reused; synthetic eligibility
  and an initially shared reliability group are fixture assumptions. No target-law
  diameter certificate or production publication is manufactured by the harness.
- Stationary reset failure retained as a negative result. Replay traced all 32
  stable confirmation first invalidations to instantaneous changepoint threshold
  crossings. No detector thresholds were patched to erase that finding.

## Limits

Fixed known binary views, full-stream correctness feedback, synthetic reference
context, one frozen fitting seed, only 32 independent streams per scenario,
approximate fixed-sample intervals, and no production service integration.
Relearning publication is research-only and not a validated safe replacement.
The steady benchmark excludes audit/reference acquisition, fitting, retrieval,
concurrency and durable persistence; it cannot establish serving p99.

## Final verification

Passed `go test -race` for observation, observationexperiment,
observationrescue, observationrescueexperiment, bayes and both experiment command
packages. This includes the full 320-stream deterministic evidence replay under
the race detector. `go vet` passed for both observation controllers, both
experiment packages and commands. Three serial microbenchmark repetitions are
preserved separately. No production package was modified or service restarted.
