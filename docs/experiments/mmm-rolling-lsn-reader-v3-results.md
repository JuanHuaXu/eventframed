# Rolling-LSN reader v3: sortable-key backend component pass

Protocol: [mmm-rolling-lsn-reader-v3-protocol.md](mmm-rolling-lsn-reader-v3-protocol.md).
This opt-in, test-only screen combines an atomically derived fixed-width
`available_at_sort` key with a captured runtime snapshot and exact durable
LSN. The private collection uses as-of `.120Z`, with 200 past rows at
`.100Z`, 16 future-only rows at `.125Z`, and 256 newly visible rows at
`.120Z`. The old variable-width SQL predicate would be wrong for this
fixture. Three rotated fresh pairs per run compared the current
`Store.Search` with the LSN SQL path at 192 k=50 offers per arm.

The final code passed the frozen component gates in three fresh
uninstrumented runs:

| Run | Current / candidate pooled read-call p99 | Ratio | Writer completion ratio | Candidate post-ack lag p99 |
| --- | --- | --- | --- | --- |
| 1 | 11.850 / 11.561 ms | 0.976 | 0.994 | 14.927 ms |
| 2 | 12.783 / 11.041 ms | 0.864 | 1.002 | 4.632 ms |
| 3 | 12.857 / 11.089 ms | 0.862 | 1.000 | 4.682 ms |

Each run completed 576 reads and 256 writes per arm, with no future or
unknown IDs, no omitted offers, correct per-version motion, final top-k
visibility of batch 15, and zero active temporal leases. Every one of the
472 stored EventFrames was checked after timing against its derived sort
key. A fresh exact-LSN query after reopen still returned batch 15 and
excluded future rows. The temporal API reported 10,064 retained bytes per
arm; this is neither RSS nor a steady-state retention bound. The measured
post-ack lag counts only reads **offered after** each acknowledgment.

The research-only batch method also passes focused tests for exact retry,
subsecond SQL ordering, runtime version/motion, close/reopen, rejection of
a missing sort key on an existing row, conflicting duplicate availability,
and a stored EventFrame whose time disagrees with its metadata key.
Ordinary store and research-batch package tests and `go vet` pass. A
dedicated race-instrumented concurrency run passed both arms without a
race report; its timings were not evaluated against the uninstrumented
gate. The unchanged v2 whole-second mode passed after the shared-harness
refactor, but remains as-of unsafe for fractional times.

**Decision:** v3 qualifies a finite, as-of-correct *research backend
component* and removes the known fractional future leak in this private
fixture. It does not prove general time-domain correctness, safe migration
of existing collections, full cross-DB/SQLite authority, learner
publication freshness, complete Recall latency, or Goal 6. The sortable
method is unwired; production `Store.Search`, collection creation and
daemon remain untouched. In particular, adding new keyed rows to an old
collection without proving every old row has the key would cause silent
omissions. A future production proposal needs a migration gate that fails
closed until coverage is complete, plus loaded service confirmation.

Reproduce with:

```sh
go test ./internal/store/libravdbstore -run '^TestResearchSortableEventBatch$' -count=1 -v -timeout 2m
EVENTFRAME_RUN_ROLLING_LSN_READER_V3=1 go test ./internal/store/libravdbstore -run '^TestResearchRollingLSNReaderV3$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_ROLLING_LSN_READER_V3_RACE=1 go test -race ./internal/store/libravdbstore -run '^TestResearchRollingLSNReaderV3Race$' -count=1 -v -timeout 5m
go test ./internal/store/libravdbstore ./internal/researchbatch -count=1 -timeout 3m
go vet ./internal/store/libravdbstore
```

Source SHA-256 at final run: research writer `49b75b00d4ccf375c40faf8e20ea8fbdc9cede2a74f37d900d425590dab28e18`;
writer test `a360a6a65c438224b49617a4459d4c455ce3ac469a41a0611699d1fb9fd4f66e`;
shared harness `9ae059272bb75476936a49ef0d0c405364398edd68a41aba5b0a5bf9df95a046`;
v3 test `5f63b5c35187ce8429d8699a321cbfa34eabfed50b49d096202d623a293a459c`;
protocol `33ef021046262c3e0e395db6b46ae6793ece3744947323ca7aa545e190ed58c8`.
