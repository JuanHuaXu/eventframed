# Persistent shadow lifecycle/load pilot

Frozen before running. Isolated embedded LibraVDB on temporary test paths,
32-dimensional hash embeddings and50 seeded public-placeholder records. Real
service Recall and Observe execute normal persistence, not a memory-store stub.
No remote backend, external credentials, production OpenClaw, or LLM calls.

Three paired trials, alternating off/on order. Each arm uses a fresh database.
Four readers each perform64 requests (recall50,pack10). One writer performs64
new observations with2ms spacing to sustain overlap. Count writes completed
while readers remain active rather than assuming overlap from simultaneous start.
Collect every read duration/error and write duration/error. Do not discard tails.

Shadow callback is the same bounded numerical diagnostic as v1, not a learned
forecast:2048 passes over packet score squares with cancellation checks. Queue16,
MaxAge250ms. It never mutates served state. Record status before and after Close.
This tests persistent lifecycle/queue behavior only; actual learner and HTTP/
remote ranking are still missing. Do not call successful diagnostics discovery.

Screen: all requests succeed; writes overlap reads; enabled per-trial read p99
<=1.10 times its paired disabled p99; at least one shadow result completes per
enabled trial; shutdown clears published results. Failures are retained, not
retried until favorable. No formal tail guarantee from three small trials.
