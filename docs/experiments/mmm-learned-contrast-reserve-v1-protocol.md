# Known-hypothesis acquisition: finite-horizon reserve

Freeze before collection. The [transfer study](mmm-learned-contrast-transfer-v1-results.md)
used about 96 of 128 requests before a clock-384 shift and about 32 after.
The unchanged learned selector reduced late recovery misses but missed the
predeclared 10-clock restricted-mean lead. This study tests whether a
**predictable** label-budget reserve, not a hidden-change oracle, improves
that late recovery without losing early-shift or stable performance.

Primary methodological context: [Zliobaite et al. (2011)](https://proceedings.mlr.press/v17/zliobaite11a.html)
document why uncertainty sampling alone may miss stream drift away from the
decision boundary. [Goebel et al. (2026)](https://ojs.aaai.org/index.php/AAAI/article/view/41191)
study explicit budgets and finite horizons. The 80/48 split and all tests
below are EventFrame research choices, not results inherited from those papers.

## Frozen arms and data

Use the transfer protocol's seven cases and generator, unchanged 11-rule
working learner, initial fitting procedure, 512 clocks, and shared
potential-outcome tape. New independent design/confirmation seed bases are
2026102601/2026102602; baseline fit offsets are 1600/1700. Each case/split
has 16 fits x 16 streams. Every arm requests exactly 128 labels. Controls
0-2 are the original random/uncertainty/learned v2 allocation; arms 3-5
apply the reserve schedule to the same three policies. Each arm has its own
posterior updated only with its requested, nonmissing, arrived labels.

At clocks 0..383, the reserve arms can request at most 80 labels and must
finish with exactly 80. At clocks 384..511, they request the remaining 48.
Within a stage, use `remaining stage budget / remaining stage clocks` as
the base rate, with the v2 uncertainty or learned-disagreement multiplier.
Force a request when remaining budget equals remaining clocks. No outcome,
future event, true change time or evaluator-only conditional probability
enters the scheduler. The stage boundary is a declared horizon policy and
is identical for all seven cases; it is not an inferred changepoint.

## Frozen screen

Use the transfer study's post-change expected/realized Brier, first64
expected Brier, 16-consecutive trailing32 expected-Brier recovery rule,
restricted horizon, fit-cluster interval (`mean +/- 3.5 SE`), and per-tick
budget/delivery auditing. The reserve learned arm passes only if **both**
new splits satisfy every condition:

1. On `late_bit2`, restricted mean recovery improves by >=10 clocks and
   recovery miss fraction falls by >=.15 versus original learned. Post
   expected-Brier gain versus original learned is >=.01 with a positive
   fit-cluster lower endpoint. Record pre/post requests and arrived labels.
2. On `early_bit2`, `noisy_bit2`, `delayed_bit2`, and `skew_bit2`, reserve
   learned's post expected-Brier harm upper fit-cluster endpoint versus
   original learned is <.01, and its first64 post expected-Brier mean harm
   is <=.005. On early and skew, its recovery-delay harm upper fit-cluster
   endpoint is <5 clocks. Delayed recovery and misses are reported but not
   required to react before labels can arrive.
3. Reserve learned gains >=.01 post expected Brier versus **both** reserve
   random and reserve uncertainty, with positive fit-cluster lower
   endpoints, in at least four of five known-bit2 cases. On `stable20`,
   reserve learned's full expected-Brier harm upper fit-cluster endpoint
   versus original learned and both reserve controls is <.01. On
   `majority_ood`, its post expected-Brier harm upper endpoint versus
   original learned and both reserve controls is <.01.
4. All six arms request exactly128 labels; reserve arms request exactly
   80/48 by stage. Reserve learned mean arrived-label gap versus each
   reserve control is <=2 per case. Source hashes, no-future-label and
   independent per-tick score audit, confirmation replay and finite cost
   checks pass. Isolated update p99 <10us, forecast-plus-select p99 <1us,
   and learner state <4KiB.

This is a finite-horizon synthetic component. Passing it would not prove
automatic hypothesis invention, optimal budget allocation, an Anti-Pigeon
certificate, agent-level value, unknown-horizon behavior or service latency.
Failure is retained; these cohorts are not used to retune the stage split.
