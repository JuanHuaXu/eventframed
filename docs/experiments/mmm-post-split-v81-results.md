# Post-split forecast diagnostic v81 results

Finding: post-shift conditional-count sparsity is a supported bottleneck.
Mixture inertia and early observation choices also contribute; neither alone
explains the large late gap between full/actual and reduced-mask forecasts.
This post-hoc diagnostic does not rescue or relabel v80's failed integration.

[Protocol](mmm-post-split-v81-protocol.md),
[artifact](mmm-post-split-v81.jsonl),
[summarizer](../../research/post-split-v81-summary.mjs),
[diagnostic code](../../internal/observationgate/forecast_diagnostic_test.go).

All640 v80 trajectories are replayed with read-only hooks. Every original
per-stream metric and prediction/outcome hash matches exactly. Diagnostic
probabilities are frozen before the revealing label; none enters the actual
forecast, audit policy, mixture update or fitted model.

Artifact SHA256:
`6a8c4b4cc3764ede16295babb3c0dd568f928948e373a4754e898252b2b71faf`.

## Member-shift confirmation

Each row summarizes4096 predictions,64 steps in64 trajectories. The generator
changes at step256 from parity of coordinates6/7/8 to coordinate2. The bit2
diagnostic therefore uses generator knowledge, not a learned feature selector.

| Steps | Split fraction | Bit2 observed | Neutral expert weight | Emitted Brier | Short actual-mask Brier | Short bit2-only Brier |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 256-319 | 8.72% | 61.43% | 72.04% | 0.27821 | 0.28892 | 0.24425 |
| 320-383 | 71.41% | 61.38% | 87.57% | 0.24952 | 0.27400 | 0.16903 |
| 384-447 | 99.32% | 75.51% | 70.85% | 0.23813 | 0.23701 | 0.10714 |
| 448-511 | 100% | 97.02% | 26.58% | 0.19096 | 0.17927 | 0.06839 |

Weights are the journaled selector weights before the tiny fixed-share update,
not claimed to be the exact effective prediction weights. The neutral expert
predicts0.5. Early after the shift, actual-mask challengers are worse than that
neutral forecast, so simply forcing more influence onto them would not be a
justified rescue. Later mixture inertia has a cost, but it is much smaller than
the observed short-model conditional-mask gap.

In the last window, the short model guides observation93.51% of the time and
the relevant bit is visible97.02% of the time. Yet its actual-mask Brier0.17927
is substantially worse than its bit2-only0.06839. Thus missing the feature is
not sufficient to explain the late failure. Conditioning on irrelevant observed
coordinates partitions the limited64-label sample into many small cells.

## More context is not automatically better

Last member-shift window, same available models and outcomes:

| Model | Mean training support | Actual-mask Brier | All9 Brier | Bit2-only Brier | Old-parity-only Brier |
| --- | ---: | ---: | ---: | ---: | ---: |
| Frozen base | 4096 | 0.26837 | 0.38029 | 0.25019 | 0.45272 |
| Short | 64 | 0.17927 | 0.24039 | 0.06839 | 0.27611 |
| Local | 113.31 | 0.20009 | 0.24341 | 0.12041 | 0.31001 |
| Pooled | 226.34 | 0.22292 | 0.24875 | 0.17584 | 0.36332 |

The local and pooled models also retain more pre-change observations. More
training support therefore does not mean more relevant current-regime evidence.
All-coordinate conditioning is not an oracle: with sparse tables it can be
worse than partial conditioning. These counterfactual forecasts have extra or
generator-selected information and are not equal-cost deployable competitors.

The pattern is not confined to member revocation. In the last common-shift
window, short actual-mask Brier is0.18251 versus0.06578 on bit2 alone. This
occurs without any split. Conversely, on stable data, the frozen model correctly
observes all parity coordinates and has late actual-mask Brier0.04618, whereas
bit2 alone gives0.24999. Hard-coding the new bit would destroy the stable case.

## Verification

- Every actual-mask expert is reconstructed from its immutable pre-label
  model and compared to the journaled expert probability.
- Missing models have availability counts, not silently included0.5 scores.
- Preparation/feedback ordering, duplicate feedback, window counts and scored
  law parity are tested. Diagnostic contract race test passes in1.693 seconds.
- Full deterministic replay passes in28.457 seconds. All original v80 source
  hashes, records and prediction tapes match. Package vet passes.
- Summarizer verifies diagnostic source hashes, availability, weight sums and
  agreement with the original emitted-law Brier. No performance gain is claimed.

## Next intervention

Test learned marginalization of irrelevant detail, retaining the original
conditional-count model as a control. The repository already has a bounded
subset-model averaging implementation in observationlearners/subset.go; it
learns feature-subset weights from past labels rather than receiving bit2.
Its earlier broader-generator failures remain relevant and must not disappear
when moving it into this integration.

First integrate a retained subset challenger on the same audit labels and
observation budget, preserving the incumbent and count-model path. Assess
both actual observed-context forecasts and resulting observation choices on
fresh streams. Then vary generator, input dependence, noise and delays: matched
uniform-input results cannot validate a general partial-observation model.
Do not fix this by feeding diagnostic masks or increasing confidence by fiat.
All seven roadmap directions remain open.
