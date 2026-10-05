# Four-Shard Quality Beyond Exact Fallback (DESIGN V2)

Freeze after V1 completed, before any V2 measurement. V1 remains unchanged.
V1 passed its finite quality, packet and cold-law screens at 1000 records, but
code inspection finds each shard <=400 records, satisfying HNSW searchExact's
size <= max(2*EfConstruction, EfSearch, k) branch. That is NOT large ANN evidence.

Repeat the SAME V1 experiment, screens, 96 public DESIGN questions, 128D SQ8,
3K nominations, frontier50/200, pack10, four phases and opposite paired orders,
with 4000 past replicas (then256future replicas). IDs remain local four-digit
keys. Read no held-out rows or expected-answer labels. Construction and exact
reference cost remain retained and separately timed. Raise test timeout to20m
and context to15m BEFORE collection, because corpus-dependent rebuild cost is
part of the tested mechanism; do not truncate work after seeing a slow result.

Require source-derived route counts for all four FNV shards to exceed600 before
serving, so neither nominal150 nor600 local request qualifies for the <=400/
<=600 exact fallback. Record the pinned HNSW source and these counts. Doubling
availability probes MAY subsequently force exact scans; do not claim every
search call stayed approximate without branch instrumentation. No branch or
quality-floor change, cache bypass, ANN effort reduction or private data.

All V1 screens remain unchanged: per-cell mean tie recall>=.95, exact-score
regret<=.005, paired tie recall difference>=-.02, packet score deficit<=.01,
strict repeat/counterfactual packet+frontier+forecast equality1e-9 and exact
reference stability. These are separate finite DESIGN outcomes, not revised
definition of any whole research goal. No learned posteriors/residuals active.
Preserve failures, even if future IDs are excluded and packed answers are good.

The exact-reference controls and evaluator are the same V1 specification,
adjusted only for the declared4000past limit. Public replicas are not4000
independent facts, and cold-law preservation is not calibration or agent utility.
Nomination instability is reported separately from retained frontier/packet
semantics. All seven WHOLE goals remain OPEN; no production or paper changes.

Reference: Malkov and Yashunin, https://arxiv.org/abs/1603.09320.
The tested fallback is implementation-specific, not a statement of that paper.
