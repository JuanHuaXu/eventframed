# Grid belief critique and patch audit

## Before patch

- Confirmed modeling limitation: a mixture of .2 and .8 cannot represent useful
  rates outside that convex hull. Clipped odds narrow it further. This is not
  bad stored data, a retrieval failure, or a calibration implementation defect.
- Multi-sampling: inspect the primitive predictive map, service BeliefLaw and
  PredictiveScore wiring, and both transactional stores. The stored likelihood
  weights are not themselves the served useful probability.
- Falsifier: a legacy raw belief below .2 or above .8 under a valid default
  policy would refute the range diagnosis. Broader calibrated scores do not.
- Scope: authenticated usefulness feedback only; no source-truth inference,
  automatic certificate approval, production access, or global data repair.
- Related fix check: fetched origin/main at d5cda75; no open/closed issues or PRs
  returned by the repository CLI queries. No upstream grid implementation found.
- Invariant: predict from committed state before outcome; a fresh admissible
  identity updates once; reads/replays never advance the hidden transition;
  policy changes discard incompatible working state. Fixed 21-state loops.

## Round 1: formulation and state

Checked the reset-HMM transition against a uniform prior and exact Bernoulli
normalization. State is the post-observation posterior; serving applies exactly
one reset transition to predict the next admitted observation. Update independently
derives that same prior, not a twice-decayed served state. Read-only tests enforce
this. No outcome, future seed, or trajectory parameter enters the prediction.
Invalid or mismatched weights start from uniform; oversized importance weights
are capped, fractional likelihoods are explicitly generalized Bayes.

The default model fingerprint must remain unchanged: the new false Grid field
is omitted from policy JSON. Grid v1 is a distinct policy, automatically included
in the existing Bayesian configuration digest and certificate versioning.
The legacy log-factor/retention fields are retained for schema compatibility;
grid v1 rejects nondefault variants and uses its own frozen share and grid.

## Round 2: integration and resource boundaries

The grid uses a fixed value array, not a slice alias. The existing memory-store
pointer-copy boundary therefore copies all weights. Durable JSON roundtrip,
restart, duplicate response mutation, and served-law equality are tested in
both stores. Split-reset is exercised for BOTH working modes and both stores:
one revealing outcome, no inherited pooled certainty, retained sibling evidence.
Trust and Anti-Pigeon gates remain unchanged; this adds no new authority.

Review caught an avoidable legacy payload expansion: omitempty does not omit a
fixed-length zero array. Changed the new field to omitzero before publishing.
The Go module's minimum version supports that encoding option. State storage
still adds 168 in-memory value bytes per allocated working state; nonzero grid
JSON is larger. No historical scan, new network call, lock, goroutine, or database
transaction is introduced. Grid updates are O(21) arithmetic plus the existing
fixed-size policy fingerprint work. Full request storage and contention remain.

## Experiment audit

Design and confirmation seeds are distinct and generated after freezing the
protocol; no parameter was tuned from either output. All 512 trajectories per
split are retained, with all three control/model metrics. Approximate normal
intervals use trajectory units, not correlated individual observations. The
Bonferroni intervals are approximate fixed-sample intervals, not exact confidence
sequences or distribution-free coverage. All scenario failures are retained.
The score-composition experiment is explicitly a simulation, while the service
tests separately verify the actual score wiring. It is not an LLM answer test.

No outstanding P1/P2 defect was found in this bounded implementation review.
Known scientific limitations remain: discrete model bias, fixed hazard, no
guaranteed truth, no full-stream selection correction, no new signed real-data
accuracy evidence, and no performance extrapolation to saturated production.
