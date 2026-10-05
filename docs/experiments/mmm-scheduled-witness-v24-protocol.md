# Scheduled Durable Witness V24: Prospective Protocol

2026-10-03. All seven original goals OPEN. V23 restored conditional learning but
all eight adoption screens failed outcome p99/max100/250ms. Profiling located
4.35/4.37s aggregate outcome blocking at admission. This new isolated experiment
tests scheduling, not a posterior formula repair or a weaker performance gate.

Research context: [Shreedhar/Varghese fair queueing](https://openscholarship.wustl.edu/cse_research/339/)
motivates explicit bounded classes and non-preemptive service rotation;
[The Tail at Scale](https://research.google/pubs/the-tail-at-scale/) motivates
measuring the complete offered latency. This implementation is NOT DRR:
it has no measured-cost deficits and inherits no throughput or deadline theorem.

## Mechanism And Invariants

New isolated scheduler and test-only wrapper. Existing V23/native/store sources
remain sealed. Both scheduled and control arms enable V23 conditional witness
reuse; only external admission ordering differs. The fixed work-conserving
rotation is Outcome, Outcome, Recall cohort, Ingest. Empty slots skipped;
FIFO within each class, Recall cohort<=8, queued callers<=512. A cohort may
fill only its original eight grants and cannot admit newcomers after a writer
is pending. Mutations remain exclusive with every admitted Recall. No score,
label value, prediction error or future data enters the scheduling rule.

The original read-to-durable-journal lease remains held, plus the scheduler
lease. Existing native journal owner performs work while readers hold leases;
it must NOT acquire another external lease. Do not deadlock a reader waiting
for its own journal ack. Event/outcome wrappers acquire external writer leases
before original admission. Cancellation withdraws queued requests or releases
racing grants exactly once; never frees an active caller's lease prematurely.
Closing rejects queued/new work but allows existing owners to finish.
Both arms use the same cancellation rule for admitted journal handoff: once
handoff starts, finish native-worker and witness durability with context values
preserved, then return caller cancellation. Already-canceled calls before handoff
do not enqueue. This is documented terminalization of an admitted log operation,
not delivery of a canceled packet or new autonomous work. An integration barrier
must prove cancellation cannot release admission while durability is pending.
The native worker's bounded deadline and bounded queue remain; this does not
establish power-loss recovery or cancellation of a hung storage device.

As-of source/query/vector/frontier checks, complete runtime witness continuity,
immutable stored epochs, backend-before-witness failure behavior and native
LSN/marker checks are unchanged. Visible changes lack complete bounds and block
reuse. All certificate coverage remains synthetic assumed input, not established.
No runtime/production/whitepaper/global tooling changes or private data.

## Controls Before Freeze

Unit rotation, reader/writer exclusion, FIFO/caps/cancellation/close, idempotent
release and concurrent cancellation races. Existing semaphore gate unchanged.
Native integration: future insert/reopen accepts; visible changes, altered
query/vector/settings, old-as-of, gap/bypass and injected interruption reject.
Private grouping keys are compared only through the actual journal wire schema,
not an inappropriate DeepEqual of json-omitted fields. Preserve failed V23 flags.

## Unchanged Workload And Gates

Two repetitions x control/scheduled x future/visible = eight NEW normal trials.
200 eligible normalized256D vectors +17 future records initially;128 event
writes and128 full Recalls offered independently every4ms, eight Recall workers.
16 distinct mixed full-stream outcomes from a durable prime journal alternate
useful/not-useful and are independently offered16ms apart to one worker. Read
AsOf=offer time; no closed-loop offered-load reduction. Event batches<=16/16ms
dwell and native journal batches<=4/1ms remain unchanged. The scheduler controls
only admission after offering, not data nomination, outcome values or cadence.

Write offer->ack p99<250ms; Recall call AND offer->done p99<100ms; outcome
offered->durably-published p99<100ms/max<250ms; published-view max age<250ms.
All128 writes/Recalls and16 outcomes must acknowledge, with no stale/future/
pin/journal/witness errors. Exact top150 oracle, no future label in a law,
ordinary Beta predictive and source epoch/identity conservation. Future mode
must show >=1 actual cross-epoch served belief; visible mode zero transport.
Keep first scored use and explicitly censor never-used labels within the
offered Recall window. No quality/calibration or real-agent claim is made.

Freeze source, protocol, collector and independent auditor before collection.
Record each queue acquisition/release plus final grant/queue statistics, actual
candidate laws/pins, full witness chain, wire-journal hashes, offered timestamps,
all p99/max gates and source/posterior states. Independently audit and exercise
corrupted-tape controls. Per-class service rotation is NOT a time deadline
guarantee; variable service costs, overload and large corpora remain open.
Adoption requires BOTH scheduled future repetitions pass every unchanged gate
and all blocking controls. An unsuccessful fairness rescue stays unsuccessful.
No tuning this cohort to force a pass; journal/witness batching is a separate
prospective experiment if needed. Full Goal6 and other six goals remain open.
