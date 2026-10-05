# v114 consumed-data error decomposition

This is diagnosis, not a rescue or fresh confirmation. Parent v113 is frozen at
SHA256 54b6220b6936227e4e1e45954a67522a5f7eb152adb32a3bc8ef5f5283550860.
Use all32 indices in both parent phases for delayed/missing parity4,
majority-to-parity and parity-to-majority:192 trajectories. Retain every block,
including pre-change blocks. Compare actual arms5 (arrival log),7 (fixed Markov)
and8 (rate mixture) against arm0 generic64. Phase names identify parent splits,
not untouched diagnostic confirmation.

Reconstruct the16 initial examples from the parent seeds, then fit all four
models using only parent as-of origins. Verify every origin against the arrival
schedule. No new policy updates, fits on future labels, or modified forecasts.
Save512 full-input predictions per fitted model and raw model predictions at
each issued mask. Independently reconstruct partial predictions by averaging
uniform-input completions and compare the generic64 prediction exactly against
the parent baseline at its own mask.

For Brier risk r(p,q)=(p-q)^2+q(1-q), let S be the actual served prediction,
G_m generic64 at that policy's mask, and G_g generic64 at its own mask. Report
the exact additive identity

    r(S,q)-r(G_g,q) = [r(S,q)-r(G_m,q)] + [r(G_m,q)-r(G_g,q)].

The first term is model/mixture difference at matched observations; the second
is observation difference for the same generic model. This is an algebraic
diagnostic, not a causal or full-policy counterfactual: changing weights would
also change future observations and feedback. Acquisition costs can differ;
report them rather than describing all masks as equal-cost.

For each mask also report all four raw model risks, full-input raw risks, and
an optimistic pointwise convex-envelope lower bound. The latter clamps the
known q to the minimum/maximum of the four raw probabilities plus neutral .5.
It can use a different mixture at each known outcome/input; it is NOT a feasible
shared-weight policy, not an attainable forecast using observed data, and not a
certificate about unseen inputs. Its headroom only distinguishes the available
prediction range from the actual mixture. Partial and full-input envelopes need
not be monotone under misspecification; make no information-theoretic claim.

Aggregate per32-frame publication block, then over32 trajectories per parent
phase/case/arm. Report paired mean +/-3.5SE descriptive intervals for differences.
No new pass/fail quality gates, rate tuning, oracle deployment or production
change. Infer a next hypothesis only after inspecting both switch directions
and the stationary control. Any subsequent candidate needs fresh evaluation.
