# Derived no-sync screen: partial improvement, overall FAIL

Protocol: DERIVED_NOSYNC_PROTOCOL.md. Artifact: `derived-nosync-results.json`, six
sidecars and retained isolated stores. The overlay changes only the new derived
index database's durability setting; authoritative writes stay synchronous.
Three full researchindex race-suite repetitions passed before the load run.

| Initial N | Repeat | Read errors | Write errors | Late responses | Successful-read misses | Acknowledged/reopened |
|---|---|---|---|---|---|---|
|800|0|0|0|0|0|512/512|
|3200|0|0|0|0|0|512/512|
|6400|0|0|21|0|0|491/491|
|800|1|0|0|0|0|512/512|
|3200|1|11|0|9|0|512/512|
|6400|1|0|40|0|0|472/472|

Three of six arms pass, unchanged from the previous partitioned screen. All3011
acknowledged writes survive the authoritative ID/revision reopen audit; no build
error or failed-write-present record is reported. Successful reads all self-hit.
All61 write errors are capacity rejections. The3200 repeat1 arm has10 admission
errors, one shard-search deadline error and9 late responses; max read128.06ms,
max write114.87ms. Error and late
counts are separate dimensions, not necessarily disjoint operations.

Mean builds at6400 take130.46/134.94ms and clear12 entries each on average,
still below100 entries/s offered input even before polling overhead. Previous
synchronous-derived means were142.17/139.77ms. Fewer rejections (61 versus98)
suggest some avoidable cost, but this historical comparison is not a randomized
paired estimate. It does not establish a robust benefit or explain tail clusters.

No sync relaxation is adopted in production. If this disposable-index strategy
is pursued, startup must rebuild from authoritative data rather than trusting
possibly partial derived files. It must never alter acknowledgement durability.
This study does not test physical power failure.

The generalized `check-partition-load.mjs` verifies this artifact's source hashes,
six sidecars,9216 operation samples, revisions, compaction accounting and gates.
No tests or other benchmark workloads ran during measurement.

Next profile construction/cleanup and admission blocking with separate diagnostic
runs before changing graph quality or budgets. The large-corpus bottleneck remains
and all seven whole research goals stay open.
