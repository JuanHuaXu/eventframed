# Public-capture larger-frontier load preflight

Frozen before execution, 2026-10-05. This advances the workload surface of goal 6
without changing its completion requirements or earlier failed timing screens.

Baseline: `f3231fab2244c5c6bca8f7f822e5669ac58d13cd`. Candidate: the frozen mask
reuse patch from `frame-mask-cache-v1`, with matching manifest hash. Add the SAME
new test to both isolated checkouts; no production implementation changes.

## Data and chronology

Use only the 288 DESIGN captures from the checked-in public-fact text corpus,
whose full byte hash is `20f85fc6d35ed2ec0fa5d68ec66b0cdcb48229dd85b6d1a7616731a9e7356de7`.
No confirmation split, private conversation or agent-outcome labels are used.
Only nested capture envelopes are ingested. Source URLs/oracle fields are never
indexed. Question text may be used as a question; relevant/obsolete IDs and
expected answers are not consumed. Corpus facts are the existing public NASA,
NOAA, USGS, NIST and Smithsonian facts, not newly invented factual content.

This is an explicitly constructed timing workload, NOT historical query replay.
Create 1,000 dataset-local replicated capture envelopes with fresh local keys and
clock values before a fixed query cutoff. The number of rows is not independent
fact count. Preserve their raw user/assistant text. This timing corpus includes
all chosen design messages before its artificial query cutoff and cannot be
used to claim prediction accuracy or original-history recall.

Concurrent writes use CaptureTurn, not structured Observe: 256 future-dated
replicas, one-ms offered cadence with a serial writer. They must remain outside
every as-of frontier. Closed-loop writer rate and actual overlap are reported;
no open-loop overload guarantee is inferred. Temporary local LibraVDB with
128-dimensional hash embeddings and SQ8, plus a separate temporary SQLite
durable research learner ledger. No networked database or live OpenClaw.

## Serving and background mechanics

Test RecallK=50 and 200, PackK=10, TokenBudget=2000, 64 sequential recall calls
per fixture while the capture writer runs. Research tap capacity=4. Require the
observed pre-packing candidate count to equal the configured RecallK, and reject
any future-write ID in both frontier and packed packet. The packing reduction
is not mistaken for candidate truncation before scoring.

One synthetic usefulness label is offered per request, to a rotating frontier
member. Other members receive no invented labels; their ordinary retrieval and
Bayesian processing still run. The deterministic i%3 label is a worker-mechanics
probe, NOT public-fact truth, real agent utility or a learned-model-quality claim.
Guard issued predictions/labels against the actual service snapshot/journal;
wait for all 64 labels to be published before Close, and check all 128 durable
admit/feedback rows plus replay counts. Measure offered-before-guard to live
observed publication age, not post-return or after Close.

## Frozen timing screen

Two matched repetitions in control-candidate then candidate-control order,
for BOTH frontier sizes, each with fresh stores. All full Recall calls and
CaptureTurn calls are timed with monotonic time. Report p50/p95/p99/max, offered
label ages, capture bytes/count/overlap, failure/retry counts, init time and
correctness assertions. Store individual samples, not only headline prose.

Finite absolute screen: each fixture recall p99 <100 ms and offered-label age
p99 <250 ms. Paired non-regression: candidate/control recall p99 <=1.10 in each
frontier-size/repetition comparison. Functional failures or incomplete traces
cannot pass. No resampling, retuning or gate change after seeing results.
Keep failed cells and negative fixture outcomes. Quantiles of only 64 calls are
descriptive maximum-like statistics, NOT population-tail guarantees.

Still missing for whole goal 6: visible mutations and cross-epoch learning
transfer/recovery, sustained/open-loop backlog, actual contract-network paths,
untouched outcome-labeled agents and robust loaded guarantees. This larger
fixture does not cover millions of independent records or all evidence labels.
