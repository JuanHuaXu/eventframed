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
retention/deletion propagation, backup, compaction, guarded migration, a
certificate-gated selective Bayesian usefulness layer, and certified law-only
residual reuse for retrieval outcomes. Phase 4B adds bounded predictive graph
publication, server-computed dependency closure, targeted stale marking, and
monotone rollback after externally confirmed split or merge proposals. Phase 5
adds signed, durable wake/notify/schedule proposals and a fail-closed OpenClaw
authority layer. It does not grant `eventframed` tool-execution authority.

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

The authority path runs beside recall: a trusted local planner submits a bounded
proposal to `eventframed`; the daemon validates its evidence, signs the exact
payload, and leases it from a durable queue. The OpenClaw adapter verifies the
signature and independently applies tenant, session, capability, consent,
causal-depth, expiry, quiet-hours, and kill-switch gates before scheduling a new
agent turn. That turn remains ordinary untrusted input and receives no tool
permission from the proposal.

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

### Bounded agency

Agency is disabled on both sides by default. Start the daemon once with:

```sh
./bin/eventframed -agency-enabled
```

This creates a mode-0600 Ed25519 private key, issuer token, and authority token
under `~/.eventframed/keys`, plus a public verification key. Give only the issuer
token to the trusted local component allowed to submit proposals. The OpenClaw
adapter receives the separate authority token used to claim and resolve them. Configure the
OpenClaw plugin with `agencyEnabled: true`, `agencyKillSwitch: false`, the public
key and authority-token paths, explicit `agencyConsentActions`, matching `agencyCapabilities`, and at
least one `agencyAllowedSessionPrefixes` entry. Empty consent, capability, or
session scopes deny every proposal. Both UTC quiet-hour endpoints must be set or
both omitted.

The current daemon accepts proposals from an authenticated planner; it does not
yet synthesize them from recall or Bayesian state by itself. Proposal issuance,
leasing, signature verification, policy authorization, scheduling, and terminal
resolution are implemented end to end.

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
  next-world-event law. Its Phase 4 forecast bundle and residual cache are a
  concrete Bernoulli specialization. The compatibility graph currently supports
  only the declared `identity_bernoulli` comparison map. It is a sheaf-inspired
  predictive scaffold, not a sheaf or evidence of causal structure.
- Predictive snap candidates and their chronological confirmation statistics are
  produced by an external slow-path auditor. The daemon independently enforces
  graph bounds, dependency closure, certificate coverage, acceptance thresholds,
  atomic publication, and rollback; it does not estimate those certificates.
- Agency can only request a wake, notification, or scheduled agent turn. The
  OpenClaw authority layer cannot execute tools through this protocol, and the
  generated turn explicitly retains normal user and OpenClaw approval policy.
- Proposal signing and queue mutation are atomic inside LibraVDB. Scheduling and
  durable resolution span two processes and therefore are not exactly-once. A
  deterministic scheduler tag, lease, and rollback path reduce duplicates, but
  crash-window behavior still requires production fault testing.
- Agency mode is restricted to the mode-0600 Unix socket. Plaintext TCP listeners
  remain available only when agency is disabled.
- The kill switch prevents new and in-flight authorization. It does not revoke a
  turn that was already scheduled and durably approved before the switch changed.
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
