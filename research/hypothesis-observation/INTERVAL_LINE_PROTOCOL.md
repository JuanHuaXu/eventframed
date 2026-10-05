# Frozen interval line-guard rescue

The closest-point interval guard terminated at its existing 10000-cycle
solver cap. Preserve that failed run and fixture; no quality conclusion follows
from it, and do not relax tolerances or use its incomplete iterate.

This separate candidate uses guardedTransfer, the existing analytic maximal
feasible move on the segment from local baseline to optimistic proposal.
It uses exactly the same Bernstein coefficient envelope over [.10,.30] from
NOISE_ENVELOPE_PROTOCOL.md. Priors, epsilon=.01, five evaluation worlds,
ten allocations, sixteen counterfeit masks and all900 gates stay unchanged.

This is a new candidate, not a silent fallback within the failed experiment.
The algebraic quadratic line bound supplies feasibility, not closest-point
optimality or a best-achievable gain ceiling. All conditional coefficient-law
checks must pass; retain and report all gain failures. Run component tests,
byte-exact replay, original-control parity and alternate scoring checks.
The continuous interval guarantee is still conditional on the common-noise
likelihood and copied-root model being adequate. No independent per-source
noise or empirical coverage claim, no production deployment.

