# Rolling-LSN reader v2: frozen captured-view Goal 6 screen

The [v1 screen](mmm-rolling-lsn-reader-v1-results.md) could not complete:
the current path starved under a post-read *current-version* requirement,
and SQL could not bind undeclared availability metadata. The separate
[schema probe](mmm-lsn-availability-schema-v1-results.md) established the
predicate for a **new private collection**. This v2 comparison is test-only;
it does not modify production, migrate existing collections, or relax an
already-run cohort.

Each read linearizes at its search-view acquisition. The candidate holds
the store write read-lock only while jointly capturing runtime snapshot,
latest durable LSN and an exact-LSN lease; it queries that LSN with
`available_at <= asOf` after releasing the lock. A concurrent write that
commits after capture may be absent from that read. The control uses the
current `Store.Search` view acquisition. Neither arm has a post-read
current-version retry. Candidate results must be contained in the captured
version and both arms must exclude future and unknown IDs. Freshness is
measured separately: from each batch's durable acknowledgment to the first
completed accepted read **offered after that acknowledgment** that contains
an event from that batch or later. This explicit offer-time requirement was
added after a harness audit found that a read offered before acknowledgment
could otherwise understate post-ack lag; the original timing run is
superseded, with the 25 ms gate and workload unchanged.

For each of three rotated fresh pairs, create the private schema-backed
tenant collection. Seed 200 past-visible vectors orthogonal to query
`[1,0,0,0]` and 16 future-available vectors equal to the query. Offer 256
as-of-visible events in 16 batches of 16 at nominal 16 ms intervals.
Batch `b` uses vector `[1, 0.8-0.04b, 0, 0]`, so later batches should move
closer to the query. Offer 192 k=50 searches at nominal 4 ms intervals to
eight reader workers. Query, capture, lease and close are timed. Each read
must return 50 distinct eligible IDs. A candidate read must not return an
ID inserted after its captured version. At completion, a fresh read in each
arm must contain an event from batch 15 and still exclude future IDs.
Require 256 acknowledged writes, per-version motion and zero active leases.

Report per-pair and pooled read-call/offer p50/p99, actual offer gaps,
writer completion and offer-to-ack p99, batch-to-visible-read lag p99 and
worst lag, correctness counts and retained bytes. Frozen component gate:
no errors or omitted requests, pooled candidate read-call p99 <=1.10x
control and <=100 ms, candidate summed writer completion <=1.25x
control, and candidate batch-to-visible-read lag p99 <=25 ms. A pass is
not Goal 6 completion: migration, sidecar authority, learner publication,
full Recall latency and power-loss recovery remain separate.
