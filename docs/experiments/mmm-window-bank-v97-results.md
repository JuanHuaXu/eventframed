# v97: retained and short-window forecast experts

## Verdict: FAIL, 102/106 gates

All48 whole-stream non-harm,46/48 late non-harm,6/6 stationary interaction-gain
and2/4 recovery-gain gates pass. Both phases still fail late parity-to-majority
non-harm and positive recovery. Do not call the increased gate count a completed
rescue or discard the remaining failures.

The [frozen protocol](mmm-window-bank-v97-protocol.md) adds generic32 and
Boolean32 alongside retained64-frame models. Four-expert Brier aggregation
uses prior[.95,.05/3,.05/3,.05/3], eta.5 and fixed share.001. All six v96 arms
remain comparisons, with nine arms reported in total.768 fresh streams cover
two phases,12 cases,32 paired streams per case and256 scored steps.

Every arm and view receives identical full-frame training evidence. Refits
remain on the32-step schedule; only already observed samples enter windows.
Online weights persist across refits. The bank never receives the simulator's
change point or probability. Raw results include every publication's window
counts, origin ranges and pre-outcome weights, checked by the evaluator.

See [raw records](mmm-window-bank-v97.json),
[all gates and publication summaries](mmm-window-bank-v97-summary.json), and
[independent evaluator](../../research/window-bank-v97-summary.mjs).
Intervals use the frozen paired32-trajectory z=3.5 normal approximation,
not exact coverage or anytime confidence sequences.

## Same-stream confirmation results

| Full-view case and segment | Generic64 Brier | Old fixed-share | New bank |
|---|---:|---:|---:|
| Parity4, all256 | 0.097235 | 0.072824 | 0.076174 |
| Parity4, late128 | 0.073149 | 0.051627 | 0.053592 |
| Majority to parity, late128 | 0.204661 | 0.202585 | 0.198229 |
| Parity to majority, late128 | 0.170529 | 0.182771 | 0.177947 |

Parity4 bank gain is0.021061, interval[0.015688,0.026434], and late expected
accuracy is94.98%. All6 stationary interaction-gain gates pass. This recovers
the criterion lost by interval-only weighting, but the fixed-share control
still has better stationary Brier on these same samples.

Majority-to-parity bank gain is0.006432, interval[0.003738,0.009126], passing.
Parity-to-majority bank gain is-0.007418, interval[-0.011314,-0.003521]. Its
mean harm is below0.01, but the upper harm bound exceeds0.01, so non-harm is
not established. Design also fails, with upper harm0.010669. These four failures
are preserved, not explained away by improved whole-stream averages.

## What the new evidence suggests

The standalone generic32 model has better late parity-to-majority Brier:
0.153458 versus generic64's0.170529. At step160 it uses only origins128..159,
but mean bank weight is0.005532 (0.55%). At step192 it is0.036797 (3.68%).
This supports investigating delayed selection of a useful short-window model;
it does not prove the best weights at every step or a deployable rescue.

Blindly boosting that model is unsafe: on stationary parity4 its late Brier
is0.149902, much worse than generic64's0.073149. The Boolean32 model instead
performs well on parity and helps the other switch direction. The selector
needs evidence-sensitive model choice, not an unconditional preference for
shorter history. All these observations are now consumed evidence.

Next perform the [read-only forecast-headroom diagnostic](../../research/window-bank-headroom-proposal.md)
before choosing another intervention. Oracle convex-hull and constant-pair
comparisons will be labeled retrospective upper limits on possible gain,
never an implementable predictor or new confirmation result.

## Verification

- Explicit normalization, singleton/equal experts, two-expert predecessor,
  duplicated-challenger equivalence and lifecycle race tests PASS.
- Combined race tests and learned-stream smoke PASS:3.495s package time.
- Generation PASS:123.953s; exact768-record replay PASS:125.803s.
- All bank and retained-control prefix-bound defects are zero in the artifacts.
- Vet, source/summary reproduction and new-file whitespace checks PASS.
- Artifact SHA256:6b4559a63d7b6c1e1b2603513a90dcf4c82281b27e48430405522a1fb48380e5.

The bank comparator bound uses its actual normalized generic prior. It applies
to cumulative realized complete-feedback loss, not arbitrary recent intervals,
pointwise risk, omitted labels or delayed feedback. The interval control keeps
its separate semantics. No production or same-instance concurrent API is added.

## Performance

Apple M4,Go1.27.1 darwin/arm64,GOMAXPROCS10. The four-expert predict/update
kernel takes159.9-160.4ns with zero allocations across three500ms repetitions.
See [kernel timing](mmm-window-bank-v97-kernel-benchmarks.txt).

The [entire research fixture](mmm-window-bank-v97-benchmarks.txt) takes
162.03-183.10ms per256-step stream and allocates21.52-21.54MB,2067-2189 objects.
It includes eight refit rounds, all five constructed models (including duplicated
fitting inside the BMA control), nine arms, two views, scoring and hashing.
The first slower repetition is retained. These are not per-request serving
latencies or a controlled performance delta against v96's different sample.
Extra model construction is measured rather than hidden behind the cheap
bank update. The full seven-direction research goal remains open.
