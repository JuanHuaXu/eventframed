# Bounded decoupled tap drainage v29: results

Date: 2026-10-01. Research-only Goal 6 component; no production change.
The [frozen protocol](mmm-recall-decoupled-drain-v29-protocol.md) keeps
the v28 full-Recall fixture and adds a separate tap drainer with a
capacity-32 selected-frontier channel and a capacity-32 ordered reorder
map. Admissions and feedback remain event-time ordered and bounded.

## Ordinary matched trials

Three rotated learning-off/on pairs completed. Every enabled trial
consumed exactly 192 committed frontiers, durably admitted and
published 64 labels, and dropped zero tap observations. All v28
exact-nomination, no-future, phase-conservation, guarded as-of,
visible-mutation, journal-reopen, ledger, and replay checks passed.

| Trial | Off offer p99 | On offer p99 | On frontier-to-published p99 | On feedback-to-published p99 |
| --- | ---: | ---: | ---: | ---: |
| 0 | 48.340 ms | 35.019 ms | 41.372 ms | 21.625 ms |
| 1 | 48.241 ms | 35.496 ms | 38.325 ms | 19.979 ms |
| 2 | 49.575 ms | 36.331 ms | 42.211 ms | 22.549 ms |
| Pooled | 48.511 ms | 35.340 ms | 41.372 ms | 21.625 ms |

The pooled on/off offer-p99 ratio was 0.728. The frozen finite
component passes every on-arm <100 ms serving gate, pooled
on/off <=1.10, and both pooled <250 ms freshness gates. The apparent
on-arm speed advantage is a matched-fixture observation, not a
causal claim that learning accelerates Recall.

## Focused race and package checks

One focused `-race` trial passed in 52.63 s. A separate two-repeat
focused `-race` command passed twice more in 53.37 s and 53.69 s.
Each run completed the fixture's 192-frontier, 64-label, zero-drop,
durable/as-of/replay assertions without a reported Go data race.
Race-build timings were excluded from the ordinary latency gates.
`go vet ./internal/service ./internal/researchmemory` passed, as did
`go test ./internal/service ./internal/researchmemory -count=1 -timeout 5m`.

Commands:

```sh
EVENTFRAME_RUN_RECALL_DECOUPLED_V29=1 go test ./internal/service -run '^TestResearchRecallDecoupledDrainV29$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_RECALL_DECOUPLED_V29_RACE=1 go test -race ./internal/service -run '^TestResearchRecallDecoupledDrainV29FocusedRace$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_RECALL_DECOUPLED_V29_RACE=1 go test -race ./internal/service -run '^TestResearchRecallDecoupledDrainV29FocusedRace$' -count=2 -v -timeout 5m
go vet ./internal/service ./internal/researchmemory
go test ./internal/service ./internal/researchmemory -count=1 -timeout 5m
```

At-run SHA-256:

```text
9b21d808cc49380865e2ee66ea0b7e1689e573a0cb9ef588b198374e9f67b4c3  docs/experiments/mmm-recall-decoupled-drain-v29-protocol.md
a2c26f9b4d5e1032359a8987e9e5f7b1a52df273a6110167d73f169290dc269a  internal/service/research_recall_decoupled_drain_v29_test.go
44101187924a1d70a8f6b3d266b2cdd1012b32a9bbdbe683dbb907d5737846ea  internal/service/research_recall_live_learning_v26_test.go
5fd0470b957e5b3ed251f34f999157168a60ec2cbfd0c72cca17d7b9989c923f  internal/service/research_recall_ordered_pipeline_v28_test.go
```

The evidence supports the narrow explanation that admission-coupled
tap drainage caused v28's loss at this offered rate: decoupling it
removes the observed drops without changing tap capacity, journal
durability, or as-of policy. It does not prove capacity at higher
rates or arbitrary completion disorder, nor does it establish
cross-epoch recovery, power-loss behavior, real outcome labels,
OpenClaw serving, or a whole Goal 6 pass. The next test should freeze
a rate/burst and missing-frontier matrix with explicit bounded
overflow behavior before considering a production pipeline.
