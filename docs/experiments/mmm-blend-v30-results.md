# Family blend V30: coverage repair, prediction rescue fails

2026-10-03. Both frozen overall screens FAIL:22/24 geometry/regime groups
fail in design and22/24 in untouched confirmation. All seven whole goals
remain OPEN. Production and the whitepaper are untouched.

## Evidence and Reproduction

[Protocol](mmm-blend-v30-protocol.md), [preflight](mmm-blend-v30-preflight.md),
[summary](mmm-blend-v30-summary.json), and the matching raw JSONL tapes.
There are768 worlds and8,448 arms per split;32 worlds per cell and11 arms
per world. All non-baseline arms receive32 distinct selected labels.

- Design SHA256: `75af34ef1567e6947233e96ee43ad2570c30e1b1d8337cf48f8126d987415a2b`.
- Confirmation SHA256: `34ec7503963736b20bf5c324173ec33812b5ed8d90dc3614176ba3fa4bbcef69`.
- Frozen checker SHA256: `cdc0ba6d187ba76bfe742e91419c62955eababb22073e69fd21c9d9b40e3f8ad`.

The independent flat-model audit reconstructs integrated likelihoods,
family weights, pre-label and final predictions, policy maxima and
probability forms, packets and all score fields for16,896 arms. Source
hashes are checked against both manifests. For family-information draws,
the JavaScript checker validates the exploration/exploitation probability
forms and exploitation maximality within numerical tolerance, rather than
regenerating the policy RNG. Exact Go seed replay separately reconstructs
all latent, label and policy draws and every trace field for BOTH splits;
only five timing fields per arm are excluded. This closes the random-draw
reconstruction gap in that independent arithmetic check.

```sh
node research/blend-v30-verify.mjs
EVENTFRAME_BLEND_V30_REPLAY=<LOCAL_ROOT>/docs/experiments/mmm-blend-v30-design.jsonl go test ./internal/researchblend -run '^TestReplayV30$' -count=1 -v
EVENTFRAME_BLEND_V30_REPLAY=<LOCAL_ROOT>/docs/experiments/mmm-blend-v30-confirmation.jsonl go test ./internal/researchblend -run '^TestReplayV30$' -count=1 -v
```

## What Changed and What Failed

Random within-stratum nomination removes V29's deterministic even-only
aliasing. Confirmation primary-arm mean odd nominations ranges14.96875
to16.875 out of32 across cells. Unit tests also cover individual inclusion
frequencies, positive conditional probability in the scheduled stratum,
and eventual exhaustion without duplicates for n=2,11,150,200. This is
finite nomination evidence, not guaranteed coverage of arbitrary external
events, causal distinctions, or source independence.

The primary predictor averages three alternative joint model families,
with prior (.98 baseline, .01 affine, .01 partition). It is ordinary model
averaging under that declared family, not an arbitrary blend of separately
fitted evidence and outcome models. Flat55-state integration agrees with
the implementation in unit tests. Correct model arithmetic does not imply
externally adequate prediction or calibrated packed confidence.

| Confirmation cell | Local Brier | Affine Brier | Partition Brier | Blend Brier | Affine usefulness | Blend usefulness | Blend bias upper |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Tight / curved | .381975 | .248785 | .208461 | .224774 | .667351 | .720752 | .078430 |
| Wide / curved | .298458 | .257376 | .220343 | .251758 | .593200 | .522958 | .495658 |
| Tight / reversed | .396831 | .206743 | .221037 | .207646 | .852114 | .847685 | .043634 |
| Wide / reversed | .390609 | .207191 | .224624 | .210543 | .840436 | .816477 | .054119 |
| Wide / independent | .300064 | .267154 | .278784 | .287013 | .651875 | .599375 | .377189 |

These are matched nomination and potential-outcome tapes, unlike a
cross-run comparison. Brier is expected future Bernoulli loss under the
declared known probability, not empirical agent answer accuracy. The
packet contains the ten highest forecast means. Bias upper is
abs(mean)+3.5SE across32 worlds, with frozen ceiling .10. Intervals are
not anytime certificates.

