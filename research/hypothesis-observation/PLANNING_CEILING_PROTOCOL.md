# Known-copy-mode planning diagnostic

Declare before evaluation: compare uniformly random unseen types, maximum
report entropy, target-Brier greedy, receding depth2, receding depth3 and exact
full-horizon planning for budgets0..8, in copied05/copied20/null. All policies
know the same true likelihood and copied mode, and can only select unseen types.
Every report costs one. This isolates planning, not provenance inference.

Enumerate the exact population law, without random training/confirmation seeds.
Score terminal multiclass Brier at every budget. At ties within1e-14 choose the
lowest numbered type. Receding policies plan min(depth,remainingBudget) steps,
execute one and replan using only reports already seen. Random chooses uniformly
among unseen types. No outcome suffix or latent truth enters action selection.

Record all risks and operation counts. Check full-budget equality, oracle lower
bounds, null invariance and depth1 equivalence to direct one-step Gini selection.
Replay deterministic output. Known-mode gains do not qualify the unknown-mode
research direction. A useful lead must beat greedy at some finite budget without
silently ignoring regressions at other budgets or the planning work it costs.
Use this diagnostic to select a mechanism for a later frozen unknown-mode test,
not to declare all earlier screens passed.
