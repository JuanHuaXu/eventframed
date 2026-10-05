# Audit uncertain writes on database copies

Before execution, freeze this audit of all32 retained durable-openloop stores.
Copy each closed database into a fresh output directory and hash its source
before/after. Original files must remain unchanged. Opening the copy may perform
normal recovery; later service calls and retries mutate only the copy.

Read all200 seed IDs and all16 writer IDs. Classify failed calls as absent,
present with matching payload, or lookup/payload anomaly. Compare the direct ID
set with a300-result ANN query (the corpus has at most216 records). A discrepancy
is a diagnostic lead, not by itself proof of corruption in an approximate index.

Run a fresh service Recall before retries with the original policy and AsOf.
Then retry only failed writes with their original idempotency keys and exact
turn payloads. Record duplicate responses, errors, final presence and snapshots.
Use a60s audit context, not a100ms performance deadline: this is correctness
reconciliation, not a latency rescue. No timing result is compared with serving.

Success of the finite audit requires all acknowledged writes and seeds present,
no lookup/payload anomalies, unchanged sources, successful fresh recall, and all
failed writes recoverable by idempotent retry. Report previously failed-but-present
writes explicitly; an error is not a rollback guarantee. Snapshot consistency,
power-loss safety and arbitrary partial-index recovery are not proven by this
finite copy/reopen test. Preserve results and copies; no production edits.
