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
Contract 14 adds reversible higher-order EventFrames: a slow-path invariant
seeker may propose a join, while the daemon requires exact Anti-Pigeon authority,
preserves every constituent, and applies explicit coarse/fine retrieval preference.
Contract 15 adds validity-constrained property fuzzing as a read-only slow-path
audit. Declared field bundles are perturbed over an as-of context, the local
embedding-nomination law is recomputed, and total-variation sensitivity plus a
one-sided stability bound is reported. The audit cannot publish an invariant or
make a causal claim. Low packing-boundary certainty can also nominate the bounded
context for an in-process background fuzz queue. Nomination is nonblocking;
one idle-gated worker executes snapshot-bound source-EventFrame bundle audits.
Contract 16 adds a separate chain-translation audit. It compares two aligned
baseline/revealed trajectory pairs, enforces exact stage maps and
unchanged-coordinate locality, and classifies higher-order invariant,
predictive-translation, or divergence candidates. It is read-only and never
turns a declared correspondence into a causal edge or accepted graph snap.
Structurally impossible candidates stop before prediction. Eligible audits use
version-matched stored EventFrame vectors and a bounded exact-score cache; the
response states whether prediction was evaluated.

## Architecture

```text
OpenClaw
  |  before_prompt_build / agent_end
  v
TypeScript contract adapter
  |  eventframe.v1alpha1 over a mode-0600 Unix socket
  v
eventframed (Go)
  |  structured state: embedded LibraVDB (mmap, optional SQ8/FSQ6/PQ)
  |  candidate contracts: Insert/SearchTextCollections/RankCandidates/Delete
  v
LibraVDB contract endpoint (optional production nomination and ranking)
```

The authority path runs beside recall: a trusted local planner submits a bounded
proposal to `eventframed`; the daemon validates its evidence, signs the exact
payload, and leases it from a durable queue. The OpenClaw adapter verifies the
signature and independently applies tenant, session, capability, consent,
causal-depth, expiry, quiet-hours, and kill-switch gates before scheduling a new
agent turn. That turn remains ordinary untrusted input and receives no tool
permission from the proposal.

The Go process owns all memory lifecycle and policy-bearing calculations. Its
embedded database is the authoritative structured EventFrame store. Production
may additionally configure a LibraVDB daemon for the frozen public candidate
contracts; `eventframed`, not the base `libravdb-memory` plugin, calls those RPCs
and owns their tenant-isolated collections. The TypeScript adapter only
translates OpenClaw hooks into the versioned protocol and injects escaped,
explicitly untrusted historical context. It submits successful conversations as
raw turn envelopes with no 5W1H fields. After that contract boundary,
`eventframed` performs bounded deterministic 5W1H enrichment before validation,
embedding, and durable indexing. At the API boundary the raw turn remains the
event payload; in the durable corpus it is retained as opaque metadata and the final recalled
payload. Canonical 5W1H text is the only corpus used for embedding, nomination,
reranking, and diversity. The derived fields remain predictive scaffolding
rather than causal evidence.

## Build and test

Requirements: Go 1.25 or newer and Node.js 22 or newer.

```sh
cd plugin
npm install
cd ..
make check
make build
```

Run the checked-in public-fact fuzzing confirmation with:

```sh
go run ./cmd/eventframe-fuzz-eval -split confirmation
```

Background fuzz nomination is enabled in the daemon by default. Inspect its
bounded in-memory queue at `GET /v1/invariants:fuzz-queue`; disable it with
`-background-fuzz=false` or tune its certainty, capacity, interval, timeout,
cooldown, event, and trial bounds with the corresponding `-background-fuzz-*`
flags.

Run the checked-in public-grounded chain-translation controls with:

```sh
go run ./cmd/eventframe-translation-eval -split confirmation
```

Run the daemon with conservative local defaults:

```sh
./bin/eventframed
```

Enable the full external candidate contract path with:

```sh
./bin/eventframed -libravdb-contract-endpoint unix:/absolute/path/libravdb.sock
```

Remote TLS and mTLS endpoints use the matching transport contract:

```bash
./bin/eventframed -libravdb-contract-endpoint tcp:memory.example:443 \
  -libravdb-contract-tls-mode tls \
  -libravdb-contract-tls-ca /absolute/ca.pem \
  -libravdb-contract-tls-client-cert /absolute/client.pem \
  -libravdb-contract-tls-client-key /absolute/client.key
```

