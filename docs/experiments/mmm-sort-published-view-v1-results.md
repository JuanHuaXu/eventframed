# Last-published-LSN read v1: private backend rescue passes

Date: 2026-10-02. Frozen [protocol](mmm-sort-published-view-v1-protocol.md).
The research-only reader atomically exposes the last fully journal-certified
LSN after marker publication. A search pins that LSN without taking the
writer gate lock, applies the same exact-LSN vector SQL and availability
filter, and rejects absent/poisoned or >=250 ms old views at return. The
writer publishes the pointer before acknowledging its batch. This is safe
only under the declared single-owner, gate-only writer contract; it is not
an out-of-band writer or multi-process authentication mechanism.

The functional probe passes: an old certified LSN remains readable while
a later DB-only commit is pending, without exposing that unjournaled event
or the future sentinel; poisoning then denies. Expiry denies. A complete
batch publishes and verifies after reopen, while DB-only interruption does
not publish on reopen.

Two fresh unchanged three-arm runs pass the frozen private backend screen.
Each run contains three rotated blocks per arm, 256 event offers and 192
searches per block at actual median gaps near 4 ms. All events are present,
sortable and journal-verified on reopen; no search error or future/unknown
ID appeared.

| Fresh run | Single/original read p99 | Batch/original read p99 | Batch/published read p99 | Rescue read ratio vs single | Rescue writer age p99 | Rescue max view age | Violations |
| ---: | :--- | :--- | :--- | ---: | :--- | :--- | ---: |
| 1 | 11.273 ms | 15.053 ms | **3.226 ms** | 0.286 | 32.841 ms | 26.099 ms | 0 / 0 / 0 |
| 2 | 11.684 ms | 14.913 ms | **3.423 ms** | 0.293 | 32.843 ms | 26.969 ms | 0 / 0 / 0 |

The unchanged batch/original-read arm still exceeds the earlier <=1.10
read ratio; the new read architecture passes it with substantial margin.
This is a *different candidate*, not a correction to the failed original
screen. A race-instrumented concurrent batch/published run completed 256
acknowledgements and 192 reads with zero violations; its 17 groups and
143.040 ms maximum view age are correctness observations only, not
comparable performance timings. Ordinary `libravdbstore` tests and `go vet`
pass.

The reader trusts an in-memory pointer certified by the gate. Every write
must go through that gate, and a writer error must poison it. External raw
Store/DB writes, concurrent independent Opens, power-loss behavior, huge
corpora, full Service Recall, feedback processing and background learning
are not covered. Reopen still verifies the full O(N) journal. Therefore
this is a Goal 6 backend component pass, **not Goal 6 completion**; all
seven whole research goals remain open. The next test must join this reader
to full Recall and durable labels under an enforced single-owner lifecycle.

Reproduce:

```sh
EVENTFRAME_RUN_SORT_PUBLISHED_VIEW_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortPublishedViewV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_SORT_PUBLISHED_OPENLOOP_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortPublishedOpenLoopV1$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_SORT_PUBLISHED_VIEW_V1=1 EVENTFRAME_RUN_SORT_PUBLISHED_OPENLOOP_RACE_V1=1 go test -race ./internal/store/libravdbstore -run '^(TestResearchSortPublishedViewV1|TestResearchSortPublishedOpenLoopRaceV1)$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore -count=1
go vet ./internal/store/libravdbstore
```

SHA-256: load test
`55798efba0428904627ee12b09159b870e5d9482b75246f7dff74885f2a8823c`;
slot test `96bc755d5ebfd345950ec925e828091d79bfc93dfca7259c23b94e5a120a95f5`;
protocol `5f42effe7410ae84409dcb441c2d9fff7fa226289f13d1ad28c558c6e7204c43`.
