# As-of guarded admission v23

Status: temporal admission partially rescued; overall performance NOT rescued.

The separate research-only guard checks committed native/publication agreement,
then complete owned future-ingestion history relative to the original query
time, while excluding mutations through validation. Exact guards remain exact.
It does not grant feedback-label or model-history authority.

Three repetitions under race passed primitive motion/input/wait/release checks,
memory/persistent service-bound admission and small persistent load accounting.
Negative controls cover equal-time ingestion, backfills, changed policy, unknown
history, bypassed backend mutation, quarantine, mismatched request time, pending
publication, missing deadline, cancellation and callback errors/panics. Vet passed.

## Frozen loaded comparison

Nine rotated arms, each with 192 recalls and 96 concurrent future writes, were
run against fresh persistent LibraVDB instances. There were no errors. Queued
exact controls rejected all 576 attempts as stale. As-of controls admitted 224
observations, validating 11,200 candidate records; 352 observations were dropped
by the bounded handoff. No as-of attempts were busy, stale or timed out.

| Trial | Accepted / 192 | Dropped | Read p99, off / as-of (ms) | Write p99, off / as-of (ms) | As-of guard p95 (ms) | Accepted-age p95 (ms) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 74 | 118 | 33.920 / 19.172 | 18.763 / 56.841 | 54.511 | 1142.768 |
| 1 | 76 | 116 | 30.950 / 25.854 | 18.226 / 59.168 | 49.453 | 1137.183 |
| 2 | 74 | 118 | 32.389 / 21.801 | 17.236 / 57.324 | 53.466 | 1166.714 |

Quantiles use nearest rank. Admission is 38.89% of submitted observations; all
attempted as-of callbacks complete, but 61.11% of observations never reach them.
The approximately 1.14-1.17 second age tails fail the prior 250ms screen. Lower
read p99 is not an overall acceleration: writes slow and observations are lost.
The callback's 50 separate full validations repeatedly read journal/query/event
inputs while holding the mutation gate. This is a candidate bottleneck, not yet
a measured causal attribution; profile or paired ablation before optimizing it.

No ledger writes, feedback, fitting or predictive accuracy were measured. These
figures cannot establish durable-learning throughput or rescue MMM's unresolved
accuracy/calibration claims. Next investigate bounded batch validation with the
same per-record authority checks, rather than relaxing gates or enlarging queues.

## Audit artifact

`mmm-asof-load-v23.jsonl` has all nine distinct cells, correct read/write counts,
conserved attempt/drop outcomes, validated=50*accepted, and verified embedded
source hashes. All ledger counters are zero. The Go test's PASS is an accounting
result, not an overall research acceptance decision.

SHA-256: `8fddc8534e584a269ed4ac27b43dd87095cda8b3261b82ab92b8989134e971b8`.
No production configuration, deployment or repository push was performed.
