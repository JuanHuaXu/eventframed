# Prequential stacking V31: completed negative study

2026-10-03. Both frozen overall rescue screens **FAIL**: 20/24 design and
21/24 confirmation cells fail. This is a completed component experiment,
not completion of any of the seven research goals. Production, the
whitepaper, and preexisting tracked changes remain untouched.

## Protocol and Evidence

The [sealed protocol](mmm-stack-v31-protocol.md) and
[preflight](mmm-stack-v31-preflight.md) precede collection. Two fresh seed
bases generate 768 worlds each, with 15 arms per world: 1,536 worlds and
23,040 independently reconstructed arms in total. Each non-baseline arm
receives 32 labels. Primary stacking uses exactly the same nominees,
outcomes, and conditional probabilities as its four stratum controls.

- Design raw SHA256: `9f850266fbf5d6efd8f51a88449269f349db03eb1435336a367b16f5f183fcb4`.
- Confirmation raw SHA256: `187b3e71b2bb8b93ab5475d4edbc58ef88722fd89d266b5a846ec395b1f16fdd`.
- Raw tapes: [design](mmm-stack-v31-design.jsonl),
  [confirmation](mmm-stack-v31-confirmation.jsonl).
- All gate values and eleven sealed source hashes:
  [machine-readable summary](mmm-stack-v31-summary.json).

The new procedure fits convex predictive weights, not Bayesian family
probabilities. Its ridge objective uses original privately retained child
forecasts issued before each outcome; the ordinary child models still
update from their coherent joint models. No label from an unnominated
member enters inference.

## Results

Confirmation means for primary random-within-stratum arms follow. Lower
Brier is better; higher ten-item packet usefulness is better. These are
model-world expected future scores, not observed agent-answer accuracy.

| Case | Affine Brier / usefulness | Partition Brier / usefulness | Blend Brier / usefulness | Stack Brier / usefulness |
| --- | --- | --- | --- | --- |
| Tight curved | .255026 / .608265 | .210116 / .761364 | .232600 / .738597 | .244212 / .735943 |
| Wide curved | .255636 / .611822 | .224003 / .658350 | .254195 / .453546 | .254148 / .474495 |
| Wide independent | .256616 / .672500 | .265600 / .651875 | .280754 / .616250 | .271454 / .636875 |
| Tight reversed | .208875 / .838305 | .221246 / .720168 | .213139 / .815755 | .230196 / .798490 |
| Wide reversed | .211749 / .836644 | .226670 / .690755 | .213541 / .826141 | .231947 / .706493 |

Wide-independent whole-Brier gain over blend is .009300, with the frozen
paired interval [.004099, .014502]. Its mean misses the .01 requirement.
Usefulness gain is .020625 with interval [-.012858, .054108], failing the
positive-lower-bound requirement. Packed bias magnitude upper .268192
exceeds .10. A positive mean is not enough to pass the declared rescue.

Wide-curved stack bias magnitude upper is .550818, also far above .10.
Its mean weights are [.617269, .183570, .199162] for baseline, affine,
and partition; the standalone partition has substantially better Brier.
These observable fitted weights do not exploit the oracle headroom found
in V30. Why they miss it is still a research question, not an inferred bug.

Wide-reversed stack loses .130151 usefulness against affine, with paired
interval [-.200217, -.060085]. This is a clear protection failure under
the frozen screen, not just weak evidence of improvement. Tight/curved
local-usefulness protection also fails despite a positive mean gain,
because its lower endpoint is -.012402 below the -.01 slack.

Only tight/calibrated, tight/baseline_matched, and wide/calibrated pass
every confirmation cell gate. Design additionally passes wide/aligned;
that fourth pass does not survive confirmation. Original thresholds and
all negatives are retained.

The .2-exploration disagreement policy is secondary and not an Anti-Pigeon
certificate. For wide-curved confirmation it yields Brier .253524 and
usefulness .505827, versus stratum .254148/.474495, but packed bias mean
.111365 already exceeds the allowed magnitude before uncertainty.
Entropy yields usefulness .110055 there. These policy means do not
establish an equal-total-cost observation win or an overall rescue.

## Audit and Performance

All of the following pass:

- Independent JavaScript reconstruction of all 23,040 arms, flat child
  likelihoods and forecasts, issued rows, nomination maximum/probability
  forms, final law, packets, metrics, timings, and eleven source hashes.
- Independent edge/interior simplex minimization agrees with Go's
  seven-face KKT solver; unit grid/KKT checks cover degenerate children.
- Exact Go seed replay for BOTH splits, including latent draws, RNG
  decisions, issued rows, forecasts, weights and scores. Only five elapsed
  timing fields per arm are excluded from exact equality.
- Seven in-memory corrupted-tape controls reject changed original rows,
  labels, issued laws, final laws, weights, packets, and source hashes.
  [Negative-control record](mmm-stack-v31-negative-controls.json) preserves
  the original checker and raw files without editing them.
- Hidden unused-label flips, owner/replay/pending validation, out-of-order
  delayed-ticket rows, partial-failure quarantine, three-package tests,
  race tests and vet.

On Apple M4, the final unaccompanied three-repeat microbenchmark records
weight solve 451.3-456.3ns, 150 forecasts 9.422-9.437us, random nomination
160.2-162.7ns, stratum nomination 22.17-22.29ns, entropy 10.538-10.650us,
and disagreement 10.755-10.932us. All these operations allocate zero
bytes; they do not include construction, issued tickets, or child updates.
See the [raw benchmark output](mmm-stack-v31-benchmarks.txt).

Maximum isolated model-arm elapsed time is 1.542083ms design and 1.252084ms
confirmation, below the frozen 10ms cap. These totals include construction,
nomination, issue/resolve and final forecasts, but exclude packet scoring,
retrieval, evidence acquisition, database I/O and loaded queues. Stack's
selection timer includes ticket issuance, unlike retained V30 controls.
They are not daemon serving measurements.

## Interpretation and Next Lead

The current rescue is rejected. Neither correct proper scoring nor a cheap
optimal solver establishes broad learning/calibration performance.
Possible causes remain the ridge anchor, small-sample variability, and
the mismatch between early immature issued-child forecasts and the final
trained child laws. This study does not isolate those causes.

An exact leave-one-out or cross-fitted predictive-score construction is a
distinct research lead motivated by the original stacking paper. Its own
outcome must be excluded from each child fit, it must operate only on
already arrived evidence, and adaptive nomination must not be silently
treated as iid validation. First test mathematical equivalence against
full omitted-member refits and actual costs; then freeze fresh outcome
cohorts. Do not rescue V31 by retuning its consumed seeds or ridge penalty.

These fixed32 known-family worlds do not test changing regimes, loaded
freshness, untouched agent answers, valid Anti-Pigeon error control, or
equal-total-cost acquisition. All seven whole goals remain OPEN.
