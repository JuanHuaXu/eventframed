# Exact advice versus idealized symmetry

The log-domain repair makes the failed original regression predict
0.50000000071831197 instead of0.000001, while the other eight root tests pass.
The unchanged test expected ideal decimal0.5 with3e-12 tolerance. That expected
value ignores that actual Float64 `1-(1-1e-6)` differs from1e-6. The failure is
preserved and is not solved by increasing tolerance.

The next regression computes the independent static batch log likelihood from
the actual two advice values using100 successes and100 failures, then takes
their exact Bernoulli mixture. Same tolerance remains. Also add200 delayed
original-law checks and a subnormal-prior expert revival case. This is a test
oracle correction, not a new model/hazard/prior selected from outcome cohorts.