Tight-curved mean gains versus affine satisfy the nonlinear targets, but
its complete group still fails local packet-usefulness protection because
the paired gain lower endpoint falls below -.01 despite a positive mean.
Wide-curved fails nonlinear gain and calibration. Linear protection and
coordinate-irrelevant controls also fail. In confirmation only both
calibrated geometries pass complete groups; design instead passes
tight-calibrated and wide-baseline-matched. No overall rescue is promoted.

## Consumed-Data Capacity Diagnostic

[Headroom JSON](mmm-blend-v30-headroom.json) and
`research/blend-v30-headroom.mjs` are explicitly post-hoc. The diagnostic
first verifies that every actual primary forecast is its recorded family
weights times the three matched child forecasts. It then solves the
three-member convex-mixture quadratic program using TRUE future rates.
Vertices, edge optima and the feasible interior optimum are evaluated;
degenerate cases and100 independent random grid comparisons pass.

This oracle must never enter inference, nomination, or a new confirmation
claim. On confirmation wide-curved, actual Brier .251758 has oracle
headroom to .217678: paired gap .034080, lower .019443. Actual mean
baseline weight is .473205 versus oracle .052755; the oracle favors
partition weight .857805 versus actual .275882. Tight-curved gap is
.017102, lower .010984; wide-independent gap is .023918, lower .011678.
Thus existing child predictions contain useful mixture headroom on these
consumed fixtures. This does not establish a learnable selector, universal
capacity adequacy, packet improvement, or an effect achievable from32
observed labels without oracle information.

## Bug Hunts and Performance

Three checks cover joint semantics/lifecycle, nomination/no-future/matched
controls, and full generator/score replay. Invalid/duplicate evidence and
caller mutation cannot change a valid law; unexpected child divergence
quarantines the parent. Flat likelihood enumeration and exact family
entropy lookahead pass. Flipping every unselected potential label leaves
each adaptive trace and final law unchanged. Three-module ordinary race
tests and vet pass. This is not a full-daemon integration or concurrency
test, and no new serving code was added.

Apple M4 ordinary-build benchmarks, three200ms repetitions:

| Operation | ns/op | Allocation |
| --- | ---: | --- |
| Construct150-member model | 47,252..47,978 | about76,200 bytes /116 allocations |
| Predict150 members | 8,985..9,024 | 0 |
| Random nomination | 91.06..91.16 | 0 |
| Random within-stratum nomination | 22.18..22.26 | 0 |
| Entropy nomination | 10,363..10,422 | 0 |
| Family-information nomination | 13,351..13,390 | 0 |

Selection benchmarks hold the16-label history fixed; they do not include
update, construction or label acquisition. Cohort maximum model-arm time
is1.190375ms design and .923959ms confirmation, below the isolated10ms
cap. Those timings exclude packet sorting, corpus creation, extraction,
retrieval, storage, queues and full serving. They are not a latency bound
for EventFrame. Fixed32 information comparisons are NOT equal-total-cost
Goal7 validation; no modeled or measured acquisition cost was equated.

## Next Research Lead

Investigate predictive-score-based family weighting, using only forecasts
issued BEFORE each observed outcome, rather than another prior-only sweep
on these consumed tapes. [Yao et al. (2018), predictive stacking](https://sites.stat.columbia.edu/gelman/research/published/stacking.pdf)
motivates proper-score combination under candidate-model misspecification.
An online prequential adaptation is ours, not their LOO/PSIS implementation,
and would not inherit their asymptotic guarantee. Weight fitting and child
updating must remain chronologically separated, with nominal inclusion and
actual computational costs retained. Freeze a fresh nonlinear, calibrated,
linear and coordinate-mismatch study before testing; a useful hindsight
oracle is not permission to tune this cohort.
