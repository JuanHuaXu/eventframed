# Tau Agency V1 Results

The finite timing policy, reference histogram fitter, durable agency binding,
and separately authorized OpenClaw scheduler worker are implemented. The
reviewed measurements are in [tau-agency-v1.json](tau-agency-v1.json); usage and
scope are described in [the policy contract](../tau-agency.md).

Two public-synthetic smoke runs triggered ordinary GPT-6 Luna turns through
the existing isolated OpenClaw 2026.9.8 gateway after daemon restart. The final
build passed eight control assertions and native job start was 1.683 seconds
after the selected time, inside its ten-second allowance. Completion was
6.460 seconds after nominal. The fresh session's assistant response contained
the expected marker even though the cron summary omitted it. The lab
configuration hash remained unchanged. One intervening test-harness assertion
failed before dispatch because control requests shared a mutable policy;
each request now gets an independent fixture.

Three review passes by a separate auditor identified and prompted fixes for
missing numeric defaults, stale rebasing, late idempotent retries, empty
candidates, stale dispatch clocks, cancellation races, multiple uncertain
handles, and aliasing of returned records. Uncertainty survives expiry, fences
same-target admission/claim/approval, and requires a current complete handle-set
acknowledgment to clear. The external scheduler's lack of a model-start expiry
is an explicit remaining integration boundary.

The isolated tracked source passed the full Go race suite and 31 adapter tests,
plus TypeScript checking. The integrated working source passed targeted tau
tests. Finite enumeration checked 3,369 perturbed-law/delivery utility bounds.
The final 32-atom arithmetic benchmark measured 4.103-4.364 microseconds/op,
4,584 bytes/op, and ten allocations. A single eligible HTTP plan took 7.480 ms;
100 no-turn HTTP calls had p99 0.172 ms in the final smoke. These measurements
do not establish loaded latency or calibration.

Decision: retain the mechanism as opt-in feedstock. No production deployment or
whole-research baseline promotion. Next experiments should independently score
future event times and no-event outcomes, measure interruption/task utility,
test epoch-scoped replanning under ordinary activity, and validate execution
expiry or quantify model-start lateness under load. Full-stream studies also
need a complete pre-action journal, including abstentions.
