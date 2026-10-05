# Sharded Pilot Audit

Parent-led scoped audit, not an independent review or a production release.

## Confirmed Findings

1. Admission self-block: the original full store suite times out in a four-record
   transaction. Physical shard views share a bounded write controller, but the
   transaction acquires a permit per physical collection. The cap-one regression
   fails at its 200-ms deadline with no committed records or leaked permits.
   Deduplicating by controller identity preserves all physical mutation locks,
   WAL preparation and publication. The regression passes normally and under race.
2. Discovery: fresh sharded reads work but logical collection discovery after
   lazy reopen omits the parent. The wrapper attempts to recreate existing
   children. The independent library regression reproduces this before repair;
   adding deduplicated parent names passes iteration/reopen/cancellation guards.
3. Quantization: the discovery-only repair then fails the unchanged topology
   assertion because the reloaded config loses SQ8. This is not confined to
   sharding: the six-case none/SQ8/FSQ matrix reproduces loss for both ordinary
   and sharded collections. Temporary declaration persistence and recovery-index
   restoration pass two reopens and defensive-copy guards. This does not prove
   actual compression, training-state equivalence or repair legacy missing fields.

The original candidate, discovery-only failure and actual before-regression
outputs remain intact. Setup failures are separated in SETUP_ERRORS.md.
No source-pin tests or chronology guards were dropped to manufacture a pass.

4. Supplemental full V2 load race: clearing the training-vector slice races
   with workers filling reserved slots. The trained flag is already atomic;
   the initial user-facing guess about that flag was corrected after reading
   the actual source/stack. This is not a data-row repair. Atomic reservations
   are not ready-sample publication. V3 serializes only bounded cold sample
   collection/training and preserves the warm trained fast branch. Its small
   pre-patch regression did not trigger the race, demonstrating a coverage
   limitation; the full V2 workload is the authoritative failing witness.

## Dependency And Format Boundary

Six temporary dependency sources change; a portable patch records them. All
six original hashes remain pinned. The existing optional length block carries
the validated quantization declaration with its own magic and JSON payload.
Empty declaration padding keeps older readers able to skip the extension.
Opening/re-writing with an old reader is NOT a compression-preserving downgrade;
legacy databases with no quantization declaration are NOT silently guessed.
Codec tests retain graph/declaration fields and the following-record boundary,
and reject truncated/invalid quantization payloads. Full storage and adjacent
library suites provide additional finite regression coverage, not a format proof.

## Data And Timing Boundary

Both timing arms use the same repaired fork and identical frozen store/service/
extractor/corpus fixtures except the one-line sharding option. New public DESIGN
replicas do not establish independent facts or untouched task accuracy. All
future captures remain after every query's as-of time; 50/200 candidates are
scored before packing at most 10. Source hashes are checked before and after
each validation and timing command. No parallel benchmark, guard bypass or
production/private-data access is used.

The writer is closed loop. A faster implementation changes offered arrival
rate, so overlap/bytes/counts are reported rather than calling the streams equal
open-loop load. All 256 writes and 64 recalls/labels must finish; publication
age starts before guarded submission and is observed live, not during Close.
Ledger replay is same-epoch only. p99 over 64 calls is the sample maximum, not
a confidence bound on the traffic population. Whole-process memory, peak RSS,
internal retries and phase-exclusive wall times are not captured. Topology
coverage uses FNV route occupancy plus public config, not direct inspection of
every physical index's byte layout. Neighbor/rank equivalence is NOT asserted.

## Promotion Boundary

The original one-line candidate FAILS its frozen preflight. Repaired V2 is a
separate post-failure DESIGN pilot whose positive result cannot overwrite V1.
119 known opt-in store skips per arm remain unexecuted, not passing coverage.
The named new regressions, topology and three service guards must explicitly
PASS. All-arm public load race, larger corpus, sustained open-loop backlog, process
kill, cross-epoch learning transfer, network paths and untouched agent outcomes
remain missing. All seven whole goals remain OPEN regardless of pilot verdict.
No live dependency/store config, private corpus or whitepaper is modified.

The supplemental V2 complete candidate/200 race run FAILED on the shared
training slice. V3 separately repeats that complete candidate/200 fixture;
its own validation artifact is the source of the new functional race verdict.
This does not retrospectively make V2 safe or validate all arms/frontiers.
The future-data claim is only selected-ID exclusion at the declared query
times; counterfactual ANN/rank/scored-law equality is NOT established.
