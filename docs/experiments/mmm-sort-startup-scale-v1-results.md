# Incremental publication startup scale v1: private diagnostic

Date: 2026-10-02. Protocol: [v1](mmm-sort-startup-scale-v1-protocol.md).
Command: `EVENTFRAME_RUN_SORT_STARTUP_SCALE_V1=1 go test ./internal/store/libravdbstore -run '^TestResearchIncrementalSortStartupScaleV1$' -count=1 -v -timeout 5m`.
The opt-in test passed in 19.65 s. One private Store owner was fully closed
before each measured reopen. Both arms opened the same Store and SQLite
sidecar; only the candidate verified the row journal. Close and post-open
READY checks were excluded from timed Open.

| Rows | Pair | First | Control Open (ms) | Verified Open (ms) | Ratio |
| ---: | ---: | :--- | ---: | ---: | ---: |
| 35 | 0 | control | 2.522 | 3.035 | 1.204 |
| 35 | 1 | candidate | 2.287 | 2.541 | 1.111 |
| 35 | 2 | control | 2.311 | 2.471 | 1.069 |
| 35 | 3 | candidate | 1.934 | 2.648 | 1.369 |
| 35 | 4 | control | 1.738 | 2.304 | 1.326 |
| 259 | 0 | control | 7.465 | 9.602 | 1.286 |
| 259 | 1 | candidate | 8.126 | 9.862 | 1.214 |
| 259 | 2 | control | 8.279 | 9.339 | 1.128 |
| 259 | 3 | candidate | 7.597 | 9.299 | 1.224 |
| 259 | 4 | control | 8.150 | 10.207 | 1.252 |
| 1027 | 0 | control | 31.164 | 35.968 | 1.154 |
| 1027 | 1 | candidate | 29.143 | 34.936 | 1.199 |
| 1027 | 2 | control | 29.759 | 34.259 | 1.151 |
| 1027 | 3 | candidate | 29.613 | 36.858 | 1.245 |
| 1027 | 4 | control | 32.867 | 38.112 | 1.160 |

| Rows | Control p50 / max | Verified p50 / max | p50 ratio |
| ---: | :--- | :--- | ---: |
| 35 | 2.287 / 2.522 ms | 2.541 / 3.035 ms | 1.111 |
| 259 | 8.126 / 8.279 ms | 9.602 / 10.207 ms | 1.182 |
| 1027 | 29.759 / 32.867 ms | 35.968 / 38.112 ms | 1.209 |

The extra journal verification is measurable, but ordinary Store Open itself
also grows with this fixture. The candidate still scans every journal row
and counts the collection at startup, so this remains an O(N) verification
design. Five pairs per size and only 1,027 rows cannot establish production
restart behavior or an acceptable bound for a large corpus. No loaded Recall,
multi-owner safety, automatic DB-only recovery, or power-loss guarantee follows.
Production was untouched; Goal 6 and all seven whole research goals remain open.

Source SHA-256: test `2686c200dd1702b9731d45e70af66404a48c1e4de3085ee4303a1804047f8709`;
protocol `b953c3f70cdfedd774c2c6f19503514303efba0dce6106b9c3e5d102d2753ba0`.
