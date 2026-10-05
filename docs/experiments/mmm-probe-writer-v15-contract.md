# Probe/writer without learner v15: frozen attribution control

Frozen 2026-10-01 before v15 outcomes. This is a sibling diagnostic for the
v14 200-candidate writer-arm overload, not a latency rescue or production
change. Use the same local persistent LibraVDB plus durable-lineage wrapper,
short same-topic synthetic records, hash embedder, four probe workers, 192
Recall offers paced 8 ms apart, and 256 future-only writes in writer arms.
Compare 50 and 200 as-of-visible records, each quiet and writer, three
paired trials with alternating arm order. Use `recall_k` equal to the live
count and `pack_k=10`; verify exact nominated live IDs and exclude future
records from the packed packet. There is **no bound worker, no admitted
prediction, and no feedback** in any arm.

Report per-cell completed calls/writes, overlap, snapshot versions, and
nearest-rank p50/p95/p99/max offer-to-return plus call and queue p99.
If the 200 writer arm again exceeds 100 ms p99 with correct full frontiers,
the bound learner is not necessary for the observed overload on this fixture.
If it remains below 100 ms, learner/guard interaction is implicated but not
yet localized. The 50 and quiet controls must be measured in the same run.
Do not classify a failed operation as a latency sample or relax a threshold
after inspection. This no-learner run has no label-age endpoint.
