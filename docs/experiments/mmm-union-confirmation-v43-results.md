# Unchanged union-read confirmation v43

Status: PASSED the same finite age/read screens, with all v42 experiment source
hashes unchanged. No tuning or instrumentation intervened. The repeat uses fresh
stores, not fresh real-agent tasks or a different overlap distribution.

[Raw artifact](mmm-union-confirmation-v43.jsonl) follows the unchanged
[v42 protocol](mmm-union-reads-v42-protocol.md). Twelve rotated cells, each with
192 recalls/four readers and 96 future writes. Times are nearest-rank milliseconds.

| Trial | Path | Complete / 192 | Dropped | Age p95 | Read p99 | Write p99 |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | Off | n/a | n/a | n/a | 32.993 | 18.024 |
| 0 | Non-durable group4 | 192 | 0 | 136.490 | 26.632 | 21.237 |
| 0 | Per-frontier reads | 182 | 10 | 288.100 | 27.403 | 32.958 |
| 0 | Union read | 192 | 0 | 117.955 | 27.954 | 28.795 |
| 1 | Off | n/a | n/a | n/a | 32.225 | 22.415 |
| 1 | Non-durable group4 | 192 | 0 | 113.956 | 26.656 | 24.354 |
| 1 | Per-frontier reads | 190 | 2 | 258.958 | 26.008 | 32.908 |
| 1 | Union read | 192 | 0 | 183.831 | 27.349 | 31.868 |
| 2 | Off | n/a | n/a | n/a | 34.988 | 20.228 |
| 2 | Non-durable group4 | 192 | 0 | 163.446 | 31.282 | 27.167 |
| 2 | Per-frontier reads | 178 | 14 | 311.943 | 25.324 | 32.954 |
| 2 | Union read | 192 | 0 | 187.931 | 28.674 | 33.728 |

Union reads again complete 576/576 observations, compared with 550/576 for the
same-run durable control. They admit/reread/discard 28,800 originals. Zero errors,
queue drops and entry expiries occur on the union path. All three age p95 <=250ms
and read-p99/off <=1.10 screens pass. Writer p99 remains above off in every trial
and slightly exceeds the durable control in trial 2. This is not a general
writer-latency protection result.

Across v42/v43, the union path completes 1,152/1,152 observations and all six
finite age/read screens pass. Age varies from about 118ms to 218ms; do not treat
this as a population latency bound. The fixture has high overlap and identical
as-of times, no labels/fits, and future-only ingestion. Distinct-query/disjoint
unit tests do not establish corresponding throughput. Earlier failures remain.

The full run passed accounting in 20.99s. Independent parsing verifies twelve
unique cells, exact source-hash parity with v42, embedded source integrity,
request/write counts, group-weighted conservation, 50 validations per completion
and exact durable counts. Artifact SHA-256:
`6e8d9f9b69458f9c60f836188c1faddce025b6ef1f88612aece47de3971fca68`.

Next work must go beyond this favorable storage fixture: audit unique durable
journal/event admission identity and retained-history/feedback authority before
claiming loaded learning, and broaden overlap/as-of traffic with no temporal
shortcut. Writer protection remains open. Neither direction 6 nor the whole
seven-direction goal is complete. No production installation, publication or push.
