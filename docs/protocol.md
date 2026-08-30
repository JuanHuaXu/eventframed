# Protocol v1alpha1

All requests use JSON and the exact version string `eventframe.v1alpha1`.
Unknown fields are rejected. Request bodies are capped at 4 MiB.

## Endpoints

### `GET /v1/health`

Returns backend, vector dimension, quantization mode, and the current version
snapshot. This is a liveness endpoint; it intentionally remains available while
an external candidate contract is down.

### `GET /v1/ready`

Returns HTTP 200 only when the durable store and every configured required
external candidate contract are ready. A required LibraVDB connection in
transient failure, shutdown, or an open circuit returns HTTP 503 with
`status: not_ready`. Load balancers should gate serving on this endpoint.

### `POST /v1/turns:capture`

The OpenClaw adapter submits a raw successful turn containing:

- capture, tenant, session, run, and optional agent identity
- user and assistant text
- occurred, observed, and available timestamps
- recalled event identifiers used by that turn

The turn object deliberately has no 5W1H fields. After JSON contract decoding,
`eventframed` constructs the immutable conversational event and performs bounded
deterministic 5W1H enrichment. Text-derived evidence records byte spans such as
`user[bytes:start:end]`; metadata fallbacks name their basis. Explicit user
values are `observed`, assistant values are `synthetic`, and compositions or
fallbacks are `inferred`. These fields support retrieval and do not constitute
causal evidence without the separate SCM-backed path.

### `POST /v1/events:observe`

Requires `idempotency_key` to equal `event.id`. Repeating an identical event is a
successful duplicate. Reusing the ID for different content returns HTTP 409.
This endpoint accepts intentionally authored structured EventFrames from direct
clients. It does not run raw-turn enrichment or overwrite supplied fields.

Every event carries:

- tenant/session identity and sequence
- event kind and untrusted content
- occurred, observed, and available timestamps
- 5W1H fields with source and confidence
- priority, tags, provenance, and attributes
- an optional canonical vector of the configured dimension

If no vector is supplied, the daemon uses its configured embedder.
Generated and caller-supplied vectors are bound to
`eventframe-5w1h-v1` in the embedding model key. A predecessor key is rejected;
the raw `content` field is never included in the canonical embedding text.

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

### `POST /v1/openclaw/context:recall`

Runs the same internal recall operation and scoring contracts, then returns the
OpenClaw projection. Candidate events contain raw content, lifecycle metadata,
provenance, and ranking outputs; daemon-internal 5W1H fields and embeddings are
omitted. The OpenClaw adapter uses this endpoint exclusively.

Optional EventFrame output policies are policy-digested and independently
switchable. Contextual and hierarchical evidence may change the committed
predictive law. With `-libravdb-contract-endpoint`, EventFrame owns the external
candidate lifecycle and invokes LibraVDB `SearchTextCollections` for nomination
and `RankCandidates(k1,k2)` for base retrieval order. Returned rank scores cannot
alter the already committed probability law. Each candidate reports that value
as `retrieval_score`, a bounded EventFrame Bayesian/residual correction as
`rank_delta`, and their clipped sum as final `score`. The correction is applied
and sorted after the LibraVDB contract returns and before the bounded packer.
By default, `packet_answer_certainty` is the normalized retrieval-score gap at
the initial `pack_k` boundary. It is rank-domain certainty, not a calibrated
probability of answer correctness. An independent
`rank_delta_correction_reliability` authorizes each correction; the current
runtime assigns nonzero reliability only to an accepted Bayesian, residual, or
graph-compatibility path. For certainty `c` and reliability `r`, the elastic
multiplier is `[min_scale + (max_scale - min_scale)(1 - c)]r`. Candidates expose
that pair, `rank_delta_scale`, and the
`rank-boundary+correction-reliability` basis. `packet_confidence` and
`rank_delta_confidence` remain deprecated aliases for answer certainty during
migration. The elastic result remains inside the absolute rank-delta cap and
can be disabled for ablation; disabling it preserves the unscaled certified
delta.
Rank deltas are materialized in a bounded RAM cache plus SQLite and are accepted
only under their exact query/event key, validity interval, and complete semantic
version tuple. The embedded search and pass-through ranker are standalone
fallbacks, not substitutes in contract-native claim tests.

The production collection is `eventframe-` plus a truncated SHA-256 tenant
digest. Recall requests may omit `retrieval_collections`; when the external
contract is enabled, EventFrame supplies the reserved collection. Any request
that attempts to override that collection or name exclusions from another
collection is rejected. Observation retries use `eventframe_digest` with
`ListByMeta` to reconcile a successful external insert after a lost response.

The default daemon policy activates every evidence-ready member of the bounded
`recall_k` frontier. "Update all" therefore means frontier-update-all, never a
scan or posterior update over the complete LibraVDB corpus. Explicit selective
policies remain supported for evaluation and constrained deployments.

