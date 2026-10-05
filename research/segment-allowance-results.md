# Budget-aware start refusal: partial improvement, screen fails

**FAIL the original52/64 completion requirement in all three pairs.** A65ms
remaining-budget requirement raises completions from1/64 to6/64 consistently.
This is a partial mechanism rescue, not a completed research direction.

| Trial | Baseline completed | Candidate completed | Explicit refusals | Candidate drops | Context interruptions before/after |
|---|---:|---:|---:|---:|---:|
|0|1|6|52|6|21/0|
|1|1|6|53|5|25/0|
|2|1|6|52|6|27/0|

All denominators64. Refusals are actual failed terminals, not removed jobs or
negative labels. Candidate has zero stale results; returned forecasts match
the reference. All16 writes overlapped readers; no request errors or unexpected
processor errors. Timing/terminal accounting verifies every offer.

Foreground p99 baseline/candidate(ms):28.577/23.001,30.064/31.663,
27.871/31.670. Trial2 fails1.10x latency non-harm. The other two passing cells
are finite measurements, not population p99 guarantees. No shadow-off comparison
is included in this diagnostic.

Total measured processor time baseline/candidate(ms):260.005/270.024,
268.643/277.302,284.159/276.085. The improvement is productive completion, not
a demonstrated reduction in total processor time. Refusing nearly expired jobs
allows later jobs with sufficient budget to run; the selected workload changes.
Do not claim that skipped evidence was learned or that selection bias vanished.

## Interpretation

The observed low-budget-start mechanism was actionable: fixed refusal avoids
wasting the worker on many fits that previously expired. It does not supply
enough capacity for64 distinct fresh fits under these deadlines.65ms is not a
certified runtime bound, and admission cannot guarantee completion under future
load or hardware variation. Failed/refused jobs still need honest pending or
unavailable outcomes in any larger design.

Raw JSON: `segment-allowance-results.json`. Verify source hashes, threshold
decisions, trace identities and verdict with
`node research/verify-segment-allowance.mjs` (passed:false).

Next work should address fresh-fit capacity or explicitly model deferred
learning, rather than sweeping allowance thresholds until one favorable run
passes. Exact incremental updates on overlapping histories remain a separate
lead; all-distinct workloads cannot be rescued by pretending they share state.
The ordinary hot-path model stays unchanged. No production edits or publication.
