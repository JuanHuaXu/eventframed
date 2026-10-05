# Conditional Anti-Pigeon evidence power v1: frozen feasibility protocol

This is a research-only feasibility screen for Goal 3, not an MMM integration
or a certificate for an unknown EventFrame target law. It tests whether the
fixed two-context audit certificate from the earlier toy can actually fire
under reference and feedback budgets closer to the current MMM harness.

## Frozen world and roles

- Context `X` is an independently sampled fair bit. Reference outcome laws
  are `P(Y=1|X=0)=0.9` and `P(Y=1|X=1)=0.1`.
- Live cases are `stable` (same law), `swapped_from_start` (the two laws
  reversed from clock 0), and `swapped_at_256` (same until clock 255, then
  reversed). Horizon is 512 live clocks.
- Compare `batch_reference`, with 256 independent pre-live labels per cell,
  against `online_reference`, with only independently nominated reference
  audits as the stream proceeds. The batch costs 512 extra labels and may
  not be treated as free. Both strategies receive the same live stream within
  a trial.
- Compare `immediate` (25% independent audit nominations, no missingness or
  delay) against `sparse_delayed` (25% nominations, 20% independent missing
  audits, uniform delay 0..31). Both reference and live online streams use
  the same schedule; batch labels are immediate before clock 0.
- Design and confirmation have separate seed bases `2026100301` and
  `2026100302`, each with 1,000 trajectories in every strategy/schedule/case
  cell. All random roles are independent. No candidate sees the generating
  probabilities; only the evaluator uses them to generate labels.

## Frozen certificate

For each context and each reference/live stream, maintain counts and successes
from unique *arrived* nominated audits. A missing, future, un-nominated or
duplicate audit is ineligible. Let `n` be that stream/context's arrived count.
For `n>0`, the radius is

`rad(n) = sqrt(log(8*(T+1)/delta)/(2*n))`, with `T=512`, `delta=0.02`.

The radius covers two-sided Hoeffding deviations over two contexts, two
streams and at most `T+1` repeated checks by a union bound, conditional on
the fixed contexts and outcome-independent audit/arrival process. A batch
reference count is fixed before live time; using the same conservative radius
is valid but may be loose. Flag the first clock where, for either cell,

`abs(mean_live - mean_reference) > epsilon + rad(n_live) + rad(n_reference)`,

with `epsilon=0.10`. Once flagged, latch the first clock. There is no learned
partition, source-authentication guarantee, Bayesian nomination or forecast
update. In `swapped_at_256`, the gate uses all arrived live evidence from
clock 0; any failure caused by prefix dilution must be preserved.

## Measures and decisions

Report per cell: flags/1,000, detection-clock median/p95 among detected
trajectories, missed shifts, mean reference/live nominated, delivered,
missing and pending labels at horizon, and full experiment wall time. Store
all trial rows with source/protocol hashes and verify them independently.
Clock-based detection must not be called a forecast benefit.

A strategy/schedule cell is feasible for *initial* separation only if stable
flags are <=20/1,000 and `swapped_from_start` flags are >=800/1,000.
For delayed-onset separation, require `swapped_at_256` flags >=800/1,000
and no flag before clock 256 in that case. These are empirical power screens,
not mathematical coverage claims; the finite-horizon false-flag bound is
conditional on the stated iid/selection assumptions. No thresholds or seed
bases may change after design results. Passing this screen would still leave
actual MMM nomination, forecast Brier, simultaneous active-bucket error
allocation, dependent sources and serving costs unproven.
