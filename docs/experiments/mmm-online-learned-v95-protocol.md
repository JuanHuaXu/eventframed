# v95 prospective learned-expert experiment (frozen before outcomes)

Scope: immediate complete feedback, fixed full/mask63 observation views,
generic subset versus skeptical family BMA versus online generic/parity mixing.
This changes v93's static evaluation protocol and cannot erase its failures.
No adaptive acquisition, delayed evidence or real-agent claim is tested.

Two phases, 32 paired streams per case, 12 cases, 256 scored steps and 16
unscored initial labels: 768 streams. Cases are the v93 ten generators plus
majority-to-parity4 and parity4-to-majority at step128. Each phase uses its
disjoint v93 rule pool; switch cases declare separate majority/parity masks.
Seed = 2046119500 + phase*1000000 + case*10000 + index*10.
Roles0/1/2 are rule/input/outcome RNGs. These are new streams; deterministic
case semantics are reused, not old outcomes. There is no parameter sweep.

Fit at steps0,32,...224 from the latest at most64 already observed full frames.
All arms receive the same complete training evidence, including for mask63
evaluation; that evidence cost is not disguised as six-bit-only training.
Predict before obtaining the current outcome. Keep online weights across fits.
Online eta=.5, generic prior.95, rho0 and rho.001; only rho.001 is primary.
Skeptical BMA parity prior.1 is a comparison, not a fallback selected by results.
Views maintain separate online weight states. Each state's feedback is the
same actual outcome, exactly once per prediction. Oracle probabilities enter
evaluation only, never fitting, weight updates or forecasting.

Report all-stream and late-half expected single-coordinate Brier and expected
accuracy at observed inputs, plus realized losses and max prefix-bound defect.
Accuracy is expected against the simulator label law, not a sampled agent rate.
All-stream means include adaptation costs. Partial views retain dependent-input
misspecification. Tape hashes bind training windows and pre-outcome forecasts.

Primary gates (approximate paired trajectory normal bounds, z=3.5):
- Both phases: full-view all-stream Brier gain lower bound >0 and mean gain
  >=.005 for parity3, parity4 and complement4 (six gates).
- Every case, phase and view: all-stream harm upper bound <=.01 (48 gates).
- Both switch directions, phases: full-view late-half gain lower bound >0
  and mean >=.005 (four gates).
All58 gates required; fail is not grounds to change margins on these streams.
Regret checks <=1e-8 are component invariants, not replacements for these gates.
Bounds are approximate simultaneous screening, not anytime confidence sequences.

Run full exact artifact replay, source-hash checks, race smoke, vet. Benchmark
full fixture separately from v94 arithmetic-only cost. No production changes.
