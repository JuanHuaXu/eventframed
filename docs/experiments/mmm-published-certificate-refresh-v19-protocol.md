# Epoch-bound certificate refresh v19: frozen engineering screen

Date: 2026-10-02. V18 durably published feedback but scored no
updated posterior: every new event advanced `EvidenceEpoch`, while
both required certificates remained bound to the seed epoch. This
screen asks a narrower counterfactual question: if a valid audit
could supply refreshed certificates at each event-batch epoch,
would the single-owner serving architecture retain freshness and
<100 ms full-Recall offered p99?

Preserve the corrected v18 256D live-as-of fixture, 200 eligible /
17 future seeds, 128 event writes and 128 full Recalls offered at
4 ms cadence, eight Recall workers, cap-16/16-ms event batches,
native cap-four/1-ms journal batching, and 16 distinct-event
full-stream outcomes at 16 ms cadence from a committed journal.
Run two normal matched pairs: unchanged v18 feedback arm versus
the same arm with a test-only certificate publication after each
event batch. Use the same synthetic certificate values and validity
window in both arms; in the refresh arm only the bound policy and
evidence epoch may advance. The writer holds read-to-journal
admission throughout event, certificate and marker publication.
Each refresh must check exact before-LSN/READY marker, publish and
read back both certificates, advance the durable SQLite marker by
CAS, and publish an immutable view before writer acknowledgement.
An interruption or out-of-band write committed before the refresh
must fail closed. Concurrent external writers are outside this
single-owner component, as in v17. Do not reuse a stale certificate
or relax `Service.Recall` checks.

Pass criteria in each refresh trial: 128/128 writes and Recalls,
16/16 durable outcomes, zero stale/future/oracle/journal/marker
violations, all Recalls offered after first outcome certified, at
least one such Recall using a newly updated posterior in the
scored law, event-write p99 <250 ms, Recall call and offer p99
<100 ms, outcome offer-to-publish p99 <100 ms and max <250 ms,
published-view maximum age <250 ms. Record certificate refresh
count, its added latency and the unchanged v18-arm failure.
Run a correctness-only race replay; its timing is not a latency
estimate. Keep v18's failed result unchanged.

The certificates in this experiment are **assumed synthetic audit
outputs**. Reissuing their numerical bounds does not show those
bounds remain true as the corpus changes, and does not authorize
production use. A pass proves only that the runtime could serve
with a timely, independently justified certificate source. A
separate statistical procedure with simultaneous coverage, cost
and external-law validation is still required for Goal 6.
