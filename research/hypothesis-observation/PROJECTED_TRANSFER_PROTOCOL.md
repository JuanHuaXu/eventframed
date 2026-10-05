# Closest-feasible transfer diagnostic

Freeze before quality scoring. Retain the same p0, hierarchical proposal,
conditional ambiguity laws, epsilon=.01 and all180 allocation stress gates.
The candidate minimizes .5||p-p1||^2 subject to p in the simplex and
||p-q||^2 <= ||p0-q||^2+.01 for every supported declared law q. This closed
convex intersection contains p0; its Euclidean projection exists and is unique.

Use Dykstra's cyclic algorithm with one correction vector per constraint, not
ordinary alternating projections. Algorithm source: Pinto,
[On the finitary content of Dykstra's cyclic projections algorithm, introduction](https://www2.mathematik.tu-darmstadt.de/~pinto/articles/Dykstra.pdf).
Convergence in the source is not a fixed-iteration accuracy guarantee for this
implementation. Inspect numerical feasibility AND a primal-dual objective gap.

The dual lower bound for multipliers y_i is
`p1 dot sum(y_i) - .5||sum(y_i)||^2 - sum(support_Ci(y_i))`.
Ball support is q dot y+r||y||; simplex support is max coordinate. In addition
to Dykstra's multipliers, try nonnegative combinations of active-constraint
normals. They produce another valid dual candidate, not an assumed optimum.
Repair roundoff toward p0 using the already verified line guard; compute the
gap for the repaired forecast. Require gap<=1e-12, feasibility tolerance1e-12,
and stop after10000 cycles with an explicit failure, not silent row omission.

An initial component test stalled despite reaching the analytic two-class
answer: nearly redundant constraints delayed convergence of internal dual
variables. The independent active-normal certificate resolves this without
changing feasibility or objective tolerances. Component tests now cover300
analytic interval projections and2000 feasible-direction optimality checks.

Score all10 allocations/16 masks and retain every gain/protection/false-confidence
gate. Require complete deterministic replay. Report cycles and certificate gaps;
operation counts are not loaded serving latency. The same model-adequacy and
exponential ambiguity-family construction caveats remain. No production or
whitepaper changes or inferred completion of the full research directions.
