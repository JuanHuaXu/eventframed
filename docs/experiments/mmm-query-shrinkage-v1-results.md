# Frozen mass-shrinkage rescue

## Verdict

FAIL. Both training fits choose alpha=1, reproducing the original joint-value
selector. Global shrinkage is not supported by the frozen squared-loss fit.
Neutral and direct-teacher-weight heuristics also fail the primary screen.
All12 primary/supplementary screens fail; none has a positive lower gain bound
against both controls in any phase1 delayed case.

The fitted numerator/denominator are55.022831/50.431227 for binary-label targets
and53.832357/50.431227 for teacher-probability targets. Both unconstrained ratios
exceed1, so the declared [0,1] constraint returns1. Do not widen the interval or
tune on phase1 to obtain a favorable result.

Training uses4640 candidate labels across672 phase0 delayed pools, pool weight672.
These are archived generated binary labels, including labels not naturally
delivered in the source stream. They are offline synthetic supervision, not a
claim of naturally observed training feedback. The teacher-target fit is a
separate diagnostic. Offline supervision and feature-production costs are not
included in the online paid-query count; equal online counts alone cannot
establish goal7's full equal-total-acquisition-cost requirement.

Phase1 delayed means,672 records and672 paid queries per non-baseline arm:

| Rule | Actual-answer sampled Brier | Population expected Brier |
| --- | ---: | ---: |
| No query | .168800035 | .168728804 |
| Random | .167334488 | .167203079 |
| Entropy | .166629983 | .166794264 |
| Original joint | .167365847 | .166847695 |
| Binary-label fitted weights | .167365847 | .166847695 |
| Teacher-target fitted weights | .167365847 | .166847695 |
| Neutral weights | .167659920 | .166990283 |
| Direct teacher weights, oracle | .168055748 | .167163791 |

Raw/model-fitted candidate mass MSE is .045457076 versus .131128515 for neutral.
Thus the mass diagnostic does not justify calling the entire posterior globally
overconfident. Even direct teacher answer weights fail when paired with the
model's conditional-response usefulness score. Probability correction alone is
insufficient for the current acquisition heuristic.

## Verification and Boundary

[Protocol](mmm-query-shrinkage-protocol.md),
[experiment](../../research/query-shrinkage-experiment.mjs),
[summary](mmm-query-shrinkage-v1.json) and
[replay](mmm-query-shrinkage-v1-replay.json). Replay is byte-identical. All2688
records and84 cells are retained; original alpha1 choices reproduce exactly,
phase1-label poisoning leaves fits unchanged, and unit tests cover convex fit,
input ownership, degenerate/boundary cases and oracle-access traps. A unit test
caught roundoff in the alpha1 identity; that endpoint now returns p unchanged.

The [independent audit](mmm-query-weight-v1-audit.json) rebuilds calibration from
the original data, uses a second-moment expression for mixture scores, verifies
all selections, loss/interval/screen arithmetic, and checks94080 saved loss
values across this experiment and the preceding mass diagnostic. Its calibration
objective grid includes1001 points. Every input and relevant source is hashed.

[Component timing](mmm-query-shrinkage-v1-benchmark.json):2016 selections with
precomputed conditional predictions, median .000500ms and p99 .001666ms. This is
a tiny heuristic timing, not a serving benchmark; Bayesian fitting, retrieval,
feature generation, I/O and contention are excluded.

Reweighting conditional responses generally changes their implied marginal.
This experiment changes query order only and reuses the existing properly-scored
post-answer forecast branches; it does not claim the new mixture score certifies
risk reduction for the incumbent law. All seven goals remain open.

Next isolate usefulness-estimation errors: target-probe coverage and changes in
natural evidence between decision160 and publication161 are distinct remaining
mechanisms. Preserve matched controls and test them separately before adding
new critic capacity. No production, paper, commit or push changes.
