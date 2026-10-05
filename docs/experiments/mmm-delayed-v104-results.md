# Delayed routed learning v104

**Overall FAIL: 206/218 frozen gates pass.** Role carry preserves late losses
and helps stable parity, but it does not rescue delayed regime recovery. It
significantly worsens both confirmation switch cases relative to the
conservative journal under the declared approximate paired intervals.

This is a quality result, not just a retention check. It falsifies adoption of
unconditional stale-role loss carry under this experiment's full criteria.
Do not promote the variant, weaken the criteria or replace the failed seeds.

## Artifacts and integrity

- [Frozen protocol](mmm-delayed-v104-protocol.md)
- [Step-level data](mmm-delayed-v104.json): 129,748,821 bytes, created mode0600.
- [Reconstructed metrics and all gates](mmm-delayed-v104-summary.json)
- [Independent evaluator](../../research/delayed-v104-summary.mjs)
- [Experiment driver](../../internal/observationlearners/delayed_v104_test.go)

There are 768 independent latent trajectories, each run under two paired
feedback schedules: 1,536 schedule-runs and 393,216 scored frames. Do not count
the two schedules as independent replicates. All paired rules, inputs, true
probabilities and outcomes match. The 3,840 effective role seeds are unique and
disjoint from the audited v90-v103 learner/null allocations.

All 1,536 runs replayed exactly. The independent evaluator reconstructed every
score, foreground cost, fit's eligible origin list, per-step journal accounting
and final settlement. All 21 source/protocol/evaluator hashes match; the summary
regenerates byte for byte. Immediate compatibility with consumed v103 cases
0,3,10,11 is exact. Race-enabled compatibility/seed checks and vet passed.
The journal's prior race/lifecycle checks remain applicable and unchanged.

Generation: 194.91 seconds. Complete replay: 197.93 seconds.
Artifact SHA-256:
`73003cee91b7c7c8b16e6b859391fbeb7da3deb39e92d1bec118175f09f8f760`.

## Confirmation quality

Late128 means under jitter0..31/missing0.2. Lower expected Brier is better.

| Case | Generic64 | Conservative journal | Role carry | Role-carry expected accuracy |
| --- | ---: | ---: | ---: | ---: |
| Parity4 | .072022 | .069840 | .060403 | 94.47% |
| Majority to parity | .254291 | .249590 | .252541 | 65.62% |
| Parity to majority | .227787 | .226869 | .231603 | 68.52% |

Late role-carry Brier gain over the conservative journal, confirmation:

| Case | Gain | Approximate paired interval | Frozen gain gate |
| --- | ---: | --- | --- |
| Parity4 | .009437 | [.006332, .012542] | PASS |
| Majority to parity | -.002951 | [-.004484, -.001418] | FAIL |
| Parity to majority | -.004733 | [-.007976, -.001491] | FAIL |

Both phases pass the incremental parity4 gain and fail both incremental change
gains. The retention mechanism is therefore useful in one setting but harmful
after changes; preserving more history is not sufficient for adaptation.

Majority-to-parity role-carry Brier exceeds .25, the constant-.5 reference,
despite classification accuracy exceeding chance. Treat that as a remaining
forecast-quality problem, not as a successful calibration rescue. The reference
.5 score is an arithmetic diagnostic, not an additional frozen adoption gate.

On immediate feedback the journal variants are exactly identical. Confirmation
late Brier for parity4 / majority-to-parity / parity-to-majority is respectively
.060563 / .186136 / .168846. The paired delayed values show substantial recovery
degradation; a finite stable-parity success does not remove that gap.

## Gate accounting

- All192 non-harm gates pass their upper-harm<=.01 allowance.
- 12/20 generic-comparator gain gates pass. Eight fail: all four parity3 cells
  miss the .005 mean-gain floor, and both delayed switch cases fail in both
  phases. The parity3 means are positive, but do not meet the frozen target.
- 2/6 incremental carry gates pass: delayed late parity4 in each phase. All
  four delayed switch improvements over conservative fail.

Passing a .01 harm allowance does not mean zero harm. The adverse confirmation
switch intervals above exclude zero while remaining inside that allowance.
All intervals use the frozen paired mean +/-3.5 standard errors over32
trajectories. They are approximate fixed-sample screens, not confidence
sequences or a guarantee over the complete history of research variants.

## Evidence availability

Every model fit uses only arrived labels, ordered by original event index.
Delivery and expiry precede scheduled publication; zero-delay current outcomes
are delivered only after that frame's forecast. The exact fit-origin lists and
journal chronology are independently reconstructed, not inferred from summary
counts. The last64/32 arrived labels can span more than64/32 event times.

Mean confirmation final counts illustrate the gap:

| Case | Fully applied labels | Stale losses carried only by selector | Missing/censored |
| --- | ---: | ---: | ---: |
| Parity4 | 55.438 | 151.125 | 49.438 |
| Majority to parity | 50.313 | 154.938 | 50.750 |
| Parity to majority | 50.969 | 153.688 | 51.344 |

The conservative journal discards the middle column from its selector. Role
carry retains it, but neither sends it to a new publication's evidence tests.
No labels remain pending after flush. Full-frame training audits are an extra
nine coordinates per arrived label plus the16 initial labels; they are not
included in the foreground cap.

## Performance

Apple M4 darwin/arm64, benchmark suffix -10, measured after replay completed.
[Full-fixture output](mmm-delayed-v104-fixture-benchmarks.txt), three one-iteration
measurements, includes all three arms, 256 frames, fitting/publication and flush:

| Fixture | Time | Allocated bytes | Allocations |
| --- | --- | --- | --- |
| Immediate | 134.872-148.074 ms | 26.309-26.314 MB | 3,396-3,401 |
| Delay/missing | 126.675-129.760 ms | 26.293 MB | 3,346-3,347 |

These are not per-frame serving latencies. Less available training evidence
and changed update/acquisition work can make the delayed fixture faster; this
does not establish a performance improvement.

[Journal timing output](mmm-delayed-v104-kernel-benchmarks.txt), three500ms
repetitions, covers an entire256-frame fixed-model lifetime with delay8:

- Role carry: 1.333-1.341 ms,135,280 bytes,1,215 allocations.
- Conservative: 1.344-1.351 ms,about135,280 bytes,1,215 allocations.

No meaningful extra bookkeeping cost is shown on this narrow fixture; do not
infer a speed superiority from these close measurements. It excludes model
fitting/copying, storage, retrieval and network service. Benchmark source hash:
`d62211da2afc68884026e78fa4a4d5c101ad0dce31136db40d2a0aabeed5b883`.

## Interpretation and next test

Confirmed implementation facts: stale role losses receive full loss-update
weight, and fixed-share forgetting advances when a label is processed, not on
every event clock tick. Thus delayed/missing feedback changes both the age of
losses being processed and the rate of forgetting. The experiment establishes
the regression, but does not isolate which mechanism causes how much of it.
Acquisition choices can also change when selector weights change.

Follow [the origin-age diagnosis and rescue proposal](../../research/origin-age-routing-proposal.md)
before another adoption attempt. Inspect consumed update traces, then test
clock-based forgetting and origin-aged losses with explicit ablations. Do not
call the proposed discounted formulation validated by this failed run.

All seven directions remain open. The broader v88 delayed cases, v80 post-split
forecast gain, independent generators, real-agent evidence and persistence-tail
requirements are not replaced by this screen. No production changes, whitepaper
edits, commits or pushes were made.
