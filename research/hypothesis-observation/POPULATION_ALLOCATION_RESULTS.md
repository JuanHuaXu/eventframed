# Population-risk allocation: additional gains, important local tradeoff

The [frozen design](POPULATION_ALLOCATION_PROTOCOL.md) allocates the same .01
population harm allowance across1024 observed report histories. It optimizes
uniform-noise all-genuine average regret while protecting all coefficient laws
of the continuous[.10,.30] common-noise interval and all sixteen copy masks.

This is a finite schedule-conditioned design against declared laws, not fitting
with access to the actual runtime noise, latent hypothesis or copy mask.
The history-to-lambda rule is solved before scoring and is the same in all
evaluation worlds for that schedule.

## Numerical completion

The original500-sweep run terminated with feasible repaired primal but gap
1.40538e-7, above1e-8; [failure preserved](population-allocation-failure.json).
The [separate5000-sweep budget](POPULATION_ALLOCATION_LONG_PROTOCOL.md) changes
only the maximum offline work. All ten solves finish in189-1207 sweeps, with
gaps below1e-8 and coefficient risk<=.01+1e-12. No quality or numerical accuracy
threshold was relaxed. This is not a solver throughput benchmark.

## Results

Overall888/900 checks pass (FAIL). All800 population protection and50
false-confidence requirements pass;12 genuine-gain checks fail.

| Actual noise | Pass /180 | Worst population Brier harm | Genuine gain range |
| --- | ---: | ---: | ---: |
| .10 | 180 | .000239 | .008909 to .010317 |
| .15 | 180 | .003834 | .007822 to .009374 |
| .20 | 180 | .007133 | .005341 to .009194 |
| .25 | 176 | .009377 | .002009 to .007626 |
| .30 | 172 | .010000 | -.002241 to .006007 |

The prior uniform-target conditional line guard passed873/900: the new
allocation rescues15 further gates. Remaining failures are four gains at .25
and eight at .30. Some genuine .30 cases get worse, though not beyond the
population harm allowance.

## Protection is now population-level, not per-event

Maximum SINGLE history/outcome Brier loss increase is .334452, not .01.
The .01 contract bounds an expectation under each covered generating law.
It does not protect every individual memory, outcome, rare subgroup or
high-priority case. The worst pointwise value is not an average conditional
risk or an observed deployment frequency.

This tradeoff is explicitly permitted by the original population test but
may be unsuitable for high-consequence applications without extra safeguards.
Do not advertise the result as pointwise safety, a causal guarantee, or a
general authentication mechanism. Outcomes or source processes absent from the
declared model remain outside its coverage.

## Verification

Full unmodified optimization replay is byte-exact. All original control scores
and probability masses match the earlier artifact in all worlds. An alternate
Brier identity using frozen fitted lambdas agrees on3200 values, max error
8.88e-16; all gate decisions agree. Its reuse of fitted lambdas is deliberate:
full optimization was already replayed separately, and this check isolates
scoring algebra rather than independently solving the optimization.

Six analytic scalar cases, twenty feasible-grid comparisons and one rejection
test pass for both solver budgets. All176 coefficient history masses normalize
to one per allocation. Eight hundred direct population regrets match polynomial
evaluation, max error7.06e-15. These numerical checks are not formal
interval-arithmetic verification or empirical coverage estimates.

[Results, lambdas and solver bounds](population-allocation-long.json),
[verification](population-allocation-verification.json),
[solver](population-risk-allocation-long.mjs).

## Next lead and limits

The average-loss objective leaves all failures at the noisy end. Test a
worst-regime objective or explicitly balanced gain constraints while retaining
all existing protection and quality gates; an average optimum is not a
guarantee of uniform usefulness. This is a separate research rescue, not grounds
to discard noisy cases or lower the .005 minimum.

Model-family uncertainty, unseen observation structures, adaptive acquisition,
learned likelihood coverage, priority-sensitive protection and loaded serving
remain open. Exhaustive enumeration over1024 histories is not a scalable daemon
implementation. No production or whitepaper changes, benchmark claims or
publication. All seven whole research directions remain open.

