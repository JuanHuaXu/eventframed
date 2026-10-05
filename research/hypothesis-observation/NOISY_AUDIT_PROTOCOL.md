# Equal-cost noisy source audit pilot

New information model, not a rescue of the certified-infeasible old model.
All arms receive one mandatory binary source0 audit after the four root reports.
Its sensitivity and specificity for root-copying are both.8, independent of
latent target and report noise given the copy mask. This is a hypothetical
measured channel, not a implemented/validated authenticity detector. It reports
no target bit and does not reveal the true mask. All arms model it identically.

The audit costs2 credits, leaving five renewals at2 credits each: total16 with
the four roots at1 each. Random renewal selection, predictive-report entropy
and one-step target-Brier selection have equal audit access and total cost.
There are six pre-action loss measurements: before audit, then before each
of five renewals. Final loss is measured after the fifth renewal. No free
audit, extra renewal, or post-audit evidence in the pre-audit forecast.

Use the existing continuous uniform noise prior [.1,.3] and existing mask prior.
Multiply joint evidence masses by the declared audit likelihood before
normalizing; retain1e-12 lowest-index tie handling. Actual evaluation covers
five noise points and all16 masks, integrating both audit outcomes exactly.
Final nonharm remains.01 versus both controls; area gain must be positive except
mask15. This is a pilot of declared policies, not joint optimality. Failure
does not prove the new observation model infeasible.

Checks: audit outcomes marginalize to the original posterior; reliability.5
leaves forecasts unchanged; independent quadrature of the uniform-noise model;
probability/cost accounting and replay. No actual provenance pipeline or empirical
reliability estimate is claimed. No production or archived source changes.
