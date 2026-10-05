# Durable steady-state CPU profile

Freeze before dispatch. Use default synchronous local libravdb1.6.13, normal
module file and incremental-overlay-v1. Do not use the rejected zero-window
variant. Fresh retained database per ordinary/experimental packing arm,200
public seed facts, hash32 embedding, one reader and128 unique-session recalls.
Two warmups precede profiling; no concurrent event writes or admission gate.

CPU profiles cover the read loop, journal readback and one initial GC, not seed
insertion. Service durations exclude readback. Every return must retain200
nominations and matching journal explanation. Profiles are not wall-time
decompositions or benchmarks; off-CPU persistence delay will not appear as CPU.
Report flat/cumulative hotspots without adding overlapping percentages.

Use this diagnostic to select further boundary instrumentation or exact-output
optimizations. No expected speedup or latency rescue is claimed. Save raw
profiles and source hashes; no production/dependency modifications.
