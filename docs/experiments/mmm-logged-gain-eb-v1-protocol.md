# Logged gain empirical-Bernstein comparison

Freeze the existing logged-gain-v1 projection,64 logging assignments, phase0
regressions, phase1 evaluation, action collapsing and four comparison/method
coverage allocations. No new policy, outcome, seed, regression, gate or target.
Replace only the Hoeffding range-process CS with a discrete mixture based on
the empirical-Bernstein martingale in Waudby-Smith et al. section3/AppendixA.3.
This is not their continuous gamma-mixture closed form or an optimized grid.

For prepared gain X, require the entire predictable support within[-3,3].
Use fixed scale4 and predictable center clipped prior sample mean in[-1,1],
initial0. Accumulate V=sum((X-center)/4)^2 before updating the center. Mix
rates[1/128,1/64,1/32,1/16,1/8,1/4,1/2] equally, psi=-log(1-lambda)-lambda.
Invert mean(exp(lambda*b-psi*V))=2/alpha; radius=4*b/n. Intersect only with
[-1,1], not previous intervals. Return an exact interval only if every prepared
increment has singleton support, not merely because observed residuals are zero.

Alpha remains.05/4 per comparison/method with two-sided internal allocation.
Report running-average conditional gain coverage, final widths and unchanged
point errors versus the original run. A narrower interval is a measurement
result, not policy improvement. No post-result rate/scale changes.

Require exact finite MGF checks for both signs/predictable centers, adaptive
tree coverage and a positive-signal power control; sequence/atomicity/range
checks; identical original point estimates/training/assignment controls;
independent prefix variance and interval reconstruction; deterministic replay.
Retain negative policy gains and consumed-data scope. No deployment or paper
promotion, no production calls, dependencies, commits or pushes.
