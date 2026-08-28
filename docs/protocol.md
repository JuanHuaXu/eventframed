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

## Compatibility

`v1alpha1` is intentionally strict and may be replaced. Clients must reject
unknown response versions and should fail open for recall rather than blocking an
agent. Write retries must preserve the same event ID and payload.
