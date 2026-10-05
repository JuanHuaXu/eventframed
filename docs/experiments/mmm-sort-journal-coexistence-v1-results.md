# Sort publication and Recall journal coexistence v1: failed control

Date: 2026-10-02. Frozen [protocol](mmm-sort-journal-coexistence-v1-protocol.md).
The opt-in private probe calls the real `Store.PutBayesianJournal` against
a verified three-EventFrame publication gate, as `Service.Recall` does.
It is deliberately retained as a failing test.

The valid journal was durable, but LibraVDB moved from LSN 21 to 24 while
the EventFrame runtime snapshot stayed unchanged. The event-only marker
remained at LSN 21, so `capture` returned not READY and the next authorized
EventFrame append failed with `LibraVDB moved outside the batch gate: 24 != 21`.
This confirms an integration blocker, not a corruption or future-data leak:
a metadata-only write cannot be treated as an EventFrame, and the direct
Recall journal path cannot coexist with the event-only publication gate.

The separate [journal-through-gate study](mmm-sort-journal-gate-v1-results.md)
tests an explicit metadata transition. It does not make this direct-path
failure pass or authorize changing production Recall.

Reproduce the expected failure:

```sh
EVENTFRAME_RUN_SORT_JOURNAL_COEXIST_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchSortJournalCoexistenceV1$' -count=1 -v -timeout 5m
```

SHA-256: test file
`7278f4dc779ef3afd8cf7422d7afc8f83eb4062b6c01d606a461aa1bc01bf941`;
protocol `0fd81c4f341ff37709131948a6bd6ddff94b9ae16aae42c6a0e53ef5dbbda377`.
