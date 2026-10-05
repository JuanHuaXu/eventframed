# Queued exact-guard load v22

Status: availability FAILED; accounting and artifact integrity PASS.

The frozen protocol compared off, immediate exact validation and queued exact
validation in three rotated trials. Each arm used a fresh persistent LibraVDB,
192 recalls, four readers and 96 concurrent future-ingestion writes. Queued
acquisition had a 20ms deadline. There were no labels, fitting or ledger writes.

| Arm | Attempts | Accepted | Busy | Stale | Expired | Dropped |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Immediate validation | 576 | 0 | 576 | 0 | 0 | 0 |
| Queued validation | 576 | 0 | 0 | 576 | 0 | 0 |

All request and write counts match; no execution errors were recorded. Thus the
Go accounting test passes, but queued entry alone does not rescue admission.
It moves rejection from contention to the exact-snapshot requirement after
future ingestion. It does not measure admitted-record validation or learning
cost, because neither enabled arm admitted an observation.

Per-trial recall p99 (nearest rank, milliseconds):

| Trial | Off | Immediate | Queued |
| --- | ---: | ---: | ---: |
| 0 | 31.956 | 36.065 | 33.001 |
| 1 | 42.083 | 31.967 | 29.977 |
| 2 | 31.045 | 33.990 | 34.043 |

These noisy request timings are not a throughput or learning success claim.
The independently checked artifact has ten JSONL rows, all nine distinct cells,
valid embedded source hashes and conserved request/admission counts.

Artifact: `mmm-queued-load-v22.jsonl`.
SHA-256: `383672cdea213e364240ecb29f57de5b251c4d9d2d35280ad4af1d9f6f2289e2`.

Next hypothesis: a separate as-of guard may admit complete, owned future-only
ingestion history while rejecting backfills, changed policy, unknown history,
quarantine and native/publication disagreement. That hypothesis is not tested
by this frozen artifact. Existing exact guards must remain exact.
