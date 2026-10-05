# Rolling-LSN reader v1: frozen visible-write Goal 6 screen

This is an opt-in, test-only design screen. It makes no production change.
The preceding [fixed-LSN result](mmm-pinned-lsn-reader-v1-results.md) used
future-only writes, so it cannot establish freshness for a normal new turn.
This screen asks whether a new durable LSN can be paired with the runtime
publication snapshot for *immediately visible* writes without losing the
reader-tail benefit.

For each of three fresh, rotated matched pairs, seed 200 past-visible and 16
future-only same-tenant records. Offer 256 as-of-visible records in 16
ordered raw batches of 16 at nominal 16 ms intervals. Offer 192 search
requests at nominal 4 ms intervals to eight reader workers. Query k=50 with
the same vector in both arms. The candidate obtains the runtime snapshot,
exact latest durable commit LSN, and an LSN lease while holding the store's
research write read-lock. It then releases that lock and executes the SQL
vector search `AS OF LSN` with an `available_at <= asOf` filter. The control
captures the same runtime snapshot and uses the existing `Store.Search`.
Both arms check `ResearchSnapshotCompatible` after search and retry on a
visible concurrent mutation, up to four attempts; a failed or exhausted
request is a failure, not an omitted latency sample. The lease is closed on
each candidate attempt. Query, guard, LSN capture and lease cost are inside
the measured request latency. Preseeded future records must never be returned.

For each accepted read, verify 50 unique IDs from the captured version's
eligible prefix. Once any visible write is published, at least one visible
result must appear. Track each batch's durable acknowledgement time and the
first accepted read response that reflects its version or later. Record
read-call/offer p50 and p99, actual offer gaps, retries/rejections, each
batch's acknowledgment-to-visible-read lag, writer completion and writer
offer-to-ack p99, and temporal lease/retention stats. At the end, a fresh
read must reflect visible writes and exclude future-only rows; all 256
versions/motions must be present, and active leases must return to zero.

Frozen design gate: zero incorrect, exhausted or omitted requests; all 192
reads and 256 writes per arm; candidate pooled read-call p99 <=1.10x
control and <=100 ms; candidate summed writer completion <=1.25x control;
candidate batch-to-visible-read lag p99 <=25 ms. Report all pair values and
the worst lag. A pass qualifies only a backend component for separate
service-level confirmation. It does not establish durable learner freshness,
full Recall latency, power-loss recovery, or Goal 6 completion. Do not
retune the four-attempt limit, 25 ms lag guard, or offered rates on these
cohorts.
