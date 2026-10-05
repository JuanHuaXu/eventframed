# Cold prepared-prefix service test

Actual persistent service,64 recalls/four readers,16 future-dated writes,
100ms age,capacity16,temporal reuse,GOMAXPROCS4. Three alternating table/prefix
pairs for each of two workloads. Distinct=false means64 new terminal outcomes
sharing63 ordered labels. Distinct=true also changes an earlier eligible input
per request, invalidating every prefix. No completion cache in either arm.

Prefix preparation occurs inside processor context on first use or mismatch.
Reference forecasts are verification only and do not populate prepared state.
Eligibility is selected from current fixture packets and checked again by the
full fitter. The single worker owns one retained prefix (>5.5MB); failed or
canceled preparation yields no replacement. Preparation allocations are not
bounded by retained memory: failed rebuild attempts may generate GC pressure.

Record preparations as Fits, prefix reuse attempts as Hits, and verified
returned forecasts as Checked. Reuse attempts can be canceled; these counters
are not additive completed forecasts. All-invalidating workload must have zero
Hits. Track processor deadlines and all offered terminals including drops.

Screen unchanged: candidate>=52/64 completed in every pair, at least control,
p99<=1.10*control, zero request errors and overlap. No off-control inference,
population tail guarantee or general streaming learner validation.
