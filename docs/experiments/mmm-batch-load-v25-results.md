# Bounded batch validation v25

Status: substantial partial rescue; 250ms age screen still FAILED.

The profile-supported batch validator reads shared journal/query inputs once,
checks every requested record and fetches all requested events in one backend
API call. It rejects duplicate IDs, mixed journal/tenant/snapshot/time bindings,
malformed records and any invalid member. The final dependency check remains;
the guarded entry point invokes work only after the whole batch passes. No
cross-call cache, label authentication or durable atomicity is implied.

Seventeen multi-record valid/negative cases passed three race repetitions.
Additional race-tested controls cover guarded admission on memory and persistent
stores, policy invalidation, callback suppression for bad records and small
read-only/concurrent-write load. Tests assert one shared journal read and one
event API call for a valid batch. The independent single-record validator remains
unchanged. Vet passed before the frozen comparison.

## Loaded results

Nine rotated arms each used 192 recalls and 96 concurrent future-ingestion
writes. All records were public synthetic infrastructure fixtures; no feedback,
fitting or ledger operations were performed.

| Trial | Single accepted | Batch accepted | Batch dropped / expired | Read p99 off / batch (ms) | Write p99 off / single / batch (ms) | Batch guard p95 (ms) | Age p95 single / batch (ms) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 0 | 75 | 173 | 19 / 0 | 32.222 / 32.057 | 18.062 / 54.708 / 18.976 | 16.154 | 1148.015 / 367.745 |
| 1 | 76 | 186 | 5 / 1 | 32.083 / 27.992 | 18.058 / 55.609 / 22.795 | 16.595 | 1158.896 / 313.533 |
| 2 | 74 | 171 | 21 / 0 | 31.984 / 27.841 | 18.406 / 60.981 / 19.529 | 17.683 | 1132.580 / 387.824 |

Nearest-rank quantiles. Single mode admitted 225/576 (39.06%); batch admitted
530/576 (92.01%), validating 26,500 candidate records. Batch dropped 45 queued
observations and one attempt hit the entry deadline. No busy/stale batch
rejections or execution errors occurred. All enabled attempts are accounted for.

All batch read-p99/off ratios are below the prior 1.10 screen, but all three
accepted-age tails exceed 250ms. Write tails improve versus single validation;
relative to off they remain higher, especially trial 1. Therefore this is NOT
a full timely-admission rescue, much less validated durable learning or improved
MMM accuracy. Do not relabel dropped observations as suppressed redundant work.

The paired ablation supports shared-input repetition as a material contributor
to the prior bottleneck. Remaining guard duration includes admission waiting,
native snapshot acquisition and callback processing; this run does not separate
them. Next instrument those boundaries before deciding between shorter guarded
work, bounded multi-observation admission or a different persistence design.
Do not increase queue/deadline caps merely to pass this fixture.

## Artifact verification

`mmm-batch-load-v25.jsonl` contains all nine unique cells, runtime/dependency
metadata and selected source text/hashes. Independent parsing verified every
embedded hash, 192 reads/96 writes per cell, outcome conservation, zero errors,
validated=50*accepted and zero ledger counters.

SHA-256: `95c570afacba4b32447ae66585b135a4c09a54e86ddd46255a084d78515a7f5a`.
The passing Go experiment test proves accounting, not the failed age criterion.
No production configuration, deployment or push was performed.
