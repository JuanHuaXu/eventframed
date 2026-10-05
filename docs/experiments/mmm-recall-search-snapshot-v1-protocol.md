# Recall search/snapshot interleaving v1: frozen diagnostic

Run the real `Service.Recall` against both memory and LibraVDB stores. Seed
one as-of-visible EventFrame. Wrap `Store.Search` to complete its search,
then synchronously ingest one more as-of-visible EventFrame before returning
the original search result. Use `recall_k=pack_k=50`, so the new event cannot
be omitted by a small retrieval budget. The wrapper injects once only, so
the service can retry if its snapshot discipline detects the interleaving.

The consistency condition is: if the returned packet's runtime snapshot
includes the new write, its candidate set must include the newly visible
event (or Recall must fail closed). The durable frontier journal must not
claim a newer snapshot for an old-only candidate set. This is about the
Search-to-Snapshot interval, not the previously tested
Snapshot-to-Journal interval. Report search attempts, event IDs, packet and
durable journal snapshots. Preserve a failing result. No production change
or performance claim follows from this deterministic two-record test.
