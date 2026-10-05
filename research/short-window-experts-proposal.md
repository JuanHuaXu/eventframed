# Short-window experts after v96

Proposal only. Interval-local weighting alone passes96/96 non-harm gates but
fails all6 stationary interaction-gain gates and both parity-to-majority
recovery-gain gates. It is not the completed rescue.

The controlled comparison keeps model forecasts identical, so it establishes
a selector tradeoff, not proof that stale model windows are the sole cause.
On the same v96 confirmation streams, the fixed-share control's parity4
all-stream gain is0.022709; interval weighting retains only0.004082. Conversely,
late parity-to-majority harm falls from0.009916 to0.001400. Do not tune interval
entry weights on this consumed confirmation to advertise a fresh result.

## Next distinct intervention

Add a predeclared32-frame training window alongside the existing64-frame
generic and Boolean models. Keep identical fit times, received evidence,
observation masks and existing weight logic in the first comparison. A32-frame
window is motivated by the existing32-step publication cadence: at the first
post-change publication it can contain only new-regime evidence, without
knowing the change time. It is not guaranteed to be better and loses sample
support on stationary streams.

Do not replace every64-frame expert with32-frame experts. Preserve longer-term
models and compare bounded four-expert aggregation to the existing two-expert
control. Freeze prior mass and account for each expert in the cumulative bound;
adding experts cannot inherit the old comparator budget by assertion. Keep the
generic64 comparator, all106 quality gates, full as-of ordering, both views,
and separate production boundaries. Existing interval weighting remains a
comparison, not an automatically selected winner.

Before fresh confirmation, unit-test the generalized bounded expert update
against explicit normalization, singleton/equal experts and the two-expert
predecessor. Check no reset or omitted-label loophole. Record per-publication
window ages and expert weights, so another failure can distinguish model
availability from delayed selection. Include all new fitting costs in timing.

Use fresh seeds and untouched confirmation. No changes to v95/v96 outcomes,
margin reinterpretation, selective reporting or whole-direction success claims.
If the added models do not rescue recovery without losing stationary gains,
inspect held-out attainable forecast headroom before adding further machinery.

This is an empirical bounded-window hypothesis, not a claim that the SAOL
theorem supplies a statistical test or that shorter memory universally improves
learning. The [interval proposal](interval-brier-proposal.md) records the
primary research and the limits of transferring its results.
