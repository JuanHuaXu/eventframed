# Same learner, different Anti-Pigeon gates

## Finding

Faster isolation still does not deliver the required forecast improvement when
both gate arms use the stronger retained-subset learner. This closes a comparison
gap between the earlier count-only gate studies and newer subset-learner studies.
It does not weaken the failed gain criterion or complete Goal 3.

[Protocol](mmm-subset-gate-pair-v1-contract.md),
[raw records](mmm-subset-gate-pair-v1.jsonl),
[summary](mmm-subset-gate-pair-v1-summary.json),
[replay](mmm-subset-gate-pair-v1-summary-replay.json).
SHA-256: `382d267a4fded68fca44c5cecfac077011abbf59cab218ec85b647875100340b`.
All 805 captured source hashes match current and embedded text. Summary replay
is byte-identical. This is analysis replay, not a full rerun of all trajectories.

## Design and verification

Two cohorts, five scenarios and16 independent fitting/live trajectories per cell:
160 trajectories,81,920 frames. Each base is fitted on4096 independently seeded
labels. Both subset states receive the same immutable fits and audit labels,
but maintain independent issued-forecast journals and mixture weights. Only
authorization differs: old versus mixture gate, each conjoined with nomination.
All forecasts are issued before labels; fitting follows feedback.

Five-scenario parity testing preserves every original count-control metric and
tape, and the original mixture-gate subset metric/tape. Split times match the
corresponding count-gate control, fit counts match, and base support stays4096.
Focused state/parity tests pass under race detection (5.100s); vet passes.
Collection passes in15.36s (15.556s package). This is total multi-arm experiment
time including fitting, not serving latency. No runtime implementation changed.

## Second cohort

Post-change Brier (smaller is better); recurring uses steps128-511, others the
latter256 frames. Gain is old-gate minus mixture-gate with mean +/-3.5SE over
16 paired trajectories, not an anytime or simultaneous coverage guarantee.

| Scenario | Old subset | Mixture subset | Paired gain interval | Old / mixture splits |
| --- | --- | --- | --- | --- |
| Stable | .0503022 | .0503022 | [0,0] | 0 / 0 |
| Member shift | .2009015 | .2009534 | [-.0009167,.0008129] | 16 / 16 |
| Common shift | .2104898 | .2104898 | [0,0] | 0 / 0 |
| Recurring | .1750203 | .1745828 | [-.0005178,.0013929] | 15 / 16 |
| Null | .2517219 | .2517219 | [0,0] | 0 / 0 |

Member restricted delay improves112.5625 ->68.3125 frames, about39.3%; first
cohort118.1875 ->83.5, about29.3%. Restricted recurring delay improves150.4375
->76.9375; the old arm misses one split. Missing/premature splits receive the
remaining horizon, not exclusion from the delay denominator.

Member mean gain is-.00005193 in the second cohort and+.00008274 in the first;
neither reaches.005 or has a positive lower bound. Recurring mean gain changes
sign across cohorts and also fails. All diagnostic .01 full/post non-harm screens
pass, but this small pilot cannot establish rare false-revocation coverage or
broad robustness. No-change intervals of zero reflect identical paired outputs,
not a claim that either model's population forecast risk is known exactly.

The stronger learner itself remains useful in this narrow setup: second-cohort
member Brier is about.201 versus.238 in the count controls. That does not make
the independent gate-timing effect useful. Member foreground cost rises4.6841
->4.6907 coordinates/frame; monitoring, auditing and fitting remain additional
work as recorded in each original trajectory. No compute-free upgrade is claimed.

## Interpretation and next step

Source inspection of `subset_state.go` shows revocation switches only the long
forecast slot from pooled to local and caps its weight; the incumbent and short
subset learner remain. The gate does not reset the independent short learner or
create new observations. This limited intervention is consistent with the small
quality effect, but aggregate scores do not prove whether slot weight, local
model quality or acquisition differences dominate.

Next useful evidence is a read-only mediation trace around the split: record the
actual pooled-slot weight, local/pooled forecast contrast and observation guide,
with counterfactual quantities strictly outside emitted predictions and updates.
Keep the actual tape identical. This can distinguish an already-irrelevant slot
from an ineffective local replacement before proposing any new action. Do not
lower evidence thresholds, erase incumbent knowledge, or repeat gate-only samples
until a gain happens. Retain broader learner/dependence/delay failures.

All seven goals remain OPEN. Production, private data, whitepaper and remotes
were untouched. The original recovery-protocol decision remains pending.
