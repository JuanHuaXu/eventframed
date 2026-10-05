# Snapshot-read integration v36 results

Status: PARTIAL infrastructure improvement; the frozen age screen FAILS in all
three snapshot-read trials. No learning or real-agent accuracy claim follows.

## Protocol and integrity

[Frozen protocol](mmm-snapshot-reads-v36-protocol.md) and
[raw artifact](mmm-snapshot-reads-v36.jsonl) contain twelve rotated cells: three
trials each of off, non-durable grouping, durable write batches with per-key
reads, and the same durable batches with transactional snapshot reads.
Each cell executes 192 recalls and 96 future-dated writes. The frontier has
50 candidates. Both durable paths retain FULL synchronous SQLite, complete
original-record readback and actual typed-discard writes. There are no labels
or fits in this fixture.

Artifact SHA-256:
`0a56ba86fe25d50382696ff132e02a4b8a2bcc81808253f955ca84f1e9380453`.
Independent parsing verified twelve unique cells, embedded source hashes,
read/write counts, group-weighted attempt counts, zero recorded errors,
50 validated candidates per accepted observation and matching durable counts.
The original test process handle was unavailable after context restoration;
these conclusions use the completed artifact rather than a recovered exit code.

## Results

All timings below are milliseconds; percentiles use nearest rank.

| Trial | Path | Accepted / 192 | Dropped | Age p95 | Read p99 | Write p99 |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | Off | n/a | n/a | n/a | 33.979 | 22.225 |
| 0 | Non-durable group4 | 192 | 0 | 126.187 | 30.814 | 20.685 |
| 0 | Per-key reads | 138 | 54 | 420.824 | 21.123 | 43.940 |
| 0 | Snapshot reads | 138 | 54 | 399.265 | 25.366 | 37.414 |
| 1 | Off | n/a | n/a | n/a | 34.159 | 17.653 |
| 1 | Non-durable group4 | 192 | 0 | 150.650 | 28.303 | 22.102 |
| 1 | Per-key reads | 133 | 59 | 411.407 | 24.697 | 39.630 |
| 1 | Snapshot reads | 142 | 50 | 393.943 | 26.613 | 39.524 |
| 2 | Off | n/a | n/a | n/a | 33.108 | 18.843 |
| 2 | Non-durable group4 | 192 | 0 | 146.147 | 24.962 | 20.340 |
| 2 | Per-key reads | 136 | 56 | 410.746 | 25.351 | 39.036 |
| 2 | Snapshot reads | 138 | 54 | 399.078 | 23.665 | 40.878 |

Snapshot reads admit 418/576 observations (72.57%), versus 407/576 (70.66%)
with per-key reads. They persist and reread 20,900 original predictions and
persist 20,900 discards. Neither durable path has entry expiries in this run;
all losses are queue drops. Age statistics cover accepted observations only.
All snapshot-read trials pass the paired read-p99/off <= 1.10 screen, but all
fail the 250ms age screen. Write tails remain higher than off. Three trials
do not establish a population latency guarantee or a significant throughput gain.

Total readback time per trial falls from 108.96/118.24/112.28ms to
79.60/74.51/69.54ms, despite slightly more admitted records overall. Total
guard callback time falls from 603.10/613.41/609.61ms to
564.67/566.86/564.17ms. Admission/discard and validation work remain substantial.
Transaction and statement reuse were changed together; their individual causal
contributions are not identified. No durability or original-record check was
removed to obtain this improvement.

## Correctness and next step

Warm forecast parity, malformed-response rejection and commit-recovery tests
passed three race repetitions before the run. Small load accounting passed
three race repetitions. Full ledger, learner and service race suites passed
after the run. The unsupported-read test was subsequently tightened to preserve
batch-write support, ensuring that missing batch-read capability is the reason
for rejection; experiment source files and the artifact remain unchanged.
The tightened test passed three race repetitions, and vet passed again.

Next profile the remaining callback work before another optimization. In
particular, distinguish validation/serialization CPU from durable SQL costs;
do not assume further read batching can close the remaining age deficit.
Retain bounded queues, complete originals and fail-closed admission authority.
Warm loaded learning, history validity, verified feedback, checkpointing and
real-agent evaluation remain open. No production installation or push occurred.
