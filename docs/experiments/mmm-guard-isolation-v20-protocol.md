# Guard-entry isolation v20

Freeze before execution. V19 observed1,152 busy rejections and no callbacks.
Test whether writer activity or journal-read placement drives entry failure.
No guard, publication, service or store implementation changes.

Three rotated trials of four cells:64 recalls/four readers with either zero or
32 future writes, crossed with journal prefetch before guard or journal read
inside guard. Fresh persistent stores with50 visible fixture events,64-slot
handoff, full50-candidate validation if entered. No ledger or model fitting.
Writer spacing remains2ms. No unbounded retry or artificial consumer delay.

Record all original load counters/timings. Verify64 attempts+drops, accounting
of accepted/busy/stale,50 validated candidates per accepted observation and
exact write counts. Read-only cells must accept at least one observation;
otherwise persistent callback validation itself remains suspect. Report every
cell; this is mechanistic diagnosis, not an availability/latency adoption gate.

The v19 helper is factored with explicit options; its original entry point still
uses requests/2 writes and prefetch=true. V19's artifact retains its embedded
original source rather than being overwritten or reclassified.

First run small8-recall read-only and writer cells under race, then12 non-race
arms. Exclusive-create JSONL with current helper/protocol source hashes.
Per-arm fsync preserves failures. A successful read-only control cannot stand
in for a successful loaded durable-learning path.
