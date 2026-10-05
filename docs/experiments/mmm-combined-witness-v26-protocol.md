# Combined Scheduling And Publication V26: Prospective Protocol

2026-10-03. All seven original research goals OPEN. Both component studies
failed adoption. V24 scheduling outcome391-407ms but offered Recall173-192ms;
V25 joined publication offered Recall54-75ms but outcome459-489ms. A new
combination is an unproven research hypothesis, not a rollout or gate relaxation.

## Reasoning And Scope

Confirmed components: admission rotation changes who waits; joined publication
removes a second per-reader witness transaction. Competing remaining causes:
read-to-durable-ack writer exclusion, native journal/wire serialization and
variable service cost. Prior profiles plus future/visible loaded controls justify
testing their interaction, not claiming it is the unique root cause. Falsifier:
combination still fails original offered-load gates or permits premature ack,
incomplete evidence or unsafe posterior reuse. No production upstream patch;
the sealed V24/V25 sources are reused unchanged in an isolated test binary.

[The Tail at Scale](https://research.google/pubs/the-tail-at-scale/) motivates
measuring offered tail latency; [SQLite WAL](https://www.sqlite.org/wal.html)
motivates retaining commit-before-ack and distinguishing per-database atomicity.
No scheduling theorem or cross-database atomicity inherited from these sources.

## Candidate And Controls

Both arms use V25 joined native-journal/SQLite marker+witness publication.
The only factor is V24 fixed Outcome,Outcome,Recall-cohort,Ingest rotation,
configured before requests. No new slot weights, cache rules, epoch retags,
prediction formulas, label filters or source repair. Control external scheduler
disabled; candidate enabled. Recall cohort<=8; queued<=512. Original shared
lease and external lease stay held until admitted journal handoff terminalizes.
Worker must NOT acquire external writer admission while readers await its ack.
Native batch<=4/1ms; FULL marker+witness transaction, complete wire readback;
event batches<=16/16ms. Both databases must remain publication-ready before ack.

Before freeze: combined race lifecycle for future/visible insert/reopen,
changed query/vector/selection, future label rejection and unaccounted runtime;
blocked handoff cancellation must retain the external read lease, persist one
binding, return canceled and reopen; queued cancellation must not persist or
release someone else's owner. Existing sealed scheduler rotation/cap/exclusion
and joined conflict/rollback/reopen tests remain applicable. No claim of power-
loss recovery, hung-device handling, general certificate coverage or robustness.

## Fresh Normal Cohort And Unchanged Gates

Eight NEW trials:2reps x external scheduler on/off x future/visible. Identical
200eligible normalized256D+17future initial rows,128writes and128fullRecalls
independently offered every4ms,8read workers;16distinct mixed full-stream
outcomes offered16ms apart to one worker. AsOf=offer; same query/settings.
No dropped work, load throttling, restarted failed trial or success filtering.

Both Recall call/offer p99<100ms; write offer p99<250ms; outcome offer p99<100ms
and max<250ms; published-view max<250ms. All128/128/16 items must acknowledge.
Exact150nominees, as-of no-future labels, ordinary Beta predictive, immutable
stored epochs, actual pins and durable wire/provenance checks unchanged.
Future mode>=1genuine cross-epoch scored belief; visible zero transport.
First-use and censored sources reported without calling later drain a fixed-
window freshness gain. BOTH scheduled future repetitions must pass every gate
and technical controls before component adoption. All seven whole goals stay
open unless their actual full criteria are separately established.

Freeze internal Go source/protocol/runner/checker before data collection.
Independent checker reuses V23 nomination/law/timing arithmetic unchanged,
verifies V25 batch-to-chain/wire/acknowledgment invariants, and additionally
checks scheduler shared/exclusive intervals, all grants/counts and drained caps.
Corrupt tape controls reject future/pin/source/chain/wire/missing-binding,
batch/ack/grant/overlap/timing/gate violations. No consumed cohort tuning.
All previous source/tape/protocol freezes remain unchanged.

Production, private sessions, whitepaper, installed tooling, global instructions,
Git commits/pushes and cleanup are outside this experiment's boundary.
