# Batch cap v2: frozen reader-protection design screen

The v1 batch-16 acknowledgement load screen failed its <=1.10 search-call
p99 ratio despite a large writer gain. The reader-factor diagnostic points
to cross-tenant global write exclusion plus additional same-tenant effects.
This v2 screen changes only the maximum batch size to 4 or 8; it does not
alter raw LibraVDB, the sidecar, service code, the 4 ms coalescing bound,
event times, or the as-of/read workload.

Run three fresh rotated triples of single-write control, batch-4 and
batch-8 using `runBatchAckLoad`. Each arm seeds 200 past-available events,
offers 256 future-only writes nominally every 1 ms, and offers 192
concurrent `Search(k=200)` calls nominally every 4 ms with eight readers.
Require 256 nonduplicate durable acknowledgements, 192 as-of-correct
searches, all motion entries and no drops in every arm. Record p50/p99
write offer gaps and batch-size distributions to expose backpressure.

The frozen v1 screening criteria remain unchanged for each candidate:
pooled writer completion ratio <1, pooled offer-to-ack p99 ratio <=1.25,
and pooled search-call p99 ratio <=1.10 versus the pooled single-write
control. Report per-triple values and all pooled ratios. This is a design
screen; even a pass needs a fresh independent timing confirmation and full
`Service.Recall`/learner validation before any service integration. The
same deterministic fixture family is reused with fresh stores, so this
screen cannot establish workload generalization. Race-instrumented timing
must not be used to assess the performance gates.
