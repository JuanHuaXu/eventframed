# Bayesian grouping and changepoint upgrade, 2026-08-28

This is synthetic implementation evidence, not a real OpenClaw replay. The
production comparison and changepoint code was run on fixed design seeds and
then on distinct confirmation seeds. The v2 and v3 design iterations remain in
this directory rather than being overwritten. They produced 12 and 9 unmatched
alarms over their stable confirmation trajectories, but no numeric rejection
boundary had been frozen before those runs.

## Frozen v4 policy

- Beta-Bernoulli group comparison: split prior 0.5, decision probability 0.95,
  minimum support 8 per member, and at most 64 members.
- Changepoints: hazard 0.05, exact threshold 0.30, run-length cap 32, EWMA rates
  0.25/0.025, two-sided CUSUM slack 0.10 and threshold 8, a 20-observation
  ordinary running-mean warm-up, and a 20-observation reset cooldown.
- Abrupt and recurring detections use a 20-observation window. Gradual drift has
  a stable prefix and a separately declared 60-observation window.
- Changepoint design/confirmation seed bases: 982,451,653 and 67,867,967.
- Group-comparison design/confirmation seed bases: 984,451,656 and 69,867,970.
  Both confirmation runs occurred after the v4 policy was frozen.

## Confirmation results

### Shared-versus-split comparison

| Scenario, 100 samples/member | Share | Split | Uncertain | Wrong | 95% Wilson interval for expected decision |
| --- | ---: | ---: | ---: | ---: | ---: |
| Same rates, 0.8 / 0.8 | 0.0% | 0.0% | 100.0% | 0.0% | n/a: expected share was not observed |
| Moderate split, 0.65 / 0.35 | 0.0% | 87.5% | 12.5% | 0.0% | [77.23%, 93.53%] |
| Strong split, 0.9 / 0.1 | 0.0% | 100.0% | 0.0% | 0.0% | [94.34%, 100%] |

The comparison is deliberately asymmetric in practice: divergent groups become
visible much sooner than compatible noisy groups become safe to merge. At this
sample size it proposed no false shares, but it also did not affirm ordinary
compatible groups. Each 64-trajectory divergent scenario has a 95% Wilson
upper endpoint of 5.66% for its unobserved false-share rate. A separate
deterministic integration fixture with 50
all-useful outcomes per member does produce `share`, proving that the branch is
reachable. Every proposal remains non-authoritative until an external
Anti-Pigeon certificate passes.

### Changepoints

| Scenario | Detected (95% Wilson) | Miss rate | Unmatched alarms (total; per trajectory) | Mean delay among detections |
| --- | ---: | ---: | ---: | ---: |
| Stable 0.8 | n/a | n/a | 1; 0.0156 | n/a |
| Abrupt, noiseless | 64 / 64 ([94.34%, 100%]) | 0.0% | 0; 0.0000 | 0.0 |
| Abrupt, 0.9 to 0.1 | 56 / 64 ([77.23%, 93.53%]) | 12.5% | 17; 0.2656 | 12.3 |
| Gradual after stable prefix | 60 / 64 ([85.00%, 97.54%]) | 6.25% | 10; 0.1563 | 44.6 |
| Recurring, noiseless | 128 / 128 ([97.09%, 100%]) | 0.0% | 0; 0.0000 | 0.0 |
| Recurring, 0.9 / 0.1 | 99 / 128 ([69.36%, 83.74%]) | 22.66% | 21; 0.3281 | 13.8 |

A detector trigger is matched at most once to a declared change, in chronological
order, when it falls between the change and the inclusive detection-window end.
Triggers left after matching are unmatched alarms. In the noisy abrupt,
recurring, and gradual scenarios they were 23.29%, 17.50%, and 14.29% of all
triggers, respectively. Mean delay is computed over matched detections only;
missed changes are not entered as zero-delay observations. Wilson intervals are
fixed-sample descriptive intervals, not sequential confidence sequences.

Compared with the original detector, which missed 96.88% of noisy abrupt, all
gradual, and 93.75% of noisy recurring changes, v4 is a substantial correction.
It is not a proof of production reliability. One stable false alarm, repeated
unmatched alarms in changing streams, recurring misses, and long gradual delay
remain explicit limitations.

Raw reports:

- `2026-08-28-additional-claims-v4-design.json`
- `2026-08-28-additional-claims-v4-confirmation.json`
- intermediate confirmations: `2026-08-28-additional-claims-v2-confirmation.json`
  and `2026-08-28-additional-claims-v3-confirmation.json`
