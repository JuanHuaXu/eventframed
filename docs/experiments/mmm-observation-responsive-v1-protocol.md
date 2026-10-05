# Evidence-responsive observation budget, v1

Freeze before collection. The fixed 80/48 reserve rescued a clock-384
change but harmed earlier and delayed shifts. This study tests a budget
schedule triggered by **arrived evidence** rather than a fixed stage clock.
It is research-only; it does not authorize a causal or Anti-Pigeon decision.

Methodological context: [Zliobaite et al. (2011)](https://proceedings.mlr.press/v17/zliobaite11a.html)
show why uncertainty-only stream sampling may miss drift away from a decision
boundary. [Ramdas et al. (2023)](https://arxiv.org/abs/2210.01948)
describe anytime-valid e-process reasoning. The specific monitor and budget
policy below are our hypotheses, not implementations of those papers.

## Fixed data and arms

Use the same parity-before-change and bit2/majority-after-change rules,
5%/20% outcome noise, missingness, delay, and independently fitted 4096-label
baseline as in the [transfer protocol](mmm-learned-contrast-transfer-v1-protocol.md).
Use **new** design/confirmation seed bases 2026102701/2026102702 and fit
offsets 1800/1900. Each case/split has 16 baseline fits x 16 streams,
512 clocks, a shared potential-outcome tape, and exactly128 requested labels
per arm. Eight cases are:

| Case | Change | Noise | Missing | Delay | Contexts | Post rule |
| --- | ---: | ---: | ---: | ---: | --- | --- |
| early160 | 160 | .05 | .20 | 0 | uniform | bit2 |
| late360 | 360 | .05 | .20 | 0 | uniform | bit2 |
| late400 | 400 | .05 | .20 | 0 | uniform | bit2 |
| noisy256 | 256 | .20 | .20 | 0 | uniform | bit2 |
| delayed256 | 256 | .05 | .25 | 32 | uniform | bit2 |
| skew256 | 256 | .05 | .20 | 0 | bit2=1 at .10 | bit2 |
| stable20 | none | .20 | .20 | 0 | uniform | none |
| majority_ood | 256 | .05 | .20 | 0 | uniform | majority(bits0,1,2) |

Arms 0-2 use unchanged v2 random, uncertainty and learned-disagreement
selection. Arms 3-5 use those same policies with the responsive scheduler.
Every arm owns a separate 11-rule working state and sees only its own
requested, nonmissing, arrived outcomes. No evaluator truth or true change
time enters any model, monitor, selector, or budget calculation.

## Frozen monitor and budget rule

The responsive arms monitor a fixed mixture of 8 origin-clock starts
`s in {0,64,...,448}` and the ten known alternatives `h=1..10`. For each
selected label `(x_i,y_i)` delivered at clock `t` with origin `i`, update
the wealth of component `(s,h)` only if `i >= s` by the Bernoulli likelihood
ratio `p_h(y_i|x_i) / p_B(y_i|x_i)`. Before its start a component holds wealth
1. The denominator `p_B` is the independent, frozen *fitted* pre-change
baseline; the numerator uses the declared .05/.95 rule. Average, never
multiply, the 80 overlapping component wealths. The first average >=100
sets an irreversible observation-budget alert. Delivery at clock `t` may
affect requests only from clock `t+1` onward.

Under an **exact conditional null** `p_B=P(Y=1|X,history)` and predictable
selection, this predeclared mixture is a nonnegative martingale, so a
100 threshold has a Ville-style <=1% anytime crossing bound. The fitted
`p_B` here is approximate, and its uncertainty is not certified; therefore
the study claims **no** 1% guarantee. Measure empirical false alerts on
`stable20`. The monitor is not an Anti-Pigeon certificate.

Before an alert, a responsive arm nominates with base rate `.20`. For the
64 clocks *after* its first alert, its base rate is `.60`; then it returns
to `.20`. Use the unchanged v2 uncertainty/disagreement multipliers and
nominate-before-feedback ordering. Never exceed the remaining budget;
force selection when remaining requests equal remaining clocks so every
arm finishes with exactly128. The same monitor and budget rule applies to
all three responsive policies. Record first alert, pre/post-change requests,
arrived-label counts, and the fitted baseline probability at each context.

## Frozen screen

Use the transfer study's expected/realized Brier and recovery rules, with
the declared case-specific change clock. Paired fit-cluster intervals are
`mean +/- 3.5 SE` over16 fits. The responsive learned arm passes only if
**both** new splits satisfy:

1. On `late360` and `late400`, post expected-Brier gain versus original
   learned is >=.01 with positive lower fit-cluster endpoints, restricted
   mean recovery lead is >=10 clocks, and miss fraction decreases >=.10.
   Neither late case is aligned to a monitor start or budget phase boundary.
2. On `early160`, `noisy256`, `delayed256`, and `skew256`, post expected-
   Brier harm upper fit-cluster endpoint versus original learned is <.01
   and first64 post expected-Brier mean harm is <=.005. On `early160` and
   `skew256`, recovery-delay harm upper fit-cluster endpoint is <5 clocks.
   For delayed feedback, report detection and recovery without pretending
   to react before a revealing label arrives.
3. On `stable20`, first-alert fraction <=.05 and full expected-Brier harm
   upper endpoint versus original learned and both responsive controls
   <.01. On `majority_ood`, post expected-Brier harm upper endpoint versus
   original learned and both responsive controls <.01. The responsive
   learned policy's >=.01 post expected-Brier gain with positive lower
   endpoint over both responsive controls holds in >=5 of 6 known cases.
4. All arms request exactly128 labels. Responsive learned mean arrived-
   label gap from each responsive control <=2 per case. Source hashes,
   independent no-future-label/score/monitor audit, exact confirmation
   replay, and isolated cost checks pass. Monitor-plus-update p99 <10us,
   forecast-plus-select p99 <1us, total working state <4KiB.

Failure is retained without tuning the threshold or budget rates on these
cohorts. Passing would support only this bounded finite-horizon synthetic
component, not external-law validity, real agent outcomes, unknown
hypothesis invention, or loaded service latency.
