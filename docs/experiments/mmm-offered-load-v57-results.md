# Matched offered-load v57 results

**FAIL the full offered-load non-harm screen:0/9 resolved cells pass jointly.**
Observation age is acceptable at the two slower offered rates, but writer
latency fails its relative screen in eight of nine cells. This is a capacity
and contention diagnostic, not a successful rescue of v55 or direction6.

[Frozen protocol](mmm-offered-load-v57-protocol.md),
[raw artifact](mmm-offered-load-v57.jsonl). SHA-256:
`112b76a61c464a6c2e331e3185c8ec2ad28bafa35cc5c589a9b15752483ea246`.

## Verification

```sh
go test -race ./internal/service -run '^TestResearch(OfferedTimingChecks|OfferedLoadAccounting|ResolvedAdmissionAccounting)$' -count=3
go test -race ./internal/service -count=1
go vet ./internal/service
EVENTFRAME_OFFERED_LOAD_ARTIFACT=<LOCAL_ROOT>/docs/experiments/mmm-offered-load-v57.jsonl go test ./internal/service -run '^TestResearchOfferedLoadExperiment$' -count=1 -v
```

Targeted race tests PASS24.401s; full service race suite PASS60.707s; vet clean.
Measurement Go PASS67.761s means accounting/integrity, not latency-screen success.
No test execution overlapped measurement. Source snapshots/hashes match recorded
and local files. Only test fixtures changed; no runtime implementation or default.

Thirty-six fresh isolated public cells contain6912 recalls and3456 future writes.
Every offered index appears exactly once; recorded due/start/end times preserve
ordering, scheduled spacing and exact inside-call duration. All cells drained.
Ordinary observation/guard/admission/terminal accounting passes:258650 accepted
originals and terminals, zero unexpected errors or original mismatches. Eleven
observations are dropped overall, all in the fastest source-owner cells. No
entry expiry. No labels, fitting, private data, production or agent-outcome test.

## Resolved arm

Nearest-rank quantiles, milliseconds. Scheduled latency means end-minus-original
due time, including waiting behind previous work in the producer lane. Age starts
at frontier enqueue. These are distinct populations/boundaries, not an estimate
of total offered-request-to-background-completion latency.

| Read / write offered interval ms | Trial | Completed | Age p95 | Off / resolved scheduled read p99 | Off / resolved scheduled write p99 |
| --- | --- | --- | --- | --- | --- |
| 2 / 4 | 0 | 192 | 211.856 | 964.755 / 442.848 | 717.215 / 790.057 |
| 2 / 4 | 1 | 192 | 286.731 | 1004.597 / 433.920 | 758.806 / 813.084 |
| 2 / 4 | 2 | 190 | 299.238 | 985.565 / 401.033 | 725.314 / 812.330 |
| 5 / 10 | 0 | 192 | 114.512 | 623.539 / 29.839 | 153.856 / 350.207 |
| 5 / 10 | 1 | 192 | 155.969 | 586.584 / 31.876 | 134.833 / 392.751 |
| 5 / 10 | 2 | 192 | 109.763 | 620.596 / 28.982 | 146.836 / 406.129 |
| 10 / 20 | 0 | 192 | 25.793 | 19.574 / 18.859 | 13.986 / 23.029 |
| 10 / 20 | 1 | 192 | 26.866 | 17.694 / 19.965 | 11.999 / 13.364 |
| 10 / 20 | 2 | 192 | 29.410 | 19.573 / 20.381 | 13.187 / 21.084 |

Age passes7/9 cells, scheduled read non-harm8/9, scheduled write non-harm1/9,
and full192-observation completion8/9. No cell passes all four. In particular,
the fastest trial0 writer value narrowly exceeds1.10x off; keep that failure.
All three100-read/s,50-write/s trials finish every observation with age below30ms,
but this does not erase their writer overhead or prove a general sub100ms bound.

At the middle rate, resolved inside-call write p99 is28.066-29.874ms, while
scheduled write p99 is350.207-406.129ms. Writer start-minus-due p99 is340.207-
396.130ms. Much of the experienced delay is backlog, not time inside one call.
At the fastest rate, resolved read lateness p99 is391.118-434.753ms and write
lateness782.064-806.074ms. Reporting only inside-call times would hide overload.

## All controls

Ranges span three trials, not confidence bounds. Completion counts are summed;
off has no background observations by design.

| Read interval ms | Arm | Background completed / offered | Age p95 range ms | Scheduled read p99 range ms | Scheduled write p99 range ms |
| --- | --- | --- | --- | --- | --- |
| 2 | Off | Not applicable | Not applicable | 964.755-1004.597 | 717.215-758.806 |
| 2 | Raw durable | 576/576 | 158.037-212.127 | 431.808-548.293 | 805.844-866.365 |
| 2 | Combined source | 567/576 | 253.133-314.926 | 361.845-463.126 | 787.750-818.549 |
| 2 | Resolved source | 574/576 | 211.856-299.238 | 401.033-442.848 | 790.057-813.084 |
| 5 | Off | Not applicable | Not applicable | 586.584-623.539 | 134.833-153.856 |
| 5 | Raw durable | 576/576 | 77.335-122.008 | 30.259-33.447 | 386.987-417.666 |
| 5 | Combined source | 576/576 | 128.918-159.005 | 28.995-30.264 | 357.202-395.521 |
| 5 | Resolved source | 576/576 | 109.763-155.969 | 28.982-31.876 | 350.207-406.129 |
| 10 | Off | Not applicable | Not applicable | 17.694-19.574 | 11.999-13.986 |
| 10 | Raw durable | 576/576 | 19.800-23.798 | 19.356-21.508 | 17.330-24.682 |
| 10 | Combined source | 576/576 | 26.273-36.797 | 19.425-21.471 | 16.485-20.437 |
| 10 | Resolved source | 576/576 | 25.793-29.410 | 18.859-20.381 | 13.364-23.029 |

At the middle rate, even raw durable shifts the large off read backlog toward
writers. This is not evidence that resolved admission or MMM improves reading.
All active arms also install the publication-store serialization wrapper; off
does not. Wrapper effects, guarded work and persistence are not separated by
these four arms. The improvement is an observed read/write tradeoff, not a
proven scheduling mechanism or an unqualified throughput gain.

## Next lead

Add an idle wrapped control: identical serialization wrapper, but no frontier
consumer or learner persistence. Compare it with unwrapped off and the current
active arms under the same fixed schedules. This isolates a plausible cause of
the read/write shift before changing lock ownership, fairness or durability.
Keep all source-identity, actual-original, atomic-commit and uncertainty checks.

Retain v55's failed closed-loop screen and v56's isolated cost success separately.
The new offered-load test is stricter about writer delay, not a retroactive
threshold change. Feedback/history authority, fitting under load, broader data
and true agent outcomes remain unproven. All seven directions remain IN PROGRESS.
Nothing pushed, deployed or changed in the whitepaper's accuracy claims.
