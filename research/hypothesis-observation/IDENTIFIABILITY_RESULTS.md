# Copied-report information ceiling

This exact simulator audit changes the next research question; it does not
complete direction7 or validate a production observer.

## Model and lower bound

Use the original 16 uniformly likely hypotheses, four target classes h modulo4,
eight observation types and original likelihood tables. In the fully copied
regime, each type supplies exactly one Bernoulli draw. All later reports of that
type repeat it. First draws are conditionally independent across types given h.

Let Y be the complete eight-report vector. Any adaptive acquisition transcript
made solely from these reports is a function of Y and independent policy
randomness. It cannot contain more target information than Y. Therefore its
expected proper multiclass Brier is bounded below by

`E_Y[1 - sum_c P(c | Y)^2]`.

The analogous minimum classification error is
`E_Y[1 - max_c P(c | Y)]`. These are population quantities under the declared
simulator law, not confidence bounds estimated from the experimental episodes.
The oracle knows the true copied mode and likelihoods. Unknown mode, additional
cost, model misspecification or weaker policies cannot improve this optimum.
New independently informative side evidence would change the bound's premise.

## Exact results

| Regime | Full-report Bayes Brier floor | Classification error floor |
| --- | ---: | ---: |
| Copied05 | 0.094190356 | 0.062597000 |
| Copied20 | 0.471909825 | 0.336260800 |
| Null | 0.750000000 | 0.750000000 |

Both non-null tables have 16 distinct hypothesis likelihood vectors. They are
not structurally identical models: finite noisy evidence, rather than identical
observation laws, limits this experiment. The null table has one common law.
Distinct laws do not imply reliable identification from one eight-report vector.

The copied20 floor explains why near-perfect recovery is unavailable from this
report family. It does NOT explain away all earlier failures. The v10 copied20
full-policy confirmation mean Brier was 0.531646, above the exact floor. That
sample mean versus population optimum is not a paired improvement estimate or
a confidence interval; it merely leaves room for a better policy/model.

## Finite-horizon planning diagnostic

An exact dynamic program enumerates all 3^8 partial-report states, with the
oracle posterior and no repeated queries. At each state it minimizes expected
terminal Brier over remaining observation types. Copied20 optimum for budgets
0 through8 is:

`0.750000, 0.660000, 0.537600, 0.527020, 0.482605, 0.472347, 0.472347, 0.471910, 0.471910`.

Copied05 has a flat optimum between budgets3 and4, followed by improvement
at5. Thus marginal value need not decrease monotonically with budget here;
intermediate nuisance-bit observations can support later useful combinations.
This is a planning lead, not proof that a cheap one/two-step rule achieves the
full optimum or that the noisy source-mode model is correctly specified.

## Verification

[Audit implementation](identifiability-audit.mjs) checks probability
normalization, child-state mass conservation, monotonic optimal risk and the
full-vector endpoint against separate enumeration. It visits 6,561 states per
regime. [Independent Python reference](identifiability_reference.py) imports
the ORIGINAL experiment likelihood, class and Brier functions, enumerates all
256 complete report vectors, and computes direct expected squared loss instead
of the Gini shortcut. Maximum cross-language difference is below 3e-16. Full
audit replay is byte-exact. Sources are hashed in the artifacts.

Artifacts: [exact audit](identifiability-audit.json),
[independent reference](identifiability-reference.json).

## Next work

Separate two aims. First compare a deployable observation planner against this
known-mode optimum at equal acquisition cost to measure remaining policy loss.
Second, for performance beyond this information floor, introduce observations
that actually carry independent information and charge their acquisition cost.
Source IDs alone are not that evidence. Unknown-mode and misleading-provenance
controls remain required; an oracle copied-mode policy cannot replace them.
No old acceptance gate is relaxed, and no high accuracy claim is rescued by
calling an impossible target complete.
