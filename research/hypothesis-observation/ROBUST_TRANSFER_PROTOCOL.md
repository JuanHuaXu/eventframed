# Conditional robust transfer guard

Freeze before quality scoring; retain every allocation and mask from the failed
180-gate stress test. No tuned mixture prior, cost, noise or gate tolerance.

General inspiration is finite ambiguity-set robustness, as surveyed by
[Rahimian and Mehrotra](https://arxiv.org/abs/1908.05659). Only that broad concept
is attributed here; the following Brier line-segment guard is our direct
derivation, not an inherited DRO theorem or a verified source-authentication rule.

Given local forecast p0, proposed hierarchical forecast p1 and d=p1-p0, let Q
contain the conditional target law under EACH declared counterfeit mask with
nonzero likelihood for the acquired report vector. These laws come from the
declared measurement model, uniform hypothesis prior and observed reports.
No true mask is supplied. Exclude a mask only for exactly zero likelihood, not
small estimated probability. This is not the simulator's single chosen oracle.

For p(lambda)=p0+lambda*d, conditional expected multiclass Brier regret against
p0 under q is a*lambda^2+b_q*lambda, where a=||d||^2 and
b_q=2*d dot(p0-q). Set b=max_q b_q and choose the largest lambda in[0,1] with
a*lambda^2+b*lambda<=.01. Zero correction is always feasible. Use the positive
quadratic root, with stable evaluation and conservative roundoff handling.

If the actual conditional target law is in Q (or its convex hull), the guard
bounds conditional regret by.01, hence integrated regret by.01. It does NOT
control realized individual losses, classification errors or model misspecification.
Wrong likelihood/noise, unknown source roots or omitted masks void that premise.
Extending to adaptive histories requires modeling the actual acquisition and
arrival process; this test uses the fixed paid schedules only.

Evaluate all10 renewal allocations and16 masks, same180 gates as before, but
replace hierarchical with its guarded proposal. Retain genuine gains>=.005 and
all-counterfeit false-confidence reductions>=.05; identity fallback alone is
not success. Record clipping frequency, retained correction, worst conditional
regret, all final risks and gates. Exact enumeration, independent bisection unit
tests and full replay required. No production/paper promotion from this pilot.

For M laws and C classes, guard work is O(M*C), but constructing Q is a separate
cost. Enumerating M counterfeit patterns scales exponentially in the number
of measured source types; it is not a corpus-independent free certificate.
