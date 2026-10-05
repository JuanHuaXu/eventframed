# Bounded event transaction screen

Research-only, normal libravdb v1.6.13 and synchronous durability. No service
integration, production deployment, WAL change, private data, or queue yet.

Hypothesis: sharing event inserts and the final runtime-state upsert in one
transaction reduces per-event persistence cost. Alternatives include index work,
encoding and fixed journal costs dominating enough to make batching insufficient.

Before timing, require race-enabled tests (three repetitions): payload/vector
parity with ordinary Put, final snapshot parity, duplicate retries, in-batch and
existing digest conflicts, malformed/cross-tenant/over-cap rejection, canceled
context rejection, mixed concurrent ordinary Put, and orderly reopen durability.
These are not crash or injected commit-failure tests. A commit error remains an
unknown outcome requiring reopen/reconciliation before reuse. Collection setup
may create an empty collection even if later batch preflight fails.

Measure 128 already-available distinct events per operation, fresh embedded
database, same four-dimensional public-fact vectors and content for every arm.
Compare original Put (size1) against batch sizes2,4,8,16. Three fixed one-operation
repetitions, CPU4. Setup and close excluded; first collection creation included.
Record full Go benchmark output and SHA256 of implementation, tests, original
store, go.mod, and this protocol. Do not rerun over the same output artifact.

Continue this lead if any batch size reduces median time per128 events by at
least20%, with all correctness tests passing. This is an engineering screen, not
a statistical confirmation or end-to-end claim. All batch members already exist
at the start: it excludes accumulation latency, deadlines of separate requests,
read contention and synchronous journal commits. Passing cannot close goal6.

If promising, implement a bounded queue that batches only already-pending writes,
then replay the same fixed-arrival durable workload including arrival-to-response
latency. Preserve synchronous acknowledgement, final snapshot semantics and
individual cancellation/unknown-outcome accounting. Do not batch across an
intervening read's linearization requirement merely to obtain larger batches.
