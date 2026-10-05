# Real build cache: performance rescue REJECTED

Protocol: BUILD_PAIR_CACHE_PROTOCOL.md. Artifacts: `build-pair-cache-off-results.json`
and `build-pair-cache-on-results.json`, each with a `.pairs.json` sidecar. Same
binary, pinned research dependency copy and fixture; modes execute sequentially.

The selected-pair hook now invokes BuildPairCache with the original metric closure
on misses. Keys are directed node IDs; capacity remains65,536 entries per build.
The kernel accepts a per-call computation closure to use already-resolved vectors
without a second ID lookup. Its immutable-build contract still requires identical
metric meaning for each pair. The hook is removed after workers join.

## Correctness

Three full race-suite repetitions pass. An activated integration test separately
passes three race repetitions, recomputes every hooked distance and requires
bit-identical float32 results, exercises real hits and verifies serving never
calls the retired cache. Both eight-partition modes retain every self hit and
100% recall on128 exhaustive-oracle probes. Source hashes/oracles/merges verify.

## Performance

| Mode | Partition replacement ms, repeats0/1 | Total initial8-partition build ms, repeats0/1 |
|---|---|---|
|Off|74.82 / 74.21|649.05 / 616.15|
|On|374.54 / 365.33|2307.11 / 2379.78|

The partition cache records15,365,988 hits and23,736,889 misses across18 builds:
39.30% hits, close to the observed reuse estimate. Yet replacement time worsens
about4.9-5.0x, exceeding the320ms static gate. Whole-base replacements also regress
from about1.16s to4.37-4.47s. This implementation fails the rescue hypothesis.

No offered-load experiment follows this failed construction gate. The cache is
not enabled in daemon or existing serving controls. High hit rate alone did not
justify admission. All correctness results remain distinct from performance.

The experiment has not isolated mutex contention, map lookup and callback/closure
cost separately. These are leads, not proven causes. A possible next candidate
is a bounded allocation-free lookup/store path, but it must first demonstrate
lower overhead and preserve concurrency/identity contracts. Do not enlarge the
cache or weaken graph quality just to make this fixture pass. All seven goals open.
