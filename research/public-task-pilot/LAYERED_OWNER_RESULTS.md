# Bounded layered snapshot owner

Research-only `layered_owner.go` adds a single prepared-candidate slot, a bounded
number of reader handles, and a bounded number of retired version objects.
Tests use two readers/one historical version and eight readers/two versions.
One current version is separate from these limits. Version objects may share
the same underlying tree, so this is conservative counting, not unique-byte
accounting. No daemon or backend code was changed.

Preparation reserves the writer and a possible historical slot before releasing
the mutex. This matters because a reader can acquire the current root while
preparation runs. A full retired-version budget rejects preparation even when
the current root has no readers at that instant. Publication is then independent
of reader release timing. Failed preparation releases its reservation.

Leases return copied records, not raw roots. Releasing the final reader removes
the retired version reference. Finished candidates drop their prepared snapshot
reference. Close refuses live readers or candidates rather than invalidating
them. Initial snapshot aliases outside this owner remain the caller's obligation.

## Verification

Three race runs of `TestLayeredOwner*` passed in 1.336s. Covered invariants:

- Reader capacity and single-candidate admission reject excess work.
- A retained-version limit blocks publication preparation until the last old
  reader releases, not merely until one of several readers releases.
- Old leases see old records; new leases see the published records.
- Commit/abort cannot be repeated; completed candidates drop their root.
- Cancellation releases the writer reservation; a subsequent prepare succeeds.
- Concurrent lookup/release is race-free and subsequent reads reject released
  handles. Close rejects unfinished work and rejects acquisition after closing.

## Remaining work

This is neither a durable transaction owner nor an enforceable RAM budget.
Record counts, vector dimensions, and retained version sizes still matter; one
stalled reader can retain a large old tree. Total memory, preparation scratch,
identifier buffers, and external aliases need separate accounting. Commit
currently allocates a version object and may grow a map; it must not be used as
an allocation-free durable publication callback.

Lookup copies a record while holding the owner mutex. No concurrency-throughput
claim is made; trace/benchmark before considering serving integration. Context
cancellation does not interrupt mutex acquisition. This model still receives
precomputed edits and does not run HNSW update discovery. Next work must address
those load-bearing integration requirements, not treat lease-unit success as a
sustained-load rescue. All seven whole goals remain open.
