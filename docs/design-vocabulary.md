# Design Vocabulary

This note maps informal names from the EventFrame design sessions to concrete
runtime behavior. The private working transcript is not distributed because it
contains unrelated personal and project context. The whitepaper remains the
formal specification; this note only preserves the intent behind shorthand used
in code and discussion.

## Skepticism

"Skepticism" is inverse-certainty rank plasticity. When the packing boundary is
clear, an accepted cached correction is allowed to move scores conservatively.
When the boundary is uncertain, the same correction may move scores farther so
the system can adapt quickly.

This is not a probability that an answer is true. Boundary certainty never
authorizes a correction: the correction must already pass its reliability,
version, and applicability gates. Reliability can reduce the movement to zero.

## Curiosity

"Curiosity" is the bounded nomination of deferred sensitivity work after a
low-certainty recall. The request path copies a small, already retrieved
EventFrame frontier into a nonblocking queue. A single background worker runs
at most one audit at a time, starts it only when no recall is active, and
discards work that no longer matches the captured snapshot. A recall that begins
after admission may overlap the audit, so deployments must include that resource
contention in their latency measurements.

Curiosity is not an emotion, an online reasoning loop, or permission to mutate
memory. Its output is proposal-only evidence for later review.

## Anti-Pigeon and shock

"Anti-Pigeon" is short for anti-pigeonholing. It prevents events with divergent
downstream behavior from continuing to share an abstraction or posterior merely
because they once looked similar.

An Anti-Pigeon shock combines validation-eligible changepoint evidence with a
shared-versus-split recommendation so stale shared confidence can be revoked
quickly. Selected evidence may update a posterior, but it cannot certify its own
split or merge. Anti-Pigeon retains final authority over sharing.

## Fuzzing and the invariant seeker

Fuzzing constructs declared synthetic changes to internal 5W1H fields and
measures movement in a bounded predictive law. Repeated stability can nominate
a higher-order invariant. The result is a sensitivity finding, not a causal
claim and not an automatic ontology update.

## Predictive chain translation

A translation requires an aligned upstream variable change to propagate through
the whole mapped chain while non-target fields remain fixed. Similar endpoints
alone are insufficient: each aligned stage and the terminal effect must agree.

## Sheaf snap

"Sheaf snap" evokes fitting locally compatible puzzle pieces into a coherent
larger structure. The runtime checks local compatibility and reconciliation, but
the implementation remains a sheaf-inspired scaffold until a deployment
instantiates the required base spaces, restriction maps, identity laws, and
composition laws.

None of these names grants subjective state, agency, tool authority, or execution
permission. They describe bounded scheduling, scoring, and validation policies.
