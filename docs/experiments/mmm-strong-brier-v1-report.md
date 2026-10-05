# Strong Brier substitution and selection diagnosis

Status: FAIL as a broad rescue. Consumed independent-v1 source only. No
thresholds changed, no new confirmation, and all seven goals remain open.
See the [prospective contract](mmm-strong-brier-v1-contract.md).

## Diagnosis preceding the experiment

`research/segment-selection-diagnostic.mjs` decomposes the previous adaptive
segment mixture's excess loss into selection within its two-head convex hull,
additional headroom from the six-head hull, remaining approximation error,
and the control's distance from the noise floor. It checks all 344,064
identities against Markov and Boolean controls, with maximum error 1.11e-16.
All cases are retained and the output replays exactly.

This uses inaccessible generator Q strictly for evaluation. Oracle hull
headroom is not necessarily learnable, nor is the decomposition causal.
For phase1/case19/delayed terminal64, the selection gap is .019666693,
additional-head gap .023118621, and remaining approximation gap .034033963.
The mean excess over Markov is .001876494. Thus more than one mechanism
remains plausible; changing weights alone need not eliminate the deficit.

Five of the six failed Markov protection screens still have favorable means
with wide uncertainty. All 35 failed Boolean-control screens are retained
in the diagnostic, not dropped in favor of the easier Markov comparison.

## Published method and implementation scope

[Vovk and Zhdanov (2009), Algorithm 1 and Theorem 1](https://www.jmlr.org/papers/volume10/vovk09a/vovk09a.pdf)
give a strong Brier aggregation rule. The earlier v94 implementation used
a conservative convex average, not this substitution. For our scalar binary
Brier, the paper's rate translates to eta=2. Define
`z_y = sum_k w_k exp(-2*(p_k-y)^2)` and issue
`p = .5 + (log(z_1)-log(z_0))/4`.

The one-step inequality is `2*(p-y)^2 <= -log(z_y)`. With static experts'
identities, no sharing, immediate complete feedback and uniform two-expert
prior, cumulative scalar regret is at most `log(2)/2`. Evolving predictable
expert forecasts are allowed. That is not a future-risk or calibration
certificate, and the unmodified bound is NOT claimed under delayed/missing
feedback or Fixed Share.

The research adaptation keeps the existing origin-order refiltering and
sharing schedule, using eta2 losses. Strong and linear variants share identical
weights, which isolates the substitution from the higher learning rate.

## Quality and controls

All 672 trajectories and four arms were scored. Criteria keep .01 upper
non-harm and .005 mean gain with positive lower endpoint. Approximate +/-3.5SE
uses eight available indices, not original32-index or simultaneous coverage.

| Arm | Whole Brier | Terminal64 | Protection / 672 | Gains / 96 | Harm windows / 5376 |
| --- | ---: | ---: | ---: | ---: | ---: |
| Segment strong | .156037200 | .144056609 | 608 | 7 | 165 |
| Static strong | .157409663 | .145390148 | 611 | 1 | 112 |
| Segment linear eta2 | .155982284 | .144003898 | 612 | 7 | 150 |
| Static linear eta2 | .157390797 | .145383975 | 611 | 1 | 122 |

The previous segment linear eta1 arm was .156123153 whole Brier, with 606
protection passes and 7 gains. Strong aggregation slightly improves that
pooled number, but is numerically worse than the matched eta2 linear control.
No confidence claim is made about that small between-substitution difference.

Strong segment stationary Brier is .119853883 versus Markov .119723242;
changing Brier is .214835092 versus Markov .218930662. The linear eta2 control
nearly matches stationary Markov in the pooled mean (.119724692), but that
aggregate does not pass every cell or establish general non-harm.

## Mathematical and leakage checks

`research/strong-brier.test.mjs` completes 6,758 checks, including:

- Literal bisection of the paper's substitution equation versus the binary
  closed form, outcome complement symmetry, ties and invalid input rejection.
- Both outcome inequalities, maximum numerical defect 4.44e-16. The eta2
  linear negative control has defect .189387, so this test distinguishes the
  valid substitution from simply substituting a weighted average.
- Independent latent-path enumeration of eta2 delayed weights: maximum error
  5.55e-16; eta1 weights also equal the old reference exactly.
- 256 outcome paths with current/future/unavailable-label poisoning and the
  immediate, no-sharing cumulative bound at every checked prefix.

These are finite numerical checks of the stated implementation, not a new
delayed-feedback theorem. The full screen's control scores agree with source
metrics and its numerical replay is byte-identical. Test strengthening after
the screen did not change candidate or scorer logic. Outcomes and Q never
enter the substitution before they are eligible; Q is only for scoring.

## Compute, artifacts and next step

One-pass mixing totals 458.68ms; replay totals 458.03ms, or 1.33 microseconds
per head forecast for strong and matched-linear outputs together. Both include
cold/JIT effects, exclude I/O, input preparation, scoring, all fits and loaded
serving, and do not establish a speedup over another separately timed run.
The original 545.32s source-model collection remains charged. Reference
refiltering remains O(T^2), not a production constant-time kernel.

Artifacts: `mmm-strong-brier-v1-{screen,replay,timing,replay-timing,checks}.json`
and `mmm-segment-selection-v1{,-replay}.json`. Generation scripts refuse
existing output paths. No failed result was overwritten or deleted.

Next inspect whether a declared multi-head candidate can use the identified
missing-head information while preserving the original evidence budget.
Check the older multi-expert composition failures before testing; do not
recycle them with new names or choose priors by these outcomes. Representation
and statistical precision remain separate problems. No further eta sweep is
justified by this screen. Production, Go, private data, paper and remotes are
unchanged.
