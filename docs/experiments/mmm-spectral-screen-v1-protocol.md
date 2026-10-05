# Boolean interaction screening feasibility, v1

Frozen before execution, 2026-09-15. This is a new synthetic component diagnostic,
not confirmation of the seven research goals or reuse of a teacher oracle in
the learner. No production changes. Existing main-effect logistic features
cannot express pure degree-two-or-higher parity; missing labels, obsolete labels,
and posterior weighting remain separate possible causes of broader failures.

Enumerate all 246 degree-two-through-four signed monomials on nine bits. Retain
at most 16 whose absolute empirical signed-label correlation reaches
sqrt(2 log(2*246/0.1)/N), ordered by magnitude then mask. N is 32 or 64.
All nine main effects would remain in a subsequent logistic learner; this test
does not fit that learner. Its falsifier is inadequate interaction recovery at
these label budgets, even when the correct interaction is in the candidate set.

512 independent SHA-256-derived replicates per cell; uniform nine-bit inputs,
randomly permuted target coordinates, degrees 2/3/4, independent label flip
probabilities 0/0.1/0.3/0.45. A fair-label degree-zero control is also included.
Record target recovery, any selection, any spurious selection, total selections.
No threshold tuning after outcomes. Orthogonality, cap and input immutability
are executable assertions. The hash sampler is a reproducible pseudorandom
simulation, not a cryptographic independence proof.

The union-Hoeffding threshold bounds false inclusion of zero-population-
correlation features under iid samples. It is not a selected-posterior
calibration certificate and does not extend automatically to selective,
dependent, or shifted evidence. Complexity is O(246*N) bounded-degree products
plus sorting, before any logistic fit. No serving latency claim.

Source: [Heidari and Szpankowski (2023)](https://proceedings.mlr.press/v206/heidari23b.html)
studies Boolean polynomial/Fourier learners. This screen is our own restricted
heuristic, not their agnostic PAC algorithm, and inherits none of its theorem
guarantees. See their paper's summary and algorithms before attempting a faithful
implementation. Signed products here use coordinates 2*x-1.

Run: node research/spectral-screen-v1.mjs
