# Real build pair reuse: useful lead, no speed claim

Artifacts: `pair-observation-results.json` and `.pairs.json`; generated observer
overlay and runner sources are hashed in the main artifact. The diagnostic uses
the isolated pinned dependency copy. No module-cache or normal dependency change.

Instrumentation is installed only on a fresh unpublished HNSW index, counts the
non-pointer-SIMD selected-pair heuristic calls, and is cleared after build workers
join. Every original distance evaluation remains. Directed numeric node IDs avoid
pointer-address aliasing. A65,536-entry first-seen table bounds each build's state.

| Build group | Builds | Calls | Recognized repeats | Untracked calls after cap | Recognized repeat fraction |
|---|---|---|---|---|---|
|757-857 records (partition builds)|18|39,102,311|15,356,985|22,565,678|39.27%|
|6400-6432 records (whole bases)|4|122,568,974|4,988,833|117,317,997|4.07%|

For each build, Calls = Repeated + Untracked + Entries. Every table reached its
65,536-entry bound. Untracked pairs may themselves repeat, so recognized repeats
are a lower bound on repetition in this trace, NOT a guaranteed cache hit rate.
A real concurrent cache may compute the same miss twice or admit pairs in a
different order; callback locking and table lookup have nonzero cost.

The exact-oracle checker confirms both eight-partition arms retain all self hits
and100% probe recall. One-base controls retain the previous approximate misses.
Observer timing is deliberately excluded from performance claims: locking on every
pair slows construction substantially. Passing its static timing printout does
not validate a cached serving implementation.

Three full researchindex race-suite repetitions pass with the observer overlay.
A separate activated-observer test passes three race repetitions and verifies
callbacks run during construction, finish before publication, and are not invoked
by subsequent search even when its context retains the observer factory.

Next connect the bounded cache to the real selected-pair calls, compute exactly
the same distance on misses, clear it before publication, and compare actual
build time/allocations/recall and unchanged offered load against uncached control.
No cache is enabled in HNSW yet; all seven goals remain open.
