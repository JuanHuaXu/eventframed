# Continuous noise envelope: partial rescue and a solver limit

The [interval envelope](NOISE_ENVELOPE_PROTOCOL.md) is derived from nonnegative
Bernstein coefficients, so its inclusion claim covers every common channel-noise
value in [.10,.30], not just a grid. That interval and the copied-root likelihood
family are assumed; no empirical real-world coverage has been established.

## Closest-point attempt: numerical failure

The original projection solver hit its unchanged10000-cycle cap on a72-law
envelope. The raw iterate's constraint violation was2.40834e-8 and the last
reported repaired-point objective gap was1.17015e-8. Both exceed the solver's
acceptance tolerances. The experiment terminated, emitted no quality artifact,
and was reproduced with the same fixture. No tolerance relaxation or replacement
of the incomplete iterate was made.

[Failure artifact](interval-guard-projection-failure.json) preserves inputs,
source hashes and stderr. This shows the solver did not meet its finite
contract; it does not falsify feasibility (the baseline is feasible), the
Bernstein inclusion argument, or asymptotic projection convergence. Similar
nearby coefficient-law constraints are a possible cause of slow convergence,
not a proven diagnosis. The closest-point version's quality remains unknown.

## Separately declared line-restricted rescue

The [line protocol](INTERVAL_LINE_PROTOCOL.md) instead uses the existing analytic
maximal safe interpolation from the local baseline to the optimistic target.
This is not a silent fallback and makes no closest-point optimality claim.

| Actual noise | Pass /180 | Protection failures /160 | Gain failures /10 | Worst Brier harm | Genuine gain range |
| --- | ---: | ---: | ---: | ---: | ---: |
| .10 | 180 | 0 | 0 | -.000971 | .006526 to .007161 |
| .15 | 180 | 0 | 0 | .001678 | .005652 to .006513 |
| .20 | 173 | 0 | 7 | .004257 | .004073 to .005471 |
| .25 | 170 | 0 | 10 | .006967 | .001812 to .003840 |
| .30 | 170 | 0 | 10 | .009707 | -.001063 to .001456 |

All50 false-confidence gates also pass. Overall873/900 passes: FAIL because
27 genuine-gain requirements remain unmet. This improves the universal guard's
850/900 count without abandoning protection, but is not a complete rescue.
These are consumed finite research cases, not untouched confirmation.

## Verification

- 107100 polynomial-likelihood and normalized-law comparisons against direct
  products pass, max absolute error5.55e-16. They cover three allocations,
  deterministically selected bit vectors and seven endpoint/interior noise
  points. The continuum guarantee follows from the coefficient derivation,
  not from those sampled verification points alone.
- Five malformed-contract rejection checks pass.
- Full line-run replay is byte-exact, and original controls/masses are identical
  to the preceding mismatch experiment in every world.
- All3200 Brier values agree with an alternate conditional-risk calculation,
  max error6.11e-16. This verifies scoring algebra, not an independent complete
  inference implementation.
- All1781280 coefficient-law regret checks pass; maximum computed regret is
  .010000000000000342. Counts repeat identical predictions across five scoring
  worlds and are not independent observations or statistical coverage trials.
- 34600 of51200 report-vector evaluations alter the optimistic target; these
  unweighted counts are not real-data activation rates.

[Line results](interval-line-noise.json),
[verification](interval-line-noise-verification.json),
[envelope implementation](noise-envelope.mjs).

## What remains

The interval coefficient hull may be conservative because its vertices need
not each equal a physical noise-conditioned law. Investigation could tighten
that hull, diagnose the numerical projection limit, or use population-risk
allocation while retaining the original .01 harm and .005 gain thresholds.
Independent per-source errors, out-of-interval noise, adaptive acquisition and
real likelihood estimation remain out of scope of this result.

No source-model authenticity claim, production change, performance benchmark,
whitepaper promotion or publication. All seven research directions remain open.

