# Partitioned growth load: partial rescue, overall FAIL

Protocol: PARTITION_LOAD_PROTOCOL.md. Artifact: `partition-load-results.json`,
six per-arm sidecars and retained authoritative/derived stores. The original
GROWTH protocol's arrival rates, deadlines and total64-record delta cap remain.

| Initial N | Repeat | Read errors | Write capacity errors | Late responses | Self misses among successful reads | Acknowledged/reopened |
|---|---|---|---|---|---|---|
|800|0|0|0|0|0|512/512|
|3200|0|0|0|0|0|512/512|
|6400|0|0|47|0|0|465/465|
|800|1|0|0|0|0|512/512|
|3200|1|4|0|3|0|512/512|
|6400|1|1|51|0|0|461/461|

Only three of six arms pass the unchanged combined gate. All98 write failures
are `research index delta full`; all5 read failures are lease-admission busy.
Late counts include all response outcomes, not only successful requests. At3200
repeat1, maximum read response152.13ms and write response113.13ms. No build error,
failed-write-present record or reported reopen audit error occurs. All2974
acknowledged writes reopen. This audit checks ID presence and global revision,
not a physical crash or exhaustive reopened vector equality.

At6400 the earlier one-base screen rejected339+352=691 writes; the partitioned
screen rejects47+51=98 under the same workload. That is a historical comparison,
not a randomized concurrent A/B estimate. It establishes substantial observed
improvement, not robust validation or a precise causal speedup.

## Why Static Timing Was Insufficient

At6400,34/36 partition builds average142.17/139.77ms and remove an average
12.15/11.92 delta entries. Incoming writes arrive at100/s. The implied build-only
clearance rates are about85/s, below offered writes; polling adds further cost.
The bounded queue cannot absorb that mismatch indefinitely. A build below320ms
does not suffice when it clears fewer than32 entries.

Removed-entry counts are derived as before-total + global-revision-advance -
after-total. This is valid only for this single-compactor, unique-single-append
workload, where each successful revision adds exactly one delta entry. General
updates, deletes or batching require explicit counts instead.

The checker independently validates all9216 operation records, source hashes,
sidecars, revisions, accounting bounds and gate outcomes. Tests/builds did not
run concurrently with measured load. No serving defaults or production changed.

## Next Leads

Reduce candidate construction cost without weakening authoritative durability;
derived graphs need not themselves be durable if rebuilt from the authoritative
snapshot. Also attribute lease/deadline outliers before changing admission limits.
Do not enlarge delta/lease caps or lower arrival rates merely to clear this screen.
All seven whole research goals remain open.
