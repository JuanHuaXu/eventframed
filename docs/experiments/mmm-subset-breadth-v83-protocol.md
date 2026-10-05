# Retained subset breadth v83

Frozen before fresh data. Do not change v82's state, subset prior, model fitting,
mixture weights, gate, audit cadence or observation budget. Exercise ten cases:
stable_uniform, bit_uniform, xor2_uniform, majority3_uniform, mux_uniform,
parity4_uniform, bit_dependent, xor2_dependent, stable_dependent, null_uniform.
The seven nonstationary cases change at256 from parity6/7/8 to the named rule.
Rules are bit2; bit1 XOR bit2; majority0/1/2; bit0?bit1:bit2; parity1/2/3/4.
Each rule can be inspected within the six-coordinate view budget. Keep5% label
noise; null labels are independent fair coins. Dependence sets bit0=bit2 and
bit4=bit1 after drawing a uniform9-bit vector. The candidate STILL assumes
uniform inputs, intentionally testing misspecification rather than fixing it
with generator knowledge. Stable dependent is a mandatory protection case.

For each case fit a fresh4096-label base on its initial regime/input law with
seed2026118300*1e6+case. This base is frozen and shared across both evaluation
phases. Confidence statements condition on that fitted base, not an estimated
distribution over all possible fits. Future stream bases are
2026118301+10*case+phase, phase0/1, with the existing role-separated Seed function
and case index.64 streams per case/phase,512 steps,1280 trajectories total.
Record train seed/configuration and stream seed base explicitly. No phase tuning.

Keep the five v80 controls and retained subset candidate, same audits/labels.
Every changed case requires post-Brier gain>=.005 versus the mixture-gate count
control with paired z3.5 lower>0. Every stable/null case requires full and post
harm upper<=.01. Report case/phase results, do not average away a failure.
Costs, accuracy and absolute Brier remain visible even where relative gates pass.
This is an exploratory breadth screen, not simultaneous universal coverage.

Verify rule truth tables, dependency mapping, attainable feature costs, seed
separation, no parameter/generator access in subsetState, replay/source hashes,
unchanged labels/cost caps, focused race/vet. Preserve any v82 matched pass and
v83 failure as distinct findings. Delayed/missing feedback and actual agent
integration remain separate open work; this study does not claim to cover them.
