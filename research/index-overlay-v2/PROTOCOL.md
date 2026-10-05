# Aligned-Key Rescue Of Prepared Overlay V1

Freeze before V2 timing. V1's checkpoint-lock and key-arena failures remain
preserved. Clone V1 control/candidate into fresh isolated copies; do not mutate
frozen V1 sources. Both V2 arms receive the same eight-byte key-capacity repair
in the common transaction touched-state builder. Candidate otherwise retains
the repaired64-slot immutable-base overlay. Control remains full HNSW rebuild.

Run the12case key-capacity regression before/after patch, ordinary and race,
and the same5standalone +5database selected preflight cases, before timing.
Keep the V1 timing design unchanged:3repetitions,32/128D,256/1024 initial
records,128individual durable inserts and top50 queries percell, SQ8,
GOMAXPROCS10, alternating arm order. Include initialization and compactions.
Quality frozen mean tie recall>=.95 and mean exact-cosine regret<=.005.
No removed records, bulk-load shortcut, protocol threshold or family change.

Full serving/freshness, asynchronous compaction, retained RSS, kill/recovery,
future-counterfactual rank equality, learned law and agent utility remain
UNPROVED. Canonical vector generation costs O(ND) payload; overlay preparation
cost is bounded in changed IDs, overflow is synchronous O(N graph build).
Peak retained payload accounting was not instrumented by V1's frozen harness;
report it NOT CAPTURED, not a memory pass. Whole six science goals and Goal6
remain OPEN. Reuse V1 primary-source references and source-derived root cause.
