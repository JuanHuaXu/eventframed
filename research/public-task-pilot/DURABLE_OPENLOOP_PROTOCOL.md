# Local libravdb fixed-arrival screen

Freeze before dispatch. Use the incremental-overlay-v1 and unchanged service
semantics with the actual local libravdbstore adapter instead of memorystore.
No network database, production configuration, private data or LLM generation.

Repeat the previous32-arm grid:200 public-fact events, hash32, no quantization,
memory mapping enabled; default adapter write concurrency2/queue64. Four reader
permits or no admission; experimental or ordinary packing; visible/future writes;
20ms/5ms read intervals with half-rate writes; two repetitions with arm order
reversed.32 reads+16writes per arm. Each arrival independently scheduled with
100ms deadline from due time; callbacks, errors and arrival delays all recorded.

Use a fresh, retained database path for every arm under NEW_OUTPUT.stores.
Never reuse or remove an existing directory. Preserve both warmup records even
if they fail. Warmup failures count against qualification, not as grounds to
silently change nomination expectations. Successful nomination must still be200
in future mode or200..216 in visible mode; journal and availability checks remain.

After timed work, close and reopen each database. Compare complete JSON digests
of every successfully returned journal and check every acknowledged new event
by ID. Report close/reopen errors separately. Do not count failed or uncertain
writes as acknowledged success. This is orderly reopen evidence, not crash,
power-loss, torn-write or disk-full testing. Retained files contain only public
fact fixtures and research metadata; they are not intended for source publication.

Finite success requires all admitted arms to meet the prior no-error/no-stale,
100ms checks, both warmups succeed, and all acknowledged records survive reopen.
Do not relax deadlines if persistence is slower. Preserve failed artifacts and
the exact local adapter source hash, including existing worktree changes.
Close/reopen verification is outside service latency and measured separately.
The short burst is not long-run capacity evidence; no whole goal closes here.
