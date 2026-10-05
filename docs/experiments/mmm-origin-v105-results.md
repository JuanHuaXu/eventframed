# Origin timing diagnosis v105

Consumed-data diagnosis, not a new confirmation or rescue pass. All128 delayed
switch trajectories from v104 were replayed with read-only traces: both phases,
both directions and all32 indices. The complete original records remain
bit-identical. No individual adverse trajectories were selected afterward.

## Evidence

- [Diagnostic protocol](../../research/origin-v105-protocol.md)
- [Trace artifact](mmm-origin-v105.json)
- [Independent reconstruction](mmm-origin-v105-summary.json)
- [Tracing code](../../internal/observationlearners/origin_v105_test.go)

There are26,230 selector-update records. All traces replay exactly, all24 source
hashes match, and the summary reproduces byte for byte. Original probabilities,
acquired masks, scores, fit origins and journal counts match v104 completely.
The evaluator reconstructs fixed-share updates from original losses and checks
weight continuity, arrival/release clocks and diagnostic risk. Race-enabled
profile/reference and trace smoke tests pass; vet passes.

Generation took16.90 seconds; replay16.97 seconds. Artifact SHA-256:
`eba90d4ff17c299a5aec11c42ab4ef7426de979a65da2b9c983e6058f69fc060`.
Parent v104 SHA-256:
`73003cee91b7c7c8b16e6b859391fbeb7da3deb39e92d1bec118175f09f8f760`.

## What the clocks reveal

Confirmation, updates released after the change but before the final forecast:

| Case | Origin regime | Bank-only updates | Mean age at release | Mean arrival delay | Extra prefix wait |
| --- | --- | ---: | ---: | ---: | ---: |
| Majority to parity | Before change | 719 | 28.34 | 14.98 | 13.36 |
| Majority to parity | After change | 2,095 | 28.64 | 15.48 | 13.16 |
| Parity to majority | Before change | 689 | 28.64 | 15.93 | 12.72 |
| Parity to majority | After change | 2,096 | 28.81 | 15.87 | 12.94 |

The waiting queue adds roughly13 ticks beyond arrival. Old- and new-regime
losses have similar ages at release. This is why simply treating every late loss
as unreliable could discard useful new-regime evidence as well.

## Diagnostic risk, not served-law attribution

At each scored release clock, the trace evaluates the current four fitted
models over all512 uniform inputs against the current simulator regime. It
compares raw-bank mixture Brier before and after that one selector update.
The profile is a quadratic form independently checked against direct Brier
enumeration. Oracle truth is used only in this diagnostic, never in learning.

This is neither the routed-gate forecast nor its partially observed served law.
It is not a counterfactual rerun, and sums of these changes are not the measured
v104 Brier regression. No uncertainty or causal attribution claim is made from
the correlated update counts below.

Confirmation bank-only updates after the change:

| Case | Origin regime | Updates increasing raw risk | Mean raw-risk change per update |
| --- | --- | ---: | ---: |
| Majority to parity | Before change | 593/719 | +.00002256 |
| Majority to parity | After change | 972/2,095 | -.00008659 |
| Parity to majority | Before change | 576/689 | +.00007640 |
| Parity to majority | After change | 1,052/2,096 | -.00004260 |

Both design groups have the same signs. Pre-change losses usually move the
current raw mixture in the wrong direction; post-change losses improve it on
average. This supports separating unnecessary prefix waiting from genuinely
old evidence. It does not quantify how much of the served regression comes
from weighting, changed acquisition, or the evidence gate.

## Next action

Prioritize an arrival-time selector with the original ordered, version-scoped
gate. It removes the extra queue wait without first discounting all delayed
losses. The [component implementation](../../research/arrival-routing-component-results.md)
passes lifecycle tests but has no fresh quality result yet. Keep clock-based
forgetting and origin-age discounting as separate subsequent ablations rather
than changing three mechanisms at once.

V104 remains FAIL. All seven directions remain open. No production changes,
whitepaper edits, commits or pushes were made.
