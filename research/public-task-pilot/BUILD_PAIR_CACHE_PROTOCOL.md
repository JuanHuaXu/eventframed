# Real build-cache screen

Use the same selected-pair hook as the observer, but return a saved float32 on
hits and execute the original distance closure on misses. Each fresh build owns
one65536-entry directed-pair cache, cleared after workers join and before serving.
No cache or metric crosses builds; query distances are not cached here.

Run the exact-oracle static screen in off/on modes, same binary/config/dataset,
two repeats per mode. Off bypasses the hook entirely. Report hit/miss/entry counts
and build cost, preserving negative results. Sequential mode runs are discovery,
not a randomized causal speedup estimate. Do not run load unless recall gates
hold and build times show a promising improvement; keep any regression visible.
