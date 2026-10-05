# Durable run load screen

Same offered workload as candidate-load: CPU4,768d, initial800/3200/6400 records,
two fresh repeats;1024 reads at200/s and512 append writes at100/s,100ms deadlines
from scheduled arrival, no retries,64 pending records, flush trigger32,5ms poll,
eight reader leases and two retired GRAPH slots. One builder at a time.

Start with one whole-base graph. At trigger32, flush if one run exists; otherwise
consolidate the two current runs plus captured delta. This avoids increasing the
retirement limit to accommodate large batches. It is a different layout from
eight ID partitions: initial whole-base ANN limitations remain visible. At most
two current graphs plus two retired and one unpublished graph are intended.

Use real synchronous authoritative record/revision transactions and the same
isolated bulk/candidate-only derived graph overlay. Log every read/write/build,
consolidation versus flush, before/after pending counts and revisions. Drain
requests and builder, close/drain lease pool, then close current graphs; record
final drain separately. Reopen authority and check every acknowledged ID and
revision; failed-present writes are a failure. No model/private data or production.

Pass requires no request errors, no late response, no successful self-query miss,
no build/audit errors, all acknowledged IDs recovered and revision=1+acknowledged.
Require at least two successful consolidations per arm to count as multi-cycle
evidence. Gates do not excuse failures because a component ran faster.
This remains an index/durability workload, not full EventFrame semantic validation.
