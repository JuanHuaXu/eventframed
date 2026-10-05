# Allocation stress breaks the single-schedule pass

The [predeclared exact stress test](RENEWAL_ALLOCATION_PROTOCOL.md) retains the
same model, priors, noise and16-credit cost. Four ordinary first reports are
followed by six renewals, with at least one renewal for each measured type.
All10 count allocations and16 genuine/counterfeit masks are enumerated.
The screen FAILS:173/180 gates pass, seven nonharm gates fail.

## Confirmed failures

Counts are ordered by measured types[0,1,2,7]. Mask bit i means renewals for
the i-th measured type are counterfeit; it is not the bit for the original
type number. Positive harm means hierarchical Brier exceeds local-mode Brier.

| Renewal counts | Counterfeit types | Exact Brier harm |
| --- | --- | ---: |
| [1,2,1,2] | 0,1 | 0.010545 |
| [2,1,1,2] | 0,1 | 0.011908 |
| [2,1,1,2] | 0,1,2 | 0.011214 |
| [2,1,1,2] | 1,7 | 0.010988 |
| [2,1,1,2] | 2,7 | 0.010006 |
| [2,1,1,2] | 1,2,7 | 0.011712 |
| [2,2,1,1] | 1,2,7 | 0.010292 |

All all-genuine gain and all-counterfeit false-confidence checks still pass.
The failures specifically expose partial-counterfeit pooling risk. One failure
uses the original count allocation[2,2,1,1] but a counterfeit mask absent from
the earlier four-mechanism control. Thus that earlier6/6 pass was too narrow to
establish robust protection, even without changing the acquisition budget.

These are exact population expectations under the specified finite laws, not
sample intervals. More repeated simulation seeds cannot remove these particular
failures. Their absolute size is small, but exceeds the frozen0.01 tolerance;
do not raise the tolerance after seeing the output. Conversely, this does not
prove that every shared-mechanism model or adaptive policy must fail.

## Verification

Each allocation enumerates all1024 report vectors and16 target hypotheses under
each of16 masks. Paid-slot uniqueness, total credits, true-law mass normalization
and the true-law oracle risk floor are checked. The reference uses batch latent
model evidence rather than online updating, and direct squared loss rather than
the production diagnostic's algebraic loss expansion. All800 independent
mass/oracle/model-Brier comparisons agree within7.22e-16. Source hashes bind
the model and both protocols to their artifacts.
The complete160-cell result and all180 gate decisions reproduce exactly on replay.

Artifacts: [all160 cells and180 gates](renewal-allocation-exact.json),
[independent reference](renewal-allocation-reference.json).
The independent reference covers the Brier failures, not classification tie
behavior or every false-confidence value. No runtime benchmark is claimed.

## Research consequence

Do not promote this shared-prior mixture as a universal replacement for local
freshness. Its local component gives partial-counterfeit mechanisms support,
but Bayesian support alone does not guarantee finite-evidence nonharm. Future
research must explicitly preserve these allocation/mask counterexamples when
testing a different acquisition policy or a guard on transferring source trust.
The former single-schedule pass remains recorded with its actual scope; it is
not silently erased or generalized. All seven goals remain open. No production
or whitepaper changes.
