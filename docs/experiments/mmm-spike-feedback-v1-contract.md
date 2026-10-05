# Delayed-feedback expert aggregation: frozen exploratory contract

Scope: postprocess the already-issued 84 cadence forecast records, aligned
to the original v120 source. No new fit, retrospective expert forecast, prior
tuning, production change or claim of fresh confirmation. Existing challenger
fit cost remains part of the method, even though this replay reuses it.

Experts are Markov (control index 12) and the arrival-refitted spike model.
At clock 128 initialize weights equally. Keep log weight ratio challenger /
incumbent, initially zero. Before predicting clock c, process each earlier
issued forecast j exactly once if its outcome has arrived: j<c, not missing,
and j+Delay<=c. For each, add loss_incumbent-loss_challenger to log ratio,
using squared loss and the probabilities actually issued at j. Learning rate
is frozen at 1; no forgetting, clipping or weight floor. Predict the convex
mixture with challenger weight sigmoid(log ratio). Tied arrivals process in
origin order. Current labels never update their own forecast. No labels from
before 128 update this aggregator, because no stored paired candidate
forecasts exist there. Pending labels beyond clock 159 do not affect results.

This is a generalized exponential-loss mixture, not an ordinary Bayesian
posterior. No no-delay regret guarantee or BOLD theorem is claimed for this
delayed/missing feedback implementation. Missingness may bias observed loss.
Pending storage is bounded by the 32-forecast block. Full-stream operation
would need a separately specified bounded retention policy.

Compare mixture, incumbent, challenger and fixed half-mixture on expected
and realized Brier. Report all 84 per-record deltas and phase/schedule means,
including number harmed by more than .01 versus incumbent. A pooled win is
not non-harm. Rejection screen: any per-record expected harm above .01 means
this pilot does not support unconditional incumbent protection. Passing would
only justify a broader experiment, not establish statistical non-inferiority.

Verification: hand-computed update, no current/future/missing label leakage,
positive control on available changed labels, no double counting, delayed
arrival, and exact zero-feedback fixed-mixture equivalence. Independently
recompute weights from scratch at every clock as a second implementation.
Report postprocessor runtime separately from retained fitting/prediction cost.

Audit classification: cadence regressions confirmed; their full causes remain
unresolved (prior regularization, variational mode choice, sampling noise and
drift all remain plausible). Aggregation is a research recommendation, not a
bug fix. Falsifier: it retains material scenario harm or loses average utility.
No upstream software patch is implicated; this operates only on local,
untracked exploratory artifacts.
