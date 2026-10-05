# Query-scoped posterior validity v20: frozen component

Date: 2026-10-02. V19 showed that refreshing certificates cannot
make an older posterior valid after event ingestion advances the
global `EvidenceEpoch`. Blind epoch retagging is forbidden. This
research-only component tests a conservative alternative for a
**fixed query and retrieval-usefulness horizon**. It is not wired
to `Service.Recall` and cannot rescue the v19 visible-frontier
churn workload by itself.

Let a posterior record bind a tenant, posterior key, query digest,
model/policy identity, a base runtime version, an exact top-k
frontier ID digest and a certified lower bound on the k-th score.
An append-only mutation witness is complete only if it accounts
for every runtime-version step through the current snapshot.
For a request as-of time, a new event is non-impacting only when
it is unavailable as of that time, or an independently computed
upper bound on its query score is **strictly below** the frozen
frontier cutoff and it has no graph/abstraction dependency into
the posterior key. A relevant outcome on the same posterior or
dependency group is impacting. Changes to query, model, policy,
graph, abstraction, horizon, source identity, or an unclassified
runtime step fail closed. A certificate publication can advance
runtime version without changing the posterior, provided its own
validity is checked by the caller. No decision may depend on
outcomes available after request as-of.

Implement this as a pure research package with a bounded input log,
not a production Store or Service change. Tests must accept a
future-only append and a strictly below-cutoff visible append;
reject equal-cutoff/competing visible appends, graph dependency,
same-key outcome, missing version, mixed query/model, stale source
and after-as-of evidence. Check monotonicity under harmless appends
and benchmark cost at 128 and 10,000 mutations. Retain an explicit
``unknown'' result when score/dependency coverage is incomplete;
never coerce uncertainty to compatibility.

This is a **conditional proof obligation**, not empirical target-law
coverage. The exact top-k and score bounds must be supplied by an
independent, versioned oracle; the current service does not yet
record all required provenance, and its event-ID posterior key can
pool different queries. In the v19 fixture, high-relevance writes
enter the top-150 frontier and should fail this guard. Passing unit
tests establishes the guard's internal logic only. A later full
Goal 6 rescue would need durable provenance, feasible incremental
certification, scored-law integration and loaded latency.

The design is informed by dependency tracking in
[provenance semirings](https://www.cs.ucdavis.edu/~green/papers/pods07.pdf)
and [differential dataflow](https://www.cidrdb.org/cidr2013/Papers/CIDR13_Paper111.pdf),
but no theorem from those papers is claimed for EventFrame's
posterior or retrieval semantics.

## Audit amendment

Before the final component run, the state audit added three fail-closed
requirements. Every witnessed mutation must account for its resulting
`EvidenceEpoch`; otherwise an epoch jump can hide inside a certificate
publication. The requested frontier digest must equal the record's digest.
Outcome availability must be recorded so feedback later than the request
as-of cannot invalidate an immutable old posterior merely by existing.
These additions tighten the original screen; they do not weaken any
predeclared rejection condition. The caller must independently establish
that the retained posterior itself was built only from as-of evidence.

The score bound is over the exact, frozen nomination-score space, including
all ranking terms capable of changing the top-k cutoff. Approximate ANN
membership or a raw cosine bound alone is not such a certificate when later
ranking stages can change membership. This component does not construct
those bounds or attest the caller's frontier digest.
