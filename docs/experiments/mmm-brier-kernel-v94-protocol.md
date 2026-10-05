# Online Brier kernel v94 protocol

Frozen before execution. This validates a numerical/lifecycle research kernel,
not prediction quality on the existing learned-model datasets. All whole-goal
and v88-v93 criteria remain open and unchanged.

Two experts, initial masses(0.95,0.05), single-coordinate Brier loss(p-y)^2,
learning rate eta=1/2. One prediction may be outstanding. Its outcome must be
applied exactly once before another prediction. No cancellation, skipped label,
retroactive prediction, restart or future outcome. Instances are single-owner.

Test no-sharing and fixed sharing rho=0.001. After each observed outcome, form
the normalized exponential-loss posterior, then mix(1-rho)posterior+rho*initial
prior for the NEXT prediction. No-share uses retained log weights so numerical
underflow in a displayed probability does not erase a recoverable expert.

## Explicit bounds

For every prefix N>=1, the generic expert's never-switched path has prior mass
at least0.95*(1-rho)^(N-1). Concavity of exp[-eta(p-y)^2] and the potential
argument in research/online-brier-aggregation-proposal.md therefore give

sum(loss_mix-loss_generic) <=
[log(1/0.95)+(N-1)*(-log(1-rho))]/eta.

The added term is a declared cost of sharing, not the original no-share bound.
At rho0 the bound is0.1025866. At rho0.001 its average at N16 is approximately
0.008288. These are cumulative realized-loss bounds, not per-event or future
population-risk certificates. No statistical claim about true model identity.

## Frozen stress panel

32 paired seeds, seven cases,4096 sequential steps, both variants: random
forecasts/random outcomes; common adaptive adversary maximizing sum of both
variants' one-step excess over generic; fixed opposite extreme experts with
alternating outcomes; one regime flip at2048 with opposite extreme experts;
predictable changing experts; generic always better; challenger always better.
Each seed uses base2042119400 + case*10000 + index*10, role0 forecast RNG,
role1 outcome RNG. Both variants see identical experts/outcomes. The adversary
chooses its outcome only AFTER both forecasts and may inspect them.

Store448 variant records (224 paired streams): cumulative losses, maximum
regret, maximum prefix-bound violation, maximum Jensen violation, final weights,
post-flip accuracy and first post-flip challenger-dominant prediction lag, plus
prediction/outcome tape hashes. Assert every prefix bound within1e-8 and every
one-step Jensen inequality within1e-12. These are numerical tolerances, not
confidence bounds. Do not count deterministic duplicate cases as independent
statistical confirmation. Measure recovery; do not infer it from regret alone.

## Unit and negative controls

Reject invalid/NaN probabilities, invalid sharing parameters, wrong sequence,
duplicate feedback, double issuance, zero-value state and counter overflow
without mutation. Returned forecasts must not permit mutation of stored
evidence. Test log-space recovery after >1000 losses of an expert, unchanged
state on invalid calls, and predictable expert changes. Check all forecasts
before reading outcomes. A deliberately skipped-feedback counterexample must
violate the no-share bound, while the real API rejects that lifecycle.

Run source-hashed generation and exact replay, focused race tests, vet and
same-fixture per-step benchmarks. No production, durability, concurrent access
to the same instance, adaptive observation policy, missing/delayed-label,
real-data quality, commit or push claim. Next comes a separately frozen
prospective learned-model experiment, not automatic adoption of the kernel.
