# Learning-load stage diagnostics

Two separately preserved instrumented replays of the failed v1 workload. Neither
changes thresholds, queue size, labels, model cadence or the original artifact.
Both still fail the completion criterion; these are diagnostics, not rescues.

## Stage timing

[Stage artifact](mmm-learning-stage-v1.jsonl), mean milliseconds per admitted
frontier (50 candidate labels). N is32,34,34 across trials:

| Trial | Admission | Feedback submission | Remaining worker wait |
| --- | ---: | ---: | ---: |
| 0 | 7.650 | 0.039 | 4.018 |
| 1 | 8.110 | 0.041 | 3.651 |
| 2 | 8.087 | 0.041 | 4.191 |

Admission is the largest measured stage. Feedback submission is small in this
workload, contrary to the earlier candidate explanation that per-label temporal
checks were the primary bottleneck. Worker fitting can overlap submission, so
the remaining wait is not a measurement of all fitting CPU or wall time.

## Journal-read hypothesis

The [journal diagnostic](mmm-learning-journal-v1.jsonl) adds a read-through timing
wrapper without skipping the store's optional temporal compatibility interface.
It measures actual durable journal retrieval/decode, inside total admission:

| Trial | Admission ms | Journal read ms | Fraction of admission | Admitted |
| --- | ---: | ---: | ---: | ---: |
| 0 | 7.661 | 0.203 | 2.65% | 33 |
| 1 | 7.409 | 0.218 | 2.94% | 34 |
| 2 | 7.753 | 0.197 | 2.54% | 34 |

This does not support bypassing the durable journal check as the remedy. The
unattributed admission interval includes dependency checking, scheduler/lock
waits, membership validation and prediction journaling. It is not automatically
prediction CPU. Instrument these separately before editing runtime behavior.

Every timing-array cardinality equals its admitted-frontier count. Raw request
and write latencies, labels, errors, sources and failures remain in each artifact.
The journal-instrumented trial2 also exceeds the paired p99 allowance (38.550ms
on versus31.988ms off). Do not carry the original run's latency pass over to this
run or use instrumented timings as an unbiased production performance estimate.

The stage tests are separate source files so the frozen original load source and
raw outcomes remain intact. No functional rescue, publication, generation-model
call or deployment was performed. The full research objective remains open.
