# Journal gate phase timing v14: frozen diagnostic

Date: 2026-10-02. V13 measured a 6.9 ms median serial
`appendJournal` interval on the unchanged 4 ms full-Recall offer
workload. This test-only diagnostic copies the existing gate sequence
and adds timestamps; it must preserve all of its marker, LSN,
snapshot, readback, transaction, interruption and READY checks.
Production and the v13 source remain untouched.

Retain v13's 256D corpus, 128 write/Recall offers, four workers,
published-LSN Search, read-to-journal admission, full residual-enabled
service, exact top-150 oracle, durable journal verification and
unchanged <100 ms call/offer and <250 ms writer/view gates. Run two
normal fresh trials. Every successful Recall must cover exactly one
gate mutex acquisition, marker read, before-LSN check, LibraVDB
journal write, after-LSN check, journal readback, and SQLite marker
commit. Record p50/p99 for each phase and unaccounted gate time;
do not add independent percentile cells as one observed call.

Any correctness or gate failure remains a failure. Normal profiled
timings are diagnostic, not a latency promotion. Before accepting a
design inference, check a test-only gate duplicate and interrupted
commit path against the original sequence, run race-correctness
mode, ordinary package tests and vet. A different architecture must
be separately specified and tested for crash recovery, durable-
before-ack semantics, per-entry as-of validation and a stable
event/metadata LSN relationship.
