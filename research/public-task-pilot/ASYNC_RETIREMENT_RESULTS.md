# Bounded background retirement: partial rescue, overall FAIL

Protocol: ASYNC_RETIREMENT_PROTOCOL.md. Artifact: `async-retirement-results.json`
with six sidecars and retained isolated stores. Authoritative durability stays
synchronous; disposable derived no-sync remains the previous research setting.

## Lifecycle Changes And Tests

New AsyncPartitionServing is separate from the frozen synchronous control. A
single bounded cleanup worker executes old-graph close/checkpoint. Pending and
closing handles remain charged against the same global two-handle limit; errors
keep the resource charged. Release now means scheduling, not completed cleanup.
Shutdown blocks new admissions, joins all cleanup and reports failures.

Review caught an ordering hazard before the load run: making the serving retirement
slot reusable before the queue slot was released could cause a false queue-full
result. Completion now releases queue capacity before publishing reclamation to
the serving owner. Tests exercise completion-triggered resubmission at capacity1.

Tests also block actual close while requiring lease release to return, check that
capacity remains charged, verify shutdown joins the blocked worker, and inject
cleanup failures for published and cancelled/unpublished graphs. Three full
researchindex race-suite repetitions pass before the performance run.

## Load Results

| Initial N | Repeat | Read errors | Write capacity errors | Late responses | Successful-read misses | Acknowledged/reopened | Final drain ms |
|---|---|---|---|---|---|---|---|
|800|0|0|0|0|0|512/512|82.05|
|3200|0|7|0|1|0|512/512|248.20|
|6400|0|0|13|0|0|499/499|305.17|
|800|1|0|0|0|0|512/512|84.90|
|3200|1|0|0|0|0|512/512|252.80|
|6400|1|0|11|0|0|501/501|309.03|

Three of six arms pass. All seven read errors are lease-admission busy. The one
late response is a3200-record write at106.53ms. No build/cleanup audit errors,
successful-read self misses or failed-write-present records occur. All3048
acknowledged writes survive the existing ID/revision reopen audit.

Capacity errors decline from61 in the previous no-sync/synchronous-retirement
screen to24 here. These are sequential historical runs, not randomized paired
evidence; do not claim a precise causal improvement. Large-corpus build means
remain about124ms, so construction still needs work. Cleanup CPU/allocations are
not removed and can contend with reads/builds even though release no longer does
that work directly.

Final drain includes pending retirement and closing all current graphs. It is
outside per-request deadlines but explicitly measured; shutdown waits for it.
The verifier accounts for9216 operations, hashes, sidecars, positive drain times
and the unchanged combined gates. No concurrent tests during load.

Next investigate the private immutable candidate-only interface identified by
the allocation profile, preserving public search semantics and exact-oracle
controls. All seven whole research goals remain open; no production changes.
