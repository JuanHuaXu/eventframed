# Archive Request Identity V32

V30 deadline error preserved. V31 moves the unchanged500ms operation timer after
owner acquisition: interruptions/ownership and vet/core race pass, but concurrent
test stops at `lost/duplicate archive`. That assertion does not record whether
the ID was empty or repeated, nor request timestamp multiplicity. Root cause
needs investigation; do not call it a lost forecast or confirmed clock collision.

New independent technical diagnostic, default native fixtures,16captured calls
per scenario. Keep V31 implementation unchanged. Record ALL request timestamps,
complete capture wires, complete response packets and errors before identity
assertions. Barrier permits two outcomes/one visible insert before archive ack.
Scenarios: actual concurrent `time.Now()` requests; deterministic distinct past
AsOf nanoseconds; intentional identical-request retries. Full original wires and
captured snapshots must persist; every response must bind to its submitted AsOf.
Identical ID is valid ONLY with identical full wire. Distinct scenario requires16
unique; retry scenario requires1. Reopen must validate exact distinct journal count.

Record raw even on post-drain invariant failure, and preserve nonzero runner code.
No relaxed original latency/quality gates, speed claim, or production changes.
Elapsed diagnostic duration includes artificial barrier, scheduling, native append,
and verification; it is NOT the independent offered-load p99 benchmark.
