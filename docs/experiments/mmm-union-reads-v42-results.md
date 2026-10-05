# Per-guard event union reads v42 results

Status: PASSED the finite completion-age/read screens on the highly overlapping,
same-as-of fixture. Writer overhead remains; this does not complete direction 6
or validate warm loaded learning or real-agent quality.

## Change and boundary

The group validator reads the union of event IDs once under the existing service
mutation guard, for at most four frontiers and 256 record references. All records
must have the same tenant and as-of time, with distinct prediction IDs. Each
frontier still independently checks its journal, query digest, baseline, feature
extraction and publication compatibility. Only event data is shared, never
query-specific features or authority across calls. The original batch validator
delegates to the same checking implementation using its normal event getter.

Focused tests cover overlapping and disjoint event sets, different valid queries
and journals, missing/duplicate/unknown event returns, invalid features/baselines,
mixed tenant/time, repeated prediction/member identities, empty/oversized input,
cancellation and stale publication. The valid cases make one event read and two
journal reads; a later policy change prevents reuse. Three focused race and
small persistent-load repetitions pass, as do full ledger/learner/service race
suites and vet. These are correctness checks, not disjoint-workload throughput
measurements. Admission and FULL commit remain guarded; full original readback
and typed discard remain post-guard and included in completion age.

## Same-run results

[Frozen protocol](mmm-union-reads-v42-protocol.md),
[raw artifact](mmm-union-reads-v42.jsonl). Twelve rotated cells, each with
192 recalls/four readers and 96 future writes. Durable arms both use prepared
SQL and differ only in UnionReads. Times are nearest-rank milliseconds; age is
conditional on completion and does not assign fast values to dropped requests.

| Trial | Path | Complete / 192 | Dropped | Age p95 | Read p99 | Write p99 |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | Off | n/a | n/a | n/a | 32.340 | 17.696 |
| 0 | Non-durable group4 | 192 | 0 | 112.076 | 24.065 | 22.930 |
| 0 | Per-frontier reads | 190 | 2 | 248.966 | 24.544 | 36.909 |
| 0 | Union read | 192 | 0 | 209.836 | 28.744 | 28.902 |
| 1 | Off | n/a | n/a | n/a | 38.955 | 17.922 |
| 1 | Non-durable group4 | 192 | 0 | 126.283 | 25.722 | 24.860 |
| 1 | Per-frontier reads | 175 | 17 | 297.959 | 26.677 | 35.033 |
| 1 | Union read | 192 | 0 | 210.062 | 26.064 | 30.057 |
| 2 | Off | n/a | n/a | n/a | 37.146 | 19.305 |
| 2 | Non-durable group4 | 192 | 0 | 107.037 | 26.937 | 21.847 |
| 2 | Per-frontier reads | 181 | 11 | 280.872 | 28.407 | 36.268 |
| 2 | Union read | 192 | 0 | 217.937 | 26.714 | 31.609 |

Union reads complete 576/576 observations versus 546/576 for separate reads.
All 28,800 experimental original records are admitted, reread and durably
discarded. Zero errors, queue drops or entry expiries occur on the union path.
All three age p95 <=250ms and paired read-p99/off <=1.10 screens pass. Write
p99 improves against the durable control in all trials but remains above off.
This supports a bounded overlap optimization, not general zero serving overhead.

Independent parsing verified twelve unique trial/mode/PreparedWrites/UnionReads
cells, embedded source hashes, request/write counts, group-weighted conservation,
50 validations per completion and matching durable counts. The full run passed
accounting in 20.91s. Artifact SHA-256:
`4d8f8102e9e9efadc2f1df2856c6cf008ee1b68a30ddfc0c9f401a59651cd49f`.

## Confirmation and limitations

Repeat the exact unchanged source/protocol in a new exclusive artifact before
calling the finite result repeatable. This fixture deliberately has high event
overlap and identical as-of times. Actual traffic can have disjoint frontiers or
different times; those groups cannot be merged by silently weakening temporal
validation. Fresh overlap/time distributions, writer protection, warm loaded
learning, durable service identity, verified feedback and retained-history
authority remain open. No production installation, publication or push occurred.

The [unchanged v43 repeat](mmm-union-confirmation-v43-results.md) subsequently
passed the same finite screens; it does not remove these scope limitations.
