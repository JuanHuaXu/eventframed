# Delayed-evidence V36: frozen component protocol

Freeze before either outcome split. Bases 2026103603 design/2026103604
confirmation; seed=base+geometry*100million+regime*1million+world*1000.
Two V35 input geometries, six regimes, 16 worlds/cell: 192 worlds/split.
Regimes: aligned, independent, curved, mean_shared, symmetric_noise,
mid_shift. First four reuse V35 generators on NEW seeds. Noise starts with
independent .2/.8 rates and flips each observed label with probability .1
using seed+404. Shift starts with independent .2/.8 and swaps rates after
eight rounds; redraw the 16 rounds using seed+202. These two deliberately
challenge the stationary clean-label likelihood, not an automatic rescue.

Issue all 16*150 trials in round/member order at ticks 0..2399. Every arm
has identical trials/outcomes, one/member/round, with pending cap 2400.
This fixed nomination does not establish equal-total-cost Goal7 sampling.
Schedules, chosen before outcome analysis:

- immediate: delay zero;
- fixed150: delay 150 ticks;
- uniform299: iid integer delay 0..299 using seed+505;
- burst300: feedback for each block arrives at that block's last tick;
- reverse_flush: trial t arrives at tick 4800-t, after all issues;
- outcome_coupled: positive observed labels have delay zero, negatives
  delay 599. This is a NEGATIVE MODEL-ASSUMPTION CONTROL, not a justified
  ignorable-delay posterior or recommended observer.

Issue precedes any resolution at the same tick. Equal-time arrivals resolve
by ascending issue identity. The scheduling evaluator knows when feedback
arrives; the learner receives no future label, schedule, true rate or epoch
oracle. Keep issued forecasts and all returned original-forecast receipts.
Snapshots at ticks 599,1199,2399 and after the last arrival; coincident
2399/final snapshots are stored once for immediate feedback. Nomination/
schedule building, setup, Issue, Resolve and snapshots have separate costs.
Scoring is evaluator work outside learner phases.

## Frozen Technical Screen

For every arm/world in BOTH splits: receipt identity/epoch/time/outcome/
forecast exactly bind to Issue; one resolution per identity; no pending
after drain; all counts agree. At every snapshot, laws agree within 3e-10
with independent batch Beta integrals/direct atom products over the actual
arrived set. Final laws agree within 3e-10 across ALL schedules for the
same full set (stationary model algebra, not model adequacy).

Every issued law must equal the exact replay as of its issue. Future-label
flips must preserve all earlier emitted laws before the first changed
feedback arrives. Opaque ticket mutation, replay, owner, cancellation,
epoch, time and cap/normalization controls must pass. Hash/replay/metric
corruptions must be rejected. Maximum accounted learner work <=50ms per
arm; allocation per150-member constructor <=2MiB. This is an offline
bounded component cap, NOT serving/freshness proof or a relaxation of V35.

Report future whole/priority Brier, top10 usefulness/bias and original
issued expected Brier at all schedules. Report delay harm relative to
immediate, but NO threshold is tuned to its observed magnitude. Noise,
shift and outcome-dependent arrival can expose model-adequacy failure;
retain these rather than treating identity correctness as quality success.
Intervals use16 independent worlds/cell and +/-3.5SE, not AP coverage.

## Reproducibility

Freeze all V35 sources plus adapter, adapter tests, collector and both V36
protocol/preflight sources in the manifest. Exclusive-create raw JSONL.
Postcollection audit independently reconstructs batch snapshot inference,
all costs/metrics and exact fresh RNG/issued-law/receipt replay. Keep source
hashes, auditor hash and negative controls. No private data or external
agent/provider runs. All seven whole goals remain OPEN; production untouched.
