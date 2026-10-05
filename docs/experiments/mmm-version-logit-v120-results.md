# Version-local calibration and cadence-matched control

Status: FAIL. Consumed v120 diagnostic, not fresh confirmation or production
evidence. Both strict-version candidates and moving-history controls fail.

## Intervention and attribution

The [strict protocol](mmm-version-logit-v120-protocol.md) prevents corrections
from using labels attached to older predictor versions. Each original expert
publishes every 32 frames. Calibration resets to its original prior at those
boundaries and refits on same-version arrived labels every 8 or 16 frames.
The underlying experts remain unchanged. These are research slow-path fits.

Because the earlier calibration study used a 32-frame cadence, comparing only
against it would confound version isolation with update frequency. The
[cadence-matched control protocol](mmm-version-logit-v120-cadence-control-protocol.md)
was declared after seeing strict-version results and before scoring the control.
It uses the latest 64 eligible labels at the same 8/16 cadence. It is an
exploratory attribution check, not an independently confirmed rescue.

## Complete gate outcomes

Each experiment retains all 2,688 runs and 21 cases. Each produces 2,752,512
forecasts and 258,048 fits. The full candidates are compared with BOTH retained
variational and segment caps, so their denominator exceeds the earlier study's.

| Model / cadence | Strict nonharm | Strict gains | Moving nonharm | Moving gains |
| --- | ---: | ---: | ---: | ---: |
| Baseline / 8 | 204/336 | 3/64 | 236/336 | 0/64 |
| Baseline / 16 | 228/336 | 2/64 | 221/336 | 0/64 |
| Full / 8 | 615/1176 | 14/96 | 672/1176 | 8/96 |
| Full / 16 | 663/1176 | 11/96 | 555/1176 | 7/96 |

No row meets all requirements. Counts are gates, not accuracy. Neither
increased gain counts nor identity-like abstention substitutes for broad gains
and stable-case protection.

For the full / 8 candidate, original confirmation-phase delayed terminal64:

| Case | Strict Brier | Markov Brier | Strict gain versus cadence-matched moving control |
| --- | ---: | ---: | ---: |
| Additive stationary | 0.237146 | 0.221755 | -0.010062 [-0.022020, 0.001896] |
| Null | 0.278343 | 0.258525 | -0.012216 [-0.018605, -0.005826] |
| Majority to parity | 0.092457 | 0.060262 | -0.009511 [-0.034379, 0.015357] |
| Parity to majority | 0.087009 | 0.102272 | 0.023242 [0.000749, 0.045735] |

Positive paired gain favors strict version restriction. Intervals are mean
+/- 3.5 SE over 32 trajectories, not simultaneous or anytime-valid intervals.
The phase label is inherited: these data are already consumed. Do not describe
the isolated positive cell as untouched validation.

The cadence-matched comparison weakens a simple stale-version explanation.
Strict isolation helps one switch direction but harms null behavior. It also
reduces evidence volume, so this comparison does not separate version alignment
from sample scarcity. A larger online-fitting budget alone is not a rescue.

## Support and verification

Under delayed/missing feedback, strict calibration averages 3.249 labels per
8-cadence fit and 1.895 per 16-cadence fit. Empty fits are 14,064/43,008 and
10,872/21,504 respectively per candidate in that schedule. Immediate-feedback
means are 12 and 8. Empty fits intentionally return zero correction.

Each experiment passes 1,344 poisoned-prefix checks. Strict calibration also
passes 1,008 prior-version label isolation checks with issued expert laws held
fixed; these do not assert that the original learner ignores past evidence.
Independent verifiers import no fitter, reconstruct every admitted origin,
check all fitted gradients/objectives and reproduce Brier, accuracy and log-loss
scores with zero discrepancy. Maximum gradients are below 1.001e-8; objective
differences are below 2.132e-14. The first collection attempt hit a 128 MiB output
buffer limit and saved no artifact. The same experiment was rerun with adequate
collection capacity; no failed trajectory was omitted.
Both full diagnostic outputs subsequently reproduced byte-for-byte, verified
by SHA-256 against the saved artifacts while streaming replay output.

Artifacts: [strict diagnostic](mmm-version-logit-v120-diagnostic.json),
[strict reference](mmm-version-logit-v120-reference.json),
[moving control](mmm-version-logit-v120-cadence-control.json),
[moving reference](mmm-version-logit-v120-cadence-reference.json),
[all paired comparisons and support](mmm-version-logit-v120-comparison.json).

## Research implication

Do not deploy strict resetting as the rescue. It trades stale information for
very small fitting samples and has not met the stable/shift requirements.
A next model would need to represent uncertainty in the correction, not merely
fit a confident point correction to two or three labels. That is a lead, not
evidence of success; a new proposal must distinguish itself from the already
failed variational base-learner study and retain these controls. No production
changes, runtime claims, whitepaper promotion or completed research directions.
