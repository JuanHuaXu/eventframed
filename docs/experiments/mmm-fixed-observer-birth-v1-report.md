# Fixed observation does not rescue the fresh prior

## Decision

Close the fixed10% birth-prior approach as an unsuccessful quality rescue. Its
failure is not solely caused by changed observation choices. With observation
held fixed, member-shift Brier slightly worsens in both archived cohorts; no
changed scenario passes the required gain screen. No runtime changes or promotion.

[Contract](mmm-fixed-observer-birth-v1-contract.md),
[raw replay](mmm-fixed-observer-birth-v1.jsonl),
[summary](mmm-fixed-observer-birth-v1-summary.json),
[summary replay](mmm-fixed-observer-birth-v1-summary-replay.json).
Raw SHA-256:
`2f675e9d854d67475a039053c70868ddf6293175804b1c61309fe5ba9f9e8347`.

## Isolation and checks

All160 original trajectories, control/candidate metrics, tapes, fits and birth
activation times match the prior archive. Two added states have no Reader API:
they use only the corresponding inherited-weight control's issued same-step
mask, values and expert forecasts. They keep their own weights and forecast
journal, then update on the outcome. Disabled states reproduce controls exactly.
Acquisition costs match controls exactly in both full and post windows.

Focused race/parity tests pass2.054s, vet passes, full replay17.72s (17.945s
package). All811 captured source hashes match local/embedded files, and the
summary replay is byte-identical. This is consumed-data diagnostic replay, not
new confirmation or serving-performance evidence.

## Fixed-observer birth versus inherited action, mixture gate

Gain is control minus candidate post-Brier; intervals are paired16-trajectory
mean +/-3.5SE, not anytime guarantees or independent-frame confidence intervals.

| Cohort/scenario | Control | Fixed-observer birth | Gain interval |
| --- | --- | --- | --- |
| Design/member | .2036040 | .2037543 | [-.0002465,-.0000541] |
| Second/member | .2034951 | .2036586 | [-.0002055,-.0001215] |
| Design/recurring | .1764658 | .1764824 | [-.0001049,.0000717] |
| Second/recurring | .1680326 | .1678842 | [-.0002209,.0005177] |

All fail the .005 mean-gain plus positive-lower-bound requirement. All primary
.01 non-harm diagnostics pass, but small non-harm is not a beneficial upgrade.
Stable/common/null outputs remain identical with no birth activation. The old
gate variants and gate-within-fixed comparisons also yield no gain-screen pass.

Fixed versus coupled birth effects vary: on design member it removes much of
the coupled action's adverse mean; on second-cohort member it removes the small
favorable mean instead. Thus acquisition is part of the dynamics, not a uniformly
harmful confound whose removal guarantees benefit. Small diagnostic intervals
do not establish population causality or broad repeated-selection validity.

## Why the earlier mediation was insufficient

Local beating pooled is not the same as local beating the ensemble. For outcome
y=1, let another expert forecast.9, pooled.2 and local.6. At weight.01, replacing
pooled with local improves the ensemble forecast.893 ->.897 (Brier.011449 ->
.010609). Raising local weight to.1 moves it to.87 and worsens Brier to.0169.
The local expert is better than pooled but worse than the other strong expert.

The exact whole-mixture gain for q=p+a(l-p) is:

`(y-p)^2 - (y-q)^2 = 2a(y-p)(l-p) - a^2(l-p)^2`.

A positive standalone local-versus-pooled score does not establish a positive
first term relative to the ensemble. `research/mixture-benefit-identity.mjs`
checks2662 finite grid identities (maximum residual3.331e-16) and the explicit
counterexample; [artifact](mmm-mixture-benefit-identity-v1.json). This is an
algebra check, not a learned policy or a proof of population gain.

## Next boundary

Do not increase birth mass or lower Anti-Pigeon thresholds on these consumed
data. Predictive promotion needs evidence against the full emitted mixture,
while structural divergence remains a separate sharing decision. Investigate
whether scoring/selection is aligned with the evaluated proper loss, first
checking the existing strong-Brier and logged-gain studies so failed variants
are not repeated as new techniques. Better local learning or genuinely new
task evidence may be needed; neither follows from a prior reset.

All seven goals remain OPEN. Production, whitepaper and remotes remain untouched.
