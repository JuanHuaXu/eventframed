# Interval-aware correction after v95

Status: proposed, not implemented or validated. v95 passes54/58 gates but
fails all four recovery gates. Preserve that result. Improving whole-stream
loss does not certify the post-change segment.

## Evidence and competing explanations

Confirmed: confirmation parity-to-majority late-half expected Brier rises from
generic0.172585 to online-share0.184354, despite whole-stream improvement.
Skeptical BMA is also worse in that segment (0.190748), so merely substituting
it as the rescue is not supported. The retained parity expert is substantially
worse there (0.264845).

Possible causes are old expert weights, stale 64-frame model windows, and the
challenger family's inability to represent majority. No scalar sharing-rate
sweep on consumed streams resolves their causal contribution. The diagnosis
must distinguish selector delay from both experts being temporarily wrong.

## Relevant research

[Daniely, Gonen and Shalev-Shwartz (2015), Sections1.1-2](https://proceedings.mlr.press/v37/daniely15.pdf)
define regret on every interval and construct SAOL using geometrically
scheduled learner instances. Theorem1 gives logarithmic per-round overhead
and an interval-dependent regret bound. The distinction from fixed-share
tracking is directly relevant to our lifetime-versus-recent-regime gap.
The full-information setting matters: after a binary outcome we can score
all issued forecasts; unanswered or selectively labeled agent tasks do not
automatically satisfy it. This is motivation, not a theorem for our runtime.

[Hazan and Seshadhri (2009)](https://icml.cc/Conferences/2009/papers/75.pdf)
study adaptive regret in changing environments. Their exp-concave setting
is relevant to the bounded squared-loss construction. Their results do not
authorize uncharged resets of the v94 lifetime comparison.

## Proposed bounded implementation

First test interval-local weight learners over the SAME pre-outcome generic
and parity forecasts. Start instances on geometric intervals and retain their
weights only for the declared lifetime. No reset observes future labels or
the simulator's change point. This isolates selector history without adding
extra model refits. All interval and outer updates wait for the actual label.

For the deterministic squared-loss specialization, mix active interval
forecasts p_i with normalized meta-weights pi_i. Use the surrogate loss
ell_bar=sum_i pi_i*(p_i-y)^2 for the paper's meta-update; actual mixed-forecast
loss is at most ell_bar by convexity. Do not silently substitute a different
weight update and inherit the original proof. Maintain positive weights and
explicit issue/update ordering. Verify against a literal reference algorithm.

Predeclare a finite evaluation horizon and interval coverage. Capping or
evicting long-lived instances changes the comparison outside that horizon;
logarithmic growth is not constant memory for an unbounded daemon. Account
for every active instance, update and restart rather than calling it free.

The published worst-case bound can be loose at128 outcomes. Before making a
finite1% guarantee, evaluate the actual constant and comparator; empirical
non-harm tests remain necessary. Keep stable-case and interaction-gain gates,
but additionally test recent windows so old gains cannot hide new harm.
Preserve v95's four failed gain gates in reporting, not retroactively relaxed.

If interval weighting alone fails because both forecasts are stale, the next
distinct experiment adds shorter bounded model windows, charging every fit.
Do not combine both changes in the first experiment: that would hide whether
the rescue came from selection or newly learned models. Use fresh design and
confirmation streams; old stream diagnostics are explicitly consumed evidence.

This lead does not solve missing/delayed feedback, selective labels, Anti-Pigeon
external certification, real-agent validation, or persistent concurrent serving.
