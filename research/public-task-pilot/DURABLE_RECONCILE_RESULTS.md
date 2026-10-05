# Retained-store cancellation reconciliation

The audit inspected copies of all32 closed databases from the durable load run.
Source hashes before and after match, and the verifier rechecked them. No
original experiment database or production state was modified.

## Results

- All401 acknowledged writer events are present with matching payloads.
- All111 writes that returned errors are absent before retry; none was observed
  committed despite the error in this retained sample.
- All111 retries using the original ID and exact payload succeed. They are new
  insertions, not duplicate responses, consistent with prior absence.
- All6400 seed records (200 per store) are present. Direct ID sets and a
 300-result search agree in every store before retries.
- All32 fresh service recalls succeed before retries with the expected
 availability-aware nomination counts.
- After retries, every copy contains all216 expected event IDs. The fixture's
 evidence epoch is216 and runtime version217; before retries those counters
 match the directly observed record count plus the initial policy-binding step.

The finite reconciliation criterion passes. These checks resolve the uncertain
writer outcomes from this specific retained run without finding a remaining
lookup/index or version-count discrepancy. They do not turn a generic write
error into a rollback guarantee. The source workload did not inject power loss,
arbitrary crash timing, disk-full errors or torn writes, and version counts alone
are not a general semantic consistency proof.

## Scope and next step

Retries used a60s audit context and ran serially on copies, not under the failed
100ms load deadline. Their success does not rescue the performance result.
Fresh recall writes an audit journal to the copy, so this is deliberately not
a read-only audit of the copies themselves. Original hashes remain unchanged.
The final state was checked before closing the copies; the retry results were
not subjected to another power-loss/recovery experiment.

The next useful lead is measuring durable request phases: search/decoding,
journal serialization/commit and event insertion/index work, including waiting.
The high-rate throughput failure remains open. No production patch follows from
this audit, and all seven full research directions remain open.

## Artifacts

DURABLE_RECONCILE_PROTOCOL.md preceded execution. The runner and input artifact
are hashed in durable-reconcile-results.json; original database hashes are stored
per row. Copies are in durable-reconcile-results.json.stores and are not intended
for source distribution.

```sh
node research/public-task-pilot/check-durable-reconcile.mjs
```
