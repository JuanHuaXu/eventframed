# V44 eager publication: complete mixed-load result

2026-10-04. **Correctness audit PASS; original latency adoption FAIL (0/16).**
After preserved single-string reader and pre-lock timing-check failures, a new
owner-time source freeze and complete repeat passed all thirteen native technical
root tests under race detection, adjacent admission/validity tests, vet,
52 scientific corruptions, eleven envelope negatives and five temporal negatives.
The actual repeat completes two repetitions of future/visible writes,
joined/archive storage and eager off/on. All 128 Recalls, 128 writes, sixteen
mixed outcomes, 150 nominees, eight workers, arrival schedules and full durable
acknowledgements remain unchanged. Warm search/core reuse are enabled in both arms.
The actual trace contains 154 of 1,024 core reuses where another owner built the
core after caller entry but before lock acquisition. All 1,024 are available at
the recorded acquisition time. This exercises the repaired lifecycle boundary
on real loaded traces, not only a synthetic positive control.

Two-repeat observed p99 ranges, milliseconds:

| Writes | Storage | Eager | Active call | Offer-to-Recall | Outcome |
| --- | --- | --- | ---: | ---: | ---: |
| Future | Joined | Off | 57.25-59.54 | 181.20-185.33 | 398.44-403.15 |
| Future | Joined | On | 50.26-54.90 | 154.84-180.58 | 364.01-391.93 |
| Future | Archive | Off | 75.55-77.58 | 261.20-279.89 | 463.30-470.32 |
| Future | Archive | On | 84.31-90.29 | 302.17-306.86 | 417.92-451.61 |
| Visible | Joined | Off | 50.83-56.84 | 147.51-148.39 | 349.14-364.17 |
| Visible | Joined | On | 45.21-54.30 | 146.75-158.71 | 359.24-367.83 |
| Visible | Archive | Off | 75.29-75.83 | 247.89-249.76 | 455.44-460.11 |
| Visible | Archive | On | 87.13-90.23 | 290.80-302.24 | 407.29-417.93 |

All active-call p99 values remain below 100 ms, but every trial misses the
100 ms offer-to-Recall and outcome p99 gates, and the 250 ms outcome maximum.
Writer p99 and view-age maximum remain below their 250 ms limits. Eager publication
improves joined future-write timing descriptively, worsens archived Recall timing,
and does not rescue the complete contract. Fixed arm order and two repetitions
are not population tail-latency certification or a universal causal effect.

Read-side owner acquisitions fall to one in eager trials, but this moves work:
each enabled trial has 158-160 separately counted publication attempts/owners,
30-32 publisher builds, and 13-15 unused publisher builds. ALL owned copies,
committed roots and acknowledgement windows are audited. Retained physical copy
totals per trial rise from roughly 0.54-0.58 MB to 0.91-0.99 MB. These are total
copied bytes, not peak RSS. Callback aggregate duration includes owner waiting;
it is not CPU time or the complete request cost. Returned acknowledgements include
trace overhead. Do not advertise the read-only owner count as total cost savings.

Future-only trials retain nonzero learned decisions/transport. Visible mutations
retain zero transported/learned decisions under the existing fail-closed epoch
guards. This remains an integration gap, not a successful cross-epoch learner.
The prime journal has no independent returned-ack timestamp in the old workload;
loaded operation acknowledgements are bound, but the prime scope gap is explicit.

Evidence: `research/eager-load-v44-owner-time/{freeze,completed,audit,readback}.json`,
exact raw trace and source copies; new technical preflight in
`research/eager-load-v44-preflight-owner-time-stream-repair`. Original negative
trace/source/tool failures remain in `eager-load-v44` and
`eager-load-v44-audit-stream-repair`. The repeat uses new `OwnerAt` measurements,
not fabricated timestamps for the old trace. No latency-gate relaxation occurred.

Next: profile the queue-inclusive durable acknowledgement path under the SAME
offered work before selecting a further ownership/persistence rescue. Merely
warming more metadata or quoting sub-100 ms active calls cannot satisfy Goal 6.
This rejects the eager-publication latency rescue at this scope, not all possible
background learning. All seven whole goals stay open; production/private data and
whitepaper remain untouched by this isolated study.
