# Raw as-of Store.Search under future writes v16

Frozen 2026-10-01 before v16 outcomes. Research-only diagnostic after the
v14/v15 overload. Use the same local persistent LibraVDB plus durable-lineage
wrapper, hash embedder, short same-topic 50 and 200 live records, and 256
future-only `Service.Observe` writes in writer arms. Four workers take 192
paced offers at 8 ms. Each probe calls only `Store.Search` with the same
as-of time, fixed 8-dimensional query vector, tenant, and result limit equal
to the live count. No `Service.Recall`, Bayesian journal, packing, learner,
or feedback occurs in the probe path.

Run 50/200 by quiet/writer, three paired trials each, alternating arm order.
Every result must contain exactly the expected live IDs and no future ID;
all 192 probes and all writer-arm writes must complete. Report p50/p95/p99/max
offer-to-return and p99 call and pre-worker queue separately, with backend
version and writer overlap. This is a diagnostic threshold comparison to
100 ms, not a frozen promotion rule. If raw search overloads, backend search
plus writes is sufficient under this schedule. If it does not, trace the
remaining service phases; do not identify Bayesian journaling as the sole
cause without a separate phase measurement.

The raw call avoids service work and may change relative scheduling. It does
not measure agent serving, remote LibraVDB contracts, or production p99.
