# Future-arrival persistent load ablation

Same load and screens as persistent v2, with an explicit temporal intervention:
initial50 events available before each query cutoff;64 concurrent writes become
available strictly AFTER that cutoff. No feedback/model changes. The added data
must not enter earlier queries. Changing temporal visibility also holds candidate
count at50 rather than growing it; cross-workload latency is therefore confounded.
Compare on/off only within this workload. Do not replace the failed backfill run.

Run with EVENTFRAME_RESEARCH_FUTURE_INGEST=1 and a new artifact path. Three pairs,
256 reads and64 writes each. Same queue/callback. Added invariant after shutdown:
Accepted=Completed+Stale+Failed+Cancelled. The cancellation accounting fix is in
effect, but no snapshot or journal acceptance rule was modified.

Companion deterministic tests inject writes immediately before journal commit:
future writes commit first try;5 successive backfills exhaust5 attempts;3
backfills then quiet permit fourth-attempt commit. Both memory and persistent
stores must behave alike, without future evidence in the returned packet.

This ablation can identify a temporal boundary, not establish service reliability
under all writes or make the original zero-error requirement disappear.
