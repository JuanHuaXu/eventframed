# Six-millisecond confirmation and four-millisecond phase trace v31

Date: 2026-10-01. Research-only Goal 6 follow-up to the
[frozen protocol](mmm-recall-cadence-v31-protocol.md). All runs use
the unchanged v29 decoupled learner, full 200-event Recall, eight
workers, 256 future-only writes, one guarded SQLite WAL/FULL journal,
192 offers and 64 bound durable labels per enabled trial.

## Fresh 6 ms confirmation

Three rotated off/on pairs finished with zero tap drops, all 64
labels per enabled trial, exact nomination, no-future/as-of,
visible-mutation rejection, durable journal reopen and replay,
and phase conservation. Measured median offer gaps remained close
to 6 ms in every arm.

| Trial | Off offer p99 | On offer p99 | On frontier age p99 | On feedback age p99 |
| --- | ---: | ---: | ---: | ---: |
| 0 | 56.136 ms | 36.792 ms | 74.141 ms | 24.133 ms |
| 1 | 48.607 ms | 39.598 ms | 62.735 ms | 26.055 ms |
| 2 | 49.130 ms | 42.071 ms | 78.408 ms | 27.066 ms |
| Pooled | 50.945 ms | 38.721 ms | 75.088 ms | 26.055 ms |

The pooled on/off offer-p99 ratio is 0.760. This passes every frozen
finite 6 ms component gate, including <100 ms per-trial/pooled
serving, <=1.10 pooled ratio and <250 ms pooled freshness. The
apparent on-arm serving advantage is not evidence that learning
speeds up Recall; the comparison is a capacity/safety gate.

## Fresh 4 ms diagnostic

A separate enabled-only fixture consumed all 192 frontiers,
published 64 labels, dropped zero observations and passed the
same lifecycle/phase checks. Median actual offer gap was 3.997 ms,
offer p99 41.944 ms, frontier-to-publication p99 **417.664 ms**,
and feedback-to-publication p99 25.779 ms. The v30 freshness failure
therefore repeated on a new fixture.

| Phase | p50 | p99 |
| --- | ---: | ---: |
| Tap enqueue to take | 0.114 ms | 2.099 ms |
| Tap take to feedback offer | 305.184 ms | 402.371 ms |
| Guarded feedback | 18.010 ms | 25.766 ms |
| Publication wait | 0.004 ms | 1.227 ms |

The five oldest labels (IDs 63, 64, 61, 62, 59) each had total age
402–418 ms. In those same individual labels, tap wait was
0.012–1.140 ms, pre-feedback wait 387.730–402.371 ms,
guarded feedback 12.247–14.638 ms, and publication wait at most
1.227 ms. Per-label phase conservation passed; this comparison
does not add different labels' p99s. The dominant delay is after
tap drainage but before feedback offer. That interval still combines
ordered admission, channel waiting and earlier feedback work, so
these data alone do not assign all delay to one of those operations.

Commands:

```sh
EVENTFRAME_RUN_RECALL_CADENCE_V31=1 go test ./internal/service -run '^TestResearchRecallCadenceV31ConfirmSixMillis$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_RECALL_CADENCE_V31_TRACE=1 go test ./internal/service -run '^TestResearchRecallCadenceV31TraceFourMillis$' -count=1 -v -timeout 5m
go vet ./internal/service ./internal/researchmemory
go test ./internal/service ./internal/researchmemory -count=1 -timeout 5m
```

The package tests and vet also passed after the cadence fixture change.

At-run SHA-256:

```text
f1d82bf805e15958a3fbf70709ffccf98d80a4aa48c452ef3b37ea429178e4ee  docs/experiments/mmm-recall-cadence-v31-protocol.md
dad12eea72224bccb5e22d8196b97e908f5c48b3849089dc5eb9992a44db8864  internal/service/research_recall_cadence_v31_test.go
f93928442b993f7f91d33661668a434f62f51772ecc2c639f7dcc9887b5b0d96  internal/service/research_recall_live_learning_v26_test.go
a2c26f9b4d5e1032359a8987e9e5f7b1a52df273a6110167d73f169290dc269a  internal/service/research_recall_decoupled_drain_v29_test.go
```

The repeated 6 ms pass establishes only a finite synthetic
component. It does not prove burst tolerance at other shapes,
cross-epoch/power-loss behavior, true outcome labels, agent-answer
benefit, or whole Goal 6 completion. A safe next step is finer
research-only attribution inside pre-feedback before changing
admission or update semantics.
