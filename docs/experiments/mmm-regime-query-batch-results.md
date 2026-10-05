# Exact batched regime-query results

Status: **component equivalence and isolated speed gates PASS**. No learning
efficacy, production serving latency, or whole research direction is completed.

The [frozen plan](mmm-regime-query-batch-plan.md) is implemented in research-only
`regime_query_batch_test.go`. Shared interval posterior moments and query-tilted
prefix sums replace repeated hypothetical refits. The family prior, boundary
process,63-label support, real past query origin and next-clock probe law remain
unchanged. The slower conditioning implementation is retained as a reference.

## Equivalence

- [Core race checks](mmm-regime-query-batch-contracts.txt):32 tiny-history queries
  and96 probe comparisons agree with exhaustive partitions/direct Beta integrals.
  Empty-tail and same/different-cell covariance checks pass. Maximum discrepancy
  against explicit conditional refits1.28e-14.
- Two full63-label histories, including internal gaps, old queries and queries
  after the last label:16 queries and128 probe comparisons agree. Maximum
  discrepancy1.21e-13. Input ownership, hidden-outcome erasure, reversed pool
  ordering and winner consistency pass. Core package time12.152s.
- Invalid caps, duplicate/known/out-of-range query origins, probes and packets,
  cancellation and four detached concurrent batches pass under race.
- [Replay race checks](mmm-regime-query-batch-tape-contracts.txt):six stationary/
  transition views, both consumed phases, delayed schedule. All41 candidate
  queries (328 probe comparisons) agree with the slow reference. Maximum
  discrepancy2.56e-13; every selected origin agrees. Package60.371s.

The absolute numerical tolerance is1e-10. Test selection uses the lowest origin
within1e-10 of maximum utility to make negligible numerical ties explicit; this
is not a new statistical confidence threshold. Both pool orders obey that rule.
The full-history numerical comparison uses the slow reference, not independent
enumeration of all large-history partitions. The tiny-history check is exhaustive.

## Performance

Three paired isolated repetitions on AppleM4, history80, retained63 labels,
eight candidate queries and eight probes. Both arms include base evidence work:

| Metric | Reference: base plus16 refits | Batched joint moments |
| --- | ---: | ---: |
| Median time | 724.812ms | 46.837ms |
| Observed range | 690.917-736.644ms | 45.416-47.077ms |
| Allocated bytes per call, median | 8080248 | 350064 |
| Allocations per call, median | 233 | 98 |

Median speedup15.48x and allocation ratio.0433 pass the frozen4x/half-allocation
gate. This is about95.7% fewer allocated bytes. "Cold" means no borrowed fitted
state; shared immutable lookup tables may already be warm. It is not a cold
process, loaded p95/p99, durable transaction or complete agent-serving benchmark.

[Raw benchmark](mmm-regime-query-batch-benchmark.txt),
[machine-readable summary](mmm-regime-query-batch-benchmark-summary.json).
The summary pins benchmark/source/protocol hashes and replays byte-identically.
No claim that the remaining computation is free or suitable for the hot path.

## Remaining test

The algorithm still predicts query usefulness under a working model. Unknown
regime truth, natural feedback arriving before the paid answer, and support
changes at publication can invalidate that practical ranking. Next run the
[all-case one-decision diagnostic](mmm-regime-query-outcome-protocol.md), with
actual revealed labels changing the fitted law and random/entropy/no-query
controls. This is a precursor to the unchanged full-stream research requirements,
not a substitute for them. All seven whole directions remain open.

No production, OpenClaw, dependency, whitepaper, commit or push changes. Existing
tracked edits and all negative research evidence were preserved.
