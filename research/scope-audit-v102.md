# Scope audit during the v102 replication

This is a completion-boundary audit, not a new experimental result. Its initial
version was written before inspecting v102 quality outcomes; later revisions
corrected coverage of the historical v79/v80 gate results. The preceding v101 turn made
progress through implementation, a complete fresh experiment, replay and a
verified negative result. None of the seven directions is complete or exhausted.

The requirements below come from the numbered sections of
[research-direction.md](../research-direction.md), not from whatever the latest
small-domain experiment happens to pass. Historical reports were inspected for
this audit; their tests were not all rerun here and are not new current-runtime
certificates.

| Direction | Required evidence | Inspected evidence and remaining gap |
| --- | --- | --- |
| 1. Evaluation | Recovery and stationary protection across generators, fitting samples and feedback schedules | v101 preserves protection but fails two recovery intervals. v102 tests precision only. [v9 dependent-input transfer](../docs/experiments/mmm-dependent-v9-results.md) and [v88 breadth](../docs/experiments/mmm-age-breadth-v88-results.md) retain failures. Fresh seeds alone are not independent generator validation. |
| 2. Challenger windows | Faster recovery at equal evidence volume without stationary harm | [v101 blocks](../docs/experiments/mmm-routing-v101-results.md) show a useful short-window interval followed by short-window harm. Routing mainly reduces initial overconfidence. Adaptive fitting/window decisions and reliable transfer remain unproven. |
| 3. Anti-Pigeon gate | Faster split detection without sacrificing error control or deadline reliability | The earlier [gate-tail v2](../docs/experiments/mmm-gate-tail-v2-results.md) failed, but the later [v79 evidence mixture](../docs/experiments/mmm-rate-mixture-v79-results.md) passed its finite gate screen. [v80 member integration](../docs/experiments/mmm-member-integration-v80-results.md) reduced split delay by35% on confirmation while failing forecast improvement; its2% false-revocation target remained inconclusive. Aggregate-mean and version-scoped forecast tests do not certify target-law diameter. |
| 4. Challenger learning | Better shifted accuracy with measured bounded memory/update cost | [v83 retained-subset breadth](../docs/experiments/mmm-subset-breadth-v83-results.md) passed its immediate-feedback screen, but [v88](../docs/experiments/mmm-age-breadth-v88-results.md) fails combined-stress parity and [v101](../docs/experiments/mmm-routing-v101-results.md) remains weak in one switch direction. Small conditional/Boolean models are executable; their95% stationary ceiling is not general agent accuracy. |
| 5. Prospective tasks | Better answers or retrieval on untouched agent tasks with honest outcome labels | [Generation status](public-task-pilot/GENERATION_STATUS.md) records zero actual generation calls and an unresolved model/endpoint choice. Prepared public requests and mocked transport are not an agent experiment or untouched domain confirmation. |
| 6. Shadow serving | Useful results before staleness without harming loaded serving tails | [Exclusive-ledger v69](../docs/experiments/mmm-exclusive-ledger-v69-results.md) fails its speed rescue. Arithmetic microseconds and offline fitting fixtures do not certify durable concurrent serving, p95/p99 or production readiness. |
| 7. Falsification observation | Faster learning than random/uncertainty acquisition at equal total cost | [Stopping v10](hypothesis-observation/STOPPING_RESULTS.md) saves reports but fails accuracy protection. Known hypothesis models, provenance misspecification and omitted evidence remain unresolved. Forecast-weight routing does not itself choose better observations. |

## What a v102 pass would authorize

Only a statement that the unchanged routing candidate passed the declared
finite-scenario screen on one larger, independently generated run. Preserve
all earlier failures. It would justify the next integration/transfer experiment,
not completion of the roadmap, publication of general learning claims or
production promotion. Partial masks63 and full training feedback are not the
adaptive six-coordinate MMM acquisition loop.

## What a v102 failure would require

Keep the same success criteria and stop the proposed sample-size ladder after
this one replication. Distinguish insufficient mean benefit, an interval still
crossing zero, and new protection failures. Use block diagnostics only to choose
the next separately frozen mechanism experiment, never to retrofit an oracle
schedule or score a selectively favorable window.

## Integration boundaries that still need explicit treatment

The [v85 delayed-feedback report](../docs/experiments/mmm-delayed-learning-v85-results.md)
distinguishes received training labels from stale selector feedback. The current
routing kernel has one outstanding forecast and a fixed immediate-feedback
publication clock. A delayed integration must define evidence ordering, version
ownership, censoring and the conditional null anew; it cannot silently apply
old-version credit to a new model or inherit an immediate-feedback martingale
claim. Missing labels must not become negatives. These are required contracts,
not optional language cleanup.

This distinction is consistent with the explicit delayed-feedback formulation
in [Joulani, Gyorgy and Szepesvari (2013)](https://proceedings.mlr.press/v28/joulani13.pdf),
Sections1-2: feedback carries the origin of the decision and may arrive out of
order; non-delayed learning guarantees do not automatically cover that process.
Their regret reductions are not a theorem validating our version-scoped
falsification or evidence-routing policy.

The goal remains active. No whole-project blocker is declared while these
research leads remain available; the generation-model choice is a direction5
boundary, not permission to install a model or discover credentials.

## Subsequent v102 outcome

The [completed replication](../docs/experiments/mmm-replication-v102-results.md)
passes its106 finite quality gates and four null screens, with complete replay.
That resolves the narrow precision test, not the requirements above. Next is
the [adaptive-observer integration](routed-observation-proposal.md); all seven
directions remain open.
