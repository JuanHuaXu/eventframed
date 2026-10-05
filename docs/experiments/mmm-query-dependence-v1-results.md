# Dependence calibration: global shrinkage does not rescue

## Design and Attribution

Follow the [protocol](mmm-query-dependence-protocol.md). The motivation is
separating predictive dependence from marginal uncertainty, as in
[Wang, Sun and Grosse (2021)](https://proceedings.mlr.press/v130/wang21g.html).
Their cross-normalized likelihood operates in Gaussian regression. Our
Bernoulli mixture is a separately derived adaptation, not their XLL algorithm.

Mix the current pairwise joint law with its independent-marginal product:
Q_lambda=(1-lambda)Q_query Q_target+lambda Q. This keeps both marginals fixed
and scales dependence. Fit lambda in[0,1] using phase0 actual query/target labels
only. No teacher probabilities or phase1 labels enter the optimizer. Missing
query labels from the synthetic archive are offline supervision, not free
online feedback. All data have been consumed before; this is not confirmation.

## Result

The fit selects **lambda=1**, retaining the original law. The training objective
is convex; its derivatives at0 and1 are -5.749762 and -2.023869, so the constrained
optimum is the original endpoint. Independent derivative reconstruction agrees.

Phase1 delayed:672 histories, equal weight per history/candidate/target.

| Metric | Independence | Original = fitted |
| --- | ---: | ---: |
| Actual joint log loss | 0.993932 | 0.988835 |
| Actual conditional log loss | 0.510273 | 0.505177 |
| Realized target Brier | 0.168697 | 0.166881 |
| Target-law expected Brier, actual query answer | 0.170226 | 0.168410 |
| Teacher-product expected joint log loss | 1.006259 | 1.001522 |

Lower is better. Dependence helps on average, but both predeclared all-cell
improvement screens fail. The fitted model is identical to original, so it
cannot satisfy strict improvement over original on transition cases. Nonharm
against independence also fails in multiple cells. Case17's expected-Brier gain
is -0.000594 with the existing descriptive bound [-0.001038,-0.000151]. These
are consumed-data, trajectory-level bounds, not new significance guarantees.

This rejects global dependence shrinkage as a rescue on this training objective.
It does not show every modeled link is good, nor that stronger-than-original
dependence is valid or useful. A positive global shrink factor scales Brier
acquisition gains by lambda^2 and cannot change their exact mathematical order.
The mixture is a valid pairwise law, not an instantiated full multivariate
process. Teacher-product expected loss is a fixed-generator evaluation, not an
oracle posterior over unknown teachers.

## Verification

-180 joint normalization/marginal/relabeling/gain-scaling tests; five known
 convex optima plus ownership, oracle-field traps and invalid inputs pass.
-All2688 records retained, including1344 complete histories with no query pairs.
 There are287587 query-target pairs, not287587 independent trajectories.
-Training uses143840 pairs from672 phase0 histories, aggregate weight672.
-Independent audit reconstructs joint cells from covariance rather than the
 mixture helper:24192 stored score values,756 cell means and504 paired bounds
 pass. It rebuilds the phase0-only derivatives from raw labels and verifies
 the endpoint optimum; teacher quantities are used only in evaluation.
-Replay is byte-identical. Input, protocol, implementation and audit hashes
 are recorded. No additional posterior fits or serving performance claim.

Artifacts: `mmm-query-dependence-v1.json`, `-replay.json`, `-audit.json`.

## Next Lead

Test an evidence-conditioned, case-blind dependence calibration rather than
uniform shrinkage. Freeze allowed as-of features and the low-capacity model
before evaluation; do not use generator case IDs to select a winning branch.
Pairwise validity, preserved marginals, no-future evidence boundaries, actual
query-policy performance, and total acquisition cost remain separate gates.
No production or paper promotion; all seven research goals remain open.
