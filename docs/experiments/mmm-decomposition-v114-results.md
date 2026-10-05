# v114 model/mixture/observation diagnosis

Status: DIAGNOSTIC, not fresh validation and not a rescued candidate.
All192 selected consumed-parent trajectories completed, covering both parent
phases, both delayed switch directions and stationary parity4. All eight
publication blocks and arms5/7/8 are retained. See the
[protocol](mmm-decomposition-v114-protocol.md) and
[complete summary](mmm-decomposition-v114-summary.json).

## Findings

The reverse-switch late harm persists at matched observations. In the final
32-frame block of parent-confirmation parity-to-majority, rate-mixture Brier is
.081967, versus generic64 .057021 on its own observations and .058197 on the
mixture's observations. The exact decomposition of .024947 total harm is:

- Model/mixture term: .023770, descriptive interval [.009913, .037628].
- Observation term: .001176, interval [-.002659, .005011].

Parent-design repeats this pattern: .027979 total harm, comprising .025344
model/mixture difference and .002635 observation difference. These are paired
mean +/-3.5SE diagnostic intervals across32 trajectories, not new confirmation
tests or causal effects. The mixture spends5.535 coordinates on average in the
confirmation block versus4.874 for generic64; this is not an equal-cost claim.

The raw short generic model is useful earlier but inferior later:

| Parent-confirmation reverse switch | Served mixture | Generic64, matched | Generic32, matched |
| --- | ---: | ---: | ---: |
| Frames192-223 | .109960 | .148273 | .100622 |
| Frames224-255 | .081967 | .058197 | .105996 |

This locates remaining selection/weighting headroom, but does not prove exactly
which state transition causes it. The diagnostic does not record the posterior
weights and must not claim it does.

Early recovery also requires better available predictions, not only switching.
At frames128-159 in the same case, all four raw model mean risks at the served
mask exceed .31 (generic64 .385337, parity64 .458347, generic32 .310474,
parity32 .455481). Even the optimistic pointwise raw-plus-neutral envelope has
risk .149240, above the noise floor .0475. That envelope knows the target q and
is not an executable shared-weight policy or a general lower bound on learning.

The other switch still has observation-sensitive error: at frames192-223 of
majority-to-parity, the short parity model scores .089475 at the served mask but
.050837 with all inputs. The served mixture is .121876. A generic-model-only
matched-mask comparison would miss that model-specific observation gap.

Stationary parity remains strong: final-block served Brier .049134, parity64
at the same mask .048232. Replacing all history by aggressively fresh evidence
could sacrifice this strength. Preserve stationary gates in any next experiment.

## Verification

- Race-enabled smoke tests: PASS,3.327s, three scenarios.
- Deliberate future training-origin corruption is rejected.
- Vet: PASS.
- Generation: PASS,22.70s; complete replay: PASS,22.82s.
- All47 source hashes verified, including44 unchanged parent sources.
- Independent completion averaging checks589,824 partial forecasts from1,536
  full-input model profiles; all49,152 parent-step rows covered.
- Generic64 parent predictions match exactly in Go; independent reconstruction
  verifies both additive risk identities and the optimistic convex-envelope bound.
- Summary recomputation is byte-identical.
- Artifact SHA256:
  `8646aec8f8b1f1a9dde4f8d8a6afcb905bb1cc66723567a155521d27baace75e`.

Only research diagnostic code changed. No serving-performance improvement is
claimed; diagnostic runtimes above are not request latency. No production,
private-data, whitepaper or publication changes.

## Next lead

Investigate [bounded snapshot specialists](../../research/snapshot-specialists-proposal.md):
give immutable published predictors distinct identities and forward evidence,
rather than transferring a role's history to a changed model. This is a new
hypothesis, not the demonstrated cause or a validated rescue. It must address
delayed evidence, retirement, gate ownership, fixed resource budgets and fresh
complete comparisons. All seven roadmap directions remain open.
