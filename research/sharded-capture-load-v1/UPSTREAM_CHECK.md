# Read-only related-fix check

2026-10-05 GitHub public API returned LibraVDB master
`18515d4ce0620b24fd2512b0f9023b8b1f4d8355`. The fetched `libravdb/tx.go` byte hash
was `179c2f226585a6f28fcde58489621f77d76d9ca2502a907c34fe55dd0894e6be`, identical
to the pinned v1.6.13 source hash already recorded in the previous diagnostic.
Both index-delta and rebuild paths remain. This is a source comparison, not an
assertion that there is no unpublished or differently scoped upstream fix.

The all-state issue/PR listing includes closed [PR 6](https://github.com/xDarkicex/libravdb/pull/6)
(HNSW/async indexing), closed [PR 9](https://github.com/xDarkicex/libravdb/pull/9)
(unified SQL/graph/vector support), and open [issue 10](https://github.com/xDarkicex/libravdb/issues/10)
(daemon worker-pool startup panic). None establishes that this specific atomic
event/runtime transaction has bounded incremental HNSW construction. The
existing four-shard option is tested without an upgrade or dependency fork.

Cached browser pages were not authoritative: the issue page was months old
and a closed-PR URL could not be fetched. The public API was used for the
current list and commit. No credential discovery or repository write occurred.
The isolated eventframed clone's `origin/main` contains no `WithSharding` use;
the live project branch and working tree were not changed by this check.

Separately, native sampling of the already-confirmed isolated candidate test
process was collected while waiting for its bounded Go test timeout. It shows
mostly parked native threads and cannot resolve Go goroutine ownership or
prove a particular deadlock. The eventual Go test transcript is authoritative.
