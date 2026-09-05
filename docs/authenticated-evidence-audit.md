# Authenticated Evidence Audit

Date: 2026-09-05. Scope: outcome admission, working usefulness filter, dependent
forecast reuse, persistence, configuration, and latency. No production operations.
The preexisting `.learnings/ERRORS.md` change is outside this patch.

## Round 1: Admission and State Ownership

- Confirmed during implementation: a signed fallback `Useful` field could be
  overwritten by explicit feedback before persistence, leaving an unverifiable
  durable request. The signing schema now binds resolved usefulness together
  with every original signal field. Normalized-request verification is tested.
- Confirmed during implementation: returning the new working-state pointer from
  the memory backend could expose the stored state to caller mutation. Copies
  now isolate reads and both original and duplicate update responses. The
  service-level regression mutates a returned state and checks the next retry.
- Verified: signature modifications, feedback-source escalation, tenant changes,
  cached expiry, explicit revocation, and derived claims fail closed.

## Round 2: Integration and Invalidation

- Confirmed: existing `BindBayesianPolicy` advances the policy version but does
  not itself erase posterior statistics. A registry change therefore needed its
  own stored trust fingerprint, recall gate, and first-update reset. Added these
  for member and parent posteriors; children retain the relevant fingerprint.
  The final sibling-path check also gates group-comparison evidence by the same
  fingerprint, with a regression assertion after registry replacement.
- Confirmed during implementation: split-reset materialization would lose the
  new working filter's revealing outcome. The triggering child now initializes
  from that outcome; plain splits deliberately start working filters at prior.
- Confirmed in the touched durable load branches: testing epoch mismatch before
  a non-not-found read error could hide that error. Both member and parent paths
  now propagate the error before deciding to reset state.
- Verified: the working hypothesis probability is mapped to usefulness before
  entering the scored bundle. It is not a separate unscored decoration.
- Verified: registry validation happens before startup creates state directories
  or opens a database; invalid working/hierarchical combinations are rejected.

## Round 3: Concurrency and Persistence

- Eight concurrent deliveries of one signed observation produce exactly one
  update in both memory and embedded LibraVDB fixtures.
- Re-signing the same observation for a new journal cannot count it again.
- Key rotation preserves the observation ledger namespace; revocation changes
  trust identity. Cache entries never bypass validity checks.
- Database close/reopen preserves the signed observation's duplicate status and
  working state. Duplicate protection shares the existing atomic transaction
  with posterior, residual, and version updates, not a second database.
- Focused race checks include the verifier, Bayes code, service, both stores,
  and API. The full Go suite also exercises existing Anti-Pigeon and residual
  behavior. These checks do not establish absence of every possible bug.

## Explicit Remaining Boundaries

- Trusted observers can lie or invent fresh IDs; signatures do not establish
  independence, completeness, or real-world truth.
- Unknown and parent-derived claims are withheld, not resolved automatically.
  There is no new provenance-research background worker in this patch.
- The filter is opt-in, generalized rather than exact Bayes, and has no new
  real-data accuracy or calibration validation. Default likelihood parameters
  must be evaluated on held-out, authentically collected outcomes before rollout.
- No production keys were enrolled and no production deployment was changed.
- The on-disk benchmark includes internal lock/storage waits but uses a small
  synthetic corpus and local hash embeddings. It is not a production SLA.
