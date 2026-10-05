# Durable allocation and scheduling diagnostic

Freeze before dispatch. Use normal modules/WAL and incremental-overlay-v1.
Repeat the durable CPU fixture:200 public records,128 sequential recalls per
ordinary/experimental arm, hash32, no concurrent writes or admission. Preserve
nomination and journal checks. These traces do not qualify the failed load case.

After seeding and warmups, force GC and write a baseline heap profile. Collect
an execution trace during the read loop and journal readbacks. Stop tracing,
force GC and write the after heap profile. Subtract the baseline for cumulative
allocation attribution; these are sampled allocation estimates, not exact byte
accounting. Trace/profile machinery contributes overhead. The second arm's
baseline subtracts earlier allocations instead of treating the process-global
heap profile as arm-local by itself.

Use trace scheduler and synchronization profiles plus creation stacks to locate
goroutine sources. Cumulative blocked time across goroutines is not wall latency.
CPU samples from the prior run cannot be uniquely assigned by matching symbols
alone. Preserve raw traces/profiles and source hashes. Pick an optimization only
after identifying its call boundary and an exact behavioral falsifier.
