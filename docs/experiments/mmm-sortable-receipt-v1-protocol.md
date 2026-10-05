# Sortable event receipt v1: frozen private writer contract

Use only the research-only sortable EventFrame batch writer and a private
declared-key LibraVDB collection. Add an opt-in API returning the exact
LibraVDB transaction commit LSN while keeping the existing batch APIs and
production writer behavior unchanged. The receipt is zero when a batch is
entirely exact duplicates and no transaction commits.

The finite test inserts two subsecond events in one batch and requires a
positive receipt equal to `LatestCommitLSN`, two runtime-version/evidence-epoch
increments, both per-event motion entries, durable correct sort keys and an
exact-LSN as-of result that excludes the future event. An exact retry must
return duplicate results, a zero receipt, no new LSN and no snapshot change.
A mixed duplicate/new batch must commit only the new row and return its exact
receipt. A conflicting duplicate or a legacy row missing the sort key must
fail closed without changing LSN or snapshot. Close/reopen must preserve the
last receipt boundary and exact-LSN result. Repeat the finite test under
`-race` and run the package's ordinary tests and vet.

Passing proves only receipt correctness for this private single-process
batch writer. It does not implement marker renewal, cross-process writer
fencing, unknown-outcome reconciliation, loaded service latency or Goal 6.
