# Pre-search snapshot under repeated writes v1: frozen service screen

Run three arms on separate fresh private LibraVDB stores with the same
synthetic two-event-per-tenant sequence, 32 tenants/arm. Rotate arm order
across three blocks. Each Recall uses `recall_k=pack_k=50`; the first event
is present before the call. In the two write arms, a test wrapper inserts
the second as-of-visible event immediately after the first underlying
`Store.Search` completes. The arms are: current Search-then-Snapshot order,
pre-search snapshot pin with the same injected write, and pre-search pin
without an injected write. Production code is not modified.

For every call record search attempts, packet candidate IDs, packet and
durable journal snapshots, raw Recall duration, and injected write duration.
The current-order arm is a negative control and must not be mislabeled as
safe if it silently omits the new event while claiming its version. The
pin/write arm passes correctness only if all 96 calls return both events
with matching current packet/journal snapshots after exactly two searches.
The pin/quiet arm must return one event after one search. Any as-of-future
event or journal mismatch fails the study.

Report pooled p50/p95/p99 Recall and injected-write durations by arm;
subtract write duration only as a sequential diagnostic, not a real
latency claim. This screen has no concurrent requests, learner, external
retriever, certified published-LSN reader, or production authority. A
performance result here cannot complete Goal 6.
