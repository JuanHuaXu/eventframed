# Imperfect and changing input dependence

**Improvement screen FAIL; non-harm screen PASS.** The unchanged held-out
forest does not establish robust adaptive benefit on this broader input family.
All seven research goals remain open. This is a research result, not a runtime
bug report or authorization to tune thresholds on these data.

## Protocol and integrity

The [frozen contract](mmm-forest-dependence-v1-contract.md) covers 384 fresh
trajectories: 12 cases, 16 independently fitted initial models per case, and two
cohorts. Each trajectory has 512 forecasts. Changed cases combine bit or XOR2
outcomes with noisy copied fields, dependencies appearing, disappearing, or
reversing. Stable and null controls are retained.

The candidate, fitting schedule, Anti-Pigeon gate and label availability remain
unchanged. The fixed-observation arm uses the control's observations; the coupled
arm chooses its own. Neither receives the generator's change clock.

Before collection, seed auditing found that 12 case indices would overlap under
adjacent cohort bases in the legacy seed formula. Cohort spacing was increased
to 100, and exhaustive case/trajectory/random-role uniqueness was checked.
No experimental outcomes were collected before that correction.

Artifacts:
- [Raw observations and source snapshots](mmm-forest-dependence-v1.jsonl)
- [Summary](mmm-forest-dependence-v1-summary.json)
- [Independently regenerated summary](mmm-forest-dependence-v1-summary-replay.json)
- [Evaluator](../../research/forest-dependence-summary.mjs)

Raw SHA-256: `4f102bf924a448140b8806dcecc5c41e6706f0be9d5818cacd560dd116fcd07d`.
All 832 captured source hashes match their snapshots and current source files.
Summary regeneration is byte-identical. The evaluator checks cell counts,
training-seed uniqueness, costs, fit accounting, shared split times and tape
digests. These checks do not substitute for replaying every forecast.

## Results

Both fixed-observation and coupled arms pass 24/24 full/post non-harm screens
and 0/20 changed-case improvement screens. Every coupled changed-case gain
interval includes zero. Non-harm allows up to .01 Brier degradation; passing
does not prove equality or an absence of harm.

Second-cohort coupled post-change Brier gains (control minus candidate):

| Input condition | Bit gain | XOR2 gain |
| --- | ---: | ---: |
| Copy noise .1 | .001635 | .001843 |
| Copy noise .3 | -.000166 | .000556 |
| Dependency appears | .000818 | .000084 |
| Dependency disappears | -.000799 | -.000167 |
| Dependency reverses | -.000129 | .000089 |

For noisy-copy .1, paired intervals are [-.000488, .003758] for bit and
[-.001534, .005220] for XOR2. The frozen gain gate requires mean >= .005 and
a positive lower bound. Intervals are exploratory mean +/- 3.5 standard errors
over 16 trajectories, not simultaneous confidence sequences or population
guarantees. See the summary for every interval, including stable/null controls.

The maximum absolute mean foreground acquisition-cost change across the 24
cells is .044556 coordinates per frame. This is an acquisition count, not
wall-clock latency or total learning compute. No fresh serving benchmark was
performed in this screen.

## Decision and next boundary

Preserve the candidate and negative results; do not promote or retune it.
Failure in the fixed-observation arm means changing acquisition alone is not
supported as a sufficient rescue by this experiment. It does not identify a
unique cause: estimator sensitivity, finite evidence and mixed pre/post-change
windows remain distinct hypotheses.

Immediate feedback was used throughout. Delayed-feedback robustness remains
untested for this candidate. Existing delayed-role-carry experiments already
show that retaining late losses can harm regime recovery, so a delayed test
must preserve issued forecasts and availability timestamps rather than merely
delaying a call on a single-pending-state learner.

Collection completed successfully in 67.17 seconds (Go package 67.355 seconds).
The precollection contract race test passed in 2.258 seconds; package vet passed.
On resumption, source integrity and summary reproduction were checked again.
No production, whitepaper, remote, commit or push changes were made.
