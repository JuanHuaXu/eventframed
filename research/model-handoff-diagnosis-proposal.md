# Model handoff diagnosis after v109

Subsequently completed as [v110](../docs/experiments/mmm-handoff-v110-results.md).
The trace supports a late weighting lag in the failed direction while retaining
an acquisition effect elsewhere. Next is the separately specified
[compatibility-handoff candidate](compatibility-handoff-proposal.md).
The original pre-diagnosis hypotheses below remain for provenance.

Research lead, not a confirmed causal explanation or implemented rescue.
The [v109 result](../docs/experiments/mmm-log-v109-results.md) passes every gate
outside delayed parity-to-majority, while that case has early and late harm.
Its log/neutral variant is worse. Do not tune a neutral prior or discard those
trajectories to make a candidate pass.

## Competing explanations

1. Late-arriving losses favor roles that were good when issued but are no
   longer best after a fit/publication. Historical role identity is preserved,
   but fitted probability laws change.
2. Fast weighting makes a short-window model useful mid-recovery, then retains
   it after a long-window model improves. This is a hypothesis until weights
   and the relevant raw forecasts are inspected.
3. Changed mixture weights choose different observations, so apparent selector
   gains/harm may actually involve omitted coordinates. Full-input model risk
   and served partial-observation risk must remain distinct.
4. Neutrality may remain influential after useful models return, or may change
   acquisition. The block differences alone do not identify either pathway.

## Required consumed-data trace

Use all delayed switch trajectories from both v109 phases, not only failures.
Reconstruct original trajectories exactly. Record pre-issue advice weights,
post-gate mixture weights, actual acquired masks, raw issue-time predictions,
origin/arrival/publication identity for each delivered loss, and per-publication
model movement. The trace must not change the original outputs or gate counts.

For the fixture's uniform512-input law, compute exact mean absolute differences
between successive fitted forecasts per role, alongside full-input expected
Brier. These are simulator diagnostics; truth never enters advice, gate or
model publication decisions. Use paired fixed-mask counterfactuals only as
diagnostics and label them distinctly from a fresh adaptive-policy run.

The falsifier for the stale-weight explanation is that the correct role already
has weight and the regression persists under matched views. The falsifier for
the acquisition explanation is that identical masks retain essentially the
same harm. Report mixed findings instead of forcing a single cause.

## Subsequent rescue options, not yet chosen

Consider version-aware transfer of advice evidence only after establishing
which forecast changes invalidate useful history. Preserve stationary evidence;
v108 already showed why blanket discounting toward the initial prior is costly.
Alternatively, newly fitted models can be treated as fresh experts with explicit
entry priors, but this requires accounting for the expert cap, gate budget,
pending feedback ownership, model storage and additional prediction cost.
It cannot be slipped into the existing four-expert guarantee.

[Mourtada and Maillard (ALT2017)](https://proceedings.mlr.press/v76/mourtada17a/mourtada17a.pdf)
study efficiently tracking newly arriving experts and give MarkovHedge-based
constructions. This motivates a possible new-model admission mechanism, not a
guarantee for our delayed, censored and evidence-gated policy. Freeze any new
choice, costs and tests before fresh outcomes; preserve v109 as FAIL.
