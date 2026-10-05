# v111: forecast-compatibility handoff, frozen quality protocol

Freeze before fresh generation. Keep the complete v109 experiment: 12 cases,
two phase-disjoint rule pools, 32 trajectories per cell, paired immediate and
delay 0..31/missing .2 schedules. This is 768 latent trajectories and 1,536
schedule-runs. Initial 16 labels, 256 scored frames, changes at 128, as-of
long64/short32 fits every32, six-coordinate acquisition and full training audits
are unchanged. Both phases are evaluated without interim candidate tuning.

Fresh base2140111100 + phase*1000000 + case*10000 + index*10, roles0..4 for
rules/input/outcome/delay/missingness. Audit effective seeds against archived
learner and null allocations through v110 (v110 reuses v109). No optional
sample-size enlargement, seed replacement or parameter sweep after inspection.

## Nine arms

0 generic;1 conservative;2 full carry;3 arrival Brier;4 Brier neutral;
5 log/no-neutral;6 log/neutral. All seven controls must match frozen v109
exactly on consumed compatibility trajectories, including outputs and clocks.
7 publication handoff;8 publication plus pending-loss transfer. Both new
variants use log/no-neutral's original prior, .001 share, models, gate and
acquisition. They add no neutral expert, relaxation of rejection or oracle reset.

For each role and successive models, under their unchanged common input law:

```text
a_j = E_mu[sqrt(p_old p_new) + sqrt((1-p_old)(1-p_new))]
gamma_j = a_j^32
w_new,j proportional to prior_j * exp(gamma_j * log(w_j/prior_j))
```

The exponent32 is fixed before this run. Use normalized log weights and
preserve zero-prior support. Identical laws are an exact identity. Changed
input measures are rejected. This compatibility modulation is heuristic,
not a reuse certificate, ordinary posterior or effective-sample-size theorem.

Arm8 also multiplies compatibility across publication boundaries from the
issue version to the arrival version, per role. It attenuates the fixed-reference
log likelihood ratio `log(P_issue(y)/.5)` before the unchanged .001 share.
Arm7 retains the original issued log update without this attenuation.
The .5 reference is not a competing forecast. Old outcomes are never scored
under a new model or admitted to that model's comparative tests.

Advance clock, release earlier labels, expire unresolved age>=32 entries,
then publish as-of models. Acquire all nine forecasts before sampling the
current outcome; deliver zero-delay labels afterward; flush clocks256..287.
Brier is primary; expected/realized log loss and accuracy remain secondary.

## Gates:718 per candidate,1,436 total

The paired mean +/-3.5 standard errors over32 trajectories remains an
approximate fixed-sample screen, not a confidence sequence, exact coverage
result or a guarantee across the whole research history. Paired schedules
are not independent trajectories.

- 672 non-harm gates: each control0..6, every phase/case/schedule/full-or-late
  segment. Upper Brier harm must be <=.01.
- 20 generic gains: full parity3/4/complement4 and late both switches in both
  phases and schedules.
- 6 conservative gains: delayed late parity4 and both switches, both phases.
- 4 gains against each control2..6: delayed late both switches, both phases.
- Every gain requires mean>=.005 and lower>0. No Brier gate is replaced by
  accuracy or log-score improvement, and favorable cells cannot be combined
  across different candidates to declare success.

A candidate passes only all718 gates plus integrity. All v109 requirements
remain; protection and recovery against its two log policies are additional.

## Integrity and performance

Require component/race tests, seven-control equality on consumed cases, full
replay, source hashes, independent per-frame score/cost, label clock, as-of fit
and journal-accounting reconstruction. Component source remains unchanged
after its successful checks. No model selection using the simulator's truth.

The paired component performance batch is already recorded separately. Any
whole-fixture benchmark must run only after other research processes stop;
include fits and nine-arm bookkeeping explicitly. Neither benchmark proves
database serving latency. No production calls, new private data, deployment,
commits, pushes or whitepaper promotion. Even a complete finite-screen pass
does not close all seven research directions or establish real-agent utility.
