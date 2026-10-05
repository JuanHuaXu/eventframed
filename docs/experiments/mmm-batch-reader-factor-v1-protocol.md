# Batch reader-factor diagnostic v1: frozen protocol

This is a Goal 6 *diagnostic*, not a rescue or a pass/fail promotion screen.
It follows the failed [batch-16 load screen](mmm-batch-ack-load-v1-results.md)
without changing that artifact or any production code. The observed search
tail could reflect same-tenant future-corpus growth, writer exclusion/CPU
contention, or both.

Use the real LibraVDB plus private SQLite batch-intent prototype. In each
fresh arm, seed 200 past-available events in tenant A. Issue 192 tenant-A
`Search` calls at a nominal 4 ms cadence with eight readers, `k=200` and
as-of before all future events. The four cells are:

1. `base_static`: 200 tenant-A events, no writer.
2. `expanded_static`: preinsert 256 future tenant-A events, then no writer.
3. `other_tenant_writer`: write 256 future tenant-B events in 16-event batches
   during tenant-A searches.
4. `same_tenant_writer`: write 256 future tenant-A events in 16-event batches
   during tenant-A searches.

The two writer cells offer one batch every nominal 16 ms, starting with the
read offers. Rotate cell order in three fresh blocks. Record search-call
p50/p99, offer-to-complete p99, actual read offer gaps, writer completion,
and whether all 192 searches return exactly the 200 past-eligible tenant-A
events with no future leakage. Verify +256 runtime versions and complete
future-only motion where writes occurred. Use the same query and event IDs
across cells. No hidden labels, fitting, or adaptive threshold changes.

`expanded_static/base_static` estimates the cost of a larger same-tenant
future-heavy corpus without concurrent writes. `other_tenant_writer` tests
writer interference while tenant-A search corpus stays fixed, but cannot
separate global lock exclusion from CPU use. `same_tenant_writer` includes
both effects. Treat relative ratios as descriptive only: scheduler timing,
index state and measurement noise prevent additive causal decomposition.
Do not interpret a result as full `Service.Recall`, learner freshness, or
production throughput.
