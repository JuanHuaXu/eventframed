# Two-step acquisition diagnostic v8

Freeze before execution. Inspect the first16 episode seeds per case in v7's
split0 only, after0,4,8,12 acquired reports: seven cases,112 episodes,448 states.
Use only each stored prefix to rebuild the current model; do not consume its
future outcomes, target label or actual source modes for action values. This is
reuse of design data for a mechanistic diagnostic, NOT untouched confirmation.

For available first action a define V2(a) = G1(a) + E_y[max_b G1(b | a,y)],
where b excludes the used source slot. Group identical source slots by test,
keeping the smallest available slot to preserve the tie rule. Exact binary
branches and the full reliability-mixture update, no rollout of true outcomes.
Compare max_a V2(a) with V2(a_greedy), not merely with G1(a_greedy): both choices
must get the same two-observation budget and optimal one-step continuation.

Record all action values, selected actions and model-implied value gain. An
action change counts as meaningful only when value gain exceeds1e-10. Aggregate
by case and by prefix length without pretending the four states per episode are
independent samples. No uncertainty or predictive-success claim is made.

Warrant a fresh full rollout if either independent20 or matched_misleading20
has mean model-implied value gain>=0.002 AND meaningful action changes in>=10%
of its inspected states. This is a computational screening rule, not an adoption
gate. Preserve low-value outcomes rather than lowering this threshold afterward.

Verify values against a separate four-leaf enumeration of expected terminal
posterior squared norm, empty/single-action boundaries, no state mutation,
stored-prefix forecast agreement, source/artifact hashes and full diagnostic
replay. Runtime measurements are Python diagnostic execution costs on this host,
not serving benchmarks. No production changes or additional data acquisition.
