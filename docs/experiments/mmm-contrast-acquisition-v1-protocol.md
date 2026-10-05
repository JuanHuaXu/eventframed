# MMM contrast acquisition v1: frozen protocol

This isolated study follows the failed [bit-2 transfer](mmm-member-pair-feasibility-v1-results.md).
It asks whether a *declared* baseline-versus-local hypothesis contrast can
focus the same acquisition budget on informative contexts, improve an
unchanged working forecast updater, and provide the existing bounded
paired gate with enough matched evidence. This is not an automatically
learned hypothesis family or an actual member-state split.

Use the existing `integrationLabel` generator, five scenarios (`stable`,
`member_shift`, `common_shift`, `recurring`, `null`), and 512 clocks.
Each independent trajectory has a pre-generated, shared potential-outcome
tape across three arms. The policy receives only the current nine-bit
context, its frozen baseline probability and its own pre-outcome state;
it never sees current or future outcomes, missingness, delay or future
contexts. The live context is acquired at all 512 clocks in *every* arm.

The two declared hypotheses are `B`: the frozen fitted baseline model's
full-context probability, and `L`: the local bit-2 rule with 5% label
noise (`p_L(x)=0.95` when bit 2 is one, otherwise `0.05`). Let `q(x)`
be the parity of bits 6,7,8. A contrast context has `q(x) != bit2(x)`.
This is an oracle-supplied candidate rule for the synthetic generator;
no claim is made that EventFrame constructs it from real events.

Every arm nominates exactly 128 clocks online, before labels. At clock
`t`, let `k` be nominations still required and `n=512-t`. If `k=n`,
nominate; if `k=0`, do not. Otherwise:

- `random`: nominate with probability `k/n`.
- `uncertainty`: nominate with probability
  `min(1, (k/n)*u(x)/mean_u)`, where `u(x)=4*p_B(x)*(1-p_B(x))` and
  `mean_u` is the frozen average over all 512 contexts.
- `contrast`: nominate only a contrast context, with probability
  `min(1, 2*k/n)`; the forced `k=n` rule ensures an exact budget.

For every nomination, all arms inspect the same four pre-generated
reference-context candidates (four charged reads). Choose the first
candidate whose `(parity,b2)` class matches the live context; if none
matches, choose the first but do not feed it to the paired gate. Request
one live and one reference label per nomination, so every arm has exactly
256 label requests and 512 reference-context reads per trajectory. Charge
missing labels and delayed labels as requests. Independently per side,
labels are missing with probability 0.20 or otherwise delayed by an
integer-uniform 0..31 clocks. Record arrived labels separately.

For a matched **contrast** pair with both labels arrived, set
`Z=(2*bit2-1)*(Y_live-Y_ref)` in `{-1,0,1}`. The unchanged bounded
candidate uses starts `0,64,...,448`, positive and negative signs,
`lambda=0.8`, `epsilon=0.1`, and flags when mean wealth reaches 50.
Only complete same-origin matched pairs update the gate; partial labels
never enter its state. The frozen one-pair null is conditional equality
of the two outcome laws within the declared class under the generator's
independent noise and outcome-blind selection/arrival. Historical flags
are not current-law certificates after a recurring regime reverts.

Every arm uses the same fixed *working* live forecaster. Before the
clock's outcome or due feedback, predict with
`p_t=(1-sigmoid(L_t))*p_B(x_t)+sigmoid(L_t)*p_L(x_t)`, initially `L_0=0`.
At the end of each clock, for each arrived nominated live label from
origin `o`, update `L <- clip(0.95*L + log(P_L(y_o|x_o)/P_B(y_o|x_o)), -8, 8)`.
This discounted log-odds statistic is not an ordinary Bayesian posterior.
Reference labels and gate flags do not update this forecast law. Score
every issued live forecast against the outcome at that clock, including
unnominated clocks; compute full, post-256 and first-64-after-256 Brier.

Use independent seed bases `2026100801` design and `2026100802`
confirmation, 1,000 trajectories per scenario/split, with the same tape
across arms within a trajectory. Preserve every arm/trajectory row and
source hashes. Primary success requires, in *both* splits:

- On `member_shift`, contrast-minus-random and contrast-minus-uncertainty
  Brier gains at least 0.02 for post-256 **and** first-64, each with
  paired mean-minus-3.5-SE lower bound above zero.
- Contrast gate detects at least 800/1,000 member shifts by clock 511,
  with zero prechange flags; stable, common-shift, and null flags each
  at most 20/1,000.
- Stable full-stream Brier harm versus each control at most 0.01 and
  absolute difference in mean arrived-label counts at most two.

Report all three arms and all scenarios, including recurring failures,
conditional flag clocks, matched/delivered pairs, arrived labels,
requested labels, context reads, fit and experiment wall time. No
threshold or policy is altered after a split is inspected. An
independent verifier must reconstruct summaries and cost invariants;
the confirmation split must replay exactly apart from timing. Passing
would show only a known-hypothesis synthetic acquisition component,
not Goal 7 completion, target-law diameter, or deployed MMM benefit.

Research context: [Golovin, Krause and Ray](https://arxiv.org/abs/1010.3091)
motivates discrimination between competing hypotheses, but this
selector is not EC2 and inherits none of its approximation guarantees.
