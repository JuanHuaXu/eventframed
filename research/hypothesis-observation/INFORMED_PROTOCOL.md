# Informed source-mode priors v4

Frozen before evaluation. This is a new conditional-provenance experiment, not a
revision of v3's failed gates or a test of cryptographic authenticity.

## Model and isolation

Retain v3's 16 hypotheses, eight tests, four source slots, observation likelihoods
and 16-query maximum. A binary signal for each test indicates independent mode.
Calibrate P(signal | mode) from a separate 4096-record synthetic sample with
balanced-in-expectation modes and signal accuracy 0.9 (seed 731946281). Use
Jeffreys 0.5 pseudo-counts and equal prior odds to initialize each test's mode.
Calibration never sees evaluation hypotheses, observations or outcomes. This is
a plug-in posterior predictive approximation; shared uncertainty in signal
reliability is NOT propagated through the factorized model.

All arms receive eight signal checks before acquisition, charged one unit each.
Compare fixed-50/50 joint Gini, informed joint Gini, independent-only Gini and
one-source-per-test Gini. Same evidence tapes, source signal and initial class
prior per episode; no oracle modes enter informed selection. Charge each selected
observation one unit. The one-source control naturally costs less; report actual
cost and do not claim equal-cost dominance over it.

Two independently seeded splits, 128 episodes per case. Evaluation base seed
202609120417; split/case/episode offsets fixed in source. Matched-signal cases:
independent05, independent20, copied05, copied20, mixed20. Stress cases retain
mixed20 observations but degrade signal accuracy to 0.5 and 0.1, without refitting.
Ground-truth mode in mixed cases alternates by test, as in v3; the model's equal
prior is thus a misspecification stress, not a perfectly matched population prior.

## Frozen screening criteria

Primary confirmation split only. Against fixed joint Gini:

- independent20 mean curve-Brier gain at least 0.01, with paired z=3.3 lower
  descriptive bound above zero;
- all five matched-signal cases: mean curve and final Brier harm at most 0.01;
- copied20 and mixed20: no increase above 0.02 in confidently-wrong frequency.

Also report comparisons against independent-only and cheaper one-source controls.
These criteria test rescue of v3's independent-stream harm, not universal victory.
Stress cases are mandatory diagnostics, not retroactively added success gates.
Report their failures explicitly even if the matched-signal screen passes.
Normal paired bounds are descriptive finite-sample approximations, not confidence
sequences or established simultaneous coverage. No serving changes authorized.

## Verification

Check informed factorized Bayes against enumeration over all 16*256 joint states,
including unequal priors and conflicting observations. Check calibration replay,
normalized forecasts, unique acquisitions, signal-free fixed-arm invariance,
full experiment replay and source hashes. Exclusive-create artifacts retain all
outcomes and raw traces. Do not tune thresholds/seeds after inspecting results.
