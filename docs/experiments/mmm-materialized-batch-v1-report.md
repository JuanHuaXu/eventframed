# Materialized batch lookup: modest scoped improvement

All 12 paired cells favor materialized source keys, with mean time ratios
.867-.952 (approximately5-13% lower). This is substantially smaller than the
37-44% point-lookup reduction. The existing control already amortizes transaction
and statement costs across each batch; point timings were not a proxy for this
batch path. No admission/write rescue or full-service latency result is claimed.

| Layout / size | Original means ms (trials0/1/2) | Candidate means ms |
| --- | --- | --- |
| Grouped50 | .357 / .349 / .345 | .314 / .324 / .317 |
| Scattered50 | .383 / .340 / .354 | .339 / .324 / .307 |
| Grouped200 | 1.534 / 1.504 / 1.487 | 1.394 / 1.401 / 1.395 |
| Scattered200 | 1.453 / 1.427 / 1.419 | 1.320 / 1.329 / 1.303 |

## Method and evidence

[Contract](mmm-materialized-batch-v1-contract.md),
[raw artifact](mmm-materialized-batch-v1.jsonl),
[summary](mmm-materialized-batch-v1-summary.json).
Raw SHA-256:
`aebe7950cc103d891c2185e4433bf57add9bd17f2eb3194d5563bd6bc3f4dbe8`.

Three rotated trials,6400 originals per ledger,32 calls per cell,24 cells.
All96,000 returned originals verified by source,sequence,key and exact payload.
Keys are grouped or scattered across the history; no missing-key timings in
this experiment. The same read transaction, prepared lookup, bounded projection,
shared row validator, caller order and independent8MiB request/output caps are
retained. Canonical lookup construction is included in timing. The candidate's
admission-only schema and its known write/compatibility limitations remain.

Collection1.45s (package1.693s). Summary verifies44 captured source files and all
cell counts; a second independent invocation produces byte-identical output.
The postcollection snapshot test and summarizer are not claimed as precollection
captured sources. Race tests passed2.159s; ledger vet passed.

Contract tests cover count/key/request-byte bounds,8MiB exact returned-payload
boundary, overflow without partial results, zero-budget misses, cancellation,
panic cleanup, and subsequent reuse. A separate SQLite connection commits a
new source between two reads: the active batch retains its earlier snapshot,
and the next batch sees the new source. Duplicate output payloads do not alias,
and mutating returned bytes does not mutate the stored original.

## Decision

Keep as an optional read-optimization lead, not a replacement ready for deployment.
Migration, feedback, crash recovery and concurrent full-service evaluation still
remain. The measured write-heavy regression is not solved, and no quality claim
changes. Do not justify further schema integration by the earlier larger point
gain alone. Preserve this evidence and return attention to unresolved learning
quality rather than continuing storage microbenchmarks without a stronger
admission mechanism. All seven goals remain open; production/paper/remotes
unchanged.
