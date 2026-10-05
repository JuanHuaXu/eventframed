# Bounded ready-group admission v27

Frozen after the v26 phase diagnostic and before execution. Twelve rotated arms:
off, prior batch, group1 and group4, three trials each. Retain fresh persistent
LibraVDB, 192 recalls, four readers, 96 future writes at 2ms spacing, 50 visible
candidates, 64-slot handoff and 20ms entry deadline. No labels, fitting, ledger
operations, retries or production installation.

Group1 and group4 use the SAME new consumer. It prepares original forecasts and
request bindings before acquiring the guard; group1 controls this ordering change
against prior batch. Group4 drains at most four already-ready observations, never
waits for a full group and holds at most 200 pending original records. It acquires
one as-of guard and independently validates every observation's complete batch
under that guard. All fixture as-of times must match. No group is counted accepted
unless every member validates. A rejected group is discarded without evidence.

Per-group phases and GroupSizes record acquisition/processing work once; accepted
age and admission/drop/expiry counts remain PER OBSERVATION. PrefetchNS in the
group modes also includes original-record preparation, unlike prior batch.
Verify weighted phase accounting, group cap, 50 validated records per accepted
observation, request/write counts and no invented learning. Small read-only and
write controls run under race first. Then exclusive-create the full raw artifact.

Report nearest-rank read/write p99, accepted-age p95, guard duration p95 and
group-size distribution. Retain the 250ms age and 1.10 paired read-p99 screens;
also report any write-cost transfer. No success claim from fewer acquisitions
alone. This is a research scheduling experiment, not a completed durable bridge
or predictive improvement, and no larger cap is tuned on these results.
