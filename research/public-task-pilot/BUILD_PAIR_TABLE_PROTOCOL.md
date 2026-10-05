# Atomic table construction screen

Preserve the rejected map-cache outputs. Use a separate build-only overlay with
Lookup/Store instead of a heap-escaping distance callback. Directed key and exact
float bits occupy a single atomic word. Capacity remains 65,536 slots; collisions
evict and unsupported IDs bypass. Misses compute the original distance directly.
Each fresh unpublished build gets its own table, cleared after workers join.

First run lifecycle tests plus an instrumented overlay that recomputes original
distances on cache hits and rejects unequal float bits. Do not time this overlay.
Then run ordinary off/on static screens (two repeats each), exact-oracle verifier,
and report all hit/miss/build results. Same configuration and data as the map
screen. Sequential off/on runs are discovery, not randomized speedup evidence.
No load experiment unless correctness holds and construction cost improves.
No production or normal module changes; all seven whole goals remain open.
