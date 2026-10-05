# Kalman early-recovery v2: frozen component protocol

This isolated Goal 2/4 study tests whether larger process noise repairs the
early-recovery failure in the consumed v1 study. It is not a serving change or
an exact Bernoulli posterior. The v1 records and implementation remain intact.

## Split and arms

- Use the existing nine-bit `observationlearners` generator, 512 ticks per
  stream, and scenarios `stable05`, `shift128`, `gradual`, `delayed_missing`,
  and `interaction`. Fit four independent frozen baselines per scenario using
  fit indices 200..203 for design and 300..303 for confirmation. Run eight
  independent streams per fit. Stream seed bases are 2026102101 and
  2026102102; neither appears in v1.
- Every arm sees the same context, potential outcome, 25% outcome-blind audit,
  missingness, and delay tape. Forecast before generating the current outcome;
  deliver due labels only after that forecast. Only arrived audited labels
  update the last-64 model or the filters. Delayed filter updates use origin
  features at delivery time and remain approximate.
- Arms: frozen baseline; last-64 audited-label refit after the 32nd audit and
  every 16 audits thereafter; and three 11-coordinate residual filters with
  process noise `q` in {0.0005, 0.002, 0.01}. All filters start with zero
  mean/identity covariance, use observation variance 0.25, identical fixed
  features and [0.01,0.99] forecast clipping. No per-case tuning or reset.

## Frozen decision

Report paired mean Brier for full512, first64 after change, and last128;
mean-minus/plus 3.5 standard errors over 32 independent trajectories per
case, plus audit counts, invalid states and isolated per-update p99. A candidate
`q` passes only if **both** splits meet all of the following:

1. On `shift128` and `delayed_missing`, first64 Brier improves by at least
   0.01 over `q=0.0005`, with a positive paired lower endpoint, and is no
   worse than last-64 by more than 0.005.
2. On `shift128`, `gradual`, and `delayed_missing`, last128 Brier is no worse
   than last-64 by more than 0.005.
3. On `stable05`, full512 Brier harm against the frozen baseline has paired
   upper endpoint below 0.01. On `interaction`, last128 harm against last-64
   has upper endpoint below 0.01. No numerical state failure is allowed.
4. Isolated update p99 is below 50 us and state remains at most 4 KiB.

The two larger-q candidates are evaluated together; selecting whichever passes
after viewing design is not permitted. If neither passes, preserve both splits
as negative evidence. Even a pass is a fixed-generator component result, not a
whole-goal success, selective-observation result or full EventFrame latency
claim. This protocol is frozen before reading the new stream outcomes.
