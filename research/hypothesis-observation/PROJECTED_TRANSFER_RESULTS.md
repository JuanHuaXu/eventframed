# Closest-feasible projection does not rescue the gain failures

Under the [frozen protocol](PROJECTED_TRANSFER_PROTOCOL.md), full convex
projection again passes175/180 gates: all160 protection gates, all10
false-confidence gates and five of10 genuine-gain gates. The same five gain
allocations fail as with the straight-line guard. Overall status: FAIL.

Genuine-data gains range from0.003713 to0.005622 Brier. Compared with the line
guard, differences range from-0.000018 to+0.000085: eight allocations improve
slightly and two worsen slightly. Nearest to the hierarchical proposal does
not mean best under the true genuine mechanism. Maximum exact population harm
versus local inference is0.005656, still below the unchanged0.01 tolerance.

## Solver integrity

The feasible region is the intersection of the probability simplex and one
closed Brier-regret ball per supported conditional law. It contains the local
forecast. Dykstra's method preserves the closest-point objective through its
correction variables; see [Pinto's introduction and algorithm](https://www2.mathematik.tu-darmstadt.de/~pinto/articles/Dykstra.pdf).
No finite iteration guarantee is borrowed from asymptotic convergence.

An initial two-class test stalled at the correct analytic point while internal
dual variables moved slowly between nearly redundant constraints. Rather than
loosening tolerance, the solver now also builds a dual candidate from active
normals and checks its objective bound. Multipliers need not be optimal for
that bound to be valid. The final forecast is repaired toward the feasible
baseline before checking the gap, so the returned object is what is assessed.

Component verification passes300 analytic interval cases,2000 sampled feasible-
direction optimality checks and three boundary/redundant-order checks. Maximum
analytic error is6.89e-15; the general sampled directions are checks, not a
complete proof over all directions. Maximum component cycle count is1329.

Across all10,240 stress vectors, the solver uses at most123 cycles, with maximum
numerical primal-dual gap9.54e-13. All37,152 supported conditional risks satisfy
the0.01 bound within floating-point roundoff. No failed projection is omitted
or silently replaced by a baseline. Full output reproduces byte-for-byte.
These are numerical certificates at declared tolerances, not interval-arithmetic
proofs of machine-level roundoff bounds.

Artifacts: [all projected cells/gates](projected-renewal-allocation.json),
[component and replay checks](projected-transfer-verification.json).

## Implication

Merely removing the line restriction is not enough to recover useful gain in
these tests. It also adds iterative computation. The shared proposal itself,
the conservative ambiguity family and the objective remain possible limitations;
this result does not isolate one as the sole cause. A different forecast target
would need a new declared objective and the same counterexamples, rather than
an increased epsilon or relaxed gain threshold. All seven directions remain
open. No production, dependency installation or whitepaper changes were made.
