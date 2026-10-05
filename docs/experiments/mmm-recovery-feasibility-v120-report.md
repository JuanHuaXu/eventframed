# Confirmed infeasible recovery gates in the original protocol

Classification: CONFIRMED protocol contradiction for an all-gates success on
the archived v120 records. Existing failed experiments remain failed. No
threshold, sample, source artifact, or success classification was changed.

## Scope and proof

The original [v120 protocol](mmm-soft-learners-v120-protocol.md) requires
terminal64 expected-Brier mean gain at least .005 against three controls in
every changing scenario. There are96 recovery requirements per candidate cap,
192 across caps64/32, using32 indices in each phase/case/schedule cell.

For a Bernoulli outcome probability Q, expected scalar Brier is

```text
L(p;Q) = (p-Q)^2 + Q*(1-Q).
min_p L(p;Q) = Q*(1-Q), attained at p=Q.
max_p [L(control;Q)-L(p;Q)] = (control-Q)^2.
```

This holds even for an oracle with access to Q and with no model, safety-guard,
compute or evidence restriction. Averaging the last quantity over the exact
terminal frames and trajectories gives a necessary upper bound on attainable
mean gain. If it is below .005, no predictor can pass the prescribed mean gate;
changing a confidence interval cannot repair that contradiction.

## Seven unattainable requirements

All affected cases are majority-to-parity. Phase0/1 are the original design/
confirmation partitions, not new untouched evaluations in this audit.

| Phase | Feedback | Cap | Control | Maximum possible mean gain |
| --- | --- | ---: | --- | ---: |
| 0 | immediate | 64 | Markov | .002619815 |
| 0 | immediate | 64 | no-change64 | .000627418 |
| 0 | immediate | 32 | Markov | .002619815 |
| 0 | delayed/missing | 32 | no-change32 | .002723956 |
| 1 | immediate | 64 | Markov | .002175045 |
| 1 | immediate | 64 | no-change64 | .000736737 |
| 1 | immediate | 32 | Markov | .002175045 |

These are seven protocol requirements but five distinct control/cell
comparisons: the Markov comparisons recur for both candidate caps. Cap64 has
four impossible requirements and cap32 has three, so neither can pass its
complete original recovery contract on these fixed records.

Example: phase0 immediate no-change64 terminal Brier is .048127418. The noise
floor is .047500000. Even a perfect probability forecast can improve by only
.000627418, far below .005. This is not merely low statistical power.

The remaining185 requirements satisfy this necessary mean-ceiling condition;
that does not prove their confidence gates attainable or the headroom learnable.
The many other failures remain substantive. This finding does not vindicate a
previous candidate or demonstrate continuous learning, calibration, real-task
benefit, or runtime readiness.

## Reproducibility and independent check

- Audited all2688 original records, all unique identities, all32 indices per
  changing cell, and both candidate caps. No generated outcomes were changed.
- Source SHA256 matches the original summary:
  `5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f`.
- `research/recovery-feasibility-v120.mjs` computes direct squared-probability
  error ceilings. It also checks1331 elementary probability-grid inequalities.
- `research/recovery-feasibility-crosscheck.mjs` independently subtracts a
  recomputed noise floor from the ARCHIVED control Brier metric. All192 values
  agree, maximum difference1.249e-16. This is independent arithmetic, not an
  independent outcome dataset.
- Full direct audit replay is byte-identical. No fitter was rerun or selected.
- Artifacts: `mmm-recovery-feasibility-v120.json`,
  `mmm-recovery-feasibility-v120-replay.json`, and
  `mmm-recovery-feasibility-v120-crosscheck.json`.

## Boundary and proposed next action

Stop attempting to obtain all-gates success by tuning learners against these
records. It is impossible under the existing objective for this experiment.
Do not remove the cells, lower the threshold, swap the seed, or retroactively
promote results. The broader seven research goals have not been narrowed or
declared achieved.

User approval has been requested before drafting a replacement recovery
protocol. A candidate direction is to measure post-change recovery speed and
accumulated prediction error, retaining stationary non-harm and fresh frozen
confirmation. This is a recommendation for review, NOT an adopted protocol.
Terminal performance after a control has already recovered is not by itself
a measurement of adaptation speed.

Other goals remain actionable while that decision is pending. In particular,
the recent incremental-kernel speedup does not settle loaded durable-learning
latency. Existing v62 service-load and v69 ledger results remain failed;
repeating their rejected locking variants would not constitute new progress.
No production, Go runtime, private data, whitepaper, commits or push changes.
