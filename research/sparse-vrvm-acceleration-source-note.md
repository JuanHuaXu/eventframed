# Fixed-point acceleration lead (not implemented)

Primary source read: Du and Varadhan, *SQUAREM: An R Package for Off-the-Shelf
Acceleration of EM, MM and Other EM-Like Monotone Algorithms*, JSS 92(7), 2020,
[published article](https://www.jstatsoft.org/article/view/v092i07),
[author manuscript section 2 / Table 1](https://arxiv.org/html/1810.11163).
This builds on Varadhan and Roland (2008), not a newly invented accelerator.

Source mechanism: two fixed-point evaluations give r=F(z)-z and
v=F(F(z))-2F(z)+z. Extrapolate z-2 alpha r+alpha^2 v, with the discussed
choice alpha=-norm(r)/norm(v). Compare the objective with the ordinary
two-step result and stabilize accepted proposals. At alpha=-1 the proposal
equals the two ordinary steps. The source distinguishes near-monotone defaults
from strict monotonicity, and counts mapping evaluations rather than outer
iterations. Its smoothness/contraction and convergence assumptions cannot be
inherited by calling our variational loop EM. No reported source speedup is
an EventFrame result.

## Proposed application and its unresolved obligations

For this variational model, let z contain positive Gamma rates and xi; the
Gaussian factor G(z) is the exact conditional coordinate optimum. Let F(z)
update rates and xi using G(z). Define the profiled objective
E(z)=ELBO(G(z),z). The old coordinate inequalities imply
E(F(z)) >= ELBO(G(z),F(z)) >= E(z), in exact arithmetic.
This makes a full-state objective available, unlike comparing an extrapolated
rate vector with a Gaussian fitted against unrelated rates.

A proposed log-coordinate accelerator would preserve positivity; it remains
our adaptation and needs round-trip and domain checks. Every objective trial
requires a Gaussian solve unless that exact state is already cached. Charge
all trials, fallback steps and the final coherent Gaussian. Reject nonfinite
states and compare against ordinary-step progress, not just the previous
objective. A finite rejection budget must fall back without accepting a
worse or partially updated state. Model priors and evidence stay unchanged.

Before testing quality: verify the profiled-objective inequality, alpha=-1
identity, matched ordinary trajectory, independent final-state reconstruction,
rejection/fallback branches, and equal Gaussian-evaluation accounting. Match
the old returned coordinate phase explicitly; its recorded Gaussian precedes
the final Gamma/xi update. Do not silently score a different state as though
it were the original 64-step control.

This is a source-grounded lead only. The ongoing 1,024-step diagnostic should
be completed, audited, and replayed first. No global-optimum, convergence-rate,
calibration, or forecast-improvement claim has been established.
