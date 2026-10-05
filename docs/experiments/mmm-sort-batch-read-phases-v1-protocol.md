# Batch open-loop read phases v1: frozen diagnostic

Repeat two matched 4 ms offer pairs from the batch open-loop v1 fixture,
rotating arm order. Keep the same 200 eligible base rows, four future
sentinels, 256 visible writes, 192 exact-LSN searches, one owner per Store,
and the same one-event versus cap-16/16 ms writer policies. Replace only
the read call with an instrumented equivalent that records elapsed time
for publication capture, temporal snapshot lease, exact-LSN vector SQL,
and lease close. The SQL text, parameters and result validation must match
the existing private `sortPublicationGate.search` path. The phase sum must
account for almost all measured read-call time; report any discrepancy.

Report per-arm p50 and p99 for each phase, full read-call p99, offer-to-ack
p99, completeness and all integrity violations. This is a diagnostic with
no acceptance threshold. Do not reinterpret the two failed frozen
open-loop v1 screens as successes, extrapolate to full EventFrame Recall,
or change production. Preserve an unexpected result as evidence.
