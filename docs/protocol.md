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
schema is additive. An explicit vector must name the exact active
`embedding_model`; model and dimension mismatches are rejected.
