# Synthetic Bidirectional Reranking Results

Date: 2026-08-29

This experiment tests whether EventFrame can alter the final packed top 10 when the
answer is present in the 50-item recall frontier but initially ranked outside the
packet. It also tests the reverse operation: removing a historically harmful item
from the initial top 10 without disturbing useful controls.

The experiment exercises the Go `Service` path through recall, certificate gates,
Bayesian outcome updates, posterior prediction, rank-delta application, final
ranking, and packing. A deterministic controlled rank contract replaces only the
remote LibraVDB ranker so that initial ranks and score margins are known exactly.
It therefore tests the EventFrame reranking mechanism, not retrieval recall,
natural-language relevance, remote database latency, or production traffic.

## Frozen protocol

- Recall frontier: 50 candidates.
- Packed packet: 10 candidates.
- Evidence: 16 full-stream outcome observations for each designated candidate.
- Design seed: `29082901`.
- Untouched confirmation seed: `29082902`.
- Recoverable cases use narrow score margins inside the configured correction
  envelope.
- Negative controls use wide score margins outside that envelope.
- Pass-through and active arms receive the same candidates and contract scores.

See `PREDECLARED_PROTOCOL.md` for the frozen thresholds and decision rules.

## Results

| Criterion | Design result | Confirmation result | Decision |
| --- | ---: | ---: | --- |
| Promote useful item from ranks 11/15/25/40/50 | 50/50, 100%; lower 95% CI 92.87% | 200/200, 100%; lower 95% CI 98.12% | Validated in the synthetic mechanism test |
| Remove harmful item from ranks 1/3/5/7/10 | 50/50, 100%; lower 95% CI 92.87% | 200/200, 100%; lower 95% CI 98.12% | Validated in the synthetic mechanism test |
| Perform both operations in the same packet | 50/50, 100%; lower 95% CI 92.87% | 200/200, 100%; lower 95% CI 98.12% | Validated in the synthetic mechanism test |
| Retain useful controls initially at ranks 1/5/10 | 30/30, 100%; lower 95% CI 88.65% | 90/90, 100%; lower 95% CI 95.91% | Confirmation passes; design sample alone was inconclusive |
| Promote target across a deliberately wide score margin | 0/10, 0% | 0/50, 0%; upper 95% CI 7.13% | Bounded-envelope negative control behaves as declared |
| Improve packed useful-item precision | 3.33% to 8.89% | 2.65% to 8.53% | Validated for this constructed corpus |
| Mean unnecessary packet churn | 0 | 0 | Pass |
| In-process active recall p99 below 100 ms | 0.825 ms | 0.837 ms | Pass under the synthetic local setup |

All five target-rank strata and all five harmful-rank strata produced 40/40
successful promotions and demotions in confirmation. The upper and lower Wilson
intervals are stored in `report.json`.

## Interpretation

The confirmation run shows that EventFrame can rescue an answer from anywhere in
the bounded 50-item frontier, including rank 50, while simultaneously ejecting an
item at any tested top-10 rank. It did not damage any of the 90 useful controls.
This directly confirms that packing at `k=10` need not discard a useful candidate
before EventFrame's learned correction sees it, provided recall supplies a larger
frontier and the required score movement lies inside the declared envelope.

The wide-margin controls are equally important. EventFrame did not promote any of
those 50 targets, so this experiment does not support a claim that posterior
correction can overcome arbitrary retrieval-score gaps. Such misses should be
classified as outside-envelope failures, not as evidence that reranking never ran.

These data do not establish production answer quality or generalization to organic
OpenClaw sessions. Those require blinded natural or adversarial queries with
independent relevance labels and the real LibraVDB service. The sub-millisecond
latencies measure in-process Go execution with an in-memory store and controlled
ranker; they are mechanism overhead measurements, not end-to-end production
latency.

## Reproduction

```sh
go run ./cmd/eventframe-rerank-experiment \
  -dataset docs/experiments/synthetic-bidirectional-rerank-2026-08-29/dataset.json \
  -report docs/experiments/synthetic-bidirectional-rerank-2026-08-29/report.json
```

The complete generated corpus is in `dataset.json`; paired per-case packet outcomes,
rank strata, intervals, and latency distributions are in `report.json`.