`auto` (the default) uses plaintext only for Unix sockets and loopback TCP,
and TLS for remote TCP endpoints. TLS channels are pooled by gRPC rather than
opened once per retrieval request.

Remote contract traffic defaults to 16 concurrent RPCs, a 2-second per-attempt
deadline, at most two attempts for retryable transport failures, and a circuit
that opens after five consecutive failures for five seconds. Reads may proceed
concurrently, while insert/delete RPCs use a single-writer lane. Configure these bounds with
`-libravdb-contract-concurrency`, `-libravdb-contract-timeout-ms`,
`-libravdb-contract-attempts`, `-libravdb-circuit-failures`, and
`-libravdb-circuit-cooldown-ms`. Use `/v1/health` for liveness and `/v1/ready`
for serving readiness.

The deprecated `-libravdb-ranker-endpoint` flag remains an alias during
migration, but now enables the complete contract lifecycle rather than ranking
alone.

Defaults:

- socket: `~/.eventframed/run/eventframed.sock`
- database: `~/.eventframed/data/eventframe.libravdb`
- rank-delta sidecar: `~/.eventframed/data/rank-deltas.sqlite` with a
  100,000-entry write-through RAM cache
- elastic rank corrections: enabled, scaling certified raw deltas from 0.5x at
  a certain packing boundary to 2.5x at an uncertain boundary, multiplied by
  independent correction reliability and still subject to the hard cap
- vector dimension: 768
- traversal quantization: SQ8
- embedder: deterministic development hash (`-embedder hash`)
- recall budget: 50 candidates
- packing budget: 10 records / 2,000 estimated tokens

OpenClaw usefulness learning is explicit. The plugin exposes the
operator-scoped gateway method `eventframe.outcome.observe`, bound to one durable
recall journal and one nominated event. Packing or an otherwise successful turn
does not silently train the posterior.

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
  only the declared `identity_bernoulli` comparison map. Edges may be symmetric
  `compatible` relations or directional `supports` and `supersedes` rank
  relations; an omitted effect retains legacy `compatible` behavior. These are
  predictive controls, not causal edges. The graph remains a sheaf-inspired
  scaffold, not a sheaf or evidence of causal structure.
- Predictive snap candidates and their chronological confirmation statistics are
  produced by an external slow-path auditor. The daemon independently enforces
  graph bounds, dependency closure, certificate coverage, acceptance thresholds,
  atomic publication, and rollback; it does not estimate those certificates.
- Higher-order composition is a derived abstraction, not destructive compaction.
  `POST /v1/invariants:compose` retains an exact member set and representative;
  `POST /v1/invariants:decompose` deletes only the macro frame. Replaced or
  revoked Anti-Pigeon authority immediately removes the macro from recall even
  before physical decomposition. The daemon validates proposals but does not yet
  discover candidate joins autonomously.
- Background fuzz jobs contain a query vector, snapshot, bounded EventFrame IDs,
  and synthetic 5W1H replacement bundles, not raw query or transcript text. The
  current queue is bounded and in-process: pending work is lost on restart. It
  executes only while no recall is active, drops stale snapshots without retry,
  and can produce audit summaries only; it cannot publish an invariant, graph,
  posterior, residual, composition, or rank delta.
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

Version 12 changes the semantic corpus from raw-plus-fields to canonical 5W1H.
Existing durable databases must be re-embedded with a separate backup:

```sh
./bin/eventframed -database /absolute/events.libravdb \
  -migrate-eventframe-corpus \
  -migration-backup /absolute/events.pre-eventframe.libravdb
```

When the external LibraVDB contract is enabled, rebuild its versioned candidate
collection after the local migration:

```sh
./bin/eventframed -database /absolute/events.libravdb \
  -libravdb-contract-endpoint tcp:127.0.0.1:50051 \
  -reindex-eventframe-contract
```

See [docs/architecture.md](docs/architecture.md),
[docs/protocol.md](docs/protocol.md),
[docs/design-vocabulary.md](docs/design-vocabulary.md),
[docs/roadmap.md](docs/roadmap.md),
[docs/release.md](docs/release.md), and the
[claim rescue/replacement results](evidence/claim-rescue-v1/RESULTS.md).

## License

Apache License 2.0. See [LICENSE](LICENSE), [NOTICE](NOTICE), and
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
