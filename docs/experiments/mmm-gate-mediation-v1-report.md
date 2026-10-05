# Split mediation: local improvement with little mixture influence

## Finding

The local replacement is better than the pooled forecast in the member-shift
interval, but the mixture gives that slot little weight. Faster authorization
therefore has little direct influence on the emitted law. This supports a
specific next experiment about expert identity and initialization, not a lower
Anti-Pigeon threshold or a claim that splitting is useless.

[Contract](mmm-gate-mediation-v1-contract.md),
[raw per-frame trace](mmm-gate-mediation-v1.jsonl),
[summary](mmm-gate-mediation-v1-summary.json),
[replay](mmm-gate-mediation-v1-summary-replay.json).
Raw SHA-256:
`dd8b21196f69c73c83e7ad33dd1ff7ce0f47106002719218afcdd52d005b474f`.
The source archive is the previous paired-gate study, not fresh data.

## Verified isolation

All160 trajectories and81,920 frames are replayed. Every original control metric
and tape, both subset-arm metrics and tapes, and all fit hashes/counts match the
archive exactly. All807 captured source hashes match current/embedded files;
summary replay is byte-identical. The independent JS calculation reconstructs
all full/post Brier sums from raw forecasts and labels and verifies123,042
available-model affine identities, maximum error2.082e-16.

The diagnostic race/parity test passes1.964s; vet passed before collection.
Full replay completes15.44s (15.619s package). This is experiment overhead, not
serving latency. Models absent before first fit are marked unavailable. Diagnostic
forecasts are frozen before labels, scored afterward, and never enter any update.

For the actual observed mask and current mixture state, compute both local and
pooled slot forecasts while holding every other expert fixed. Effective slot
weight includes the actual fixed-share transformation, measured through the
existing Forecast API. With its clipping function c, the verified identity is:

`p_with_local - p_with_pool = w_effective * (c(p_local) - c(p_pool))`.

This is a same-mask one-step substitution, not an alternative observation or
learning trajectory. Descriptive frame averages do not supply independent-frame
confidence bounds or establish a full causal mediation decomposition.

## Earlier-split interval

Only frames where the mixture-gate arm has split and the old-gate arm has not
are included below. Values are for the mixture-gate arm's own pre-label state.

| Cohort/scenario | Frames | Mean slot weight | Mean absolute probability movement | Local-vs-pool slot Brier gain | Direct mixture Brier gain |
| --- | --- | --- | --- | --- | --- |
| Design/member | 555 | .005049 | .000554 | .035881 | .0000668 |
| Second/member | 708 | .014184 | .001183 | .037309 | .0001090 |
| Design/recurring | 1159 | .045506 | .004142 | .010774 | .0005710 |
| Second/recurring | 1175 | .037136 | .003221 | -.000134 | .0002254 |

Positive Brier gain means local is better. The standalone local forecast need
not rank the same as its contribution inside a mixture, as the recurring second
cohort illustrates. We do not infer that multiplying the slot gain by weight
equals the mixture gain; squared loss includes cross terms.

For the second member cohort, the short expert guides55.51% of these forecasts,
the incumbent33.47%, and the long slot11.02%. The local replacement is not absent
from acquisition, but its effect on the scored mixture remains small. The full
earlier study's emitted post-change gain remains negative/tiny; these conditional
diagnostics do not reclassify it.

## Next candidate, not a proven fix

Current split behavior changes the long-slot expert from pooled to local while
retaining its accumulated mixture weight (capped above at.1). Thus a newly used
local expert can inherit penalties earned by a different pooled forecaster.
This is a confirmed mechanism, not necessarily a software defect: the existing
rule deliberately favors conservative influence.

A bounded fresh prior for the newly activated expert is worth testing, with
incumbent/short histories preserved and unchanged split authority. Do not grant
it retrospective credit for the revealing outcome or force it to dominate.
Use fresh trajectories, explicit false-split controls and the original gain/
non-harm screens; distinguish the gate-timing effect from the new action effect.

Relevant conceptual source: Freund, Schapire, Singer and Warmuth (1997),
[Using and combining predictors that specialize](https://cseweb.ucsd.edu/~yfreund/papers/SpecializedExperts.pdf),
sections1-2. Their framework distinguishes awake specialists from abstaining
ones and scores only awake specialists. This motivates explicit expert identity
and activation accounting, but does not supply a theorem for assigning a local
replacement an arbitrary reset weight. A bounded birth-prior experiment would
be a declared adaptation, not a faithful specialist algorithm or inherited bound.

All seven goals remain OPEN. Production, whitepaper and remotes are untouched.
