# V10 issued-expert bottleneck

2026-10-01. This is a consumed-data diagnosis of the failed
[v10 outcome-family test](mmm-retained-truth-v10-results.md), under the
[frozen expert protocol](mmm-retained-truth-v10-experts-protocol.md). No
forecast, fitted model or served behavior changed. The experiment scored
each archived expert only on the view that actually issued it, with no
per-frame hindsight selection.

## What the experts show

On the **replacement arm's own view**, confirmation Brier in the first 64
post-change frames (lower is better):

| Input/case | Issued mix | Tree | Short count | Neutral |
| --- | ---: | ---: | ---: | ---: |
| Uniform shift128 | 0.27274 | 0.25329 | 0.27427 | 0.25000 |
| Uniform shift256 | 0.28016 | 0.26287 | 0.29984 | 0.25000 |
| Latent shift128 | 0.27704 | 0.25854 | 0.28293 | 0.25000 |
| Latent shift256 | 0.27760 | 0.27658 | 0.31476 | 0.25000 |

The tree is better than the issued mix in each early cell, but a neutral
0.5 forecast is better still. The short-count expert is especially poor
early. The archived frozen-base expert scores about 0.37-0.44 across the
post-change replacement-view cells, consistent with severe stale-regime
exposure. This is not proof of the mix's internal weight trajectory; those
weights were not archived in this diagnostic.

Across the **entire post-change window** on that same replacement view:

| Input/case | Issued mix | Tree | Short count | Neutral |
| --- | ---: | ---: | ---: | ---: |
| Uniform shift128 | 0.24625 | 0.24639 | 0.24737 | 0.25000 |
| Uniform shift256 | 0.25732 | 0.25519 | 0.26618 | 0.25000 |
| Latent shift128 | 0.24531 | 0.24647 | 0.24378 | 0.25000 |
| Latent shift256 | 0.25621 | 0.26130 | 0.27523 | 0.25000 |

The early mix has avoidable loss relative to its own tree and neutral
experts, so online expert tracking is worth a **separate** investigation.
But in both full shift256 windows, neutral beats every learned replacement
expert. Better weights alone cannot realize the much lower selected-view
oracle Brier (about 0.17-0.19) from the prior diagnostic. A more useful
conditional predictor is also needed. The existing research-only ridge
logistic model is one candidate for the majority-rule family, but its
[earlier broad standalone test](mmm-soft-learners-v118-results.md) failed;
it must remain an optional specialist behind forward evidence, not replace
the incumbent by assumption.

[Herbster and Warmuth's fixed-share method](https://mwarmuth.bitbucket.io/pubs/J39.pdf)
is relevant to tracking a changing best expert. Its regret setting does not
imply EventFrame's delayed-feedback, selected-view, Anti-Pigeon or forecast
quality requirements are satisfied. We have not implemented or validated it
here. No retrospective expert choice is a deployable policy.

## Verification and scope

The streaming scorer read all 48 confirmation shift trajectories per input
mode, checked the fixed/short, replacement/tree and static-blend expert
mappings, and reproduced every arm's archived post Brier to within 1e-12.
The independent check recomputed 136 aggregate fields per mode and matched
the original v10 final-arm summary. Focused accumulator tests passed under
the race detector; `go vet` passed. Diagnostic scans took 3.76 and 3.65
seconds for the roughly 25 MB compressed journals. This is offline file
analysis, not daemon request latency or a new confirmation sample.

Complete per-record and per-arm results are in the
[uniform](mmm-retained-truth-v10-uniform-experts.json) and
[latent](mmm-retained-truth-v10-latent-experts.json) diagnostics. Their
[uniform check](mmm-retained-truth-v10-uniform-experts-check.json) and
[latent check](mmm-retained-truth-v10-latent-experts-check.json) are preserved
alongside the source journals. The next causal discriminator is whether
an as-of specialist plus a bounded switching rule can improve fresh
multi-family streams without the parity and stationary harms observed in
earlier research. That test must keep all original controls and cannot reuse
these consumed outcomes as confirmation. All seven research directions
remain open.
