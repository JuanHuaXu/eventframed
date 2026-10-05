# Frozen loaded screen: mixed result

The unchanged guarded durable freshness fixture completed both arm orders on
the repaired `f3231fa` control and isolated mask-reuse candidate. All four test
commands passed their functional and absolute finite timing checks. The stricter
paired non-regression screen FAILED; no loaded-speed promotion is justified.

| Pair | Workload | Control recall p99 ms | Candidate recall p99 ms | Ratio | <=1.10 |
| --- | --- | ---: | ---: | ---: | --- |
| 0 | Quiet | 7.245333 | 6.651083 | 0.917982 | Pass |
| 0 | Future writer | 54.416916 | 54.575458 | 1.002913 | Pass |
| 1 | Quiet | 6.182167 | 7.191625 | 1.163285 | FAIL |
| 1 | Future writer | 52.626958 | 56.720709 | 1.077788 | Pass |

Each command used three fresh trials per workload, 192 calls and 192 labels.
Each writer arm reported 768 structured Observe writes and 762 overlaps with
the request/feedback-active interval. Writer-arm offered-label age p99 was
12.26-15.47 ms, below the frozen 250 ms screen. All 64 labels per trial completed
before Close; the test audited 128 durable admission/feedback rows and replay.

The quiet-pair failure is retained, not dismissed as noise or attributed to the
candidate without evidence. This fixture has one eligible event (RecallK=3,
PackK=2), uses structured Observe rather than CaptureTurn, and excludes future
writes from the frontier. It does not establish large-frontier behavior,
capture ingestion cost, visible mutation safety, cross-epoch learning transfer,
untouched agent utility, sustained backlog performance or population-tail
guarantees. No production implementation was changed. All seven goals remain
open. The subsequent public-capture experiment addresses a different workload;
it cannot retroactively change this verdict.

Commands, raw transcripts, source hashes and parsed measurements are preserved
in `manifest.json` and `results.json`. Run `node verify.mjs` to verify hashes and
recompute the finite verdict from the transcripts without repeating measurements.
