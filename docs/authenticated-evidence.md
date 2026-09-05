# Authenticated Evidence and Reversible Working Beliefs

## Boundary

This opt-in layer protects `ObserveBayesianOutcome`, after the OpenClaw contract
boundary. Raw chats and extracted 5W1H frames remain untrusted memory; they do not
become authenticated observations merely because the daemon ingested them.
Neither a signature nor this filter certifies factual truth or independence.
The existing Anti-Pigeon, selection, omission, and residual gates retain authority.

Enable `--evidence-trust-file /absolute/path/public-evidence-keys.json` to require
signed outcomes in every feedback class. Add `--working-belief` for the reversible
filter. Without a registry, unsigned compatibility behavior remains enabled;
attested requests are rejected rather than silently treated as verified.
Working belief requires authentication and currently rejects hierarchical
posterior mode: no undeclared composition of those two models is performed.

The registry is an array of objects with these fields:

| Field | Contract |
| --- | --- |
| `issuer`, `key_id`, `tenant_id` | Operator-enrolled, nonempty identifiers, at most 256 UTF-8 bytes |
| `source` | Exactly `selected`, `independent_audit`, or `full_stream` |
| `public_key` | Base64 Ed25519 public key, 32 decoded bytes |
| `not_before`, `not_after` | RFC3339 validity window, exclusive upper endpoint |
| `revoked` | Boolean; revoked entries cannot authenticate outcomes |

Private keys stay with the observer, never in this registry or chat memory.
The registry permits at most 4096 entries and 4 MiB. Provision keys independently
of retrieved content. Replace it through a controlled restart, not an agent tool.
Registry changes alter the policy fingerprint. Old-trust posteriors cannot be
served in authenticated mode, and new outcomes start new-trust statistics.
This is conservative invalidation, not selective historical subtraction.
Removing the registry intentionally returns to weaker legacy behavior.

## Envelope and Replay Rules

The outcome request gains `attestation` containing `issuer`, `key_id`,
`observation_id`, optional `parents`, and base64 `signature`.
`internal/epistemic.OutcomeSigningBytes` specifies the exact versioned JSON-array
encoding. It is not arbitrary JSON serialization or JCS. Sign its bytes using
Ed25519. The ordered array contains:

1. Schema string `eventframe-outcome-attestation-v1`, protocol version.
2. Issuer, key ID, observation ID, sorted parent IDs (empty array if none).
3. Tenant, event ID, journal ID, resolved usefulness Boolean.
4. Explicit usefulness (Boolean or null), packed, cited, successful-downstream,
   correction, and rejected signal fields, in that order.
5. Observation and availability times, UTC RFC3339Nano strings.
6. Source and inclusion probability encoded as lowercase hexadecimal IEEE-754
   float64 bits without padding or prefix.

The standard Go JSON encoder's escaping rules apply. Sign the interpreted
usefulness plus the signal fields, so stored normalized outcomes remain verifiable.
Request retry IDs and the signature itself are excluded. All evidence-bearing
fields, source scope, and target journal are bound. Keys are checked at verification
time and at the claimed observation time. The 4096-entry successful-verification
LRU never bypasses key validity or admission checks on a cache hit.

After verification, a domain-separated digest of `(tenant, issuer, observation_id)`
replaces the transport idempotency key in the existing atomic outcome transaction.
It deliberately excludes key ID and journal ID. An admission-eligible exact retry is a duplicate;
changing the journal, outcome, or signing key for an already consumed observation
conflicts instead of learning again. Concurrent submissions and database restart
cannot consume that ledger entry twice. Replay protection depends on retaining
the database ledger; deleting or rolling back the database removes that history.
The `authenticated-observation:` namespace is reserved against unsigned callers.

At most eight unique parent references are accepted syntactically. Parent-bearing
claims are withheld from independent outcome admission, not recursively resolved
on the request path. Their text can still live in ordinary untrusted memory.
There is no new autonomous provenance-investigation worker in this change.
An authorized observer can still lie, omit parents, or invent fresh observation
IDs. Separate observers can repeat the same underlying source. Authentication is
not a Sybil defense, a source-independence proof, or a substitute for external
audits and the existing correlated-evidence gate.

## Reversible Filter

The declared hypotheses are Bernoulli usefulness models with probabilities
`p_low = 0.2` and `p_high = 0.8`, equal prior odds, retention `lambda = 0.98`,
log-odds bound `L = 6`, and per-observation log-factor bound `c = 2`.

For admitted outcome `x` and effective update weight `w`, let:

```text
factor = log P(x | H_high) - log P(x | H_low)
ell'   = clip(lambda * ell + min(1,w) * clip(factor,-c,c), -L,L)
p_useful = p_low + (p_high-p_low) * sigmoid(ell')
```

Successful replay exclusion is the binary novelty gate: a consumed observation
never reaches this update again. Shared-pool discounting is retained; selection
importance weights above one cannot manufacture multiple observations here.
This capped, discounted filter is a working/generalized update, not an exact
selection-adjusted posterior. In particular it does not inherit unbiasedness or
calibration guarantees from inverse-probability weighting. The parameters are
explicit implementation defaults, not fitted or established empirical optima.

The predictive map above enters the existing belief-conditioned scored forecast
before residual correction. It does not confuse `P(H_high)` with `P(useful)`.
Raw Beta/member evidence remains available for Anti-Pigeon and monitoring.
Authorized reset starts from equal odds and includes the revealing outcome once.
A plain split starts new child working filters at their prior without importing
pooled certainty; retained per-member statistics still drive structural tests.
A split-reset initializes the triggering child's filter with its revealing outcome.
Changing the filter policy also starts its working state at the new prior.

Neither old confidence nor source authentication makes a belief immutable:
in the unit fixture, five newly admitted negative observations reverse a filter
previously saturated by 10,000 positive observations, without a changepoint reset.
That demonstrates reversibility, not the factual correctness of those observations.

## Verification and Timing

Focused tests cover tampering, source/tenant scope, key rotation and revocation,
cached expiry, derived claims, bounded cache size, normalized-signature readback,
concurrent replay, cross-journal replay, disk restart, returned-state isolation,
registry invalidation, and propagation into the scored forecast.

The benchmarks include primitive checks and complete internal service requests.
The latter use a temporary on-disk LibraVDB, 50 synthetic events, a 32-dimensional
hash embedder, recall/pack budgets 50/10, and one or four workers. Mixed mode is
75% recall and 25% unique outcome writes. Request timing includes internal store
and lock waits; observer signing happens before timing. No production data,
production daemon, external model, or external database service is used.

These are bounded synthetic latency tests, not an OpenClaw end-to-end benchmark,
corpus-scale test, calibrated accuracy experiment, or hard real-time guarantee.

See the [audit](authenticated-evidence-audit.md),
[timing summary](authenticated-evidence-results.md), and
[raw repeated benchmark output](authenticated-evidence-benchmark.txt).
