# Origin-age diagnosis and rescue proposal

Latest outcome: [v108](../docs/experiments/mmm-advice-v108-results.md) implements
neutral-only, age-only and combined candidates. All fail their complete gates.
Age improves recovery but harms stable parity; neutral-only adds little.
The finite discounted-evidence mass cannot easily overcome the strong generic
prior. No half-life retuning is authorized by this consumed result. The next
distinct lead is [log-score advice](log-score-advice-proposal.md), while the
original event-clock fixed-share and conditional-activation leads remain open.

Latest diagnosis: [v107](../docs/experiments/mmm-mixture-v107-results.md)
computes constant-mixture oracles on every consumed v106 delayed switch run.
Early raw-model mixture risk is usually worse than neutral; later bundles
recover. Next test explicit neutral competition separately from origin-aged
selection, with unchanged gate thresholds and fresh controls. No oracle weight
enters the runtime, and no consumed result is a fresh rescue.

Latest outcome: [v106](../docs/experiments/mmm-arrival-v106-results.md) tested
arrival-time selection and FAILS16/318 gates. It partially improves full carry,
but does not meet the recovery target. Next compute a consumed convex-mixture
headroom bound, with and without neutrality, before choosing another mechanism.

Progress: [v105 diagnosis](../docs/experiments/mmm-origin-v105-results.md)
completed with exact parent replay. Prioritize the
[arrival-time selector ablation](arrival-routing-component-results.md) next:
its component is tested, but fresh quality evaluation remains outstanding.
The discounted formulation below remains a separate, unimplemented lead.

Research lead, not implemented or quality-validated. Preserve the complete
[v104 failure](../docs/experiments/mmm-delayed-v104-results.md): retaining every
delivered role loss helps stable parity but worsens delayed switch recovery.

## Establish cause before selecting a rescue

The code confirms two timing facts, not their individual contribution to the
regression: fixed share is applied per processed label, and an old forecast's
loss enters with full weight when its origin prefix finally drains. Missing
heads can postpone many already-arrived labels until after a publication.

Replay consumed v104 switch trajectories with read-only instrumentation. Record
original and release clocks, model version, raw forecast losses, bank weights
before/after updates, original acquired masks and new forecast masks. Require
bit-identical original metrics. Separate loss age, delayed arrival, additional
head-of-line waiting and publication crossing. Check whether old-regime losses
actually move weight toward the currently worse role; do not infer that solely
from a higher aggregate Brier. Simulator truth may score this diagnosis, never
enter the update rule.

## Candidate formulation

Chernov and Zhdanov's [Prediction with Expert Advice under Discounted Loss,
v2 (2010)](https://arxiv.org/pdf/1005.1918), Protocol1 and Sections1-3, explicitly
discount accumulated expert losses and study corresponding algorithms and
bounds. Their protocol reveals outcomes each round. Our missing/delayed
adaptation below is not their algorithm or an inherited theorem.

Let A_t contain outcomes available before prediction t, restricted to origins
i<t. Let l_i,j=(p_i,j-y_i)^2 use the raw role prediction actually issued at i.
A prospective origin-aged selector is:

```text
D_j(t) = sum_{i in A_t} lambda^(t-1-i) * l_i,j
w_j(t) = pi_j * exp(-eta * D_j(t)) / sum_k pi_k * exp(-eta * D_k(t))
```

For a fixed lambda in(0,1], maintain it by an exact recurrence:

```text
D_j(t) = lambda * D_j(t-1)
       + sum_{i newly available before t} lambda^(t-1-i) * l_i,j
```

The available set, not the arrival order within a set, determines these losses.
A late outcome keeps its original age; releasing an old blocked batch must not
make it as influential as a batch of recent evidence. Missing outcomes add no
loss. Updating this selector promptly on arrival can be separated from the
origin-ordered, version-scoped evidence gate. Duplicate outcomes must remain
single-use across both paths.

This rule discounts to the declared prior as evidence ages. That may sacrifice
rare but valid old knowledge and weaken stable-parity gains. It is not a claim
that age implies falsity, nor does it repair informative missingness or calibrate
a forecast. It is an online choice among stable learner roles, not a posterior
over objective truths. Carry across structural generation changes stays outside
the contract until role identity is explicitly preserved or reset.

## Required ablations and checks

Compare the unchanged conservative and full-carry controls with (a) moving only
fixed-share transitions to the event clock, (b) origin-discounting the losses,
and (c) separating arrival-time selector updates from gate-prefix release.
Use matched-acquired-mask diagnostics to separate selection and acquisition
effects before claiming the learner itself improved. Do not silently change
all three mechanisms and attribute the result to one.

For any new kernel, compare streaming state with literal batch sums across
arrival permutations, idle clocks, late batches, duplicates, expiry and model
publication. State exactly which immediate limit is reproduced: lambda=1 with
no fixed share is ordinary exponential weighting, not the existing .001
fixed-share policy. Existing regret and test-martingale claims do not transfer
automatically. Keep the current gate's stale-version rejection unchanged.

Only then freeze new constants, seed allocation, gain/non-harm gates and delay
controls before fresh quality evaluation. No half-life sweep on the consumed
v104 confirmation outcomes. The next study must still protect stable cases and
test both delayed change directions, not merely improve one favorable example.
