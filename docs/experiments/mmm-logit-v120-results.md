# Residual-logit calibration: consumed v120 diagnostic

Status: FAIL. Research-only; no production or whitepaper promotion. This uses
already-consumed v120 trajectories, including the phase named confirmation in
the original experiment. That phase is not untouched confirmation of this lead.

## Scope and controls

The frozen [protocol](mmm-logit-v120-protocol.md) compares baseline-only
calibration and eight-expert signed logit correction, each using 64 or 32
eligible labels. All candidates use generic64 as their offset baseline.
Training uses original issued forecasts and actual label arrival times, never
teacher probabilities. This is penalized likelihood, not integrated Bayes.

All 2,688 schedule runs, 21 cases, and both original phases were retained:
2,752,512 forecasts, 86,016 fits, and 1,344 poisoned-prefix checks. No failed
solver run was discarded. All four candidates fail the frozen quality gates.

| Candidate | Non-harm gates passed | Gain gates passed |
| --- | ---: | ---: |
| Baseline calibration, 64 labels | 186/336 | 0/64 |
| Baseline calibration, 32 labels | 130/336 | 1/64 |
| Eight-expert correction, 64 labels | 214/840 | 5/96 |
| Eight-expert correction, 32 labels | 277/840 | 4/96 |

Gate denominators differ because full correction must also beat its matched
calibration and other retained challengers. Do not compare pass fractions as
if they were accuracy. See the [machine-readable diagnostic](mmm-logit-v120-diagnostic.json).

## Numerical audit

Component checks cover roots, gradients, Hessians, extreme probabilities,
collinearity, ownership, and prior-only behavior. Independent reconstruction
does not import the fitter: it checks every fitted gradient and objective and
reconstructs all scored forecasts. Maximum gradient infinity norm was
9.997e-9; maximum objective difference was 2.132e-14. Brier, accuracy, and
log-loss reconstruction errors were zero. Diagnostic and reference replays
were byte-exact. This supports numerical correctness, not predictive validity.

Artifacts: [component checks](mmm-logit-v120-component-checks.json),
[independent reference](mmm-logit-v120-reference.json).

## Why old calibration can hurt a recovering predictor

The [timing diagnostic](mmm-logit-v120-timing-diagnostic.json) evaluates all
37,632 noninitial baseline-calibration fits. Teacher probabilities are used
only AFTER fitting to assess expected Brier risk. Positive gain means improvement
over the raw generic64 forecast. The following are original confirmation-phase,
delayed-feedback, 64-label fits at clock 224, evaluated on the next 32 frames.

| Case | Expected gain on training origins | Expected gain on next 32 | Raw training Brier | Raw next-32 Brier |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | 0.000061 | -0.004798 | 0.215915 | 0.223681 |
| Majority to parity | 0.090921 | -0.154874 | 0.335010 | 0.073351 |
| Parity to majority | 0.048502 | -0.113035 | 0.286910 | 0.059096 |

The switch-case next-32 gain intervals are [-0.196034, -0.113714] and
[-0.147643, -0.078426], respectively. These are exploratory mean +/- 3.5 SE
over 32 trajectories, not simultaneous or anytime-valid coverage. Effective
logit slopes average 0.0832 and 0.2305: the correction suppresses the confidence
of a baseline whose subsequent risk is much lower.

This is evidence of a training/deployment reliability mismatch, not merely
fitting noisy observed labels: expected training gains are also positive.
It does NOT isolate predictor evolution from environmental/input shift.
Some admitted origins predate the change (minimum origins 121 and 110);
we cannot describe these windows as entirely post-change. Stationary-case
observed training gain 0.007403 versus near-zero expected gain also leaves
ordinary finite-sample overfitting as a separate concern.

## Next falsifiable lead

Study calibration paired with a fixed predictor version, using only labels
from forecasts that version actually issued. Compare against identity
calibration and the existing moving-predictor correction under identical
arrival schedules and resource accounting. Do not hindsight-rescore training
labels with a predictor fitted on those labels. A rescue must preserve stable
performance AND recover both switch directions; removing all corrections or
selecting cases after seeing outcomes is not success.

Freezing a predictor creates its own adaptation delay. That cost, scarce
same-version labels, and extra shadow forecast computation must be measured.
This is a proposed experiment, not an implemented rescue or a claim that
versioning alone solves the mismatch. All seven research directions remain open.
