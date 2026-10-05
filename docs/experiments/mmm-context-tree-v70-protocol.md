# Context-tree averaging v70: frozen learner screen

Research only; no serving, MMM acquisition, or production changes. Hypothesis:
branch-specific pooling can improve sparse-label learning compared with the
existing subset-table mixture. Alternative explanations for prior failures are
observation allocation, insufficient labels, and changing regimes; this screen
isolates representation and cannot establish those causes.

Use nine Boolean covariates, depth at most three, Beta(1/2,1/2) leaves.
At each nonterminal context put prior mass 1/2 on a leaf and divide the remaining
1/2 equally among unused-coordinate splits. Children are independent conditional
on their split. Average all these trees exactly using context dynamic programming;
do not greedily choose variables. Sequence evidence has no binomial coefficient.
Empty contexts have unit evidence and predictive probability 1/2.

This is our finite adaptation, not a faithful implementation or empirical claim of
[Chipman, George & McCulloch, Bayesian CART Model Search (1998)](https://www.rob-mcculloch.org/some_papers_and_talks/papers/published/cartfinal.pdf).
Their tree priors and posterior averaging motivate the model; our bounded exact
enumeration replaces their stochastic search. No causal interpretation follows.

Compare the existing subset mixture and new tree on identical 64 labeled samples.
Two phases use seed bases 2026117001 and 2026117002; eight independent fits per
family/noise/window cell. Families: majority3, multiplexer3, parity3, parity4
(depth-mismatch negative control). Noise: 0 and 0.1. Uniform independent inputs.
Window composition: 0,16,32,64 samples from the new regime, remainder old.
Initial relevant coordinates start at bit5; new coordinates start at bit0.
Evaluate old truth only when new-sample count is zero; otherwise new truth.
Exactly enumerate expected Brier over all 512 test inputs and Bernoulli noise;
test truth is used solely by the scorer, never passed to the learner.

512 paired fits total. Primary screen: mean tree-minus-subset Brier <= -0.005
for multiplexer at 32 and 64 new labels in both noise levels and both phases.
Protection: every family/noise/window/phase mean delta <= 0.005.
Require every primary and protection cell to pass for overall finite-screen PASS.
Per-fit observations retained; no pooled-input confidence claim, tuning or seed
replacement. Report descriptive seed-level uncertainty, not monitoring guarantees.
No direction is completed by this screen; independent generators and partial-view
integration are still required. Failure preserves incumbent.

Test evidence recurrence against independent tree enumeration, label complement
and coordinate permutation symmetry, invalid input, bounded probabilities,
immutability and deterministic replay. Fit is slow-path only; retained prediction
table is 512 float64s. Up to 835 contexts, 130 applicable masks per sample, fixed
depth3. Measure fit and lookup cost separately; no end-to-end latency inference.
Record source/protocol hashes with exclusive-create artifact before final docs.
