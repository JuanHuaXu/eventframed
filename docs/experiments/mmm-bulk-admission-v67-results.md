# Bounded multi-row admission v67 results

## Verdict

FAIL: fresh200-record admission regresses2.84-8.50% across all six pairs rather
than improving>=10%. Retry non-harm passes all twelve pairs, with small gains,
but does not rescue the fresh workload. Keep the existing prepared strategy as
control/default; do not integrate this bulk variant into serving.

## Evidence

[Protocol](mmm-bulk-admission-v67-protocol.md),
[raw JSONL](mmm-bulk-admission-v67.jsonl), SHA256:
`d6d0e814a381facf4098f24d64567e65fbf35ecf590e13dbd34c5760db78d3cf`.
All36 captured source hashes match embedded and local files. Twenty-four cells,
96000 fresh originals with exact retries and96000 verified terminals, plus768
synthetic training labels. Paired full original/training hashes and final
database sizes match. Reopen and lifecycle checks pass. Go1.27.1, Apple M4,
darwin/arm64, GOMAXPROCS10; no competing task-started performance work.

Ledger race suite PASS4.002s, memory race suite PASS14.548s. Additional bulk
owner uncertain-commit recovery race test PASS1.354s; vet PASS. Experiment
PASS7.906s indicates integrity/accounting, not the performance screen.

An initial unit-test harness attempted replay pages above the existing256-row
limit; it failed before benchmarking and was corrected to traverse bounded
pages. No measured experiment was discarded or repeated to select a result.

## Timings

Ranges of three trial means,ms;32 cycles per cell. Baseline is resolved source
with prepared per-record SQL; candidate changes only all-admission batch SQL.

| Batch | State | Fresh baseline | Fresh bulk | Retry baseline | Retry bulk |
| --- | --- | --- | --- | --- | --- |
| 50 | cold | 1.061-1.175 | 1.068-1.127 | 0.806-0.919 | 0.788-0.808 |
| 50 | trained | 0.976-0.983 | 1.032-1.067 | 0.818-0.823 | 0.790-0.799 |
| 200 | cold | 3.524-3.594 | 3.624-3.797 | 3.148-3.180 | 3.046-3.082 |
| 200 | trained | 3.421-3.454 | 3.700-3.712 | 3.094-3.120 | 3.005-3.028 |

Fresh50-record changes range from4.12% faster to9.11% slower. Retry improvement
is0.54-12.07% at50 and2.12-4.23% at200. Neither slight retry gains nor fewer SQL
boundaries establish fresh admission improvement.

## Contract audit

- Shared count/byte/JSON validation still runs before the bulk branch.
- A transaction reads existing exact identities and checks bytes, including
  repeated IDs within the request batch, before any insert.
- Fresh records preserve original input order, using VALUES chunks<=128 rows.
  A later source-identity conflict rolls back previously inserted chunks.
- Receipt lookup uses identity rather than assumed contiguous row IDs; tests
  compare exact sequences, retry flags and replay pages with the old strategy.
- Mixed admission/feedback batches use the old prepared path. Feedback-first
  input still fails instead of being reordered to make it valid.
- Cancellation and panic before commit roll back. Lost acknowledgments before
  or after commit stop the owner; reopen resolves fresh versus exact retry.
- FULL WAL and commit-before-ack stay intact. No hardware power-loss claim.

No new historical-admission authority, guard relaxation or model substitution
was introduced. The variant remains explicit opt-in research code.

## Next lead

Fewer per-record statement boundaries have now failed twice. Do not assume the
Go/SQL crossing itself is dominant. Bulk preflight and receipt queries compile
variable-length IN lists, and bulk INSERT statements are rebuilt per call;
remaining costs include SQL compilation, record/index work and commit. A next
diagnostic should attribute those costs before another rewrite. The separate
staging/certificate design remains unproven, and none of these isolated tests
changes the failed mixed-service latency conclusions or completes a research
direction.
