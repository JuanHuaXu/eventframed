# Exclusive research ledger v69 results

## Verdict

FAIL: no200-record pair reaches the required10% fresh-admission improvement.
Changes range from1.43% faster to7.61% slower. Retry non-harm passes all pairs,
but does not compensate for the failed fresh criterion. Do not adopt exclusive
locking as a performance rescue, change thresholds or apply it to production.

## Correctness and boundaries

The explicit research constructor configures locking_mode=EXCLUSIVE via DSN
_pragma, followed by the pinned driver's journal_mode=WAL and synchronous=FULL
shorthand settings on every connection. Ordinary Open is unchanged. SourceOwner
still uses resolved source identities, prepared SQL, actual forecasts, service
guards, exact retries and commit-before-ack. No schema or dependency change.

Tests verify exclusive/wal/FULL readback on initial and forced replacement
connections, actual connection retirement counters, refusal of a second owner,
SQLITE_BUSY for a raw external reader and reader access after owner close.
Before/after-commit child-process exit is followed by ordinary and exclusive
reopen, checking absence/presence and exact retry state. This is process-exit
recovery, not hardware power-loss proof. External readers are excluded by design
in the candidate; that is not a hidden compatibility claim.

Ledger race suite PASS4.124s, memory race suite PASS15.829s; vet PASS. Experiment
PASS8.097s indicates integrity/accounting, not speed. All96000 originals with
exact retries and96000 verified terminals pass; paired full training/original
hashes and database sizes match, including cold and64-label trained states.

## Evidence

[Frozen protocol](mmm-exclusive-ledger-v69-protocol.md),
[raw JSONL](mmm-exclusive-ledger-v69.jsonl), SHA256:
`f77600d62fffe198e02e1886ae9a248206a2061e6e99dacbb5e3a1981ae74537`.
All36 captured source hashes match embedded and local files. Twenty-four cells,
32 cycles each;768 synthetic setup labels. Go1.27.1,Apple M4,darwin/arm64,
GOMAXPROCS10; no competing task-started performance work.

Trial-mean ranges,ms. Baseline uses normal locking; candidate uses exclusive.

| Batch | State | Fresh normal | Fresh exclusive | Retry normal | Retry exclusive |
| --- | --- | --- | --- | --- | --- |
| 50 | cold | 1.057-1.177 | 1.055-1.072 | 0.851-0.885 | 0.825-0.849 |
| 50 | trained | 0.998-1.028 | 1.012-1.151 | 0.831-0.847 | 0.840-0.870 |
| 200 | cold | 3.654-3.731 | 3.677-3.733 | 3.242-3.297 | 3.221-3.260 |
| 200 | trained | 3.589-3.600 | 3.559-3.874 | 3.164-3.203 | 3.158-3.373 |

One50-record cold pair improves10.39%, but a trained pair regresses12.78%.
No general improvement follows from cherry-picking either. Cleanup likewise
has mixed small differences; all raw cleanup timings remain in the artifact.

## Interpretation and next work

v68's lock samples were real observations, but they were aggregate across
admission, retry, cleanup, training and replay. This matched intervention does
not establish locking mode as an important fresh-admission bottleneck. It does
not prove all filesystem costs irrelevant or imply durability can be weakened.

Prepared per-record SQL remains the best supported control after conditional
insert, bulk VALUES and exclusive-mode candidates failed fresh criteria. Stop
stacking these failed variants or attempting threshold rescue. Before any new
locking optimization, require a mechanism-specific comparison showing which
calls and critical-section wall time actually disappear. A guard-staging design
remains a separate contract problem, not a reason to loosen acknowledgment.

The continuous-learning/evidence directions remain open alongside service
non-harm; these infrastructure diagnostics do not establish broad MMM accuracy
or exhaust research leads. No default, deployment or publication change.
