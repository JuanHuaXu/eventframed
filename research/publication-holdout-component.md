# Publication holdout: a distinct selection hypothesis

## Status

Research component implemented and verified, **quality untested**. Not promoted.
All seven goals remain open. No existing learning policy or production file was
changed. Implementation: `internal/observationgate/publication_holdout_test.go`.

The delayed forest experiment has early confident errors and later useful short
forecasts. Prior neutral-prior, age-discount, compatibility-transfer and strong
Brier attempts failed broader checks. This candidate instead evaluates the
newly fitted models on labels excluded from that model's training, then initializes
the new outer selector from those forecasts. It does not transfer old-model
losses into a new model as though they were the new model's own predictions.

## Research grounding and limits

[Kuznetsov and Mohri, Time Series Prediction and Online Learning (COLT2016)](https://proceedings.mlr.press/v49/kuznetsov16.pdf),
especially introduction p3 and the model-selection discussion, explicitly
identifies the problem: recent validation leaves recent information out of
training; old validation may select obsolete models. Their guarantees involve
nonstationarity discrepancy and other assumptions. None is established for our
delayed, selectively observed process. This experiment does not implement their
online-to-batch algorithm or inherit its theorem.

[Maillard, Arlot and Lerasle, Aggregated Hold-Out (JMLR2021)](https://jmlr.org/papers/volume22/19-624/19-624.pdf),
Sections2.3-2.4, separates training from validation and discusses retaining
the actual validated fitted objects instead of replacing them with full-data
refits. Its setup assumes independent identically distributed data. Our candidate
uses one chronological split and likelihood weights, not agghoo's repeated
holdout selection/aggregation, and claims none of its oracle guarantees.

These sources motivate an isolation contract and warn about its limitations;
they do not establish that the proposed rescue will work.

## Frozen component definition

Receive32-256 arrived audited samples sorted by original event index. Reserve
the newest16 for calibration; train count and subset models exclusively on the
older samples. The existing initial base must have independent pre-stream
training provenance, enforced by the eventual driver. Both live and reference
labels from the reserved origins are excluded from fitting. Do not refit on
validation labels before publishing the calibrated objects.

Use the original observed masks/values from those audited frames. Evaluate four
fixed experts: base, short .7-count/.3-subset mixture, long pooled/local according
to current Anti-Pigeon authority, and .5 neutrality. The inner mixture remains
at its default prior during calibration. Fitting both nested selection layers
on the same16 labels is deliberately excluded.

For prior pi=(.7,.1,.1,.1), validation predictions p_ij, clipped to
[1e-6,1-1e-6], and observed labels y_i:

```text
log w_j = log pi_j + sum_i [y_i log p_ij + (1-y_i) log(1-p_ij)] - log Z
```

Compute normalization by log-sum-exp. These are finite-model likelihood weights
used to initialize a selector, not a calibrated correctness certificate under
drift. A validation event's outcome may influence the next publication but never
retroactively replace its original served prediction or score.

All fitting occurs after the supplied availability clock. The function rejects
future labels, repeated/unordered origins, malformed bits and observed-value
inconsistency. It cannot authenticate a dishonest caller's claimed provenance.
Temporal disjointness does not imply statistical independence or eliminate drift.

## Checks and performance

Race tests pass in1.309s; package vet passes. Tests cover literal Bernoulli-product
weights, equal-forecast prior preservation, invalid numerical inputs, provenance
rejection, and exhaustive512-input equality of all fitted model predictions
after flipping validation labels. Selection weights change under that flip.

Apple M4 darwin/arm64, 80 supplied samples (64 training,16 calibration):
7.117773 /7.205489 /7.138048ms per complete bundle, about6.957MB allocated,
38 allocations. Three200ms repetitions of `BenchmarkPublicationHoldout`.
This includes three count fits and forest/subset fitting, not a serving request.
No performance advantage over the no-holdout control is established.

## Next experimental contract

Use the delayed experiment's six cases, independent fresh bases, both schedules,
same audit acquisitions, expiry and external gate. Compare:

1. Unchanged uniform-input learner using all available training evidence.
2. Same held-out training partition, reset inner/outer weights to default prior.
3. Same partition, default inner weights and calibrated outer weights.

Initially use uniform input weights in all three to isolate publication
selection; the forest remains a separate unvalidated addition. The implementation
accepts either for future comparisons. Do not tune holdout size after results.
Primary claims need calibrated gains over both controls, not merely over an
artificially weakened reset control. Retain full/post non-harm across stable,
null and shift cases; preserve the .005 gain and paired positive-lower-bound
requirements. Charge every fit and calibration forecast.

Tests before collection must prove original-control immediate/delayed parity,
no validation-origin training overlap, models unchanged between calibration and
publication, and no double application of old journal labels after warm start.
End-to-end quality remains untested until that driver and fresh experiment run.
