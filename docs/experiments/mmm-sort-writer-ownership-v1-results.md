# Sort-key writer ownership v1: dual-open path fails

The frozen [private ownership probe](mmm-sort-writer-ownership-v1-protocol.md)
**fails** in both ordinary and `-race` runs on this host. This is a negative
Goal 6 result, not a production incident. The test is opt-in and uses only
temporary databases.

The initial three-row sortable collection and READY marker were valid at
LSN 21. An advisory owner-file lock denied a second lock descriptor, but a
second independent LibraVDB Store ignored that lock, opened the same path,
and returned success from a legacy unkeyed write at its own LSN 25. The
already-open first Store still reported LSN 21 and accepted READY at LSN 21.
That is an old as-of snapshot, not evidence that an unkeyed row entered its
query, but it is stale relative to the acknowledged write. After closing the
second Store and then the first, reopening the path reported LSN 21 and the
new row was absent. The successful second-Store write did **not** survive
this close order.

| Observation | Normal | `-race` |
| --- | ---: | ---: |
| First Store / second Store post-write LSN | 21 / 25 | 21 / 25 |
| First Store capture after second write | accepted LSN 21 | accepted LSN 21 |
| Reopened LSN | 21 | 21 |
| Acknowledged second-Store row after reopen | absent | absent |
| Test result | FAIL | FAIL |

The race run exited nonzero because of the same behavioral assertions; it
reported no Go data race. The probe does not establish that the first Store
ever served the unkeyed event, nor does it identify which close/checkpoint
operation discarded the second Store's state. It does establish that the
current advisory lock is not a database-enforced writer fence and that
simultaneous independent Opens cannot be used as the renewal protocol.

Reproduce the expected failure with:

```sh
EVENTFRAME_RUN_SORT_WRITER_OWNERSHIP_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortWriterOwnershipV1$' -count=1 -v -timeout 3m
EVENTFRAME_RUN_SORT_WRITER_OWNERSHIP_V1=1 go test -race ./internal/store/libravdbstore -run '^TestResearchSortWriterOwnershipV1$' -count=1 -v -timeout 3m
```

Protocol SHA256: `f51acba3e6a490f9b45d1620fdbefbb160e566bd09e0f9fe8ff99dc3d38b7e75`.
Test SHA256: `20b31cb7e3e9394474b2be9bb8b4aa056a49c608eb524c59f8292390e4cb3400`.

The [sequential close-then-open control](mmm-sort-writer-serial-v1-results.md)
subsequently passed. Next, design a single-owner daemon contract so no
independent Store can open
the live path. Only after that boundary is established should receipt-bound
incremental marker renewal be tested. A cooperative lock is insufficient if
an unmodified direct Store path remains available. No production path was
changed. Goal 6 and all seven whole goals remain open.
