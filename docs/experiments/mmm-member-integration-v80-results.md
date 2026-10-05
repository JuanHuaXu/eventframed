# Paired member integration v80 results

FAIL of the pilot forecast-improvement requirement in both phases. The new
evidence trigger advances actual research MMM member revocation substantially,
but does not demonstrate a meaningful improvement in the emitted forecast.
The separate2% false-revocation requirement remains INCONCLUSIVE at this size.

[Protocol](mmm-member-integration-v80-protocol.md),
[artifact](mmm-member-integration-v80.jsonl),
[evaluator](../../research/member-integration-v80-summary.mjs),
[adapter](../../internal/observationgate/paired_investigator.go),
[integration harness](../../internal/observationgate/member_integration_test.go).

640 fresh trajectories,327,680 frames, five paired arms,68 source/evaluator
hashes. The fixed4096-label base is unchanged; future stream seeds are fresh.
Artifact SHA256:
`6755ddcc4711476bd1ae2ba9ce56dd15e5539ef88f247d830e9599057a6ceb4e`.

This uses the actual isolated observation.Run, observationpreserved.State,
ForecastMix, Bayesian Monitor and short/local/pooled fitting paths. It does not
exercise production serving, an LLM or actual chatbot tasks. Its frame space
is the existing nine-bit finite observation experiment.

## Confirmation

64 trajectories per scenario. Post-change Brier and split delays are paired by
stream. Restricted delay charges absent or premature splits at the remaining
horizon. Stable/common/null rows have no splits, not256-step observed delays.

| Scenario | Old gate split delay | Mixture split delay | Old gate post Brier | Mixture post Brier | Old / mixture splits |
| --- | ---: | ---: | ---: | ---: | ---: |
| Stable | No splits | No splits | 0.046954 | 0.046954 | 0 / 0 |
| Member shift | 117.17 | 76.16 | 0.239350 | 0.239205 | 64 / 64 |
| Common shift | No splits | No splits | 0.237862 | 0.237862 | 0 / 0 |
| Recurring | 194.58 | 97.48 | 0.192693 | 0.192584 | 62 / 64 |
| Null | No splits | No splits | 0.251396 | 0.251396 | 0 / 0 |

Member-shift delay improves35.00%, passing the10% delay target, but post-Brier
gain is only0.000145 with the frozen paired z3.5 interval
[-0.000409,0.000699]. It fails both the0.005 gain requirement and positive lower
bound. Design delay improves36.49%, but Brier gain is-0.000158 with interval
[-0.001144,0.000828], also failing. There are no premature splits in either phase.

Member-shift accuracy changes from58.423% to58.282%, a small decrease. Stable
accuracy stays95.111%. Faster split detection must not be presented as improved
accuracy, retained high accuracy under shift, or a calibrated end-to-end rescue.

Stable, common-shift and null predictions are unchanged between gate arms.
No split is observed for these cases, but0/64 has a Wilson95 upper5.66%.
That is insufficient to establish a2% false-revocation frequency. The pilot
size was explicitly declared inadequate for that requirement, not used to
relax it after seeing zero events.

## What the controls show

In member shift, frozen MMM has post Brier0.449445; no-revocation learning
improves it to0.239663; old gate0.239350; new gate0.239205; the local-slot
control0.239113. Most observed recovery comes from the existing learning loop,
not the timing difference between evidence gates. The local-slot control still
retains other mixture experts, so it is not an oracle or an upper bound on
what a better replacement model could achieve.

Common-shift no-revocation learning reaches0.237862 without any paired alarm.
That is consistent with a difference monitor missing common degradation; the
independent learning path remains necessary. The adapter is a paired reliability
investigation trigger, not a general distribution-shift or diameter certificate.

Actual member-shift foreground cost per frame is4.72284 old versus4.72629 new,
within the common six-coordinate cap, not exactly identical realized cost.
Both use the same audit schedule and labels: monitor cost8 coordinates/frame,
shadow audit average4.52362 coordinates/frame, and19.828 fitted models/trajectory.
These extra costs are reported, not hidden in the foreground cap. Null monitoring
costs12 coordinates/frame. Model fitting occurs after the revealing labels and
only influences subsequent predictions.

## Verification and repair

The initial execution stopped on exact floating-point scalar equality, before
writing an artifact. A regression reproduced q=0.249999999999999972 and tiny
nonzero correction coefficients despite identical histories. The adapter now
validates history equality structurally and materializes the exact scalar law
q=.25,m=0,c=0,eta=1. Genuinely unequal histories still fail; evidence updates
once per completed pair. No acceptance threshold or generator changed.

- Twenty random512-step scalar histories pass the rounding regression.
- Duplicate/out-of-order feedback leaves state unchanged and does not increase
  evidence support. Signed divergence and identical-member controls pass.
- Stable/member controls match the archived workflow's Brier, split time and
  acquisition cost. The base model remains unchanged.
- Full deterministic replay passes in28.395 seconds, including source hashes
  and hashes of every prediction, outcome, nomination and gate decision.
- Focused integration/adjacent race tests pass in2.616 seconds; package vet passes.
- Experiment wall time28.356 seconds is for the entire multi-arm fit/evaluation
  job, not a per-request latency measurement.

## Next investigation

Retain the measured trigger improvement but do not call the integrated rescue
successful. Inspect post-split expert forecasts, influence weights, observed
coordinates and audit support to distinguish an inadequate local model from
insufficient mixture influence or a stale observation policy. These correspond
to roadmap directions1/2/4, now tested at the actual research integration boundary.
Test any resulting intervention on fresh streams with the same foreground/audit
budgets and untouched stationary controls. Broad false-revocation coverage,
delayed/missing evidence and actual agent tasks remain open.
