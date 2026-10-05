# Dense 256D published-LSN top-k v6: finite transfer passes

Date: 2026-10-02. The [frozen protocol](mmm-published-dense-v6-protocol.md)
passed in a private 256D cosine collection with declared EventFrame payload
columns and journal-certified publication. The vectors were dense, with
deterministic row-specific components orthogonal to the dense query. The
corpus held 200 as-of-eligible events and 17 future events. No future row
was returned at the as-of boundary.

Exact candidate membership, decoded body/availability, score parity within
`1e-5`, and nonincreasing published SQL score passed at k=10, 50, and 200.
The ordinary comparator used `Store.Search` on the same committed corpus.
All 32 measured calls per arm and cap succeeded. Tied aligned genesis
rows were compared as a set.

Quiet call times after four warm calls per arm, 32 alternating measured
calls per arm and cap (nearest-rank p99 is the sample maximum):

| k | Published SQL p50 | Published SQL p99 | Ordinary Search p50 | Ordinary Search p99 |
| ---: | ---: | ---: | ---: | ---: |
| 10 | 164.583 us | 415.5 us | 355.208 us | 726.125 us |
| 50 | 293.042 us | 634.125 us | 539.333 us | 883.958 us |
| 200 | 820.334 us | 1.143708 ms | 1.792375 ms | 2.060083 ms |

Focused normal/race tests, ordinary package tests, and vet passed. The
256D fixture reduces one representation-transfer uncertainty from the 4D
v5 result. It does not establish performance or parity at a large corpus,
under concurrent publication, in the full Recall+journal pipeline, or on
agent outcome tasks. The timing excludes ingestion, publication, journal,
network, and offered-load queueing. Existing nonunit records still lack
a migration path. Goal 6 and all seven whole goals remain open;
production was untouched.

Reproduce:

```sh
EVENTFRAME_RUN_PUBLISHED_DENSE_V6=1 go test ./internal/store/libravdbstore -run '^TestResearchPublishedDenseV6$' -count=1 -v -timeout 5m
EVENTFRAME_RUN_PUBLISHED_DENSE_V6=1 go test -race ./internal/store/libravdbstore -run '^TestResearchPublishedDenseV6$' -count=1 -timeout 5m
go test ./internal/store/libravdbstore -count=1 -timeout 5m
go vet ./internal/store/libravdbstore
```

SHA-256: test `58c0765dba602b72c08e1772cc37c62da92ff1e5724a3980475e3706ae06a02c`;
protocol `4392dba31e4f6b4034d218d656e7a71a3be46d03f37c0599d7c9126bfb4d3b6d`.
