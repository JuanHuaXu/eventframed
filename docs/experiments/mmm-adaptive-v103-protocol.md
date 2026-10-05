# Adaptive routed observation v103: frozen protocol

Freeze before evaluating any fresh v103 trajectory. No interim quality looks,
retuning, seed replacement or continuation selected from the observed result.
This is a finite synthetic experiment, not production or real-agent evidence.

## Design

Two phases, 32 independent trajectories for each of the same 12 v102 cases:
parity1/2/3/4, complement4, majority3, mux3, constant, null, dependent4,
majority-to-parity and parity-to-majority. 768 trajectories total. Generator,
phase-specific rule pools, 16 initial labels, 256 scored steps, switch at 128,
32-step model publication cadence, 64/32 sample windows and immediate feedback
are unchanged. Seed = 2090110300 + phase*1000000 + case*10000 + index*10;
roles 0/1/2/3 are rule, input, outcome and random-view RNG. Check these effective
Go seeds against the archived v90-v102 learner and null-family allocations
before generating results. Distinct seeds are not a proof of statistical
independence across the entire research history or of independent generators.

Each publication fits generic64, Boolean64, generic32 and Boolean32, with the
same uniform input law; dependent4 deliberately violates that input assumption.
All arms receive identical full nine-field training audits after forecasting.
Those 272 full frames per trajectory (2448 coordinate observations including
initial training) are additional to foreground cost, not hidden in its cap.

Arms:
0. Generic64 and its archived one-step observer.
1. Routed law on arm0's acquired masks, with its own feedback state.
2. Routed law with coherent one-step acquisition.
3. Routed law with exact bounded terminal-entropy lookahead.
4. Routed law with uniformly random affordable scope/depth views.
5. Routed law on fixed mask63.

All routed arms use unchanged v102 bank, falsification and routing rules and
independent state. Adaptive observers force view(0,0), then stop at probability
<=.1 or >=.9, or a six-coordinate cap. Random selection uses only available
budget and the current forecast for stopping, never hidden fields. Fixed mask63
always costs six. Lookahead uses the same fitted joint law for branch weights,
terminal entropy, confidence stopping and its scored forecast. It does not use
simulator truth to plan. Full reference-model truth and sampled outcome become
available to scoring only after all six forecasts exist.

## Frozen quality gates

Each comparison is paired over 32 trajectories within phase/case/segment.
Interval = mean +/- 3.5*sample_SD/sqrt(32). This is a fixed-sample approximate
normal interval, not an exact, anytime or history-wide guarantee. Segments are
all256 and late128. Smaller Brier is better; use expected Brier under simulator
truth, and report realized Brier and expected accuracy separately.

- 48 gates: lookahead non-harm versus generic observer for every phase, case and
  segment; upper harm <=.01.
- 48 gates: lookahead non-harm versus routed one-step on the same combinations.
- 6 gates: all-step lookahead gain versus generic on parity3/4 and complement4,
  each phase; mean gain >=.005 and lower bound >0.
- 4 gates: late lookahead gain versus generic on both switch cases, each phase;
  mean gain >=.005 and lower bound >0.
- 2 gates: late parity4 lookahead gain versus routed one-step, each phase;
  mean gain >=.005 and lower bound >0.

Overall PASS requires all 108 quality gates, exact fixed-mask compatibility on
consumed v102 streams, and all lifecycle/data-integrity checks. Failed or
inconclusive gates remain failed; do not replace these criteria after results.
Other arm comparisons are diagnostics, not additional claims of confirmation.

## Integrity and cost

Before fresh generation, reproduce v102 routed mask63 metrics for consumed
design cases 0,3,10,11 at index0, from the same old seed allocation. This is
compatibility evidence only, not new quality data. Unit checks cover literal
lookahead values/actions, hidden-field independence, epoch rejection and
atomic issuance/feedback. Store every scored step's predictions, masks, costs,
input, truth, outcome and planning node/branch counts. Reconstruct metrics
independently; replay every record deterministically. Hash source and protocol.

Report all arms' foreground cost, extra training cost, planning work and
isolated benchmarks. Required cost invariant: 1<=paid coordinates<=6 and paid
cost equals unique acquired coordinates, except the matched arm shares arm0's
physical reads rather than paying them a second time. No serving-latency or
acquisition-cost superiority gate is claimed from this first integration run.
Approximate-null results for fixed views do not establish adaptive calibration;
this experiment tests quality and cost, not a new population null certificate.
