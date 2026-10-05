# V3 retrospective factorial: model mass drives delay; schedule also matters

This [predeclared retrospective diagnostic](mmm-learned-contrast-v3-factorial-protocol.md)
crosses the 11-rule and null-augmented 12-rule forecasters with the two
learned-selector label schedules on **already consumed** v3 design and
confirmation tapes. It is not a new confirmation run. Both original diagonal
cells reproduce every archived pre-feedback forecast, delivery and metric;
crossed cells are offline interventions using the archived potential outcomes,
not independently deployable online policies.

## Confirmation decomposition

The four cells are `old model / old schedule`, `old / new`, `new / old`, and
`new / new`. A positive delay difference means slower recovery. The intervals
below are mean +/- 3.5 standard errors over 16 independently fitted baseline
clusters, descriptive rather than simultaneous coverage.

| Shift | Delay added by new model on old schedule | Delay added by new schedule to old model | Delay added by new schedule to new model |
| --- | ---: | ---: | ---: |
| Bit2 immediate | **14.74** [11.47, 18.01] | 6.15 [3.31, 8.99] | 0.63 [-3.77, 5.03] |
| Bit2 delayed | **12.23** [8.20, 16.26] | 0.36 [-3.52, 4.24] | -0.16 [-4.55, 4.22] |
| Bit0 immediate | **16.43** [12.90, 19.96] | 8.61 [4.47, 12.74] | 3.36 [-1.46, 8.19] |

For recovery *time*, the null-augmented working forecast is the larger
measured penalty on a fixed old evidence schedule. On delayed bit2 the
schedule substitution is near zero while the model substitution adds
12.23 clocks. The model and schedule effects interact, so their listed
means are not generally additive.

Average proper-score effects are different. On immediate bit2, replacing
the model on the old schedule worsens post **expected** Brier by .00390
[.00149, .00630], while replacing the schedule under the old model worsens
it by .00763 [.00376, .01150]. Thus the schedule is not exonerated for
average score, even though model mass dominates the recovery-delay metric.
On null tapes, the new model improves post expected Brier by .12820
[.12356, .13285] on the *same old schedule*; on the out-of-family majority
case it improves by .01672 [.01368, .01975]. Those are model-family effects,
not better evidence acquisition. The majority recovery delay remains near
the 256-clock cap under every factorial cell.

Design shows the same qualitative pattern: fixed-old-schedule model delay
is 14.66 immediate bit2, 12.79 delayed bit2 and 14.91 bit0; corresponding
old-model schedule delays are 5.93, 1.40 and 5.92 clocks. No failed cell
was removed from the diagnostic.

## Reproducibility and boundary

- [Factorial records](mmm-learned-contrast-v3-factorial.jsonl): 3,584
  trajectories and four cells each, SHA256
  `be425455db068b33b1a20fcc6d74f6d0204d2041d35d33330a481ad87033e83f`.
- The [Go replay](../../internal/observationgate/learned_contrast_v3_factorial_test.go)
  checks the two original diagonal forecasts and due labels at every clock.
  The [independent summary](../../research/learned-contrast-v3-factorial-summary.mjs)
  verifies archived input/source hashes, diagonal metrics, all 128-request
  budgets, unique trial identities and fit-cluster contrasts. A full `-race`
  recollection is byte-identical to the ordinary 3,584-row output; `go vet`
  passes. The ordinary replay took about 8.2 s on this host, including
  decompression and refitting; it is not a serving benchmark.

This supports a sharper next hypothesis: an evidence-gated null *forecast*
may retain the calibration benefit without suppressing known-rule recovery.
The retrospective data cannot select a threshold or validate that rescue.
Freeze a fresh design and confirmation cohort, preserve null and majority
controls, and keep recovery delay alongside Brier before changing any
serving path. All seven whole research goals remain open; production was
untouched.
