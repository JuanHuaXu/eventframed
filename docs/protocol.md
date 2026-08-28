# Protocol v1alpha1

All requests use JSON and the exact version string `eventframe.v1alpha1`.
Unknown fields are rejected. Request bodies are capped at 4 MiB.

## Endpoints

### `GET /v1/health`

Returns backend, vector dimension, quantization mode, and the current version
snapshot.

### `POST /v1/events:observe`

Requires `idempotency_key` to equal `event.id`. Repeating an identical event is a
successful duplicate. Reusing the ID for different content returns HTTP 409.

Every event carries:

- tenant/session identity and sequence
- event kind and untrusted content
- occurred, observed, and available timestamps
- 5W1H fields with source and confidence
- priority, tags, provenance, and attributes
- an optional canonical vector of the configured dimension

If no vector is supplied, the daemon uses its configured embedder.

### `POST /v1/context:recall`

Requires tenant, session, query or vector, and `as_of`. The request may override
`recall_k`, `pack_k`, and `token_budget` within hard alpha limits. The response
contains only packed candidates, plus counts that reveal recall and eligibility
behavior.

`recall_k` is the number of candidates visible to reranking. `pack_k` is the
smaller number admitted to the context packet after scoring. Implementations must
not apply the packing cap before the scoring algorithm.

Every recall also durably records its bounded frontier, activation decisions,
audit inclusion probability, posterior-sharing key, query digest, as-of time, and
version snapshot. Query text is not copied into the Bayesian journal. The report
is `shadow` unless current external certificates support promotion. A cached
posterior may alter the score only when both selection support and omitted
influence are certified; the default Bayesian mixture weight is 0.10 and the
runtime rejects configurations above 0.25.

The journal commit compares the captured version snapshot under the durable
write lock. A concurrent semantic mutation returns HTTP 409 `snapshot_changed`;
clients may retry the recall with the same request and `as_of` value.

### Bayesian evidence and certificates

- `POST /v1/bayesian/certificates:publish-selection` publishes an externally
  audited simultaneous lower bound on nomination and activation support.
- `POST /v1/bayesian/certificates:publish-omitted-influence` publishes an
  externally audited upper confidence bound on the divergence between local and
  expanded update-all shadow forecasts.
- `POST /v1/bayesian/certificates:publish-anti-pigeon` permits listed EventFrames
  to share a posterior only when target-law diameter, support, horizon, graph,
  epoch, validity, and simultaneous-coverage checks pass.
- `POST /v1/bayesian/outcomes:observe` records availability-safe usefulness
  feedback bound to a durable recall journal. Accepted sources are full-stream,
  independent audit, and selection-certified feedback. Inclusion probabilities
  must match the journal and inverse-propensity weight is capped at 20.

Outcomes update a bounded Beta posterior. A capped Bayesian online changepoint
detector resets a shifted posterior and atomically advances evidence, graph,
posterior, residual, and abstraction versions. This invalidates stale
certificates before another forecast can use them. These guarantees are runtime
contracts; certificate quality remains an empirical responsibility of the named
external audit procedure.

### Bounded agency proposals

Agency routes exist only when the daemon starts with `-agency-enabled`:

- `POST /v1/agency/proposals:issue` authenticates with the private issuer token,
  validates a bounded draft against existing same-tenant evidence available at
  issue time, signs its exact JSON payload, and persists it idempotently.
- `POST /v1/agency/proposals:claim` requires the separate authority token and
  atomically leases eligible proposals to one named authority consumer. Priority
  orders claims before `not_before` and ID.
- `POST /v1/agency/proposals:resolve` records `approved` or `rejected` only while
  the named consumer holds a live lease and supplies the authority token.
  Approval requires a scheduler execution reference. Exact terminal retries are idempotent;
  conflicting terminal decisions return HTTP 409.

The only accepted actions are `wake`, `notify`, and `schedule`, bound respectively
to `eventframe.agency.wake`, `eventframe.agency.notify`, and
`eventframe.agency.schedule`. Every proposal carries a non-empty evidence set,
utility, priority, not-before time, hard expiry, idempotency key, causal chain,
and contract version. Default hard bounds are 32 evidence IDs, 4 KiB reason, 3
causal levels, 8 proposals per chain, 1,000 pending or claimed proposals per
tenant, 30 days into the future, and a 7-day validity interval. Claims use a
30-second lease and are capped at 50 per request.

The issue token, authority token, and signing private key are distinct local
secrets. The issue token is not sent to the OpenClaw adapter; the adapter uses
the authority token only for claim and resolution. It verifies the public-key signature and applies its
own consent policy before using OpenClaw's session scheduler. A signed proposal
is not tool authority. Event deletion and retention reject pending or claimed
proposals that cite the removed evidence in the same storage transaction.