The journal commit compares the captured semantic version snapshot under the
durable write lock. Runtime/evidence counters advanced only by events becoming
available after the request's fixed `as_of` do not invalidate that historical
decision. Concurrent policy, graph, posterior, residual, abstraction, agency,
or contract motion returns HTTP 409 `snapshot_changed`; clients retry at most
twice with the identical request and `as_of` value.

### Bayesian evidence and certificates

The OpenClaw plugin registers the operator-write gateway method
`eventframe.outcome.observe`. It accepts an exact `tenant_id`, `journal_id`,
`event_id`, `idempotency_key`, and one of `useful`, `not_useful`, `cited`,
`successful_downstream`, `correction`, or `rejected`. The method maps that
explicit control-plane signal to `POST /v1/bayesian/outcomes:observe` with
full-stream inclusion probability one. The daemon rejects events that were not
nominated by the referenced durable recall journal. Successful turn completion
and packed exposure alone remain neutral; the agent cannot certify its own
answer through an agent-callable tool.

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
- `POST /v1/bayesian/groups:compare` is a side-effect-free slow-path comparison
  of one shared Bernoulli model against independent member models. It returns a
  `share`, `split`, or `uncertain` proposal and the current posterior keys. Every
  result states that Anti-Pigeon certification is still required; this endpoint
  cannot create a group or publish a certificate.

Outcomes update a bounded Beta posterior and member-level sufficient statistics.
A certified shared posterior applies the configured `shared-evidence-weight`
(default `0.5`) to pooled Alpha/Beta support only. Member-level sufficient
statistics retain the full inclusion weight, preserving Anti-Pigeon comparison
and split authority. Event-local posteriors are never discounted by this rule.
A capped Bayesian online changepoint detector combines exact BOCPD evidence with
a warm-started, two-sided CUSUM and post-reset cooldown. A detected shift resets
the posterior and atomically advances evidence, graph, posterior, residual, and
abstraction versions. This invalidates stale
certificates before another forecast can use them. These guarantees are runtime
contracts; certificate quality remains an empirical responsibility of the named
external audit procedure.

Contract 10 joins that temporal detector to an Anti-Pigeon revision gate. After
each outcome commit, a bounded shared-versus-split comparison returns one of
`retain`, `individual_reset`, `shared_reset`, `split`, or `split_reset` in the
outcome response. A split requires sufficient member support, the predeclared
posterior threshold, and full-stream or independent-audit evidence. It atomically
revokes the old sharing certificate, disables the shared posterior and its
residual, and materializes event-keyed posteriors from the member sufficient
statistics. Revocation is fail-closed: it does not certify any replacement group.
Selected-only evidence may update beliefs but cannot revoke a sharing certificate.

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
not-useful probability mass through the calibrated baseline, accepted
belief-conditioned score, calibrated pre-residual law, corrected law, and an
aligned decision template. A frozen baseline map applies when no certified
belief is accepted; a separately frozen predictive map applies after the bounded
baseline-posterior score is formed. Each map is monotone within its branch, and
their identities and fit artifacts are part of the policy contract. The baseline
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

Outcome feedback may additionally carry boolean-only evidence signals for
packing exposure, citation, successful downstream use, correction, rejection,
and explicit usefulness. Rejection or correction dominates positive evidence;
packed-only exposure is not accepted as a usefulness observation. Free-form
feedback content is outside the contract.
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
`contract_version=8` adds separate contract-native nomination and ranking identities while preserving
the version-7 authority split. It authenticates authority claim/resolution separately, bounds
all proposal and consumer identifiers, and adds a tenant-scoped active queue
projection. Version-6 databases rebuild that projection before the new marker is
published. Agency mode requires the local Unix socket; bearer credentials are
never accepted over the daemon's plaintext TCP listener.
Pending contract-6 proposal payloads remain signature-valid and are accepted by
the contract-7 authority; only newly issued proposals carry contract 7.

`contract_version=9` separates bounded frontier-all cheap updates from selective
deep review, adds practical-equivalence and partial-pooling diagnostics, exposes
graph compatibility in recalled candidates, and permits runtime-estimated
omitted-influence certificates through
`POST /v1/bayesian/certificates:estimate-omitted-influence`. Runtime estimates
are confined to an explicit finite event-id population selected with the active
nonzero audit probability and bound to the exact durable recall-journal query
digest. The runtime rejects nominated members, stale journals, and cross-query
reuse; these are not corpus-wide certificates.

`contract_version=10` adds the atomic Anti-Pigeon revision transition and durable
revision result described above. Existing version-9 stores upgrade additively;
no prior certificate is silently reinterpreted as evidence for a new group.

`contract_version=11` adds the raw-turn capture contract and moves OpenClaw 5W1H
enrichment entirely behind the daemon boundary. The durable EventFrame schema is
unchanged, so existing version-10 stores upgrade additively.

`contract_version=12` makes canonical `eventframe-5w1h-v1` text the exclusive
semantic corpus for local vectors, remote nomination/reranking, query digests,
and packing diversity. Raw transcript remains metadata and final payload.
Embedding keys and remote candidate collection names bind the representation.
Predecessor databases require backup-first re-embedding and configured remote
candidate stores require the separate reindex maintenance operation.
