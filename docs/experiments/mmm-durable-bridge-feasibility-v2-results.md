# Guarded durable feedback feasibility v2

**PASS at finite isolated feasibility scope; goal 6 remains open.** The
[frozen v2 contract](mmm-durable-bridge-feasibility-v2-contract.md) corrects
the chronology contradiction retained in the [v1 negative result](mmm-durable-bridge-feasibility-v1-results.md).

An isolated persistent LibraVDB service issued 32 distinct as-of recalls and
explicit labels in chronological order. Every admission and feedback passed a
fresh service publication/as-of guard. A future-only insertion after label 16
left earlier feedback eligible. The independent SQLite ledger audit found 32
bound admissions and 32 feedback terminals, in order, with unique source
journal IDs, the expected event and tenant, and the declared query/label times.
Same-epoch replay recovered 32 completed labels, no failed or pending jobs,
and finite interior scores for all 512 feature codes. A later unresolved
admission survived reopening; visible deletion rejected its stale feedback
before the durable callback ran. Explicit discard restored zero pending, and
opening the old log as epoch 2 failed closed.

The focused test logged isolated replay times of 3.82 ms and 2.53 ms for the
32-label ledger; the separate validation including 512 score checks took
3.87 ms and 2.58 ms. These single-fixture timings are not request latency,
tail estimates, or a concurrency result. The test source is
`internal/service/research_durable_feedback_feasibility_test.go`.

Verification: the feasibility, durable-identity, and persistent-boundary
service tests passed under `-race` for three repetitions; the existing
background durable recovery parity test passed under `-race`; `go vet` on
service and researchmemory passed; `git diff --check` passed. No production
path or served forecast changed.

This proves neither a full service-to-durable feedback bridge nor recovery
across a general visible mutation. The newly bound post-delete bridge still
starts with a new learner; cross-epoch transfer needs an explicit validity
rule. It also does not test crash/power loss, mixed-write load, realistic
agent outcomes, or population p95/p99 latency. Those are the remaining goal-6
gates.
