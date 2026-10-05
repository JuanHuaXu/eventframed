# Guarded validation-only diagnostic v59

Frozen before measurement. v58's idle wrapper did not reproduce the active-arm
read gain. Separate grouped guarded validation/preparation from learner durable
admission and cleanup before changing runtime behavior. No runtime changes.

Six rotated arms, three trials,5ms read/10ms write offered intervals: unwrapped
off, idle wrapper, guard-only, raw durable, combined source, resolved source.
Eighteen cells. Keep192 reads/four lanes,96 writes/one lane,50-event overlap,
K50/pack10,queue64,groups<=4,20ms entry budget,FULL writes,fixed as-of,future
writes and parent context. All original offered due times and lateness remain.

Guard-only uses the existing grouped consumer and actual union validation under
WithResearchAsOfSnapshotWait, including journal prefetch and cold preparation.
It opens no learner ledger and performs no durable original or terminal write.
Temporary previews are discarded, not published or trained. Require entered
callbacks for every accepted observation,Validated=50*Accepted, and zero durable
counters/spans/post-guard work. Do not count validation-only observations as
stored forecasts or evidence. Other arms retain all existing checks. Nil/default
fixtures keep their old behavior; union reads are newly permitted for group4.

Run accounting/race and vet before measurement; drain each public isolated cell.
Capture raw timings and source/hash snapshots in exclusive JSONL. No labels,
fitting, private data, production, push or deployment. Preserve earlier failures.

Primary diagnostic: guard-only scheduled read p99 <=half same-trial idle wrapped
off in every trial would reproduce a large shift without learner persistence.
Report writer cost, completion, age and phase times regardless. This isolates a
guarded validation/preparation package, not its individual operations. A failure
does not prove any one lock or database operation is the cause.

Keep the full resolved screen visible against unwrapped off:192 completions,
age p95<=250ms,scheduled read/write p99<=1.10x off. A diagnostic pass is not
non-harm or completion of direction6. Other rates, actual learning and all seven
direction-level requirements remain open; no threshold tuning or claim rescue.
