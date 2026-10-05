# Sort-key serial handoff v1: private control passes

The frozen [serial ownership control](mmm-sort-writer-serial-v1-protocol.md)
passes normally and under `-race` on this host. Unlike the
[dual-open failure](mmm-sort-writer-ownership-v1-results.md), the first
Store and SQLite gate were fully closed before a second Store opened the
same temporary LibraVDB path.

The initial three-row READY marker named LSN 21. The second Store wrote one
legacy unkeyed event, recorded LSN 25 and closed. The reopened first Store
reported LSN 25, retained the new row and runtime version 4, denied the stale
READY marker, and refused a full-scan republish while the row lacked a sort
key. The normal and race runs produced the same 21->25 boundary and no race
report.

```sh
EVENTFRAME_RUN_SORT_WRITER_SERIAL_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortWriterSerialHandoffV1$' -count=1 -v -timeout 3m
EVENTFRAME_RUN_SORT_WRITER_SERIAL_V1=1 go test -race ./internal/store/libravdbstore -run '^TestResearchSortWriterSerialHandoffV1$' -count=1 -v -timeout 3m
```

Protocol SHA256: `fa947b1d9752ab79e82935b7cd7dd4a16359a2c6a0970f2fbc68fd8725c5bc59`.
Test SHA256: `321f434c7149b43b7254b9f9e0d63ea99566712f5bcf6eebc7a24d3da59a5168`.

This confirms a useful boundary, not a per-event restart design. A scalable
Goal 6 path needs one long-lived daemon owner, no independent direct Open of
its live database, and an incremental marker protocol within that ownership
epoch. The advisory lock is only cooperative; it cannot enforce exclusivity
against an unmodified Store caller. No production path was changed. Goal 6
and all seven whole goals remain open.
