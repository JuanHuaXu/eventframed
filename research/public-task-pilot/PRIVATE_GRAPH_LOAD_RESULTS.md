# Private graph sustained load: FAILED

Both frozen repetitions completed, including reopen checks. Raw:
`private-graph-load-results.json` and per-arm sidecars. The test intentionally
returns failure after writing evidence. Runtime 87.933s including setup/audit.

| Result | Repeat 0 | Repeat 1 |
|---|---:|---:|
| Acknowledged/reopened writes | 37/37 | 34/34 |
| Durable revision | 38 | 35 |
| Failed writes found durable | 0 | 0 |
| Reads rejected after quarantine | 8082 | 8100 |
| Writes rejected after quarantine | 4051 | 4060 |
| Other write deadline errors | 7 | 1 |
| Commit deadline error | 1 | 1 |
| Late write responses | 16 | 7 |
| Successful self-query misses | 0 | 0 |
| Reopen audit errors | 0 | 0 |

## Boundary diagnosis

Successful writes BEFORE quarantine averaged 8.729/8.904ms preparation plus
3.994/3.856ms persistence, about 12.7ms of serialized service per write. This
exceeds the 10ms arrival interval. Mean queue wait grew to 54.676/48.202ms;
some successful waits exceeded 91ms. These are survivor-conditioned early-run
diagnostics, not an unbiased sustained throughput estimate.

At write indices 44/35 (zero-based), persistence returned
`transaction commit failed: context deadline exceeded`. Their queue waits were
89.998/91.080ms, preparation 9.833/8.867ms, persistence 0.272/0.241ms. The writer
then quarantined according to its uncertainty policy, so subsequent requests
failed promptly. Aggregate p99 is therefore misleading: cheap rejections
dominate after shutdown of useful service.

All 71 acknowledgements survived, and no failed writes appeared after reopen.
This supports the finite durability check; it does NOT prove that every such
commit error can safely be treated as a known rollback. Removing quarantine
without an authoritative outcome-resolution contract would weaken correctness.

## Conclusion and rescue leads

The private graph eliminates tail rebuilds but does not meet the unchanged
offered-load serving gates. The earlier component benchmark was insufficient
because it excluded the approximately 4ms transaction cost. No whole research
direction is closed by this test.

Next prioritize measured preparation reduction (bounded immutable-vector norm
reuse while preserving exact scalar cosine arithmetic), then amortizing durable
commit cost with a bounded transaction batch if needed. Group commit must retain
per-request deadlines, ordered revisions and atomic publication; it cannot merely
acknowledge queued data. Admission-time commit headroom or resolved transaction
outcomes may contain quarantine cascades, but cannot by themselves rescue a
serialized service rate slower than arrivals. Preserve this failed baseline and
rerun the same workload after independently verified changes.

The corpus/layout differences from the old tail-merge architecture are stated
in PRIVATE_GRAPH_LOAD_PROTOCOL.md. This remains a local append-only serving
experiment, not an OpenClaw/agent accuracy test or an arbitrary-scale result.
