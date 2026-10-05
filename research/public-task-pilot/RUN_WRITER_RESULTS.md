# Durable immutable-run writer boundary

Status: write publication boundary implemented; flush and consolidation remain
unimplemented. `DurableRunWriter` owns one writer gate, a semantic revision and
bounded pending delta over a caller-owned coherent run set.

Before entering persistence it validates mutations, reserves/prepares the delta,
copies callback inputs, and builds the complete next query plan. A successful
callback must atomically commit records plus revision. Afterwards the writer
only publishes prepared state and releases exclusion; it ignores late cancellation.
New snapshots cannot cross the in-flight durable callback. Historical snapshots
remain immutable. Any callback error or panic quarantines fresh snapshots and
future writes; discarding unpublished state does not assert storage rollback.

No fallible graph build or query-plan construction happens after durable success.
Full-manifest preflight remains potentially expensive and has not been timed
under concurrent load. Graph ownership, authenticating recovery, enumerating a
complete recovery snapshot and preventing external writes remain caller contracts.

## Tests

The callback-boundary tests pass three race-enabled repetitions (1.382s):

- A blocked durable callback blocks new snapshot admission.
- Successful publication advances exactly one revision; old views stay unchanged.
- Callback mutation of its private copy cannot corrupt published vectors.
- Invalid vectors/capacity fail without calling persistence.
- Cancellation after callback success does not hide the committed state.
- Lost acknowledgement and callback panic both quarantine future operations.

The real-store test performs a synchronous libravdb transaction for two records
and revision1, then deliberately returns a lost-ack error. The writer quarantines.
After close/reopen, both vectors and revision are checked, a new graph is built
from the recovered vectors, and the recovered snapshot emits the expected ranking.
This is clean-close/reopen recovery after an uncertain acknowledgement, not a
process-kill or power-loss test. It tests inserts only; it is not a general
recovery enumerator or delete-transaction adapter.

Full bulk/candidate-only researchindex suite, including real recovery, passed
three race-enabled repetitions in17.461s:

```sh
go test -modfile=research-candidate-only.mod -overlay research/public-task-pilot/candidate-only-overlay-v1/overlay-local.json -tags research_candidate_only -race ./internal/researchindex -count=3
```

## Next step

Implement captured-version flush publication: build a run outside the writer
gate, reacquire it, remove only delta records covered by the captured revision,
retain newer writes/tombstones, and publish matching run/plan pairs without
changing semantic revision. A failed or quarantined publication must retain
proper accounting for the unpublished graph. Then add consolidation and reader
leases and test original sustained-load gates through repeated compaction cycles.
No production changes or whole-goal completion.
