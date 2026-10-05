# Origin-aged and neutral advice: component contract

Research-only ablation following the consumed v107 mixture diagnosis. Not an
ordinary Bayesian posterior, a confidence certificate, or a deployed policy.
No changes to frozen v106 files or production behavior.

Five fixed roles are neutral, generic64, Boolean64, generic32, Boolean32.
The no-neutral prior is (0,.95,.05/3,.05/3,.05/3); the neutral variant reserves
.05 for neutrality and scales those four raw prior masses by .95. This is a
declared challenger allocation, not a fit to v107 oracle weights.

The non-aged control uses eta=.5 squared-loss exponential updates, with .001
fixed share toward its prior after each arrived label. It is a five-role
extension of the frozen four-role bank. The age variant instead uses:

```text
D_j(t) = sum_{i received by t} 2^(-(t-i)/32) (p_ij-y_i)^2
w_j(t) proportional to pi_j exp(-.5 D_j(t))
```

At prediction t, every included origin is strictly less than t. A zero-delay
outcome at t is admitted only after prediction, with age zero; subsequent
event-clock advancement discounts it. Newly revealed old losses retain their
origin age. The 32-frame half-life is fixed to one model publication interval,
not selected by a quality sweep. It is a candidate, not an optimal timescale.
There is no fixed-share step in this mode: discounting already returns stale
loss weights toward the prior. The corresponding constant-clock update order
is immaterial up to floating-point roundoff. Missing labels add no loss.

Direct neutral weight stays neutral. The remaining mass is normalized into
the existing four-expert evidence-routing gate, then scaled back. Rejected
models cannot receive weight through the direct-neutral path. The coherent
mixture drives both observation acquisition and the actual issued forecast.
Publication never acts as an oracle declaration of regime change.

The journal retains the frozen 256-issue horizon, 64 outstanding entries,
32-step publication cadence and explicit expiry. Single-use delivery is
validated before advice updates; late version losses may train stable advice
roles but not the new-version evidence tests. Advice updates on arrival while
tests retain prefix ordering. The role-stability assumption does not authorize
carry across arbitrary Anti-Pigeon generation changes. Delayed/censored
martingale or regret guarantees are not inherited from immediate feedback.

Required component checks: literal discounted batch sums; arrival permutation;
idle time; no-age/no-neutral compatibility; coherent law and gate rejection;
late versions; expiry, duplicate, invalid input and reentrant failure atomicity.
Then compare original arrival, neutral-only, age-only, and their combination
on separately frozen fresh quality streams with stable-case and both switch
protections. Keep prior failures and the original conservative/full-carry arms.

Primary motivations:
[Chernov and Zhdanov (2010)](https://arxiv.org/abs/1005.1918) for discounted
expert losses; [Joulani et al. (2013)](https://proceedings.mlr.press/v28/joulani13.html)
for explicit delayed feedback. This bounded empirical adaptation is not either
paper's proven algorithm.
