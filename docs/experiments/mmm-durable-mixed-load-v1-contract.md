# Guarded durable feedback under concurrent future writes v1

Frozen before the test implementation and run on 2026-10-01. This is an
isolated goal-6 load screen, not production activation.

Use a fresh persistent LibraVDB service, publication wrapper, frontier tap,
and SQLite-backed durable learner for each trial. Run three trials per arm:
`quiet` and `future-writer`. Each trial issues 64 recalls at `now+2i seconds`
and 64 explicit labels at `now+(2i+1) seconds`, one admitted candidate per
recall, with the same deterministic labels `i%3 != 0`. The writer arm inserts
up to 256 unique events dated `now+1h`, spaced by a 1 ms ticker, in a separate
goroutine; it begins before the first measured recall. Stop it after the last
feedback. All query and label times precede the future events.

For each issued prediction, validate the envelope and persisted original
inside the service publication/as-of guard. Revalidate on feedback before
writing the durable terminal. Require that all candidates still identify the
original `seed`, that all 64 labels complete on same-epoch replay, that no
pending/failed work remains, and that the writer overlaps at least one active
recall/admission/feedback operation. Independently inspect all 128 ledger
rows for bound tenant, event, source journal, chronology, and label identity.

Record per-call wall time for full Recall and guarded durable feedback,
including wait and I/O, plus writer count/overlap and post-close replay time.
Report empirical p50/p95/p99 and max across all 192 calls per arm. The
predeclared finite latency screen is writer-arm Recall p99 below 100 ms;
failure is preserved. It is not a population tail bound or a claim about
OpenClaw, ranking, LLM work, backfilled visible writes, crash recovery or
cross-epoch state transfer. Run focused race tests and vet. Do not edit
production code or change served answers.
