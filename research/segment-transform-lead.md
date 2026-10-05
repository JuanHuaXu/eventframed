# Exact tail-transform optimization lead

Go research-only implementation and component measurements are now recorded in
[Go pilot results](segment-transform-go-results.md). The tail transform passed
numerical parity and reduced measured component time, but full fitting and
loaded-serving integration remain unverified.

Provisional algebra, not a change to the frozen v120 fitter. The component
benchmark identifies interval likelihood/tail construction as the dominant cost.
Do not claim a speedup without separately implementing and benchmarking it.

For each generic conditional-cell hypothesis m, let c(m,x) be the matching
ternary cell and put f_c=w_m*(yes_c+.5)/(n_c+1). The full forecast contribution
is sum_m f_{c(m,x)}. Starting with all3^d cell contributions, process coordinates
one at a time, adding each unknown-coordinate entry into both known-value
entries. Induction on processed coordinates shows that a final full cell holds
exactly that sum, once per compatible mask. This replaces the tail's repeated
4^d cell evaluations with3^d evaluations plus O(d*3^d) additions. Per-cell mask
metadata must be computed or cached and counted, not assumed free.

For the Boolean family, let a_m be the agreement mean and w_m include its
posterior family weight. Its contribution is

sum_m w_m/2 + sum_m w_m*(.5-a_m)*(-1)^popcount(m AND x).

The second term is an unnormalized Walsh-Hadamard transform. Pairwise sum and
difference butterflies evaluate it in O(d*2^d), rather than4^d operations.
This does not change the likelihood, model prior, boundary posterior or evidence
budget. It is an algebraic reordering, not a new learning mechanism or grokking.

The [reference script](segment-transform-reference.mjs) compares both transforms
to direct sums for every input at dimensions0..9 and two signed/positive
coefficient fixtures:2046 comparisons each. Reported errors are numerical, not
bitwise equality. Actual posterior-derived coefficients, model extremes,
ownership, first-use costs and fresh performance still need Go tests. The v120
experiment remains unchanged while it runs.
