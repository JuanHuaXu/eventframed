# Bounded materialized reader: scoped read gain, no write rescue

The read advantage survives reintroducing the original bounded projection and
the exact shared `readServiceAdmission` validator. Candidate means are 14.10-14.82
microseconds versus 22.73-25.29 for the original. Ratios across six paired cells
are .562-.626, approximately 37-44% lower mean point-read time in this fixture.
This is not end-to-end latency or a population-tail guarantee.

| Batch | Trial | Original read us | Candidate read us | Candidate/original append |
| --- | --- | --- | --- | --- |
| 50 | 0 | 25.294 | 14.222 | .779 |
| 50 | 1 | 24.204 | 14.535 | 1.002 |
| 50 | 2 | 23.663 | 14.522 | 1.022 |
| 200 | 0 | 22.731 | 14.218 | 1.024 |
| 200 | 1 | 23.770 | 14.818 | 1.019 |
| 200 | 2 | 23.617 | 14.098 | 1.049 |

All 200-record append cells are slower. The write-latency rescue remains
unsupported. Retain the first50 control outlier rather than selectively removing
it or extrapolating a gain from it.

## What changed

New test-only bounded reader uses the original identity/payload size and type
projection, validates lookup fields, constructs its canonical lookup key inside
the timer, and delegates row validation to the original shared function. The
unbounded prototype and its earlier results remain intact. Only the candidate
read strategy changed; the materialized table and counted write preparation did
not. This comparison applies to the admissible fixture domain, not every legacy
writer input or corrupted-schema scenario.

## Evidence

[Contract](mmm-materialized-bounded-v1-contract.md),
[raw data](mmm-materialized-bounded-v1.jsonl),
[summary](mmm-materialized-bounded-v1-summary.json).
Raw SHA-256:
`724b875a11c66cede4fff445c515315ffd1da46e2a179969b0fa588d1c0c2794`.

Three rotated trials, 12cells,384 appends,48,000 source reads verified after
reopen. Independent summarizer validates41 captured source files and all cells;
second summary is byte-identical. Collection2.04s (package2.259s).
Race tests for the bounded reader and transaction semantics passed1.396s;
ledger vet passed. Oversized identity/payload tests explicitly inspect the SQL
projection and require NULL rather than materialized oversized bytes. Wrong
payload type, source/owner mismatch, invalid lookup fields, absent keys and
cancellation are tested. These checks are not hardware power-loss tests.

## Decision and remaining work

Retain this as a read-optimization candidate only. No production adoption.
The admission-only table still lacks feedback, migration, process-crash and
concurrent-owner coverage; canonical writer compatibility restrictions remain.
Read gain alone does not address the measured write-heavy admission bottleneck.
Before any integration, compare batched source lookups and full-service freshness
and tail latency, preserving the original validity guard and all durability
requirements. Do not spend further fresh quality data on this storage-only result.

All seven research goals remain open. Production, whitepaper and remotes untouched.
