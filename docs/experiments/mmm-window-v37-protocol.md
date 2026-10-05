# Rolling/adaptive windows V37: frozen protocol

Freeze before either outcome cohort. Bases2026103703 design/2026103704
confirmation. Seed=base+geometry*100million+regime*1million+world*1000.
Two V35 geometries/eight regimes/16 worlds per cell:256 worlds/split.
Each world supplies16 rounds of150 genuine trials. Nomination is fixed
round/member order, same outcomes for every arm, never true-rate routing.

Regimes in order: aligned, independent, curved, baseline_matched, abrupt,
late, recurring, gradual. First four reuse V35 mean/rate generators on NEW
seeds. Last four start with the independent .2/.8 member distribution:

- abrupt swaps p to1-p at round8;
- late swaps at round12;
- recurring swaps at rounds4,8,12;
- gradual linearly interpolates p->1-p over rounds4..12, then holds.

All16 outcome rounds use seed+202. Explicit per-round truth is evaluator
data only. Seed+101 continues to define member rates, not an observation
oracle. Pure conditional Bernoulli trials; noise/selection are separate V36
negatives, NOT presumed repaired by windows.

Five modes: full64, fixed4, fixed8, fixed16 and adaptive. Adaptive priors
(.85,.05,.05,.05), widths(64,4,8,16), share1/600 are frozen. Every mode
uses the identical V36 identity ledger and original-forecast scoring, with
pending cap2400. Schedules immediate/fixed150. No arbitrary arrival-order
window: each child keeps its newest ARRIVED ISSUE ORDINALS. Adaptive expert
weights still process every original issued loss, including old late labels;
that can move today's weights toward stale performance and is a testable
limitation, not ignored evidence.

Original issued expert rows and every receipt are retained. Snapshot after
every150th issue tick149,299,...,2399 and after final drain. Coincident
final snapshots stored once. Current future truth is the round at the
snapshot tick, clamped to final round after drain. All setup/schedule/Issue/
Resolve/snapshot work is measured separately; scoring is evaluator work.

## Frozen Screen

Primary adaptive adoption component requires EVERY cell in BOTH splits:

- Stationary first4 regimes: adaptive protects full64 whole/priority issued
  expected Brier and final packet usefulness with paired lower>=-.01.
- Shifted last4 regimes: mean whole/priority issued expected Brier gain over
  full64>=.01 and paired lower>0. Whole and priority metrics are distinct.
- Abrupt/late/recurring: restricted recovery delay improves>=10% over full64
  with positive paired lower. For each changed phase, recovery is first TWO
  consecutive round-end snapshots with future Brier<=.20 and usefulness>=.75.
  Delay is completed round count since change; unrecovered phases contribute
  phase length+1. Use each world's average across declared phases, not phases
  as independent worlds. This is a finite empirical metric, not a certificate.
- All cells: final snapshot whole/priority Brier and usefulness protection
  versus full64 lower>=-.01. Fixed4/8/16 results are retained as controls,
  not used to choose a winning width after confirmation.
- No math/identity/replay/future errors. Adaptive constructor<=8MiB;
  per-arm accounted learner work<=400ms for2400 labels. This larger bank
  limit is declared up front, NOT a serving or old-study cost relaxation.

Means+/-3.5SE use16 independent worlds/cell; no time-uniform AP coverage,
post-selected significance or no-delay regret theorem is claimed. A pass
would support only this frozen bounded window-bank component, not whole
goals1/2/4/6 or unrestricted continuous learning. A failure is retained.

## Verification

Freeze V35/V36 sources plus window model/tests/collector and both new
protocol/preflight files. Independent batch Beta/atom arithmetic reconstructs
each child's retained set from arrived identities, without using its stored
counts/masks. Reconstruct adaptive weights from privately issued expert rows
and original resolved losses, then every scored mixture law. Verify all
original issuance/receipts/costs/metrics/recovery and exact fresh RNG replay.

Future flips preserve laws before changed feedback arrives. Reject changed
source, future/issued laws, expert rows, receipt identities, weights, retention,
costs and risk artifacts. Input/cap/cancel/owner/epoch/time, underflow and
atomic all-child commit controls must pass. Measure microbenchmarks including
benchmark state restoration. Existing sources/cohorts stay sealed; no
production deployment, private data, whitepaper edit or push.
