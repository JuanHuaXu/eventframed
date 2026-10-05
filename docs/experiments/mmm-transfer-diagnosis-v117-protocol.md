# Transfer diagnosis v117: consumed data only

Use every v116 record:1152 schedule runs, all phases/families/modes,32 indices.
Pin parent SHA-256 cad9f6f36879280044f0148895910f9fc10e3d0df902da95cc8276980f079d1e.
This diagnosis cannot validate a rescue. No policy, threshold, label or scored
forecast changes, fresh confirmation claim, or favorable-case selection.

Recreate the16 initial packets from the parent's declared generator spec and
verify all256 inputs/outcomes and the feedback schedule. Reconstruct each as-of
64/32 fit from recorded arrivals, not from future outcomes. Refit the unchanged
generic64, Boolean64, generic32, Boolean32 models at all8 publications. Record
each model's complete512-input table plus its probabilities on each of the three
issued views. Require reconstructed generic64 on its own view to match v116.

Independently average each stored table over uniform completions to check every
recorded partial forecast. Score against the parent's known Bernoulli probability
using expected Brier. Report all256, terminal64 and each32-tick block. Across
32 trajectories, provide paired mean +/-3.5SE diagnostic intervals; these are
post-result diagnostics, not prospective or simultaneous inferential guarantees.

For each policy view v and each raw model, compare full-input risk and view-v
risk. Also evaluate the fixed admissible mask63 (the first six coordinates).
It is diagnostically useful because these generators' relevant bits lie there;
do not pass the known relevance set to a learned policy or call this a general
observation solution. Full-input models can still overfit irrelevant inputs,
so fewer observations can sometimes reduce their loss.

For each issued policy, decompose exactly:

served risk - true-law noise floor =
(served risk - generic64 risk on that policy's view)
+ (generic64 view risk - generic64 full-input risk)
+ (generic64 full-input risk - true-law noise floor).

Also preserve served-minus-generic-own-view decomposition into matched-view
model/weighting difference plus generic view difference. These are arithmetic
counterfactual comparisons, not identified causal effects. No nonnegativity
claim is made for the first two terms.

Compute the optimistic pointwise convex-hull oracle of the four raw forecasts
and neutral .5, both on each issued view and on full input. This oracle uses
the true q to choose a different mixture per outcome/input and is not generally
realizable by a coherent evidence-based policy. A low oracle loss demonstrates
only potential headroom; a high loss bounds what those raw choices can represent
at those evaluated points. Do not infer deployed weights from oracle results.

Store an exclusive0600 artifact and pin all154 parent hashes plus diagnosis
source, verifier and protocol. Fully replay, independently reconstruct, retain
all cases, and record actual diagnosis/replay duration without calling it serving
latency. Preserve v116's FAIL and all seven open research directions.
