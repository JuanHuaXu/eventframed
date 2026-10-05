# 768-dimensional bulk-compaction screen

Repeat the eight-arm bulk-base screen with768-dimensional vectors. Generate24
separate SHA256 blocks per ID, not24 copies of one32-coordinate block. This is
still pseudo-random vector data, not actual semantic embeddings or new private
chat data. Keep200/800 corpora, fixed arrival rates,128 writes/256 reads per arm,
delta64, trigger32, leases8, retired2 and100ms response gate unchanged.

Use the same frozen bulk-base overlay. Record dimension explicitly in the new
aggregate generation-wide-results.json and retain per-arm sidecars/stores.
No backend tuning or new success threshold. Reuse the verifier and require all
eight arms to pass before claiming success at this width. A pass still does not
prove long-run, large-corpus, real semantic or full EventFrame service behavior.
Do not run other agent-started tests/builds concurrently.
