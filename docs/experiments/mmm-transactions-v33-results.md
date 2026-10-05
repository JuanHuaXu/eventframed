# Durable batch transactions v33

Status: partial performance rescue; 250ms age screen still FAILED.

The frozen twelve-arm comparison retained off, non-durable group4, individual
durable transactions and batched durable transactions. Every persisted original
was read back and compared, and explicit discard records were actually committed.
No labels, fitting, weakened synchronous durability or production changes were
introduced. Small persistent read-only/write cells passed three race repetitions
and vet before the full run.

| Trial | Mode | Accepted / 192 | Drops | Read p99 (ms) | Write p99 (ms) | Age p95 (ms) |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| 0 | Off | N/A | N/A | 32.020 | 17.670 | N/A |
| 0 | Non-durable | 192 | 0 | 27.797 | 19.854 | 141.823 |
| 0 | Individual | 78 | 114 | 29.047 | 156.395 | 828.886 |
| 0 | Batched | 129 | 63 | 26.403 | 39.540 | 428.288 |
| 1 | Off | N/A | N/A | 42.969 | 22.918 | N/A |
| 1 | Non-durable | 192 | 0 | 26.814 | 20.944 | 130.717 |
| 1 | Individual | 78 | 114 | 26.631 | 142.834 | 848.918 |
| 1 | Batched | 130 | 62 | 29.938 | 38.841 | 434.667 |
| 2 | Off | N/A | N/A | 32.029 | 18.142 | N/A |
| 2 | Non-durable | 192 | 0 | 25.790 | 22.832 | 103.890 |
| 2 | Individual | 77 | 115 | 29.946 | 169.798 | 820.841 |
| 2 | Batched | 134 | 58 | 26.752 | 39.727 | 424.427 |

Nearest-rank quantiles. Batched admission is 393/576 (68.23%) versus individual
233/576 (40.45%). The batched arm commits 19,650 admission/discard pairs with
complete original-record equality checks. It still drops 183 observations.
There were no execution errors, stale rejections or entry expiries in these arms.

All batched read-p99/off ratios pass 1.10, but all age tails fail 250ms and write
p99 remains above off. Reduced synchronous commit count is a material improvement,
not a completed timely-admission rescue. The non-durable arm continues to pass,
so its result cannot stand in for durable operation.

## Remaining cost

| Trial | Batched entry sum (ms) | Callback sum (ms) | Admit API sum (ms) | Full readback sum (ms) | Discard API sum (ms) |
| --- | ---: | ---: | ---: | ---: | ---: |
| 0 | 215.505 | 599.019 | 175.971 | 108.618 | 194.939 |
| 1 | 230.103 | 607.377 | 186.266 | 106.643 | 188.487 |
| 2 | 248.311 | 622.095 | 171.854 | 108.646 | 223.765 |

These are API durations, not isolated disk times. Batching changes the balance:
readback is now a larger share, and API preflight/retry lookups and per-entry SQL
work remain. Next profile the batched APIs and their dependencies before choosing
prepared-statement reuse or bounded bulk reads. Do not remove integrity checks
or increase queue/deadline caps merely to pass the existing workload.

No warmed learning, feedback authority, crash during live service mutation, or
actual agent-answer quality was measured here. Those direction-level requirements
remain separate. The wrapper's trained-original unit tests do not replace them.

## Artifact verification

`mmm-transactions-v33.jsonl`, SHA-256
`da43ac89e0399f70588fc3535d5a2e82958b99a759c4b2d184b89c5f33cce6fe`.
Independent parsing checked all twelve unique cells, embedded source hashes,
192 reads/96 writes per arm, group-weighted outcomes, 50 validated candidates
per accepted observation, matching admit/discard counts and phase containment.
No deployment, publication or push occurred.
