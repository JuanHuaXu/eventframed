# Alternate observation paths v12 diagnostic

Frozen before calculation. Use the192 consumed v9 trajectories, fixed audit
schedule and original tree seeds. Compare short64 count MMM, uniform tree,
empirical-joint tree, and oracle-joint tree, each choosing its OWN views through
the existing reader with the same six-coordinate budget. All forecasts precede
feedback. Count and tree fit the same last64 audit labels. Empirical input joint
uses last<=256 admitted audit inputs plus one uniform pseudo-observation, as v10.
Oracle is a diagnostic reference only. No outer mixture is applied.

Use only frames after the first supported fit. Report equal-weight per-stream
full/post Brier, accuracy and mean observation cost by scenario/generator/split.
This tests whether depriving the tree of its own path explains poor component
quality; it is not real-agent or fresh confirmation evidence. Replay the original
tree forecast on the recorded mask as an independent reconstruction check.

For the empirical tree to justify a fresh policy trial, require post clustered
shift128 mean Brier improvement>=.005 over the short-count own-path control in
both splits, and no >.01 full/post mean harm in any group. Preserve all outcomes;
do not promote oracle results. A failure deprioritizes forcing this particular
tree, not all ensembles or all attention policies. Benchmarking is out of scope
for this quality diagnostic; the previous performance evidence remains separate.
