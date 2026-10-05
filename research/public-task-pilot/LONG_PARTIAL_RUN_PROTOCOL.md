# Long partial-tail growth falsification

Freeze before run. Extend PARTIAL_RUN_LOAD_PROTOCOL to4096 append writes and8192
reads per arm at the same100/200 per second, initial6400 records, two repetitions.
All other policies stay fixed: CPU4,768d,64 pending,32 flush trigger,10 current
graphs,2 retired,8 leases,one builder,100ms scheduled deadlines. No retries.

Each measured arrival interval is about40.96s. Complete/drain all scheduled work
and reopen authority. Pass gates unchanged: no request/build/audit errors, no
late responses or successful self-query misses, acknowledged records and revision
recover, at least two successful partial merges. Report build-time growth and
request failure position, not merely aggregate medians.

This is still append-only. It specifically tests the previously unmeasured growth
of the single merged tail; it does not substitute for updates/deletes or complete
EventFrame/agent evaluation. Retain negative outputs and do not relax the budget.
