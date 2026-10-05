# V82 summary/initialization cost follow-up

Freeze AFTER bounds-repair terminal unit/race/vet/bench, before this attempt.
Existing initial failed and bound-repair successful sources/logs remain immutable.
All7 WHOLE goals OPEN. This is an algebraic cost rescue, not a new learner.

Confirmed prepared recent queries: V82 about2.94/4.40ms vsV81 about.84/1.40ms;
older49/66ms vs31/42ms. New code uses costly log-domain member motion merely to
emit a FLOAT summary, and runs an empty replay with allocated scratch buffers
to construct each starting prior. Those operations are visible in code, but no
exclusive profiler-based dominance claim is made. Do not drop long-history
data, protected paths, log probabilities or source/epoch boundaries for speed.

For OUTPUT ONLY, compute sigmoid(z) without taking log first, and compute the
next mean as (1-lambda)*sigmoid(z)+lambda*priorHigh. This is the same Bernoulli
mixture mean as sigmoid(movedOdds), but the former never changes stored odds.
Use stable sign-based logistic; summary rounding cannot feed into the posterior.
The empty-history cold result can be constructed directly without scratch replay:
its forecast is the class-prior mean, independent of hazards. Preserve validation,
parts, component counts, empty identity, support representation and caches.

Falsifiers: persistent log weights/odds or input masks changed; >2e-11 ordinary
independent/dense/law/point/cache/tower difference; long oracle failure at5e-8;
impossible evidence revived; bad empty identity; fixture omission. Keep ALL
existing tests and negative controls; same-command serial benchmarks. Output
last bits may differ; no bitwise rank/order/tie or whole-cohort equivalence claim.
Record old/new costs and allocation, not only successful recent query timing.
No quality gate, population, controls or scientific acceptance rule changed.
Original native-control/fitting/missingness/interaction/whole-core/equal-TOTAL-cost
requirements remain, along with agent utility and durable mixed-write freshness.
