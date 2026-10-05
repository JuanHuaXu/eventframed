# Post-guard discard v38 results

Status: PARTIAL. Moving typed terminal cleanup outside the service guard improves
completed observation count in all three finite trials, but FAILS the unchanged
250ms age screen. No loaded learning or real-agent accuracy claim is established.

## Invariant and verification

The prior profile confirmed prominent storage work inside the guarded callback.
Possible causes of completion delay included guard contention, durable commit
cost and validation/readback cost. This experiment isolates the scope of typed
discard, not SQLite durability: admission and complete original verification
remain guarded, and the same consumer waits for FULL durable discard before
counting completion or taking another group. No label, fit or evidence-clock
update is authorized by discard. The existing inside-guard mode remains control.

Publication/recovery tests show that a real policy change can complete between
admission and cleanup, stale admission still rejects, originals survive reopen,
closed-owner cleanup errors cannot increment completion counters, and discarded
records cannot later become labeled feedback. Three race repetitions passed,
including small persistent load. Full ledger/learner/service race suites and vet
passed before the experiment. Existing commit-error/panic/process-exit and pending
cap tests are retained; no durable wrapper or ledger semantics changed here.

The load fixture is cold and unlabeled, with 50 candidates and at most four ready
observations per group. Preview-ID mapping remains restricted to this fixture.
Post-guard errors are fatal, not misclassified as stale or expired admissions.
Guard duration excludes post-guard work; total duration and accepted age include
all terminal persistence. This is an uninstalled test consumer, not production
integration or a cross-database atomic transaction.

## Same-run results

[Frozen protocol](mmm-postdiscard-v38-protocol.md),
[raw artifact](mmm-postdiscard-v38.jsonl). Each of twelve rotated cells performs
192 recalls with four readers and 96 future writes. Times are nearest-rank
milliseconds. Age is conditional on completed observations, not dropped ones.

| Trial | Path | Complete / 192 | Dropped | Age p95 | Read p99 | Write p99 |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | Off | n/a | n/a | n/a | 36.847 | 26.628 |
| 0 | Non-durable group4 | 192 | 0 | 168.097 | 26.669 | 22.986 |
| 0 | Guarded discard | 140 | 52 | 401.298 | 21.171 | 42.909 |
| 0 | Post-guard discard | 159 | 33 | 355.863 | 25.966 | 29.952 |
| 1 | Off | n/a | n/a | n/a | 32.049 | 18.742 |
| 1 | Non-durable group4 | 192 | 0 | 155.395 | 22.612 | 21.805 |
| 1 | Guarded discard | 139 | 53 | 403.638 | 24.198 | 41.283 |
| 1 | Post-guard discard | 162 | 30 | 337.010 | 24.327 | 32.495 |
| 2 | Off | n/a | n/a | n/a | 32.066 | 21.283 |
| 2 | Non-durable group4 | 192 | 0 | 111.090 | 25.141 | 20.791 |
| 2 | Guarded discard | 144 | 48 | 380.019 | 24.657 | 39.058 |
| 2 | Post-guard discard | 158 | 34 | 360.890 | 27.440 | 42.444 |

Completed observations increase from 423/576 (73.44%) to 479/576 (83.16%),
with 23,950 actual admission/readback/discard records on the post-guard path.
All paths have zero recorded errors and entry expiries; losses are queue drops.
The experimental path passes read-p99/off <= 1.10 in all trials, but fails age
in all trials. Write p99 improves versus guarded discard in two trials and
worsens in one; it exceeds off in all three. These finite comparisons are not
population confidence guarantees or a general serving-latency rescue.

Entry totals fall from 251/264/256ms to 141/121/130ms across the three trials.
Callback totals fall from 564/572/561ms to 417/417/444ms; post-guard work adds
263/284/257ms and remains counted. More records complete on the experimental
path, so raw phase totals are not per-record cost estimates. The evidence is
consistent with better writer/consumer overlap, not removal of persistence cost.

## Audit and next lead

The full non-race experiment passed in 21.39s. Independent parsing verified
twelve unique cells, embedded source hashes, exact request/write counts,
group-weighted conservation, 50 validations per completed observation,
admission/discard equality and inside/post-guard timing containment.
Artifact SHA-256:
`be872b60168654d669fb4776de27a36cbfbb154ebd2403d1ca47e74dc3ac8b43`.

Next investigate whether immutable ledger readback can also occur after guard
release, without moving admission or service-side validity checks. It must still
compare full actual owned-worker originals, fail on missing/mismatched records,
and finish before any completion acknowledgment. This is an untested hypothesis;
its likely gain is bounded by the remaining readback cost and may not close the
age deficit. No optimization may convert historical validation into fresh
feedback/model-history authority. All seven direction-level criteria remain open.
No production change, publication or push occurred.
