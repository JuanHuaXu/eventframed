# Durable as-of witness v23: prospective integration protocol

2026-10-03. Goal6 engineering experiment, all seven whole goals OPEN. V18/v19
confirmed stale epochs despite publication, not a posterior arithmetic defect.
Competing causes of missing learned laws are stale certificates, stale posteriors,
changed nominee sets and absent labels. Test-only dependency witnesses target
the first two ONLY when all intervening inserts are invisible as of a request.
Visible/unknown transitions remain fail closed. This is not a production repair,
general epoch-retag permission or a relaxation of the original load gate.

Research inspiration: [provenance semirings](https://www.cs.ucdavis.edu/~green/papers/pods07.pdf)
motivate explicit input dependencies; [differential dataflow](https://www.cidrdb.org/cidr2013/Papers/CIDR13_Paper111.pdf)
motivates tracking changes rather than indiscriminately recomputing; neither
proves probabilistic coverage or posterior equivalence here. Measure complete
tail latency, following [The Tail at Scale](https://research.google/pubs/the-tail-at-scale/),
not just a witness scan microbenchmark.

## Isolated Architecture

Only new _test.go code; existing runtime and studies unchanged. A single-owner
libravdb/SQLite publication gate logs every runtime-changing batch/outcome in
a new SQLite witness table AFTER backend commit, BEFORE caller ack/read admission.
Hash-linked complete before/after snapshots, per-version mutations, and the
entire source/journal binding state survive
reopen. A gap, unclassified transition, hidden semantic motion, witness-write
failure or bypass poisons transport; backend durability is not rolled back.
Journal commits advance LSN but not runtime/evidence state and use the existing
audited marker gate. The witness proves runtime-state continuity, not a list of
every backend LSN. Existing marker/LSN fail-closed checks still apply.

Every request holds the existing admission read lease through durable journal
ack. Context binding derives from ACTUAL query text/model/session/selection settings and the normalized
Search vector; current exact nominated-ID digest comes from Search, not a caller
assertion. Committed journal bindings and source posterior JSON are retained.
Getters may return an epoch-aligned COPY only after the complete v20 guard accepts
the source/query/frontier/semantic-version/as-of binding. Durable posteriors and
certificates are never retagged or rewritten. No complete bound for visible
inserts is supplied: even below-cutoff visible writes must remain Unknown.
Certificates are synthetic assumed inputs, not newly established coverage.
Transported certificates are conditional reuse of those inputs, not new audits.
Updated source posteriors replace their prior record, not replayed trial evidence.
Model source timestamps/availability remain unchanged.
The request session is fixed and distinct from every event fixture session;
no cross-session or hierarchical-posterior certificate is claimed. Hierarchical
outcomes are rejected before mutation because affected-parent closure is absent.
Caps are10000 runtime mutations,512 committed journal bindings,200 source keys.
Every journal binding pays a durable SQLite transaction; include this overhead.
Checksums detect corruption under a trusted single writer, not a malicious
writer rewriting the whole chain. Interruption controls are injected boundaries,
not process-kill or power-loss guarantees.

## Tests And Frozen Workload

First run unit/lifecycle controls: future insert accepts; changed query/vector/
frontier rejects; visible insert rejects; direct bypass/gap rejects; interruption
after backend but before witness rejects after reopen; committed witness reopens;
future outcome never reaches an old as-of law; original stored epochs unchanged.
Actual database close/reopen and corrupted chain/binding controls are required.
Race and vet on new tests plus core guard. No unknown coerced to Compatible.

Normal collection: two repetitions each control versus witness-enabled under
future-only append and visible high-relevance append (eight trials). 200 eligible
256D events plus17 future records initially,128 writes and128 full Recalls each
offered every4ms, eight Recall workers, original journal batches<=4/1ms dwell,
event batches<=16/16ms dwell.16 mixed full-stream outcomes on distinct activated
events from a committed journal, alternately useful/not-useful, offered16ms apart
by one background worker. These are synthetic evidence, not private real-task
labels or predictive accuracy validation. Both arms log durable witnesses to
isolate serving-copy semantics from logging overhead. Serving AsOf=offer time;
future writes available origin+one hour, visible writes available origin.
Independent bounded producers offer writes/outcomes at the planned cadence even
when their worker blocks; no closed-loop omission of queuing delay. Initial and
inserted vectors are normalized before SQL distance ranking. Report actual
offer gaps and overall work span, not only per-call latency. Both arms share the
same no-future bounds and include journal-witness persistence after backend ack.

Unchanged timing gates: write offer->ack p99<250ms; Recall call AND offer->done
p99<100ms; outcome offered->durably-published p99<100ms/max<250ms; published
view max age<250ms. Correctness: all128 writes/Recalls and16 outcomes acknowledged,
no stale/future/journal/pin/marker/witness errors; exact top150 nomination oracle;
no forecast uses outcome available after its as-of. Witness future mode must
deliver updated source beliefs into >=1 later scored Recall despite epoch motion;
visible mode must not transport across its impacting writes. Report live
label->first observable scored use freshness separately, censor never-used
labels explicitly. No timing interpretation of race-instrumented runs.

Adoption of this CONDITIONAL component requires both future-mode repetitions
pass all gates and all blocking controls pass. Failures remain failures. Full
Goal6 needs broad visible mutation behavior, externally valid coverage, external
writers/crash recovery, realistic agent workloads and corpus scaling; this study
cannot close it. Freeze source/protocol/collector before collection, emit complete
machine-readable per-request/per-outcome timings and laws, independently audit
counts/as-of/epoch/nomination/freshness and keep negative controls. No consumed
v18/v19 rerun is a new confirmation of quality.
