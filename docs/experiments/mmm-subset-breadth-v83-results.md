# Retained subset breadth v83 results

PASS of the frozen breadth screen in both phases. The retained candidate's
benefit extends beyond the one-variable change in v82 to the tested interaction
and dependent-input cases. This is conditional on the declared finite generators
and fitted bases, not a universal generalization or production-readiness result.

[Protocol](mmm-subset-breadth-v83-protocol.md),
[artifact](mmm-subset-breadth-v83.jsonl),
[evaluator](../../research/subset-breadth-v83-summary.mjs),
[generator and tests](../../internal/observationgate/subset_breadth_test.go).

1,280 fresh trajectories,655,360 frames, ten cases, six paired arms,116 frozen
source/evaluator hashes. No change to v82's candidate, subset prior, input model,
weights, audit budget, fitting cadence or split authorization. Each case uses
a new fixed4096-label initial-regime base. Stream uncertainty is conditional on
that fitted base; it does not estimate uncertainty over all possible fits.

Artifact SHA256:
`837b964e3178d7c43864396a98fef48ff1f01d9a71e7c3da6e80aaa28307294d`.

## Confirmation

64 trajectories per case. Post-change metrics compare the mixture-gate count
control with the retained subset candidate on the same frames/audit labels.

| Case | Control Brier | Candidate Brier | Control accuracy | Candidate accuracy | Paired Brier-gain lower bound |
| --- | ---: | ---: | ---: | ---: | ---: |
| Bit, uniform | 0.240375 | 0.204525 | 58.25% | 66.90% | 0.029189 |
| XOR2, uniform | 0.239038 | 0.211322 | 58.43% | 65.66% | 0.022355 |
| Majority3, uniform | 0.238308 | 0.226852 | 58.33% | 61.76% | 0.005831 |
| Multiplexer, uniform | 0.236141 | 0.223278 | 59.81% | 62.91% | 0.008541 |
| Parity4, uniform | 0.255851 | 0.240517 | 52.51% | 57.34% | 0.009066 |
| Bit, dependent | 0.212785 | 0.198227 | 64.29% | 68.03% | 0.010171 |
| XOR2, dependent | 0.213276 | 0.206808 | 64.97% | 66.55% | 0.002489 |
| Stable, uniform | 0.048807 | 0.048799 | 94.87% | 94.87% | -0.000006 |
| Stable, dependent | 0.046865 | 0.046851 | 95.10% | 95.10% | -0.000010 |
| Null, uniform | 0.251348 | 0.251384 | 50.03% | 49.84% | -0.000228 |

All seven changed cases exceed the frozen0.005 mean Brier gain with positive
paired z3.5 lower bounds. All stable/null full and post harm bounds remain below
0.01. The design phase also passes every case. The narrowest mean-gain pass is
dependent XOR:0.005482 in design and0.006468 in confirmation. Its lower bound
is positive but below0.005; the predeclared rule requires the MEAN gain>=.005
and lower bound>0, not a lower bound>=.005.

These are paired trajectory screens, not simultaneous nonparametric guarantees
over arbitrary generators. No previous failed subset-model experiment is
retroactively reclassified by this new retained-state result.

## Limits visible in the data

Parity4 improves materially relative to control but still reaches only57.34%
post-change accuracy. Majority and multiplexer recovery also remain far below
their low-noise attainable accuracy. Passing a relative improvement requirement
does not establish strong absolute reasoning or rapid complete recovery.

The dependent cases copy bit2 to bit0 and bit1 to bit4. They are two specific
perfect-dependence structures, not a broad sampling of real-world dependence.
The candidate's uniform-input model remains misspecified; better scores here
do not make its posterior or partial-input probabilities generally calibrated.
The count path and incumbent are retained, which distinguishes this composition
from replacing all forecasts with the subset model alone.

No delay, missing labels, asynchronous publication, input corruption, source
spoofing or actual chatbot task is exercised in this study. The observed
stationary protection does not certify a2% false-revocation rate.

## Resource accounting

Foreground coordinate caps remain6 and all candidate labels come from the
same audits as control. Realized acquisition cost is not always lower: the
dependent-bit case increases from4.62799 to4.65796 coordinates/frame. Uniform
XOR decreases4.74371 to4.61871, and uniform parity4 decreases4.66962 to4.63199.
Audit/monitor counts and additional subset fits are recorded in the artifact.

The new candidate's compute code is unchanged from v82. This study does not
remeasure per-request latency or turn v82's microbenchmarks into bounds over
every case. The full multi-arm experiment takes113.951 seconds on the local
machine; that includes model fitting and is not foreground serving latency.
The nine-dimensional subset enumeration remains bounded here and exponential
if naively generalized to unrestricted dimensions.

## Verification

- Full truth tables for all rules, dependency mapping and relevant-coordinate
  bounds pass. Every new rule can be observed with at most5 coordinates under
  the existing view structure; no impossible-budget target is hidden in the set.
- Train/stream/role seed separation is checked. Configuration, training seed
  and stream base are stored in each record and verified by the evaluator.
- Candidate split times, audit cadence and training-label budgets match the
  control's records. Every forecast precedes its revealing label and refits
  occur afterward, using the unchanged v82 state.
- Full replay of1,280 trajectories passes in115.144 seconds with the frozen
  source manifest and exact outcome/prediction records. New unrelated files
  do not redefine the artifact manifest.
- Focused breadth/state/integration race tests pass in2.479 seconds; vet passes.

## Next boundary

Proceed to delayed and missing feedback without retuning the candidate on these
confirmation results. The current single-pending-forecast state deliberately
requires feedback before another prediction; it cannot honestly simulate a
nonblocking delayed stream by inventing labels or silently discarding ordering.
Introduce an explicit bounded pending-forecast journal, preserve the forecast
actually emitted, and use only available feedback for fitting and weight updates.
Require immediate-feedback parity and duplicate/censoring isolation before a
fresh delayed/missing-data comparison. All seven roadmap directions remain open.
