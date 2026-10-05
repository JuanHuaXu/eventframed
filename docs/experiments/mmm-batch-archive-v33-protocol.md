# Bounded Batched Historical Archive V33

2026-10-03. Previous goal turn concrete progress: V29-V32 historical handoff,
timer repair and preserved failures. All seven whole goals OPEN. Usage7%.

## Contract And Reasoning Gate

V26 full loaded deadline failures confirmed; V29/V32 correctness controls do not
establish latency success. Competing causes: read-to-ack exclusion, native append/
owner serialization, repeated linear ancestry/copy work. Hypothesis: bounded
batched historical journals can shorten exclusion and amortize publications.
Falsifier: incorrect source/law/nomination, lost work, premature ack, stale result,
or missed original loaded latency/freshness gates. Isolated research, not an
upstream production repair; old frozen adapters and whitepaper unchanged.

Archive-only wrapper reserves one of128total job slots BEFORE copying; copy and
seal full journal/request provenance under original read admission; release read
lease only after owning capture and reserving enqueue under the close mutex.
One FIFO worker batches1..4jobs with1ms maximum collection wait. Same500ms native
operation budget begins AFTER owner acquisition. Native byte readback, receipt,
original snapshots and marker checks retained. Capture rows and witness bindings
commit with FULL SQLite marker. Cross-store incomplete writes fail closed, not
atomic across databases. No changes to current residual/certificate validity.

Owner validates complete committed ancestry per capture; record its before/after
hash INSIDE owner because writers can now advance after read release. Queue
rejection is explicit, counts as failed work in load, and is never dropped from
metrics. Close rejects new requests and waits for accepted jobs AND registered
Recall/Put calls. Cancellation after accepted capture gets no packet and waits
for durability. Graph/parent closure and optional async service remain out of scope.

Review found sequential identical-retry root conflict in V31: record an expected
negative control against sealed V31. V33 accepts byte-identical retry with the
same original witness binding, retains first committed capture root, and checks
that root's original snapshot against full ancestry before reusing. It does not
rewrite provenance or accept forged roots. Complexity remains O(chain) per
capture and O(journal bytes) for copy/seal/readback, caps explicitly frozen.

## Technical Controls Before Load

Freeze runtime/generators/controls before tests. Race tests: capacity3 rejects a
fourth job; canceled accepted work and concurrent Close cannot return before
durability; distinct originals, sequential retry, forged root, continued service
after reopen, and after_db/before_witness/after_sqlite interruption semantics.
Vet generator/adapter; repeat scheduler/validity race checks3times. Technical
failures preserved; normal cohort not run until these controls pass.

## Fresh Normal Cohort

Mechanically copy sealed V26 offered workload and gates; both arms use its joined
publication and fixed scheduler. Factor is archived=false control (V26 combined
on) versus archived=true bounded batched historical archive. Two new reps x two
visibility modes x two arms =8trials; not replayed V26 confirmation. Same200eligible
and17future normalized256D initial rows, exactly150nominees;128writes and128Recalls
independently offered every4ms,8readworkers;16distinct alternating outcomes at16ms.
AsOf=actualoffer. No offered-rate/packing/vector/source changes.

Original gates unchanged: full Recall call AND offer p99<100ms; outcome offer
p99<100ms and max<250ms; write offer p99<250ms; view max age<250ms; all128/128/16
acknowledge. Future mode requires actual learned/cross-epoch use; visible mode
cross-epoch transport0. Recompute original as-of/top150/law/source/epoch/chain
audit and all metrics independently. All overhead counts in call/offer timing.
Collect max/censored first-use per outcome separately; later backlog can improve
eventual use without proving fixed-window freshness or predictive quality.

Archive raw includes full encoded owned journals, digests, capture roots/times,
queue counts and owner-locked batches. Independent auditor checks original
captured snapshot/laws, verified chain ancestry, batch commit/ack order, admission
exclusion, retained durable source epochs and queue conservation. Intentionally
corrupt provenance, wire, timing, source, nomination, gates and queue facts.
Freeze source and both auditors before normal run. Preserve all failed gates.

Research basis: [SQLite WAL](https://www.sqlite.org/wal.html) explains one-writer
serialization and FULL commit sync; do not reduce durability to improve tails.
[Dean/Barroso](https://research.google/pubs/the-tail-at-scale/) motivates measuring
loaded tails rather than small cached computation. Neither transfers a bound
to this prototype. Production/private data/whitepaper untouched.
