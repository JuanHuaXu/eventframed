# Real deletion transaction and reopen checks

`TestPrivateDeleteRealRecovery` uses temporary on-disk libravdb collections, normal
storage options, a flat authoritative record collection and metadata-only revision
collection. No production paths, private data, or dependency changes are involved.
Two public synthetic IDs/vectors are seeded atomically at revision1.

The private writer's callback executes one real WithTx containing the target
deletion and revision update. Four cases each passed three race runs (1.819s total):

| Case | Writer outcome | Reopened target | Reopened revision |
| --- | --- | --- | ---: |
| Success | new bundle visible | absent | 2 |
| Transaction abort after staged delete | quarantined | present | 1 |
| Commit, then lost acknowledgement | quarantined | absent | 2 |
| Commit, then callback panic | quarantined | absent | 2 |

All cases verify reopened collection count and survivor vector contents, no retry
of uncertain persistence, and continued validity of the old historical view. The
test rebuilds a tiny coherent graph/summary from the reopened known IDs and creates
a fresh writer at the durable revision. It does not simply reuse the failed writer's
in-memory candidate as recovery authority.

## Scope

This establishes real transaction/reopen behavior for the fixture, beyond mock
callbacks. It does not test process kill, power loss, partial file writes, filesystem
faults, or durability between commit and clean Close. General ID enumeration and
full ANN reconstruction are not implemented: the fixture knows its two IDs and
uses empty adjacency for its tiny rebuilt graph. Exact original topology is not
claimed recoverable from only authoritative vectors and revision metadata.

The ordinary graph-deletion capture tests separately cover nontrivial repair;
this database fixture targets the persistence/publication boundary. The two types
of evidence cannot yet be combined into a claim of complete durable ANN serving.
Next close that integration and reader-lifetime gap, and implement insertion.
All seven whole research goals remain open; no production adoption is made.
