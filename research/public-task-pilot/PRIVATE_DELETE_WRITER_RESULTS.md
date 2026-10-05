# Private deletion publication coordinator

`PrivateDeleteWriter` serializes graph-only deletion preparation and publishes one
immutable graph/entry-summary/revision bundle after a persistence callback returns
nil. It uses the existing research PersistGeneration contract: synchronously delete
the authoritative ID and commit its revision atomically. Exact graph topology is
not serialized by that callback. Recovery must restore/rebuild coherent state.

The bundle and mutation slice are allocated before callback entry. Successful
publication is a pointer replacement under the writer gate. New views use the
same context-aware gate and cannot observe a partially published bundle. Existing
views remain immutable historical snapshots. Late cancellation after durable
success does not roll back the published result.

## Tests

Three targeted race runs passed in1.359s. The callback tests cover:

- revision and deletion payload, immutable old view during blocked persistence;
- successful publication after callback-side mutation and late cancellation;
- callback error/panic quarantine, no automatic retry of uncertain persistence;
- current-view rejection while quarantined and preserved historical views;
- missing target/revision overflow rejected before persistence without quarantine.

These callbacks are test doubles. They do NOT establish actual disk durability,
crash behavior, or correctness of a libravdb transaction adapter. That integration
and reopen/recovery validation are the next required checks.

## Remaining limitations

Deletion only: no private insertion, ID index, loaded topology validation, reader
retention budget, shutdown owner, or recovery enumerator. Constructor inputs must
already be coherent. The earlier bounded lease owner is not yet integrated here;
callers can retain old views indefinitely. The persistence callback must honor its
atomic synchronous contract and must not reenter this writer. No callback duration
or total serving-latency bound is proved.

All seven whole research goals remain open. Production and standard dependencies
are unchanged; do not use callback-unit success as evidence of sustained durable
throughput or completed database functionality.
