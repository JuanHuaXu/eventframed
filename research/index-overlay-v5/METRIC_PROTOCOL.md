# Resolved Metric Correction Before Timing

The original score-before/full-ordinary/full-race traces do NOT prove the claimed
canonical metric defect: the test incorrectly called scalar CosineDistance_func
instead of the Collection's runtime GetDistanceFunc (NEON on this machine).
59/320 differences chiefly measure two reduction orders. Preserve these outputs
as a harness error and withdraw that count as valid pre-patch evidence.

Inspecting the resolved functions exposes the relevant mixed path: base HNSW
uses the runtime SIMD metric while overlay scoring explicitly uses the scalar
function. Fresh corrected regression uses the EXACT runtime metric, includes
one overlay event alongside80base events, and checks81records per query. Run
it in a fresh diagnostic V4 copy and current V5 BEFORE prospective source/timing
pins. Retain independent failing and passing witnesses. No scientific criterion
changes: canonical consistency and existing1e-9past/future equality still required.
Do not call this exact real arithmetic or independent FP64 oracle agreement.
