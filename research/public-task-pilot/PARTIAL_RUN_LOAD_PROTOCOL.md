# Partition-base / partial-tail run load

Retain the RUN_LOAD_PROTOCOL arrival rates, CPU4,768 dimensions,64 pending
records,100ms scheduled deadlines,8 leases,2 retired graph slots,one builder,
six arms and authoritative durability. Initial state is eight SHA256 ID
partitions, queried together. Never merge those base graphs in this finite study.

Flush at32 pending entries. When two tail runs exist (10 total graphs), merge
those two newest runs before another flush, even if pending delta is below32.
That merge preserves tombstones and does not drain pending writes. Up to10 current
graphs,2 retired graphs and1 unpublished graph are allowed, versus up to2 current
graphs in run-load: this resource change is explicit, not a like-for-like memory
comparison. Initial partitioning matches the earlier partitioned retrieval study.

Same pass gates; require at least two successful partial merges per arm. Log all
builds and final cleanup. This finite tail can grow without bound in record count
despite bounded graph count, so passing requires follow-up beyond512 writes before
claiming stable throughput or a general scalable consolidation policy.