Contract version 4 adds a retrieval-specific `forecast` bundle to every packed
candidate and durable frontier decision. It explicitly carries useful and
not-useful probability mass through the baseline, accepted belief mixture,
pre-residual law, corrected law, and an aligned decision template. The baseline
is labeled a plug-in Bernoulli forecast; this is not presented as a general
next-world-event model.

Outcome feedback is scored against the corrected law frozen in its referenced
journal before posterior or residual updates run. The posterior record accumulates
design-weighted Brier loss, forecast mass, observed mass, and effective
calibration weight. The same transaction updates exact action-key and general
posterior-key law residuals. Reuse checks exact before general and requires:

- matching policy, evidence epoch, and `retrieval-usefulness-v1` horizon
- bounded age and effective support
- an anytime-valid improvement lower confidence bound from unweighted
  full-stream or independently audited validation trials under repeated monitoring
- an analytic Bernoulli law-motion bound against the frozen base reference
- finite clipped correction and immutable source provenance

The law-only correction shifts probability between both Bernoulli branches and
then derives the template from the corrected law. No point residual is claimed.
Missing or failed gates return the uncorrected forecast bundle.

### Predictive abstraction graph

- `GET /v1/abstraction/graph?tenant_id=...` returns the currently published
  predictive graph together with the snapshot under which it was read.
- `POST /v1/abstraction/snaps:publish` compare-and-publishes a complete bounded
  candidate graph after disjoint chronological design and confirmation windows.
- `POST /v1/abstraction/snaps:rollback` republishes the named snap's previous
  topology under a fresh monotone version. Only the currently active snap may be
  rolled back.

The publish request identifies its exact base snapshot, finite candidate family,
unchanged candidate, externally fixed comparison obligations, net priority gain,
resource-cost and proper-risk bounds, and simultaneous external bucket/edge
certificates. The runtime computes the old/new dependency closure itself. A
statistically failed candidate returns `accepted=false` without mutation; a
stale base or concurrent publication returns HTTP 409. An accepted publication
atomically advances graph, abstraction, posterior, residual, and runtime versions
without changing the evidence epoch, and disables only posterior/residual records
named by the closure.

Contract version 5 supports only the predictive `retrieval-usefulness-v1` law
space and `identity_bernoulli` comparison map. These routes implement bounded
sheaf-inspired split/merge scaffolding, not causal-edge publication or a general
sheaf construction. Confirmation values remain empirical claims of the named
external audit procedure.

### Lifecycle endpoints

- `POST /v1/events:delete` removes one tenant event and atomically invalidates
  dependent graph, posterior, residual, and abstraction versions.
- `POST /v1/maintenance:retain` deletes the oldest events available before a
  declared cutoff, with a hard batch cap of 10,000.
- `POST /v1/maintenance:backup` requires an absolute destination and creates a
  point-in-time LibraVDB backup.
- `POST /v1/maintenance:compact` vacuums obsolete records and WAL frames without
  changing semantic versions.
- `GET /metrics` exposes bounded Prometheus request/error/latency counters.

Lifecycle calls are administrative. The alpha TCP listener has no authentication;
use the mode-0600 Unix socket and do not expose these routes remotely.

## Compatibility

`v1alpha1` is intentionally strict and may be replaced. Clients must reject
unknown response versions and should fail open for recall rather than blocking an
agent. Write retries must preserve the same event ID and payload.

Snapshot `contract_version=3` adds the durable Bayesian journal, externally
issued promotion certificates, bounded posterior state, and changepoint
invalidation. Version-2 durable state is upgraded in place because the LibraVDB
schema is additive. `contract_version=4` adds forecast commitments, delayed
calibration, and durable law-residual records; the same additive marker upgrade
applies. An explicit vector must name the exact active
`embedding_model`; model and dimension mismatches are rejected.
`contract_version=5` adds durable predictive graphs, snap audit records,
server-computed dependency closure, targeted invalidation, and monotone rollback.
It is also an additive durable-state upgrade.
`contract_version=6` adds signed agency proposal records, an independent monotone
`agency_version`, durable lease and resolution state, and evidence-lifecycle
invalidation. The schema upgrade is additive; agency remains disabled by default.
`contract_version=7` authenticates authority claim/resolution separately, bounds
all proposal and consumer identifiers, and adds a tenant-scoped active queue
projection. Version-6 databases rebuild that projection before the new marker is
published. Agency mode requires the local Unix socket; bearer credentials are
never accepted over the daemon's plaintext TCP listener.
Pending contract-6 proposal payloads remain signature-valid and are accepted by
the contract-7 authority; only newly issued proposals carry contract 7.
