# Isolated full-Recall read-load diagnostic

Freeze configuration before timing. Frontiers50/200, workers1/4, two repetitions,
32 measured requests per arm. Compare normal service versus task+lexical overlay;
reverse arm order on repetition2. Fresh memory service per arm, four verified
public landing facts repeated under distinct IDs solely to exercise cardinality.
No independent-evidence or accuracy claim follows these duplicates.

Hash embedder32 removes network/model latency. Pack10, recall equal frontier,
budget10000, adaptive and diversity enabled. Two warm-up requests precede each
measurement. Unique sessions force distinct journal entries; no learned outcomes
or writes during the measured window. Record full-Recall durations, errors,
nominated/packed counts, explanation sizes, allocated bytes and wall time.
Journal reading/validation is outside each request timer but inside arm wall time.

This is finite closed-loop concurrency, not an open-loop arrival-rate benchmark,
durable storage, fitting, write-heavy churn, OpenClaw, or production qualification.
Use nearest-rank p50/p95 and max;32 observations do not support a meaningful
p99 guarantee. Report each arm separately, retain all failures, and compare
matching repetitions. No performance threshold or success claim is invented
after observing the numbers. All new artifacts use exclusive creation.
