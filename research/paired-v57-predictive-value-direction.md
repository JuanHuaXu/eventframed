# V57 direction: observation value for prediction

RECOMMENDATION to test, not a confirmed implementation bug. V54's concentration
criterion correctly targets its declared coarse class c=(eta,family). The next
question is whether that target misses useful uncertainty in member-local state
once the class is settled. Alternatives remain dynamics/prior mismatch, delayed
measurement value and actual acquisition cost. No production change, no tuning
on consumed cohorts and no new empirical success claim.

## Primary Source Read

Chaloner and Verdinelli, Bayesian Experimental Design: A Review (1995),
[author technical report tr599](https://www.stat.cmu.edu/tr/tr599/tr599.html),
[full PostScript](https://www.stat.cmu.edu/tr/tr599/tr599.ps),
[journal DOI](https://doi.org/10.1214/ss/1177009939).
The complete author report was retrieved and converted with Ghostscript's SAFER
text device; the introduction and relevant design/prediction sections were read.
Rendered report pages3 and11 verify equation placement.
Section1.2/equations1-2 put the terminal decision inside the expected design
utility. Section2.2/equation6 discusses quadratic loss and distinguishes a
prediction objective from parameter estimation. Our finite binary derivation
below is not the report's normal-linear-model result or an inherited guarantee.

## Own Finite Joint-Model Derivation

Let F be revealed history, j a candidate measurement of its ORIGINAL outcome,
V the requested measurement, and Y_i a declared future binary prediction target.
All conditionals must come from ONE current joint model, not an unrelated
evidence likelihood and outcome kernel. Define

q_i=E[Y_i|F], q_i^v=E[Y_i|F,V=v,j], a_j(v)=P(V=v|F,j).

For predeclared nonnegative target weights omega_i summing to1,

Value(j)=sum_v a_j(v) sum_i omega_i (q_i^v-q_i)^2.

The tower identity gives sum_v a_j(v)q_i^v=q_i. Expanding binary Brier loss
shows Value(j) equals CURRENT model Bayes risk minus expected model Bayes risk
after the measurement. It is nonnegative under that SAME joint model. It does
NOT imply nonnegative risk change under unknown external truth, greedy optimal
batches, correct source independence or an equal-total-cost learning win.

The requested V shares the old latent outcome with W1; future Y_i is a new
target, not W2 relabeled as an independent clean event. An unavailable response
only has zero information if its missingness mechanism is declared ignorable.
Otherwise missingness must enter the joint model, not be silently canceled.

## Tested Algebra Only

[Script](paired-v57-predictive-value-identities.mjs),
[results](paired-v57-predictive-value-identities.json):18 exhaustive joint-table
cases across noise0/.1/.2, three rate pairs and both first measurements verify
the tower/proper-risk identity. Eight cases have ZERO coarse-class concentration
gain but positive predictive value. Example: fixed eta.1, hidden rates.2/.8,
first measurement false; next clean mean.356, risk.229264->expected.224712195,
gain.004551805. This is a finite mathematical example, NOT an EventFrame accuracy
gain. Zero noise or identical rates correctly gives zero value for repeated
measurement. These algebra cases alone establish no empirical benefit.

An isolated Go prediction-value observer is now implemented in
`internal/researchpairedvalue`, leaving the forecast law unchanged. Thirteen
module roots race-pass at the initial unit freeze, including 72 independent
joint-history comparisons and zero-concentration / positive-predictive-value
controls. A 150-origin dense batch measured approximately 7.4-7.7 ms at one
issued trial/member, 15.3 ms at eight, and 23.2-23.5 ms at sixteen, with about
307 KB allocated per query. These are isolated microbenchmarks, not complete
trajectory cost, loaded serving performance, or an external accuracy result.
The new full-fixture reference separately integrates unnormalized histories
and uses proper-risk subtraction; its values also match the independent
72-origin rebuild audit. Full all-regime evaluation is still required.

## Implementation Requirements And Falsifier

Keep the predictive law, existing posterior model and observation provenance
unchanged initially. Freeze target horizons/weights before evaluating; never
read generator rates, unknown future labels or source availability. Conditional
forecasts require original-position joint updates, including member-local state
and changed global weights. Store current per-component target means with a
dependency revision, not stale aggregate means. Cached conditional rows/means
need the same owner/local/global/epoch fences as the law.

Initial comparison should retain random, uncertainty and class-concentration
controls at identical requested-evidence budgets AND measure full total costs.
All generators, source dependence, missingness and delay stresses remain.
Myopic immediate value may be useless when evidence arrives after the target
was served: model/declare availability-time targets or report that limitation
and reject on delayed tests. Do not call immediate model value a latency-aware
value certificate. Batch redundancy also needs measurement, not an optimality
claim from one-step gains.

Computational lead: compute a snapshot of current target/component means once;
other-member hypothetical changes are global-weight reweightings. A bounded
weighted Gram matrix can amortize their squared forecast movement, while the
queried member needs its own hypothetical local-state mean. Prove equivalence
against a dense independent reference and charge construction/rebuild/request
costs before testing8MiB/400ms limits. No claimed speedup from this formulation.

Falsifier: model value remains uncorrelated with real proper-loss/usefulness,
fails external shift protection or exceeds total cost. Preserve such a result.
V54/V55 negatives, strict V56 execution comparison, sealed untouched agent
cohorts and all seven whole research goals remain unaffected/OPEN.
