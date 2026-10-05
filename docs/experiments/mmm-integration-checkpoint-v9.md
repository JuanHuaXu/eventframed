# Integration checkpoint and durable-learning boundary

Combined command passed:

```sh
go test -race ./internal/researchmemory ./internal/researchpublication ./internal/researchpublicationstore ./internal/store/... ./internal/service -count=1
```

All seven reported packages passed, including persistent store and service.
Opt-in load experiments skip without their artifact environment variables; this
command is integration/race evidence, not another performance replication.
The separately recorded v7/v8 load outcomes remain the performance evidence.

## Confirmed remaining boundary

ResearchFeedbackBridge stores seen journals and pending bindings in maps, rejects
admission at256 seen journals, and clears pending state on Close. Background
abandons queued labels at shutdown; Adapter stores samples, mixture weights and
pending original forecasts in memory. Reopening the event store restores none of
these objects. Current continuous-learning claims must remain bounded-lifetime.
Removing the cap or evicting tombstones would create replay risk, not a rescue.

## Next implementation contract

Build a separate research durable consumer, not a silent change to serving:

- Persist admission identity `(tenant, journal, event, learner-contract)` and the
  original pre-outcome expert forecasts before acknowledging admission.
- Persist verified outcome identity and availability before acknowledging it;
  learner application is replayable and keyed, not counted twice after restart.
- Preserve event ordering, model seed, fit cadence, contract fingerprint, epoch,
  samples and mixture state, or replay an equivalent complete ordered log.
  Recomputing historical forecasts with a new model is not equivalent replay.
- Commit checkpoint and applied-log position atomically. A crash between durable
  feedback acceptance and checkpoint must replay once; after checkpoint it must
  not reapply. Only a durable outcome record may generate a learning update.
- Restore dependency authority independently. A retained prediction record is
  not proof that an old publication snapshot is still reusable after restart.
- Bound in-memory working state without deleting the durable deduplication key.
  Retention/compaction needs an explicit replay horizon or durable tombstones.

Before load tests, exercise crashes at each acknowledgment/checkpoint boundary,
duplicates, conflicting outcomes, source changes, partial writes and incompatible
contracts. Compare restored predictions and counts to an uninterrupted control.
This is a proposed contract, not implemented durable recovery or a success claim.
Other research directions, especially real-task predictive quality, remain open.
