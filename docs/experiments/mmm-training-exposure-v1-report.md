# Training exposure: sparse audits stretch the short window

## Finding

The current short count/subset fit takes the last64 eligible audit LABELS, not
the last64 event origins. Audits are sampled at25%; delayed schedules additionally
lose20% of labels and delay the remainder up to31 clocks. This is not a coding
error against the declared implementation. It is a model-timescale hypothesis
that needs an intervention before being called the cause of recovery failure.

A metadata-only reconstruction matches EVERY stored fit clock and origin list
across all256 schedules in the consumed credit-learning tape. Forecast models
and training remain unchanged. Known change origin256 is used only for evaluator
accounting, never passed to a learner or used to authorize a reset.

Delayed changing cases (16 trajectories per row):

| Cohort / case | Mean fit span, events | Mean labels in last64 events | Mean new-evidence fraction during serving | Reaches half new | Reaches all new |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1 / majority to parity | 327.92 | 9.66 | 23.78% | 13/16 | 0/16 |
| 1 / parity to majority | 325.00 | 10.36 | 24.82% | 12/16 | 0/16 |
| 2 / majority to parity | 315.34 | 10.82 | 26.28% | 12/16 | 0/16 |
| 2 / parity to majority | 318.81 | 10.01 | 25.52% | 14/16 | 0/16 |

Fit span is publication clock minus oldest origin in the64-label fit. New-evidence
fraction is measured over the actual model serving each of256 post-change
predictions, then averaged over trajectories. Each row's fit-span/support numbers
average within a trajectory before averaging trajectories. A fit at clock t is
usable only at t+1. Clock511 publications cannot help the evaluated predictions.

Only51/64 delayed changing trajectories reach half-new evidence in the scored
horizon. Among those that do, mean first usable clocks are465.46,467.83,456.58,
458.71 respectively. These conditional means must NOT treat the13 censored
trajectories as having zero latency. Immediate changing schedules all reach
half-new evidence, but only8/64 reach entirely new evidence before the horizon.

Correction to the informal running update: the verified half-new counts imply
**13/64**, not11/64, delayed changing runs never reach half-new evidence. The raw
cell counts above and machine-readable artifact are authoritative.

## What this does and does not establish

This verifies persistent old-regime evidence and a major timescale distinction.
It does not prove all old evidence is harmful, that shorter windows predict
better, or that selection weights are innocent. Current late reverse short
forecasts already beat neutral while forward forecasts do not. A shorter window
may remove useful support, destroy parity learning, or merely shrink to.5.

Replacing64 labels with64 event origins would leave only about10 training labels
in delayed fits. Therefore any quality test must keep stationary controls,
report support and neutral comparisons, and distinguish better probability
calibration from learning the new rule. Reusing all full-monitor observations
as unbiased training data is NOT justified: their selection depends on the
credit/observation policy, unlike the predeclared random audits.

Prior work inspected: short-window experts, window-bank headroom and origin-age
routing proposals. Those include failed shorter-window/selector rescues on an
earlier architecture; they are not erased. The next controlled test isolates
origin-time training support in this sparse-audit architecture, not another
selector decay sweep. See [the test contract](mmm-event-window-v1-contract.md).

## Verification and artifacts

- Reconstruction checks all fit origins for audit status, nonmissing labels and
  arrival before publication, including flush-time fits after the scored horizon.
- Synthetic contracts test simultaneous triggers, single publication per clock,
  missing and nonaudit exclusion, threshold32/cadence16 and invalid arrivals.
- Both diagnostic runs completed in about0.55s and are byte-identical. This is
  metadata-analysis time, NOT a serving or fitting performance result.
- Script `research/training-exposure-diagnostic.mjs`, SHA256
  `6b74a167e5cc1343e4c848e63fde31c98526c7bb68f2b7dd21f07e2a399c0726`.
- Artifact `mmm-training-exposure-v1.json`, SHA256
  `73449685ce8e7044fc66eb81cd080f203378c83552ed591e6bf4353e1fdbbff9`.
- Parent `mmm-credit-learning-v1.jsonl`, SHA256
  `4b148306f6fc1fa0f4e8ae5f1db3b878e3f1fd20b628f305313387a91f9af655`.

Stable cases are included with an artificial origin256 accounting boundary;
they must not be described as true regime changes. No runtime modification,
adoption, fresh confirmation, performance win or completed research goal.
