# Interaction diagnostic v89 results

Status: **post-hoc diagnosis, not a rescue or confirmation**. All2,560 v88
records reproduce with unchanged forecasts, updates, latent streams and split
decisions. The parent breadth failure remains FAIL. These records represent
1,280 underlying fitted-base/stream pairs under two feedback schedules.

## Artifacts

- [Protocol](mmm-interaction-diagnostic-v89-protocol.md),
  [records](mmm-interaction-diagnostic-v89.jsonl),
  [summary](mmm-interaction-diagnostic-v89-summary.json).
- Diagnostic SHA-256:
  `16e4db9e074511feb9b7d7b2fd9abd91b1650d2e309bdddda49944d61bea627d`.
- Parent SHA-256:
  `4d97446e78948d8544c218c9e195e8d446ea5cb086c63426a7c79c54a5e14dc0`.
- The manifest pins132 source/protocol files. The post-hoc evaluator records
  its own hash and checks source hashes, parent-record equality, model
  availability, window bounds and original mixture-score reconstruction.

## Parity has more than an observation problem

Confirmation parity4, final window448..511 (4,096 forecasts per schedule):

| Metric | Immediate | Combined jitter/missing |
| --- | ---: | ---: |
| Generating-variable coverage | 82.18% | 41.99% |
| Emitted mixture Brier | 0.156039 | 0.239783 |
| Hypothetical full-input mixture Brier | 0.154411 | 0.239389 |
| Age model actual-mask Brier | 0.156302 | 0.233921 |
| Age model full-input Brier | 0.147008 | 0.222997 |
| Retained model actual-mask Brier | 0.151391 | 0.240285 |
| Retained model full-input Brier | 0.140889 | 0.235768 |
| Outer short weight | 75.47% | 17.31% |
| Outer neutral weight | 21.96% | 79.06% |

The stressed age model is available for4,014 of4,096 forecasts; its scores
are conditional on availability, not silently filled with neutral scores.
Under immediate feedback it is available for all4,096. Weights are the recorded
pre-fixed-share weights. Full-input predictions keep the same fitted models and
weights and integrate no new learning, so they are not deployment gains.

Coverage is low, but the model still scores0.222997 even when given every
coordinate. Actual-to-full age Brier improves only0.010924 in that late window.
The full-mixture difference is smaller still, but this does not prove observation
irrelevance: other components and their weights can offset individual gains.
Model quality, observation choice and weighting remain coupled.

The earlier windows explain why forcibly promoting the age model is unsafe.
Under combined stress:

| Window | Coverage | Age actual Brier | Age full Brier | Mixture Brier |
| --- | ---: | ---: | ---: | ---: |
| 256..319 | 0.49% | 0.301203 | 0.303999 | 0.347043 |
| 320..383 | 5.66% | 0.271829 | 0.275323 | 0.251616 |
| 384..447 | 24.49% | 0.245415 | 0.238005 | 0.248449 |
| 448..511 | 41.99% | 0.233921 | 0.222997 | 0.239783 |

In the first two windows it is worse than neutral even with full inputs. The
diagnostic therefore does not support an attention-only or weight-only rescue.
Design-phase late parity has the same qualitative pattern: coverage45.61%, age
full Brier0.215973, and short weight15.34%.

## Adjacent controls

Late confirmation combined-feedback XOR2 has98.05% generating-variable coverage,
age actual/full Brier0.086071/0.083261 and short weight82.41%. Single-bit coverage
is97.27%, age actual/full Brier0.066995/0.064295 and short weight83.61%.
This contrasts with parity4 and supports investigating interaction sample
efficiency, not a generic failure of every observation path.

Stationary actual/full mixture Brier is0.051683/0.066811; revealing extra inputs
can worsen the fitted predictor. Full inputs are not an optimal oracle law.
All cases, including dependent inputs and null, remain in the summary.

The coverage field measures the syntactic set of variables used by the
generator. For dependent inputs, one visible variable can imply another. For
majority or multiplexers, a particular assignment can be resolved without all
generating variables. Thus this coverage is not a universal sufficiency or
retrieval-recall metric. Uniform parity does require all four independent bits
to remove its residual parity uncertainty.

## Research decision

Investigate a more compact interaction model before changing the observer:
[provisional Boolean specialist](../../research/boolean-specialist-proposal.md).
It specifies a coherent joint likelihood/predictive model over all Boolean
subsets in the bounded9-bit domain, with no privileged test mask. Its rationale
uses Boolean Fourier structure; its bounds and non-universality are explicit.
This is a proposal to test, not an implemented or validated fix.

The next experiment must keep generic learners and controls, use unseen subsets
and permutations, preserve the feedback protocol, and charge added compute.
If learning improves but input coverage remains limiting, test joint-lookahead
observation separately. Do not claim the current greedy observer is a confirmed
code bug merely from its use of one-step information gain.

## Verification

- Original-record parity generation: PASS,269.898 seconds.
- Ordered preparation, original inner/outer forecast reconstruction, score
  parity and bounds: PASS under focused race tests,2.470 seconds.
- `go vet` and independent summary/source checks: PASS.
- Full deterministic diagnostic replay: PASS,270.384 seconds. All2,560
  original records and diagnostic windows reproduced. The prior replay's
  completion output was unavailable after context recovery, so this result is
  from a fresh explicit replay, not an inferred completion.
- Repeated focused race check: PASS,2.462 seconds; vet, independent summary
  recomputation/source checks and repository diff whitespace check pass.

No runtime learner or observer changed, no production/OpenClaw changes, and no
commit or push. All seven research directions remain open.
