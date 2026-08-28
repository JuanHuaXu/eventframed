# eventframed

`eventframed` is the first executable slice of the EventFrame prediction runtime:
a local Go daemon backed directly by LibraVDB, plus a thin TypeScript memory and
context adapter for OpenClaw.

This repository is an implementation companion to the EventFrame whitepaper. It
is an alpha runtime, not an empirical validation of the complete mathematical
specification. The current runtime implements durable event ingestion, 5W1H field
provenance, availability-time-safe retrieval, deterministic idempotency, distinct
recall and packing budgets, quantized LibraVDB traversal, and an untrusted-context
boundary for OpenClaw, model-keyed embeddings, atomic durable version state,
retention/deletion propagation, backup, compaction, guarded migration, and a
certificate-gated selective Bayesian usefulness layer.

## Architecture

```text
OpenClaw
  |  before_prompt_build / agent_end
  v
TypeScript contract adapter
  |  eventframe.v1alpha1 over a mode-0600 Unix socket
  v
eventframed (Go)
  |  ingest -> availability gate -> frontier(k=50) -> certified belief rerank -> pack(k=10)
  v
LibraVDB (embedded, mmap, optional SQ8/FSQ6/PQ traversal)
```

There is no separate LibraVDB daemon hop. The Go process owns the database and
all policy-bearing calculations. The TypeScript adapter only translates OpenClaw
hooks into the versioned protocol and injects escaped, explicitly untrusted
historical context.

## Build and test

Requirements: Go 1.25 or newer and Node.js 22 or newer.

```sh
cd plugin
npm install
cd ..
make check
make build
```

Run the daemon with conservative local defaults:

```sh
./bin/eventframed
```

Defaults:

- socket: `~/.eventframed/run/eventframed.sock`
- database: `~/.eventframed/data/eventframe.libravdb`
- vector dimension: 768
- traversal quantization: SQ8
- embedder: deterministic development hash (`-embedder hash`)
- recall budget: 50 candidates
- packing budget: 10 records / 2,000 estimated tokens

Use `-quantization none` while validating small test collections. `fsq6` and
`pq8` are available as explicit experimental choices.

For a production embedding service, use an OpenAI-compatible endpoint:

```sh
EVENTFRAMED_EMBEDDING_API_KEY=... ./bin/eventframed \
  -embedder openai-compatible \
  -embedding-url http://127.0.0.1:8080/v1/embeddings \
  -embedding-model declared-model-id \
  -dimension 768
```

The model key, dimension, and quantization contract are persisted. A mismatched
restart is rejected instead of silently mixing vector spaces.

## OpenClaw adapter

The built adapter is under `plugin/dist`. Its manifest is
`plugin/openclaw.plugin.json`. Configure the plugin with the same socket path and
an explicit tenant ID. Recall fails open if the daemon is unavailable; capture is
also skipped rather than blocking an agent run.

The adapter records retrieved event IDs on each generated turn. This lineage is
required to detect self-reinforcement and must not be removed by downstream
importers.

## Important limits

- The development hash embedder is deterministic but not semantic. Replace it
  with a declared production embedding provider before evaluating recall quality.
- Runtime, evidence, graph, posterior, residual, abstraction, policy, and
  contract versions survive restart and publish atomically with mutations.
- Selective Bayesian usefulness updates are active only when externally issued
  selection-support and omitted-influence certificates bind the current epoch.
  Anti-Pigeon sharing additionally requires an external target-diameter audit.
  Missing, expired, or stale certificates fall back to the baseline scored law.
- The Bayesian outcome model currently scores retrieval usefulness, not a general
  next-world-event law. Sheaf-inspired snapping and residual prediction remain
  later milestones.
- No proactive action is executed. The daemon defines a data-only agency proposal
  type so a later OpenClaw authority layer can approve, reject, or schedule it.
- TCP listening has no transport authentication in this alpha. Prefer the default
  local Unix socket; do not expose a TCP listener beyond a trusted loopback test.

Prometheus text metrics are available at `GET /metrics`. A Phase 1 database must
be migrated with an absolute backup path before startup:

```sh
./bin/eventframed -database /absolute/events.libravdb \
  -migrate-v1 -migration-backup /absolute/events.pre-v2.libravdb
```

See [docs/architecture.md](docs/architecture.md),
[docs/protocol.md](docs/protocol.md), and [docs/roadmap.md](docs/roadmap.md).

## License

MIT. See [LICENSE](LICENSE) and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
