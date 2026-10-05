# Guarded validation-only v59 results

**Mechanism check PASS3/3; full resolved non-harm screen FAIL0/3.** Guarded union
validation/preparation, without any learner persistence, halves scheduled read
p99 relative to idle wrapped off in every trial. It also shifts delay onto
writers. Learner persistence is not necessary for this large shift in the
measured fixture. The responsible operation within that package remains unknown.

[Protocol](mmm-guard-only-v59-protocol.md),
[artifact](mmm-guard-only-v59.jsonl). SHA-256:
`c6c9dfa1aefc9cef23d7bbdc1493f09a376d78e656c9d41711bf01afdfc198d8`.

## Verification

```sh
go test -race ./internal/service -run '^TestResearch(GuardOnlyAccounting|IdleWrapperAccounting|OfferedLoadAccounting)$' -count=3
go test -race ./internal/service -count=1
go vet ./internal/service
EVENTFRAME_GUARD_ONLY_ARTIFACT=<LOCAL_ROOT>/docs/experiments/mmm-guard-only-v59.jsonl go test ./internal/service -run '^TestResearchGuardOnlyExperiment$' -count=1 -v
```

Targeted race tests PASS25.785s; full service race tests PASS64.087s; vet clean.
Measurement Go PASS29.847s confirms accounting/integrity, not deployment readiness.
Tests and measurement did not overlap. Every captured source/hash pair matches
the measured file. Only test fixtures changed; no runtime or default change.

Eighteen fresh public cells record3456 recalls and1728 future writes, offered
every5/10ms. Every scheduled index, due/start/end tuple and inside-call duration
was independently checked. There are no drops, entry expiry, unexpected errors
or original mismatches. Durable arms verify86400 originals and86400 terminals.
Guard-only validates28800 candidate records across576 observations, with zero
durable admissions/terminals/spans. Total validated candidates115200; this is NOT
115200 stored original forecasts. Guard-only temporary previews are discarded,
never published, trained or treated as evidence. No labels/fitting/private data,
production or actual agent-outcome evaluation.

## Scheduled latency

Nearest-rank p99, milliseconds, from original offered due time to call completion:

| Trial | Off read / write | Idle read / write | Guard-only read / write | Raw durable read / write | Combined source read / write | Resolved source read / write |
| --- | --- | --- | --- | --- | --- | --- |
| 0 | 553.424 / 100.738 | 648.426 / 176.699 | 116.770 / 275.246 | 30.298 / 394.446 | 33.286 / 367.450 | 32.488 / 245.717 |
| 1 | 566.585 / 134.736 | 548.754 / 113.968 | 35.551 / 381.613 | 28.998 / 396.299 | 33.486 / 345.553 | 32.080 / 380.506 |
| 2 | 628.638 / 158.888 | 589.612 / 134.924 | 92.926 / 217.822 | 34.336 / 402.484 | 27.864 / 338.695 | 35.087 / 331.362 |

Guard-only/idle read ratios are approximately0.180,0.065,0.158, meeting the
frozen halving check in all trials. Guard-only writers are slower than idle in
all three. This is a read/write tradeoff, not unqualified throughput improvement.

Guard-only observation-age p95 is33.281,36.929,25.458ms. Resolved source age is
98.642,74.992,134.403ms. All accepted-observation age and full-completion checks
pass at this rate, and resolved scheduled reads pass against unwrapped off.
Resolved scheduled writers fail in every trial, so joint non-harm remains0/3.
No total offered-to-background-completion quantile is claimed; queue age and
scheduled service latency are separate measured boundaries.

## Phase evidence

Mean milliseconds per accepted group:

| Trial | Guard-only preparation / entry / callback | Resolved preparation / entry / callback | Resolved admission within callback |
| --- | --- | --- | --- |
| 0 | 1.955 / 9.220 / 2.019 | 1.205 / 5.240 / 7.613 | 5.883 |
| 1 | 2.585 / 10.218 / 3.081 | 0.745 / 5.187 / 8.146 | 6.112 |
| 2 | 1.490 / 8.771 / 2.060 | 1.003 / 6.162 / 7.015 | 5.193 |

Entry includes waiting plus native snapshot validation; it is not pure lock time.
Groups are formed from available observations and differ in size/frequency
between arms, so these means are not equal-size isolated operation costs.
Guard-only retains journal prefetch, cold preview preparation, the publication
guard and real union validation. This experiment does not separate those pieces
or establish a specific native-lock fairness mechanism.

## Next lead

Profile blocking/call boundaries in the unchanged guarded-validation path,
distinguishing publication-gate waiting from native-store reads and callback
work. Use the measured boundary evidence before proposing a critical-section
change or reusing validated source data. Any such change must preserve as-of
coherence, actual originals, atomic identity/retry checks and stop-on-uncertainty;
the diagnostic never authorizes dropping validation from an active learner.

v55/v57/v58 failures remain recorded; v56's isolated cost success and this
mechanism finding do not complete direction6. Other rates, concurrent fitting,
feedback/history authority and true agent outcomes remain open. All seven
directions stay IN PROGRESS. Nothing pushed, deployed or changed in whitepaper
accuracy claims.
