# Incremental publication append scale v1: frozen private diagnostic

Use two fresh private, single-owner LibraVDB Stores with identical schema,
vectors and initial three events. The control uses the sortable receipt
writer without publication; the candidate uses the test-only receipt-bound
incremental journal and READY publication gate. The control's marker becomes
stale by design after its first direct write and is never used for serving.
Never share a database path between the arms.

Grow both arms with the same event tape to 35, 259 and 1,027 rows. At each
size, run 100 paired one-event writes, alternating which arm writes first.
Measure each call from entry to acknowledgement with Go's monotonic clock;
check for errors and duplicate receipts. The two arms have equal event count
at the beginning of every pair. After each block, capture the candidate
READY marker 100 times with no concurrent writes. Report writer p50/p99/max,
candidate/control ratios and capture p50/p99 at every size.

This is an explanatory scaling diagnostic without an acceptance threshold.
It measures isolated per-event persistence and publication cost, not an
offered-rate service queue, loaded Recall, crash durability, or large-corpus
behavior. In particular, do not infer a 4 ms freshness pass from a small
writer median or bypass the single-owner contract. Preserve negative results;
production must remain untouched.
