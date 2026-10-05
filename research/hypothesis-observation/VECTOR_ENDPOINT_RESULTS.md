# Full-vector endpoint diagnostic: a genuine information ceiling

The [endpoint diagnostic](VECTOR_ENDPOINT_PROTOCOL.md) passes862/900 checks
(FAIL). All800 population protection and50 false-confidence requirements pass,
but38 genuine-gain requirements fail. Endpoint emphasis sacrifices other
regimes; it is not minimax optimization or a successful replacement.

| Actual noise | Pass /180 | Gain failures /10 | Worst population harm |
| --- | ---: | ---: | ---: |
| .10 | 170 | 10 | .008143 |
| .15 | 170 | 10 | .009267 |
| .20 | 170 | 10 | .009661 |
| .25 | 173 | 7 | .009944 |
| .30 | 179 | 1 | .010000 |

Nine endpoint gains now pass, compared with five in the fixed-line endpoint
family. The remaining endpoint failure is more important than the gate count.

## Unrestricted Bayes-risk ceiling

For counts[1,1,3,1], actual noise=.30 and all-genuine reports, the solver's
unconstrained optimum is already feasible (sweep0, gap0, maximum population
coefficient harm .008143). An independent Python calculation directly
enumerates all1024 report vectors and16 latent hypotheses, recomputes the
fixed-.20 local baseline and the TRUE conditional outcome law under the
actual .30 all-genuine process:

| Quantity | Brier |
| --- | ---: |
| Local baseline | .538247041749787 |
| Known-noise, known-mechanism oracle | .537985805114445 |
| Maximum possible improvement | .000261236635342 |
| Required improvement | .005 |

For any forecast p(x) based only on these observations X, conditional Brier
risk is minimized by q(x)=P(Y|X=x). Specifically,

    E[L(p,Y)|X=x] = 1 - ||q(x)||^2 + ||p(x)-q(x)||^2.

A randomized forecast cannot improve this conditional minimum either. Therefore
the oracle's population risk lower-bounds EVERY forecast on the same observed
history, not just a line family, a Bayesian target, or a protected solver.
It even knows the actual noise and mechanism. No correction-only policy can
satisfy the .005 gain requirement in this cell.

The independent oracle matches all4096 archived forecast coordinates within
3.33e-16 and matches both scalar risks within1e-12. The ceiling is a numerical
evaluation of the exact finite Bayes-risk identity, not an interval-arithmetic
proof; its large shortfall is not plausibly explained by roundoff.

[Independent calculation](endpoint_oracle_ceiling.py),
[oracle evidence](endpoint-oracle-ceiling.json).

## Consequence for the research

The full900-gate fixed-observation screen is unattainable as stated. Preserve
that failure; do not lower its threshold, remove the allocation or relabel a
partial pass as success. Further changing correction objectives cannot rescue
this particular requirement without changing the information available.

This does NOT rule out EventFrame, the seven research goals, or faster learning
through better observations. It points back to goal7's original equal-cost
acquisition comparison: learn whether the observation policy can collect more
useful evidence for the same total cost. Any new study must keep controls and
costs explicit and must not masquerade as passing the old screen.
The result also does not prove this schedule's absolute accuracy is sufficient:
a near-optimal baseline can still have high error when observations are weak.

## Verification and limits

All ten endpoint solves finish within1080 sweeps with unchanged tolerances.
Full optimization replay is byte-exact; original controls match all worlds.
Alternate scoring agrees on3200 values within7.77e-16;800 coefficient/direct
regret checks agree within2.35e-14. Existing four analytic, twelve feasible-grid
and two rejection checks pass. This is not an independently implemented
general optimizer, but the key oracle ceiling has a separate direct-product
implementation that does not use that optimizer.

Worst SINGLE history/outcome loss increase is .309129; population protection
still provides no per-event or priority-sensitive guarantee.

[Full results](vector-endpoint-experiment.json),
[verification](vector-endpoint-verification.json).
No real-world confirmation, performance benchmark, production changes,
whitepaper promotion or publication. All seven research directions remain open.

