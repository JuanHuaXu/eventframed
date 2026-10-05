# Delayed switching and local loss budgets: partial rescue

Contracts: `mmm-spike-fixed-share-v1-contract.md` and
`mmm-spike-ledger-ablation-v1-contract.md`. Same672 consumed trajectories and
172,032 forecasts, no model refits or rate grid. Alpha_t=1/t and eta=1 were
frozen before scoring. The origin-aware filter follows the declared transition
family in [Korotin et al., Section4.2](https://arxiv.org/html/1902.10433).
Permanent missingness and guarded predictions do not inherit that paper's
regret guarantee. This is not ordinary Bayesian inference about physical truth.

| Expert weights | Loss ledger | Whole expected Brier | Terminal64 harm versus Markov | Terminal64 records harmed>.01 |
| --- | --- | ---: | ---: | ---: |
| Static continuous | Global | .155041378 | +.000621136 | 24 |
| Switching | Global | .154381275 | -.000057377 | 21 |
| Static continuous | Reset32 | .154770655 | -.000026168 | 3 |
| Switching | Reset32 | .154433164 | -.000261524 | 2 |

Markov whole-stream Brier is .156448819. Resetting both expert weights and
ledger every32steps scores .154441468. The combined switching/local-ledger
policy is essentially tied with that control over the whole stream: difference
-.000008304, exploratory clustered interval [-.000073626,+.000060645]. It is
not a demonstrated whole-stream improvement over reset32.

Terminal64 performance is better: difference -.000447252 versus reset32,
interval [-.000545426,-.000355842], and -.000261524 versus Markov, interval
[-.000427827,-.000097632]. Intervals resample eight trajectory-index clusters;
they are pointwise, exploratory and based on consumed data, not simultaneous
safety certification or prospective confirmation.

The two-by-two experiment separates two mechanisms. Switching weights improves
the average, while restricting loss credits to fixed32-frame blocks accounts
for most of the reduction in large terminal harms. Neither ingredient alone
establishes complete non-harm. The local ledger resets only guard accounting;
expert weights still use feedback arriving from earlier blocks.

## Remaining failures

Combined policy has no whole-trajectory harm>.01 and no scenario-window mean
above.01 across all eight32-frame windows. However,27 individual window records
still exceed.01, worst .036710914 at the window starting32. This is not an
all-case safety result. The two terminal64 counterexamples are phase0/index4/
additive-abrupt/immediate (+.011502980) and phase0/index3/additive-gradual/
immediate (+.011228127). Do not tune parameters specifically to erase them.

For comparison, static/global has171 individual window records harmed>.01;
switching/global181; static/local30. The source-aware local ledger helps the
harm pattern more than switching alone, but mean and worst-case behavior are
different criteria. Global and each aligned local realized prefix bounds pass.
No arbitrary sliding-window or conditional expected-score bound is claimed.

## Verification and performance

- 384 small-prefix exact latent-path checks pass; maximum filter discrepancy
  3.89e-16. Tests cover delayed/current/missing labels, alpha=0 equivalence,
  equal experts, and the incorrect naive-arrival-update negative control.
- Constructed banked-credit test shows why a global ledger permits late harm;
  local budgets preserve the same proposed weights and bound aligned blocks.
- Invalid external weights are rejected. Static baseline regression reproduces
  every prior forecast and score. Both local cells and switching/global replay
  exactly at the per-trajectory forecast/score level. Added metadata differs
  from older files; this is not a claim of whole-file byte identity.
- A JavaScript exponentiation-parentheses syntax error was fixed before any
  tape run; it did not change the frozen mathematical rule.
- Warm256-step fixture: static/global .165-.168ms/trajectory;
  switching filter alone .380-.384ms; switching/global .547-.568ms;
  switching/local .397-.399ms, or1.55-1.56us/forecast amortized. These are
  O(T^2) reference methods, not loaded serving latency or per-request tails.
- Full research postprocessing took1.393s for switching/local, including
  source reads and controls. That scorer also computes a discarded global
  guard when preparing local weights; the component benchmark avoids it.
  All132,592 original fits remain chargeable; one capped source fit remains.
- No Go/runtime code changed this round, so no new Go race run was performed.
  Prior full model race checkpoint is83.679s PASS. No production changes.

SHA256:

- Filter: `91e29bb377034715f3dfaca66476fd3f07e234f502825709d6828ffbcd425a09`
- Budget core: `a16ac6e004560686a2f734d3e176174c4e0abb1ec2514ecf3ce6adfe6915491f`
- Scorer: `4fedc182b3b847720597ba2eb11b397891863870def0eeab08a414fce2b36a26`
- Switching/global: `f256bacf07ce009ffdda7223ab9d24bdc993f059fe951df1267fafcfb523fda6`
- Static/local: `0bb9cc6de66390b52d96e220242dbf0e8c46a4bf340de45dfc2b594c36f1c9cc`
- Switching/local: `2e9c559e3e1b38a09b1ff4839a275778794f697fe21348426c676b3ae1999eab`
- Combined summary: `7a8d4f66353ed48028239c3ce57b49071ad4a33eae376f73e6dcc914b0d02560`
- All-window audit: `38a94862d92e196e1fb37f3c37b6c3492e50fe137bd2b9ffb63b38adf0753e09`
- Benchmark: `bf28823b589ae4fc005c3b476c5f28415b9cf35183aeabfbf44cd2145a0944d8`

Next: freeze the combined policy for independent fitting/outcome draws and
shift timing, retaining reset32 and Markov controls. Broaden rather than tune
to the remaining errors. Prospective agent-task validation and loaded serving
tests remain separate. All seven goals stay OPEN.
