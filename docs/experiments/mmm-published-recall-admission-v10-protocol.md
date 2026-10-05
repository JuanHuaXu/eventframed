# Read-to-journal admission v10: frozen Goal 6 rescue protocol

Date: 2026-10-02. Research-only; production remains unchanged. The
[v9 full-Recall screen](mmm-published-recall-load-v9-results.md) failed
completion and offer-to-done latency under concurrent visible writes.
This candidate protects the existing full `Service.Recall` read-to-
journal interval from those writes without weakening durable journal
acknowledgement or as-of checks.

## Single change

Retain v9's private 256D fixture, exact-LSN adapter, 200 initial
eligible/17 future rows, 128 new visible writes, 128 Recalls, nominal
4 ms offer clocks, four Recall workers, cap-16/16-ms-dwell writer,
`RecallK=50`, `PackK=10`, overfetch 3, and per-version top-150 oracle.
Add one Go `sync.RWMutex` admission gate outside the store adapter:
each Recall worker obtains a read lease before calling `Service.Recall`
and releases it after the durable journal has been committed or the
call returns an error. The event-batch writer obtains the write lease
around its gated batch append and publication. The existing store owner
mutex still serializes journal metadata and event commits. A queued
writer blocks newly arriving reader leases according to Go's RWMutex
semantics. Offer-to-completion timing includes admission wait.

Run two fresh trials without changing inputs or thresholds. Record
actual offer gaps, lease-wait time for reads and writer batches,
stale-retry counts, durable acknowledgements, exact versioned frontier
membership, packet/journal/pin agreement, acknowledged-before-offer
freshness, future leakage, published-view age, write offer-to-ack and
Recall call/offer-to-done p50/p99. A stale rejection or any error is
not hidden by admission; it remains a failure to explain.

## Frozen gates

Both normal trials must have 128/128 write acknowledgements and
128/128 successful Recalls; zero stale-snapshot rejection, oracle,
snapshot, journal, future-leak, or acknowledged-before-offer violations;
final gate READY and all 128 writes in its snapshot; write age p99
<250 ms, Recall call and offer-to-done p99 <100 ms, and published-view
max age <250 ms. Do not relax the writer or service limits if the gate
shifts queueing to one side.

Run the same trials under `-race` with the predeclared
`EVENTFRAME_RACE_CORRECTNESS_ONLY=1` mode: counts and semantic gates
apply, but instrumented timings are reported rather than gated. Run
ordinary package tests and vet. A pass would be a finite private
service result, not multi-tenant, large-corpus, graph/posterior mutation,
label-to-forecast freshness, or real-agent validation.
