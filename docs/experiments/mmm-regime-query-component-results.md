# Joint regime-query component results

Status: **mathematical and replay contracts PASS; efficacy untested**. This is
not a rescue of the seven research directions or a production inference path.

The [protocol](mmm-regime-query-component-protocol.md) is implemented in the
research-only `regime_query_research_test.go` and `regime_query_tape_test.go`.
No existing learner or daemon default changed. The retained63-label support
reserves room for a query's hypothetical64th label. Its outcome is conditioned
at its actual past origin under the same full segmentation prior; older evidence
is not evicted. The failed one-change alternative is not used.

## What passed

- [Core race log](mmm-regime-query-contracts.txt):96 probe/query combinations
  checked against exhaustive partition enumeration with direct Beta integrals.
  This covers every four-frame availability pattern with at least one unknown
  outcome and three probes per query. Query masses, conditional forecasts,
  total probability, total expectation, and the two Brier-risk formulas agree.
- The full63-plus-one fixture preserves every original evidence origin and
  satisfies the probability/risk identities. Invalid cap, origin, input, probe
  and cancellation controls pass. Caller mutation cannot alter owned state.
- Four concurrent read-only queries reproduce the serial result under race.
  Core package test time4.430s.
- [Replay race log](mmm-regime-query-tape-contracts.txt):12 clock160 views span
  stationary and both transition cases, both phases and both delivery schedules.
  Six delayed query values reproduce after poisoning all excluded Y, every Q,
  and inputs strictly after the snapshot. Immediate complete delivery has an
  empty query pool. Input ownership and63-label support checks pass.49.639s.

The exhaustive numerical reference is tiny-history coverage. Full-cap tests
check joint-law identities and ownership, not independent enumeration of every
large-history partition. These are correctness checks, not evidence that the
model's query values predict real learning gains.

## Cost

[Three isolated benchmark repetitions](mmm-regime-query-benchmark.txt), AppleM4:

| Component | Time | Allocated bytes | Allocations |
| --- | ---: | ---: | ---: |
| Base63 fit | 39.57-39.70ms | about475184 | 9 |
| One query, two conditional refits | 81.29-81.64ms | about950632 | 28 |
| Eight-query pool, excluding base | 651.87-654.48ms | about7605064 | 224 |

A pool therefore costs roughly0.69s including its initial fit. These are local
component timings, not p95/p99, deadlines, persistence, full-request serving or
quality-adjusted acquisition cost. They do not support a sub100ms query-pool
claim. The mechanism remains slow-path-only and is not installed in serving.

## Next

The [batched joint-moment plan](mmm-regime-query-batch-plan.md) removes repeated
conditional refits algebraically, while retaining this slower implementation as
the numerical reference. Test equivalence before performance; preserve the same
63-label support,64th query slot, hazard, priors, origin semantics and probes.
Only then freeze an equal-cost full-stream acquisition experiment. Do not call
an optimization or successful risk identity a learning efficacy result.

All seven full goals remain open. No production, OpenClaw, dependencies, paper,
commit or push changes. All prior negative results remain available.
