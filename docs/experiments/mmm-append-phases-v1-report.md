# Prepared append phase attribution

## Result and scope

Diagnostic completed; no optimization or whole-goal success is claimed.
Append preflight accounts for 4.60-5.55% of measured time. Row work plus
durable commit accounts for 92.00-94.24%. Moving only append preflight outside
the admission guard therefore cannot remove most of this fixture's append cost.
This is not a bound on full-service latency or queue tails.

The frozen [contract](mmm-append-phases-v1-contract.md) uses isolated SQLite
ledgers, WAL, synchronous FULL, and the service-identity index. Three trials
rotate the original prepared append and its test-only instrumented copy.
Each of the 12 cells contains 32 fresh batches of 50 or 200 unique sources,
with 1 KiB payload padding. All 48,000 originals were checked by source, key,
and payload after reopening the databases.

## Measurements

Values are means in milliseconds; each slash-separated value is trial 0/1/2.

| Batch | Original total | Instrumented total | Preflight | Rows | Commit |
| --- | --- | --- | --- | --- | --- |
| 50 | 1.031 / .720 / .729 | .808 / .739 / .734 | .037 / .039 / .039 | .468 / .456 / .455 | .285 / .226 / .220 |
| 200 | 2.710 / 2.710 / 2.688 | 2.757 / 2.695 / 2.707 | .153 / .134 / .138 | 1.846 / 1.871 / 1.883 | .731 / .668 / .668 |

BeginTx and statement preparation account for most of the remaining time.
The original control has no phase instrumentation: its zero phase values in
the summary are sentinels, not measurements. In particular its computed
cleanupAndGapsFraction of 1 does not mean all its time was cleanup.
Rotated timings do not establish zero instrumentation overhead or a speedup.

Payload creation, source-owner resolution, worker prediction, original record
marshaling, admission-guard acquisition, concurrency, and foreground queueing
are excluded. Rows include lookup, insertion, and index maintenance; this study
does not isolate JSON expression evaluation from other SQL or storage costs.

## Verification

Raw [JSONL](mmm-append-phases-v1.jsonl) SHA-256:
`9c4a998cfa1986cd0ffeb73d937c4af967ed3d30d41c6405924092ffb7543c6b`.
The [summary](mmm-append-phases-v1-summary.json) verifies 35 captured source
snapshots and all 384 append calls. Recomputing it with
`node research/append-phases-summary.mjs` reproduced the saved summary
byte-for-byte at this checkpoint.

`go test -race ./internal/researchledger -run '^TestAppendPhasesParity$' -count=1`
passed again (1.361 seconds). The parity test compares exact fresh and retry
acknowledgments, cancellation before commit, duplicate-source rollback, and
persisted contents against the original prepared path. The earlier collection
completed in 2.23 seconds; these are test durations, not serving benchmarks.

## Decision

Confirmed: preflight alone is not the dominant append phase here.
Needs investigation: how much row time comes from identity/index processing,
SQL call overhead, and payload handling. No specific root cause is proven.
Recommendation: isolate these costs before attempting another representation
change, and compare against the existing prepared, bulk, and envelope studies
to avoid repeating a failed design. Any candidate must retain source uniqueness,
exact retry identity, rollback, crash recovery, and validity through commit.
Do not release the publication guard early on the strength of these timings.

All seven research goals remain open. Production, the whitepaper, and remotes
were not changed by this experiment or checkpoint.
