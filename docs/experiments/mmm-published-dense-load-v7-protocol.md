# Dense published-LSN loaded read v7: frozen private protocol

Date: 2026-10-02. Test-only, opt-in. No production or default-runtime change.
V6 passed a quiet 256D top-k transfer. This screen tests concurrent
journaled publication, body decoding, snapshot retention, and offered-load
tail latency together.

## Frozen setup and schedule

Use the v6 private 256D declared-payload collection, dense query and
row generator. Seed two aligned eligible genesis rows, 198 eligible
rows at angles `0.005(i+1)`, one aligned future sentinel and 16 future
rows at angles `0.001(i+1)`. Use the v6 test-only unit-normalization
boundary and single-owner journaled batch gate. Query as of 120 ms after
2026-10-02T00:00:00Z; future availability is 125 ms.

Offer 160 additional eligible writes and 160 reads at nominal 4 ms
intervals, concurrently, with eight reader workers. The one writer
drains at most 16 writes per batch and waits at most 16 ms from the
first offer before forming a group. Its new row i has angle
`0.00213(i+1)` and availability equal to as-of. A group is acknowledged
only after the journal marker and immutable `(LSN,published_at)` pointer
are advanced. On writer error poison the pointer. Run four fresh trials
in the fixed cap order `50,200,200,50`.

Each reader pins one published pointer and queries exactly that LSN,
projecting full EventFrames with SQL cosine and `available_at_sort`.
An independently constructed oracle retains the eligible row IDs and
angles for each published LSN. Compare each returned top-k set to that
LSN's oracle; require its scores within `1e-4` of `cos(angle)` and
nonincreasing order. Tied genesis rows may swap order. The test must
record actual offer gaps, per-event write offer-to-ack age, read-call and
read offer-to-done durations, view age at read completion, errors,
acknowledgements, and oracle violations.

## Gates

Pass only with zero read/write errors and oracle/identity/future/score
violations, exactly 160 durable acknowledgements and 160 read results
per trial, maximum view age below 250 ms, write age p99 below 250 ms,
read-call p99 below 100 ms, and read offer-to-done p99 below 100 ms in
all four trials. Use nearest-rank percentiles. Report failures rather
than changing the fixture or caps. Focused race, ordinary package tests,
and vet must pass or be reported separately.

This backend experiment does not include full Service Recall, frontier
journal writes, Bayesian updates, OpenClaw, network, or agent outcomes;
passing it would not complete Goal 6. The gate assumes the declared
single-owner/no-bypass writer contract and does not authenticate arbitrary
external writers.
