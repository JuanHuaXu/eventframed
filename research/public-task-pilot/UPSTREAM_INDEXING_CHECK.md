# Upstream check before backend research

Checked public HEAD18515d4ce0620b24fd2512b0f9023b8b1f4d8355 and the complete
first API pages (seven PRs, ten issues including PRs; no next-page link).
Raw replies and SHA256 are retained in upstream-indexing-check/manifest.json.

[PR6](https://github.com/xDarkicex/libravdb/pull/6) merged high-performance HNSW
and durable async indexing. However the checked [transaction source](https://github.com/xDarkicex/libravdb/blob/18515d4ce0620b24fd2512b0f9023b8b1f4d8355/libravdb/tx.go)
still loads all vectors for non-delta indexes and builds replacement indexes.
The checked interfaces file implements PrepareMutations only for flatWrapper.
No separately titled transactional incremental-HNSW fix was found in these PRs;
that is not an exhaustive review of every historical commit or external fork.

The [async indexing plan](https://github.com/xDarkicex/libravdb/blob/master/docs/research/async-wal-indexing-plan.md)
distinguishes durability acknowledgement from graph readiness and provides a
FlushIndex barrier. Its throughput figures cannot establish EventFrame's
transactional read-after-write behavior. Do not enable async indexing solely
to improve acknowledgement time while leaving search behind.

Decision: no dependency upgrade or upstream patch at this stage. First isolate
known-non-entry admission failures, then investigate scalable transaction/index
publication while preserving snapshot and search-visibility contracts.
