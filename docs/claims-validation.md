# Whitepaper claims validation

Functional tests establish implementation invariants. Claims validation asks
whether those mechanisms improve untouched chronological outcomes. A passing
unit test is not empirical support for a whitepaper claim.

## Current evidence map

| Whitepaper target | Current status | Required next evidence |
| --- | --- | --- |
| Structured frames improve prediction or interpretability | Inconclusive: matched-ablation enforcement now exists, but the reserved block still has only 2 trajectories and no ratings | At least 3 new untouched trajectories plus blinded independent ratings |
| Residual caches improve held-out loss or cost | Old five-percent harmful-reuse proposition remains falsified; cumulative-harm-bounded abstaining replacement passed confirmation | Real drifting replay with regime labels and a frozen harm budget |
| Selective Bayesian updating preserves quality at bounded cost | Fixed 5% cheap-update policy remains falsified; frontier-all plus selective deep work exactly matched frontier-all quality in the synthetic fixture | Measure actual deep-work resource savings under real load |
| Anti-Pigeon prevents harmful posterior sharing | Practical-equivalence replacement passed confirmation with 61/64 compatible shares and no wrong terminal decisions; Anti-Pigeon authority remained external | Sparse, dependent, multi-member and drifting group confirmation |
| Omitted-influence certificates attain coverage | Synthetic finite-population coverage passed 256/256; runtime estimator exists, but its mean UCB was too conservative for a typical 0.05 limit | Larger and real audit populations, tighter valid bounds, adversarial omission tests |
| Changepoint invalidation adapts under drift | Mixed across seeds: prior v6 confirmation passed, but rescue-v1 confirmation exceeded the gradual unmatched-alarm ceiling by `0.003125` | Treat the monitor as not robustly validated; add independent real replay and an ensemble detector before retuning |
| Predictive snapping improves untouched outcomes | Wiring defect repaired: graph state changes nominated-candidate features and rank deltas with exact rollback; outcome improvement remains untested | Untouched chronological snap-versus-no-snap forecast comparison |
| Fast/slow separation preserves low latency | Ten-way real-contract run: p50 76.9 ms, p95 112.4 ms, p99 606.1 ms; sub-100 ms p95 not validated | Component profiles, fresh-sidecar concurrency sweep, mixed traffic, larger persisted corpora, soak |
| Priority weighting changes deployment value | Inconclusive: a separate deployment-constraint gate now prevents aggregate utility from hiding high-priority harm, but no new real evidence exists | At least 3 new untouched trajectories with frozen pre-outcome priorities |
| Belief-conditioned post-retrieval deltas improve useful ranking | Validated on a reserved 3-trajectory, 138-case Codex block | Larger independent confirmation and task-success labels |

The dated performance baseline is in
`docs/benchmarks/2026-08-28-apple-m4.md`.
The additional residual, Anti-Pigeon, and changepoint results are in
`docs/experiments/2026-08-28-additional-claims-v1.md`.
The upgraded Bayesian grouping and v4 changepoint confirmation are in
`docs/experiments/2026-08-28-bayesian-upgrade-v4.md`.
The corrected post-retrieval ranking confirmation is in
`docs/experiments/codex-postrank-confirmation-2026-08-29/RESULTS.md`. The first
concurrent real-contract latency result is in
`docs/benchmarks/2026-08-29-concurrent-contract.md`; it does not validate a
sub-100 ms p95 claim under ten-way parallelism.

The completed remaining-claims sweep, including retained failures, is in
`docs/experiments/claims-completion-2026-08-29/RESULTS.md`.

Future synthetic runs use schema v5. Reports embed post-v4 acceptance criteria,
fixed-sample Wilson intervals, alarm and delay denominators, distinct
design/confirmation seed provenance, and the group-authority integration
control. See `docs/experiments/claims-experiment-v5-protocol.md`. These additions
improve auditability; they do not promote the v4 mechanisms to production-ready
claims.

## Evaluation contract

`eventframe-eval` consumes one JSON dataset using schema
`eventframe.evaluation.v1`. Every case declares:

- a trajectory and strictly chronological prediction time;
- an outcome availability time strictly after prediction;
- a priority fixed before the outcome;
- one common candidate universe and relevant-event set;
- the same predeclared policy family, including the baseline;
- each policy's ranked probability forecast, nomination, and activation state;
- source availability for every candidate and the state snapshot time.

The evaluator rejects future-visible candidates, future state, duplicate or
changing universes, activation outside nomination, non-chronological trajectory
records, and confirmation predictions inside the declared post-freeze embargo.

It reports retrieval-usefulness Brier loss, ten-bin calibration error,
recall@10, recall@50, MRR, nomination recall, activation rate, high-priority miss
rate, priority-weighted loss and recall, paired gain against the baseline, and a
trajectory-cluster bootstrap interval for priority-weighted Brier gain.

This is the chatbot retrieval-usefulness specialization. It does not establish
the whitepaper's complete marked next-event/no-event proper-score claim. That
requires a later evaluator over the full horizon-indexed event law.

Run an evaluation:

```sh
go run ./cmd/eventframe-eval -input /absolute/path/to/evaluation.json
```

## Evidence discipline

Design and confirmation datasets must be stored separately. Thresholds,
priority weights, candidate policies, preprocessing, and bootstrap procedure are
frozen before confirmation. Confirmation is run once for a declared release;
failed or unfavorable results remain in the report. Synthetic fixtures validate
the evaluator and known mechanisms but are never labeled as real-world evidence.
