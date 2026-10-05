# Finite-table service result: small improvement, target missed

**FAIL52/64 completion in all three pairs.** Candidate completes2/64 versus
batch1/64 each trial. Foreground p99 passes1.10x comparison in each pair.
This is a measured partial improvement, not an operational rescue.

| Trial | Batch p99 ms | Table p99 ms | Batch completed | Table completed | Table drops |
|---|---:|---:|---:|---:|---:|
|0|34.566|32.588|1|2|21|
|1|31.079|26.794|1|2|23|
|2|33.695|27.447|1|2|22|

All64 offers retained in each denominator. No request errors;16 writes overlap
readers in each arm. All successful predictions match the reference; no cache
hits. Actual service, persistent temporary LibraVDB and scheduler guards remain.

Candidate started22/22/24 processors;19/18/21 had under25ms remaining.
Successful fit durations31.154-47.015ms. These are selected successful durations,
not a calibrated runtime bound.20/20/21 contexts interrupted. In trial2 one
successful return was subsequently rejected by the scheduler, illustrating why
verified fit return and completed diagnostic are distinct quantities.

Evidence: `segment-table-load-results.json`; run
`node research/verify-segment-table-load.mjs` for source hashes, timing identities,
terminal accounting and the failed criterion. No population tail guarantee,
semantic learning claim or production change follows from these six arms.

The per-fit table helps actual computation, but ordinary FIFO admission still
spends much of its worker time on nearly expired jobs. Prior65ms allowance
results remain a separate candidate, not silently combined here. Further work
should test an explicit capacity/admission design or genuinely incremental
evidence updates, with failures retained and not relabeled as completed work.
The seven roadmap directions remain open.
