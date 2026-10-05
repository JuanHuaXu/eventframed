# Future-writer factor for guarded admission v36: results

Date: 2026-10-01. Research-only Goal 6 diagnostic against the
[frozen protocol](mmm-recall-writer-factor-v36-protocol.md). Four
fresh enabled-learning fixtures hold the full 200-event Recall,
eight workers, 192 offers, 64 selected durable labels, bounded
drainer/queues, exact as-of and one guarded SQLite WAL/FULL
journal fixed. Only the 256 concurrent future-only `Observe`
calls are removed in the writer-off arms. All four arms consumed
192 unique frontiers, published 64 labels, dropped zero tap
observations, and passed exact nomination, no-future,
visible-mutation rejection, journal reopen, durable replay
and phase conservation.

| Gap | Writer | Writes / overlap | Offer p99 | Frontier age p99 | Guard entry p50 / p99 | Candidate validation p50 / p99 |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| 6 ms | On | 256 / 69 | 37.956 ms | 45.322 ms | **13.647 / 21.338 ms** | 1.192 / 1.979 ms |
| 6 ms | Off | 0 / 0 | 8.307 ms | 4.560 ms | **0.0009 / 0.0065 ms** | 0.739 / 1.095 ms |
| 4 ms | Off | 0 / 0 | 9.093 ms | 5.460 ms | **0.0009 / 0.0084 ms** | 0.791 / 1.488 ms |
| 4 ms | On | 256 / 39 | 43.092 ms | **429.125 ms** | **14.957 / 23.823 ms** | 1.200 / 2.413 ms |

Actual median offer gaps were 5.98–6.06 ms and 3.99–4.01 ms,
respectively. With no concurrent writer, the five oldest 4 ms
labels had guard-entry times below 2.2 microseconds and total
ages below 5.5 ms. With the writer on, the five oldest labels
had guard-entry times around 11.9–13.4 ms and total ages
417–429 ms. Thus this controlled fixture strongly supports
future-write contention at the shared guard as the cause of
the admission throughput/freshness failure. The writer-on
4 ms arm still serves Recalls under 100 ms but does not meet
the <250 ms freshness criterion.

The writer-off arm is **not a rescue**: it removes a required
concurrent workload. In writer-on arms only 69 (6 ms) and
39 (4 ms) of 256 writes overlapped the offer window, so
many writes finished after offers. The measured effect is
specific to this scheduling and store fixture; no claim is
made about arbitrary burst patterns or production throughput.
Any future priority or batching candidate must report writer
latency and completion as well as Recall and learning freshness,
and preserve as-of validity and durable admission.

Command:

```sh
EVENTFRAME_RUN_RECALL_WRITER_FACTOR_V36=1 go test ./internal/service -run '^TestResearchRecallWriterFactorV36$' -count=1 -v -timeout 5m
go vet ./internal/service ./internal/researchmemory ./internal/researchpublicationstore
go test ./internal/service ./internal/researchmemory ./internal/researchpublicationstore -count=1 -timeout 5m
EVENTFRAME_RUN_RECALL_DECOUPLED_V29_RACE=1 go test -race ./internal/service -run '^TestResearchRecallDecoupledDrainV29FocusedRace$' -count=1 -v -timeout 5m
```

Vet and package tests passed after the instrumentation changes. The
focused `-race` lifecycle run also passed with 192 frontiers, 64
labels, zero drops and no reported Go data race; race-build timing
is not used for the ordinary performance comparisons.

At-run SHA-256:

```text
563c12ba3056f953d41925e2e008c45eecdfdced1e40ae196b85259b1f7566e3  docs/experiments/mmm-recall-writer-factor-v36-protocol.md
0e7144934175383046fc4ef8b8ccf9ebd0aa29260655abd7598c4aea3da36e9b  internal/service/research_recall_writer_factor_v36_test.go
84454b1f3492902975aa45da2037e1eb6f8d0dadde35e14b9c027c7267a4a35f  internal/service/research_recall_live_learning_v26_test.go
563ceaeccbc20ecc450858698f725c4e82657500513e62a9579d2bd1e0b7ec2a  internal/service/research_recall_decoupled_drain_v29_test.go
```

No production code changed. This is causal attribution in one
synthetic fixture, not whole Goal 6 completion.
