# Idle publication wrapper v58 results

**Wrapper-alone halving check: FAIL0/3.** Idle wrapping does not reproduce the
large active-arm read improvement at the middle offered rate. Its effect is
variable, however: only one trial stays within10% of unwrapped off, so the
protocol's stronger near-equivalence condition for dismissing its contribution
is not met. The exact mechanism remains unresolved; no runtime patch follows.

[Protocol](mmm-idle-wrapper-v58-protocol.md),
[artifact](mmm-idle-wrapper-v58.jsonl). SHA-256:
`c3af88636172ba872cf57388d87fc83815b0de85b2b43a8b90377370d8592c71`.

## Verification

```sh
go test -race ./internal/service -run '^TestResearch(IdleWrapperAccounting|OfferedLoadAccounting|OfferedTimingChecks)$' -count=3
go test -race ./internal/service -count=1
go vet ./internal/service
EVENTFRAME_IDLE_WRAPPER_ARTIFACT=<LOCAL_ROOT>/docs/experiments/mmm-idle-wrapper-v58.jsonl go test ./internal/service -run '^TestResearchIdleWrapperExperiment$' -count=1 -v
```

Targeted race tests PASS21.622s; full service race tests PASS62.538s; vet clean.
Measurement Go PASS24.949s is accounting/integrity, not non-harm validation.
Testing and measurement did not overlap. All captured sources match their
hashes and measured files. No runtime implementation or default changed.

Fifteen fresh public cells, each192 reads/four lanes and96 writes/one lane,
offered every5/10ms. All2880 reads and1440 writes retain complete unique
due/start/end records and exact inside-call duration identities. All nine active
cells complete192 background observations, totaling86400 originals and86400
verified terminals. Zero drops, expiry, unexpected errors or original mismatches.
Idle cells have zero attempts, phases, admissions and terminals. No learner log
or research consumer is opened for them. Ordinary Recall journal persistence
still occurs in both off and active service modes; it is not disabled by this
control. No labels, fitting, private data, production or agent-outcome evaluation.

## Scheduled latency

Nearest-rank p99, milliseconds, from offered due time to call completion:

| Trial | Off read / write | Idle wrapper read / write | Raw durable read / write | Combined source read / write | Resolved source read / write |
| --- | --- | --- | --- | --- | --- |
| 0 | 505.562 / 43.822 | 558.654 / 118.960 | 39.493 / 391.609 | 32.246 / 393.376 | 38.928 / 336.100 |
| 1 | 634.666 / 168.900 | 616.532 / 157.781 | 27.563 / 364.680 | 31.158 / 336.580 | 46.442 / 339.410 |
| 2 | 598.592 / 124.786 | 477.631 / 48.983 | 35.150 / 327.348 | 28.283 / 296.215 | 27.235 / 364.544 |

Idle/unwrapped read ratios are approximately1.105,0.971,0.798. No trial halves
read latency, while active reads remain far below both off variants. Idle
wrapping alone is insufficient in these measurements, but it is not proven
irrelevant or equivalent to off. Interaction with guard work is still possible.

Resolved observation-age p95 is94.206,160.959,79.937ms. All three age and
completion screens pass; scheduled reads pass against unwrapped off, but all
three scheduled writer non-harm screens fail. The joint resolved screen stays
0/3. Do not substitute idle off as the baseline or infer success from read gains.

The active arms again move delay toward writers. Resolved writer lateness p99
is326.100,329.411,354.545ms, versus inside-call write p99 of33.961,31.088,30.343ms.
As in v57, backlog is visible only when the original offered time is retained.
An idle wrapper's extra snapshots/serialization do not by themselves explain
the large shift, and this comparison does not identify a specific lock mechanism.

## Next controlled comparison

Retain the real publication guard and union validation, but omit learner
admission/cleanup. An existing grouped validation consumer can support that
control without weakening any active arm. Compare off, idle wrapper, validation
only, raw durable, combined source and resolved source under the same schedule.
This separates guarded validation from learner persistence before changing
guard placement, fairness or commit semantics. Guard-only observations must
never be counted as durable original admissions or as learning evidence.

v55/v57 failures and v56 isolated cost success remain separately recorded.
Other offered rates, concurrent learning and full feedback/history authority
remain open. All seven research directions stay IN PROGRESS. Nothing pushed,
deployed or changed in whitepaper accuracy claims.
