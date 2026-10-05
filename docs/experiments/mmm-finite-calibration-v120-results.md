# Exact finite posterior calibration

Status: FAIL. Consumed diagnostic only. No production or whitepaper promotion.

The [frozen protocol](mmm-finite-calibration-v120-protocol.md) tests a declared
25-atom intercept/slope correction with exact finite posterior averaging. It
includes an identity-prior atom and a frozen no-learning prior-predictive control.
This is not a numerical approximation to the previous Gaussian coefficient
posterior. Changing the model/prior alongside integration means differences
from MAP cannot be attributed to integration alone.

## Results

All 2,688 runs and 21 cases were retained, producing 2,064,384 forecasts.

| Learning mode, cadence8 | Nonharm passed | Gains passed | Status |
| --- | ---: | ---: | --- |
| Current predictor version only | 614/672 | 6/96 | FAIL |
| Latest64 labels across versions | 563/672 | 0/96 | FAIL |

Both are compared against generic64, Markov, matched baseline-MAP and the
prior-only control. Gain requirements exclude MAP but retain all eight changing
cases. Counts are comparison gates, not classification accuracy.

Original confirmation-phase, delayed schedule, terminal64 expected Brier:

| Case | Version-local | Moving64 | Prior-only | Generic64 | Markov |
| --- | ---: | ---: | ---: | ---: | ---: |
| Additive stationary | 0.222976 | 0.223283 | 0.219643 | 0.221845 | 0.221755 |
| Parity4 | 0.074274 | 0.067828 | 0.081300 | 0.070807 | 0.049066 |
| Null | 0.263327 | 0.257443 | 0.262242 | 0.265511 | 0.258525 |
| Majority to parity | 0.129871 | 0.227022 | 0.134166 | 0.127403 | 0.060262 |
| Parity to majority | 0.111221 | 0.190389 | 0.116081 | 0.108775 | 0.102272 |

The version-local finite model beats its matched baseline-MAP in the displayed
stationary and null cells by 0.008650 [0.004714, 0.012586] and 0.006897
[0.005068, 0.008725]. This does not rescue recovery: both switch scores remain
worse than generic64 and Markov. The prior-only results show that some apparent
benefit can come from static softening rather than learning useful corrections.

Intervals are exploratory mean +/- 3.5 SE over 32 trajectories, not simultaneous
or anytime-valid. The original phase named confirmation has already been used;
it is NOT untouched confirmation of this candidate.

## Verification and limits

Independent multiplicative enumeration checks 6,912 small-model predictions
against the log-domain implementation (maximum error 5.56e-16), and also checks
model evidence. Prior-only symmetry, extreme inputs, invalid inputs and returned
weight ownership checks pass. In the full run, every posterior mass is checked
for normalization and finite evidence. All 1,008 poisoned-prefix and 252
prior-version isolation checks pass. The latter holds original expert laws
fixed and tests only the correction's evidence boundary.
The complete diagnostic was replayed byte-for-byte (SHA-256
`74f61bdeb23b29cbca10caf0550db6ae4d3a925a44989c53b2ec1ce2f3a37f65`).
This replay is reproducibility evidence, not an independent full-data derivation.

Artifacts: [component checks](mmm-finite-calibration-v120-component-checks.json),
[full fit logs, scores and gates](mmm-finite-calibration-v120.json).

Exact finite integration addresses numerical posterior uncertainty within this
declared model; it does not establish that the model is appropriate. Neither
version resetting nor finite calibration averaging has solved the workload.
Do not respond by tuning the same grid or identity mass on these consumed
cases and relabeling the result as confirmation. A new direction should address
the underlying representation/recovery limitation or obtain genuinely new
task evidence, while retaining all existing failures and controls. All seven
research goals remain open.
