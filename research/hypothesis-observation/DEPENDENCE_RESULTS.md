# Joint source-dependence v3: FAILED overall screen

1,280 episodes, five cases, two splits, five arms. Frozen criteria and all raw
traces remain in DEPENDENCE_PROTOCOL.md and dependence-v3.json.gz.

## Confirmation results

| Case | Independent-only final Brier | Joint-Gini final Brier | Independent-only accuracy | Joint accuracy | Joint confident wrong |
| --- | --- | --- | --- | --- | --- |
| Independent, 5% noise | <0.000001 | 0.000126 | 100% | 100% | 0/128 |
| Independent, 20% noise | 0.15685 | 0.22671 | 88.28% | 85.16% | 0/128 |
| Copied, 5% noise | 0.05743 | 0.05982 | 96.88% | 96.88% | 0/128 |
| Copied, 20% noise | 0.68899 | 0.54166 | 61.72% | 60.94% | 0/128 |
| Mixed, 20% noise | 0.41064 | 0.28916 | 73.44% | 76.56% | 1/128 |

Two screening failures: independent20 curve-Brier harm was 0.03011, exceeding
the predeclared 0.02 ceiling; copied05 final-Brier gain was -0.00238 rather than
the required >=0.02. Copied20 and mixed20 final-Brier gains had positive paired
z3.3 lower bounds (0.03133 and 0.00582). These do not override failed conditions.

The joint model materially reduces false confidence: independent-only
confident-wrong counts were 39/128 on copied20 and 14/128 on mixed20, versus
0/128 and 1/128 respectively. But reduced confidence is not greater correctness:
copied20 accuracy slightly decreased. The stop-after-one-source control achieves
final Brier 0.52262 on copied20 using eight queries, better than joint-Gini's
0.54166 using16. This suggests residual wasted acquisition on duplicated data.

## Verification

Full replay, hashes, unique requests and normalization passed. A separate exact
enumeration test sums over all16 hypotheses and256 per-test source-mode
assignments; it agrees with both marginal hypothesis weights and conditional
mode probabilities after each observation, including disagreement. The failed
screen is therefore not explained by a faulty factorized Bayes update in these
checked cases.

Model limits: fixed 50/50 mode priors are uncertainty assumptions, not measured
provenance. Distinct source IDs still cannot prove independence. Arbitrary bias,
cross-test dependence and missing true hypotheses remain outside the declared
family. Even in-family calibration is not demonstrated by zero observed
confident errors in128 trials.

## Remaining lead

Do not enable this as a universal replacement or tune its prior against this
confirmation. A fresh experiment can condition dependence priors on separately
observed provenance signals with declared false-independence rates, and compare
against conservative one-source stopping. Another lead is cost-sensitive
acquisition that stops when additional expected target information is too small.
Both need independent validation; they may improve the tradeoff but do not
remove the information limitation when all available sources copy one error.
