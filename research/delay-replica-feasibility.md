# Delayed-feedback replicas: feasibility before quality

## Research source

Read Algorithm1 and Theorem1 of Joulani, Gyorgy and Szepesvari (2013),
[Online Learning under Delayed Feedback](https://proceedings.mlr.press/v28/joulani13.pdf).
BOLD sends a new prediction to an available base learner and returns feedback
to the learner that issued it. If none is available, it creates another copy.
The number of copies follows peak outstanding feedback plus one. Their transfer
theorem requires a suitable base regret bound and prediction-independent delay;
it is not a guarantee for EventFrame's censored, selectively observed stream.

The earlier online-Brier proposal cited this paper but did not implement its
replica reduction. This is not a claim to have discovered delayed aggregation.

## New consumed-schedule diagnostic

`research/delay-replica-demand.mjs` reads only arrival/missing fields from the
192 consumed trajectories under both immediate and delayed schedules. It does
not use labels or predictions. Prediction precedes same-tick feedback, exactly
as in the captured driver. Free-slot selection is deterministic and independent
of outcomes. An optional age32 expiry is analyzed separately, not called BOLD.

Parent SHA-256:
`1533a049c2146cc7839799e9a394691d96c3fdf4a1a8e8d1c504db86053bf8e6`.
[Per-trajectory artifact](../docs/experiments/mmm-delay-replica-demand-v1.json).
The script asserts parent hash,384 unique schedule records,512frames per record,
arrival bounds and conservation of issued/completed/expired/pending counts.
Immediate feedback, fixed delay3, all-missing and expiry boundaries have literal
self-tests. A second invocation produces byte-identical results.

| Schedule / protocol | Minimum copies | Mean copies | Maximum copies |
| --- | --- | --- | --- |
| Immediate, either | 1 | 1 | 1 |
| Delayed, never expire | 97 | 117.234 | 145 |
| Delayed, expire age32 | 24 | 26.203 | 30 |

Mean permanently missing outcomes are103.255 of512. A fully missing synthetic
64-frame check needs64 unexpired copies or33 expiry-managed copies. The33
comes from issuing before same-tick expiry, not a32-copy implementation promise.
The unexpired pool can grow with stream length. Copy counts alone do not measure
memory: replicating selector weights is different from replicating full learners.

## What a bounded adaptation would need

An age-limited scheduler must bind each delayed response to its original slot
and generation, reject late/duplicate replies, and account for every expiry.
It cannot free a slot and later apply the expired loss to an unrelated prediction.
All arrived labels remain available for the ordinary model-training path;
replicating a selector need not partition its training dataset.

For illustration, an independently proved per-copy observed-round Brier regret
budget c would give a crude full-stream bound M*c+N_missing by charging at most1
loss difference to each unobserved round. This is our accounting observation,
not the cited theorem or a validated selector. Here even N_missing/N alone is
about.202, far too loose to certify a.005-scale full-stream improvement. That
does not prove learning improvement impossible; it shows why a missing-outcome
assumption or empirical quality test cannot be replaced by this crude bound.

## Decision

Do not implement unbounded replicas as the delayed-learning rescue. Retain a
bounded selector-only variant as a possible test, explicitly distinguishing
censoring from delay and charging expiry uncertainty. It must beat the original
arrival-based and strong immediate controls on fresh full trajectories, not just
pass a scheduler test. Prior snapshot, stale-credit and publication-reset failures
remain failures. No quality gates changed, no production/paper/remote edits.
All seven goals remain open.
