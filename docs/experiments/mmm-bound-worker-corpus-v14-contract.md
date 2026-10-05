# Bound-worker frontier ceiling v14: frozen screen

Frozen 2026-10-01 before v14 results. Research-only, no serving default or
production change. The v13 50-live-event, 8 ms offer-rate writer arm passed
under a short-text hash-embedder fixture but had little latency headroom.
The causal question is whether widening the as-of-visible nominated frontier
to the research tap's 200-candidate ceiling breaks that finite rate.

Use the same persistent local LibraVDB plus durable-lineage store, guarded
bound worker, four probe workers, 192 paced Recall offers, 64 witnessed
feedback labels, and 256 future-only writes per writer trial as v13. Compare
50 and 200 live records in quiet and writer arms, three paired trials each;
alternate arm order. Set `recall_k` to the live count and `pack_k=10` in both
sizes. Each learner frontier must contain exactly the expected live IDs;
every packed probe candidate must be available as of its query. The worker
still learns only the `seed` outcome. The 50-cell is a contemporaneous control,
not evidence imported from v13.

Report completed calls, labels, writes, overlapping operations, backend
versions, and nearest-rank p50/p95/p99/max for offer-to-return, plus call,
queue, and offered-label-to-observed-completion distributions. The frozen
writer gates are Recall offer p99 <100 ms and label-age p99 <250 ms, with
all expected operations and identities correct. If a cell cannot complete
under its 90-second trial deadline, that is a failure, not censored latency.
Do not loosen the gate after observing results. `-race` is a correctness
check only, never a performance comparison.

This deliberately retains same-topic short synthetic EventFrames, local
contracts, and hash embedding to isolate the widened frontier. It cannot
establish real-agent, long-text, remote-contract, large-corpus, or population
tail claims even if the 200-cell passes.
