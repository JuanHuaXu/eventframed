# Window observer integration boundary

## Verified distinction

The existing count model uses `(yes+1)/(n+2)` independently at each partial mask.
It is NOT generally the marginalization of the subset model's uniform-input
joint law. With two positive samples at input0:

- Unobserved forecast:.75.
- Forecast after observing bit0=0:.75; bit0=1:.5.
- Uniform-input weighted child mean:.625, differing from the parent by.125.

This is a counterexample to silently inserting count forecasts into the existing
joint observer's common-input-law contract. It does not say that a count forecast
at a specified mask is an invalid Bernoulli prediction, nor that production has
a new bug. No count estimator or joint observer was modified.

The positive control exhaustively checks the actual uniform-input subset model's
parent/child identities across19,683 partial states and59,049 missing-bit edges.
Those identities pass to1e-12. Equal endpoint predictions alone would not establish
this compatibility, so the two checks are intentionally separate.

## Test-only integration component

`windowGuidePredict` explicitly distinguishes an acquisition proxy from the law
that is scored. It extends the existing subset observer's role selection:

1. Choose the strongest available base/short/long outer role, preserving the
   existing raw-weight guide convention and ignoring neutral as a read guide.
2. If short wins, compare old-short mass, event-count mass and combined event-
   subset mass. If old-short wins, use its count-versus-subset comparison.
3. Run the existing bounded observer for that one proxy. It acquires a concrete
   mask; cached or hypothetical information is not silently added to that mask.
4. Evaluate every participating predictor on the acquired mask. Return complete
   old-inner, new-inner and outer advice plus the actual scored law and guide.
   This immutable prediction step does not update evidence or weights.

The proxy is NOT claimed to optimize value of information for the full mixture.
Its choice is a heuristic to test in the closed loop, not a new sheaf/proper-score
theorem or a completed falsification-observation result. Missing refreshed models
fall back to the retained guide and neutral refreshed advice; invalid weight
states are rejected before acquisition rather than silently choosing a guide.

## Checks and timing

- All512 inputs across six guide branches (3,072 predictions) pass: returned
  values match only the returned mask; read cost<=6; scores equal the returned
  nested advice mixtures; guide is explicit; models and input weight states
  remain unchanged.
- Missing-refreshed-model fallback and malformed weight states are tested.
- Race tests pass: count/coherence audit0.04s, guide contracts0.30s, package1.583s.
  Vet passes. No outcome-model learning or quality experiment occurred here.
- Apple M4 in-memory prediction including simulated reads and forecast assembly:
  count guide10.901-10.933us,11,357B/39allocations; subset guide4.948-4.957us,
  5,587B/28allocations. Fixture models and weights are frozen; this is not
  production serving latency, a distribution-wide performance bound, or a
  comparison against the previous66ns weighting-only kernel.

Source SHA256 for `internal/observationgate/window_observation_contract_test.go`:
`bc22ff37c16a9c730aaad192fdcdac124d860e7f1796eb553a2eca07731dc3e6`.

Next is the [closed-loop contract](mmm-window-coupling-v1-contract.md). The prior
2/4 gain and16/16 nonharm results remain fixed-view evidence only. All seven
goals remain open; production, paper, remotes and existing tracked edits are
untouched. This component lives solely in a test file.
