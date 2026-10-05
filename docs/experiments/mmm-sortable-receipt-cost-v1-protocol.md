# Sortable receipt cost v1: frozen private comparison

Compare the existing research-only sortable batch writer with the opt-in
receipt-returning variant on the same host. Run three independent paired
blocks, rotating arm order, each with a fresh private LibraVDB collection,
four-dimensional vectors, 16 warm-up writes and 128 measured one-event
commits. Measure only the API call, excluding collection creation and close.
Every write must be new; every receipt-arm call must return a positive exact
LSN, and the final row count and snapshot increments must match.

Report per-block p50 and p99, pooled p50 and p99, and receipt/control ratios.
The frozen cost screen passes only if every paired-block p50 ratio is <=1.10
and every p99 ratio is <=1.20. Preserve a failure rather than relaxing the
ratios. Do not use race-instrumented times. This is an isolated sequential
writer comparison, not loaded service p99, durability recovery or Goal 6.
