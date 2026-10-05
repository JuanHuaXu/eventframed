# Research group-journal dwell diagnostic v23: frozen protocol

Date: 2026-10-01. This is a research-only Goal 6 causal diagnostic, not a
production change or a replacement for the failed v22 screen.

## Question and competing explanations

The v22 group journal's loaded offer p99 was 642 ms versus 203 ms for the
single guarded SQLite journal. The downstream queue was 593 ms; its exact
source is not established. Candidate explanations are: (1) the 8 ms dwell
causes queue buildup, (2) one worker and the shared as-of writer permit
serialize commits behind concurrent event writes, or (3) Recall computation
and its four workers saturate independently of journal scheduling. The
single-journal and graph-cache controls, plus quiet versus writer cases,
separate some but not all of these effects.

## Frozen comparison

Run the existing v22 full Recall fixture with 200 live events, exact
`RecallK=200`, `PackK=10`, four workers, 192 offers 8 ms apart per trial,
and 256 future-only writes in the writer case. Compare the existing
8 ms-dwell group worker with an otherwise identical **1 ns-dwell** group
worker. The latter is an operationally no-wait schedule but still uses the
same maximum batch size 8, independent per-entry as-of checks, one owned
writer permit, SQLite WAL/FULL transactions, post-commit acknowledgements,
idempotency/conflict rules, and reopen-count check. Include the existing
single guarded SQLite path as a concurrent-workload reference. Rotate arm
order over three trials and reverse quiet/writer order in the middle trial.

Before examining timings, require all 576 offers per arm/case, all 768
future-only writes per writer arm, no future packed event, exactly 200
distinct nominations per offer, 192 durable acknowledged journals per trial
after close/reopen, zero batch-guard rejections, and no failed as-of contract
or focused race test. Report offer, call, queue, journal p50/p99 and batch
sizes. The fixed Goal 6 candidate screen remains writer offer p99 <100 ms;
the no-wait arm must also be no worse than the single SQLite reference in
the same run. A diagnostic with broken invariants is invalid, not fast.

If no-wait removes most group-specific queueing but remains above 100 ms,
the dwell was a contributor, not a Goal 6 rescue. If no-wait remains near
8 ms-dwell, dwell is not the principal cause under this fixture; pursue
guard/publication architecture or worker contention. If both arms vary
wildly, repeat the same frozen workload before attributing causality.
Do not infer production p99, power-loss recovery, or real-agent freshness
from this finite research fixture. Preserve negative results and source
hashes. The 1 ns arm is not a claim that a new batching algorithm exists.
