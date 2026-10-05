# Margin adapter service replay and reporting caveat

Completed90 isolated service calls: focus, original ordinal adapter and margin
adapter on W3C/ESA written/ESA ISO. All consumed data, pack1, four-record
frontier, memory backend, local nomic embeddings, no updates or generation.
Service and adapter both explicitly use0.25 as the maximum additive correction.

Margin top1 matches ordinal on all30 questions: W3C7/8, ESA written8/8,
ESA ISO7/8. Every proper forecast law for every admitted event remains exactly
equal to focus control (max difference0). Source identity, metadata, order and
hash checks pass. No correction records were injected in THIS full-service run;
the separate actual delta/packing tests establish the bounded perturbation case.

## Certainty Side Effect

PacketAnswerCertainty is currently derived from the normalized rank-score gap.
The margin encoding changes that number without changing evidence or forecast
laws. Relative to the original ordinal0.25, some cases become0.8421052631578947;
others become0.052631578947368356. Maximum absolute movement is0.5921052631578947.
This is NOT a59.2-point gain in calibrated confidence or factual accuracy.

The service comments already describe rank-boundary certainty as ranking
decisiveness, not factual correctness. Nevertheless, downstream interpretation
and elastic-delta modulation must not treat a reserved score band as new evidence.
Proper-law preservation alone does not address that integration surface.
An uncalibrated flag/separate constraint-priority channel is needed before
promoting this adapter; no such reporting fix is claimed here. Do not tune a
calibration map on these few consumed cases and call that validation.

## Cost

Go1.27.1 darwin/arm64, Apple M4, GOMAXPROCS4. Five repetitions,200ms benchmark
target; medians of benchmark averages, not tail-latency statistics.

| Candidates | Original ordinal | Margin | Margin bytes/op | Margin allocs/op |
| --- | --- | --- | --- | --- |
|50 |66.674 us |116.167 us |19248 |440 |
|200 |214.415 us |402.563 us |70964 |1640 |

Margin repeats date parsing and adds allocations; measured cost is about1.74x
and1.88x the original in these component fixtures. The same repeated Rosetta
facts used by the earlier benchmark exercise cardinality, not diverse evidence.
No I/O, persistence, queueing or loaded-service p99 is included. About0.403ms
for200 candidates is not an end-to-end latency guarantee.

## Reproduction

```sh
node research/public-task-pilot/check-calendar-margin.mjs
go test ./internal/researchcalendar -run '^$' -bench BenchmarkCalendarMargin -benchmem -benchtime=200ms -count=5 -cpu=4
```

Raw benchmark: [calendar-margin-benchmark.txt](calendar-margin-benchmark.txt).
Service artifacts: margin-results.json in [W3C](w3c-focus-v1/margin-results.json),
[ESA written](esa-date-v1/margin-results.json), [ESA ISO](esa-iso-v1/margin-results.json).

Decision: preserve the conditional ranking-resilience result, but do not adopt
the margin adapter as calibrated memory confidence. Next: separate priority
constraints from confidence/modulation and establish a durable, replayable
explanation contract. Production is unchanged; all whole research goals remain open.
