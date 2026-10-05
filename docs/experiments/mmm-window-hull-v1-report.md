# Existing-expert headroom diagnostic

## Verdict

There is substantial hindsight reweighting headroom, but a weights-only change
cannot close the entire same-view forecast gap. This is a diagnostic finding,
not a successful rescue. Sparse evidence, stale evidence, and model-family
limitations remain distinct explanations for the residual error. All seven
research goals remain open.

## Method and Limits

Use the same 128 trajectories, two schedules, four 128-frame windows, and three
observer arms as the [oracle-view diagnostic](mmm-window-oracle-v1-report.md).
Let q be the exact simulator probability given the actual recorded view.
For scalar Bernoulli probabilities, the convex hull of expert forecasts is
the interval from their minimum to maximum. Project q onto this interval.
Squared distance to that projection is a lower bound on the error attainable
by any convex mixture of these fixed forecasts.

We separately relax the four outer forecasts and all underlying forecasts
(base, long, label-count, label-subset, event-count, event-subset, and neutral).
The current served gap is (p-q)^2. Subtracting the outer projection error gives
outer-weight headroom; subtracting the leaf projection error from the outer
error gives additional inner-weight headroom.

These are per-frame hindsight relaxations. They discard weight floors, evidence
delay, and learning constraints. They hold views and fitted forecasts fixed;
changing weights in a real observer can change subsequent views and training
feedback. Headroom is therefore not a predicted achievable improvement, and
the projection must never be used as a learner input. No future label is used
to change any recorded prediction.

## Results

Late reverse shift (parity to majority), original coupled arm, frames 384-511.
All entries are mean squared probability error relative to the same-view oracle,
not total Brier; the noise and missing-information terms remain additional costs.

| Cohort / schedule | Served gap | Best relaxed outer mixture | Best relaxed leaf mixture |
| --- | ---: | ---: | ---: |
| 1 / immediate | .096091 | .056285 | .043207 |
| 2 / immediate | .087857 | .055071 | .040012 |
| 1 / delayed | .115120 | .059469 | .046745 |
| 2 / delayed | .116521 | .066612 | .048778 |

The leaf relaxation still leaves roughly 40-46% of the current gap. The oracle
lies outside the leaf hull on 82.32%/80.91% of all late immediate frames and
71.44%/70.12% of delayed frames. These frequencies include all frames, not only
informative views; small nonzero distances count, so the error magnitudes above
are more informative than frequency alone.

Late forward shifts also retain leaf error: .015057/.021896 immediate and
.014503/.023913 delayed in the coupled arm. These coexist with the substantial
missing-information costs identified previously. Neither better weights nor
better acquisition alone is established as sufficient.

Do not interpret an outside-hull result as proof that the underlying model
family cannot represent the truth. Finite-sample shrinkage and contaminated
training windows can put all currently fitted forecasts on the wrong side of
q even when an adequately trained model could represent it.

## Verification

- 14,641 grid checks confirm projection is no worse than every tested convex
  combination; explicit controls cover targets above, below, and inside a hull,
  and a degenerate single-value hull.
- All four input hashes are pinned. Every per-trajectory/window/arm served gap
  agrees within 1e-12 with the previous exhaustive oracle diagnostic.
- Every frame satisfies leaf error <= outer error <= served error, within
  floating-point tolerance. Observed values and target-mask sizes are checked.
- Replay using independently regenerated native tapes is byte-identical.
- The first input hash was corrected during pre-run readback; there was no
  failed scoring collection or change to data/criteria.

Script SHA256: `247802a25075af4660d784f50bd0bfe8412f5a129ef73c3a80fdf93278bda1a6`.
Result SHA256: `9c1ff1dc3b24f2c69458f61648f03d704eb936b1ddb7f420ad0ef526f9881f01`.

Artifacts: [script](../../research/window-hull-diagnostic.mjs),
[result](mmm-window-hull-v1.json), [replay](mmm-window-hull-v1-replay.json).

## Next Investigation

Before changing priors or adding another model, use the existing consumed fit
origins to distinguish sample scarcity from mixed-regime contamination. Compare
actual windows with count-matched post-change-only diagnostic fits on the same
publication clocks, using only labels already arrived at those clocks. Where
there are insufficient post-change labels, report that constraint rather than
borrowing future labels. Known change boundaries may select a hindsight
diagnostic subset but must never enter a deployable policy. Preserve both
directions and stationary controls; check earlier family-evidence follow-ups
before implementing a duplicate experiment.

No production, whitepaper, configuration, remote, commit, or push changes.
