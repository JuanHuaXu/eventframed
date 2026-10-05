# Per-guard event union reads v42

Frozen before execution. Twelve rotated cells: three trials of off, non-durable
group4, prepared durable postverify with per-frontier event reads, and the same
prepared path with one per-guard union event read. PreparedWrites and UnionReads
explicitly identify the durable cells. Keep all v41 settings: 192 recalls/four
readers, 96 future writes at 2ms, 50 candidates, ready groups <=4, queue64, 20ms
entry deadline, fresh persistent stores, FULL commits and complete originals.

Share only event data for the same tenant/as-of time under the held mutation
guard. At most four frontiers and 256 total record references. Different journals,
queries and baselines still run independent checks and feature extraction. No
cross-call cache, candidate truncation or omitted validation. Reject missing,
ambiguous, duplicate prediction identities and mixed tenant/time groups before
admission. Every completion includes post-guard readback and typed discard.

Run overlap/disjoint, distinct query/journal, missing/ambiguous event, identity,
scope, stale/cancel and no-cross-call-authority tests, then small persistent load
under race and full ledger/learner/service race/vet. Preserve exclusive JSONL with
raw timings, selected sources/hashes, group counts and both flags. No labels/fits
or production changes. The load fixture is highly overlapping; unit correctness
on disjoint groups is not a throughput claim for disjoint real traffic.

Screens remain accepted-age p95 <=250ms and read-p99/off <=1.10 in each trial.
Report drops/expiry and writer tails independently. A finite pass does not close
warm loaded learning, service identity, history/feedback authority or real-agent
quality. Do not change any prior failure or global success criterion.
