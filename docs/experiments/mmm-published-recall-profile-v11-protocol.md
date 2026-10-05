# Full Recall admission profile v11: frozen diagnostic

Date: 2026-10-02. No behavior change or promotion gate. The v9
unadmitted and v10 admitted full-Recall screens use the same private
256D corpus, 128 write/Recall offers at 4 ms, four Recall workers,
durable journal and top-150 oracle. V10 completes all calls but its
offer-to-done p99 remains over 500 ms. This profile asks where the
roughly 26-28 ms median occupied call time and serial queueing occur.

Run the unchanged opt-in v9 and v10 tests once each with Go CPU,
block, mutex, and execution-trace profiling enabled. Keep all profiler
files in a temporary directory outside the repository. Use the same
normal (non-race) execution and preserve both tests' expected failing
status; do not lower gates or alter workload. Analyze CPU and blocking
profiles by call stack, not just function name. Compare both arms and
the previously measured quiet 256D published-search and journal
isolated costs. Record git HEAD, Go/OS architecture, sample sizes,
and which profiles are aggregate goroutine time rather than per-Recall
latency. Do not infer a causal bottleneck from a single CPU hotspot.

The diagnostic outcome is a ranked list of likely capacity owners
with explicit missing spans and a falsifying next experiment. A
profiling run is not a new Goal 6 performance result, and the v9/v10
negative outcomes remain unchanged.
