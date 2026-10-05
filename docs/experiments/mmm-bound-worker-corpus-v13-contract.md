# Bound-worker corpus breadth v13: frozen diagnostic

Frozen 2026-10-01 before v13 results. Research-only; no production change.
The v12 8 ms-offer cell passed with one as-of-visible event. This experiment
tests whether that finite operating point survives a 50-event nominated
frontier while labels still bind to the original `seed` event.

Use the existing persistent LibraVDB plus durable-lineage service, bound
worker, hash embedder, 64 source-bound admitted feedback labels, four probe
workers, 192 paced Recall offers, and 256 future-only writes in writer arms.
Run both one and 50 live records, each with quiet and writer arms, in three
paired trials with alternating order. Keep offer interval at 8 ms. The 50
records are synthetic same-topic events available before the query as-of;
future writer events must never enter the as-of frontier. Use recall 50 and
pack 10 for the 50-record arm, versus recall 3 and pack 2 for one record.
This deliberately changes packet work as well as corpus size; it measures
the full widened serving path, not an isolated database-scaling coefficient.

Report per-cell nearest-rank p99 offer-to-return, call, queue, and
offered-label-to-observed completion age, plus counts, overlap, and snapshot
versions. The one-record paired control must reproduce an ordinary-build
writer Recall p99 <100 ms and label-age p99 <250 ms. The 50-record arm passes
the same frozen bounds only if all 64 labels per trial complete, all 256
future writes finish in writer trials, all calls succeed, each learner
frontier contains all 50 expected as-of records, and both writer percentiles
remain below their bounds. A failed cell is not patched or reclassified after
seeing results. The companion quiet arm and call/queue split localize the
source of a failure; they do not by themselves prove causality.

This is a single M4 local finite workload, not population p99, a billion-
event corpus, physical power-loss recovery, or agent-outcome evidence.
