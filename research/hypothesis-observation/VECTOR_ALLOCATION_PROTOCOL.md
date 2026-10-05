# Frozen full-vector population allocation

The endpoint dual bounds excluded five allocations for the fixed-line family.
Replace lambda interpolation with a full four-class probability vector for
each of1024 histories. Keep the .20 local baseline, all176 Bernstein population
regret constraints, .01 risk allowance, interval[.10,.30], all ten allocations
and all900 gates. Optimize average all-genuine regret under uniform noise,
matching the earlier average objective rather than choosing on test outcomes.

A coefficient row j has history masses m_jx, joint class masses c_jxy and
constant k_j chosen so baseline regret is zero:
  R_j(p)=sum_xy[m_jx p_xy^2-2 c_jxy p_xy]+k_j.
m_jx=sum_y c_jxy, all nonnegative. Constrain R_j<=.01 for each row.
With nonnegative dual multipliers, the Lagrangian minimizer for history x is
  p_xy=(c_objective,xy+sum_j mu_j c_jxy)/
       (m_objective,x+sum_j mu_j m_jx).
This is a normalized probability vector; no arbitrary negative weights.
The baseline is feasible. Uniformly contract numerical primal violations toward
baseline and require repaired-primal minus dual bound <=1e-8, with risk
<=.01+1e-12. Use5000-sweep cap and the existing coordinate/bisection approach;
no quality or numerical tolerance changes. Failure terminates the run.

No actual evaluation noise, copy mask or latent h enters the fitted history
mapping. This is finite design against an assumed law family, not a posterior
authenticity certificate or real-world calibrated uncertainty estimate.
The final forecast is a constrained decision law, not ordinary Bayes.

Check analytic binary examples, two-history feasible-grid problems and input
rejections. Verify coefficient/direct risk and original controls; independently
rescore all forecasts and replay optimization. Preserve all failed gates.
Measure worst history/outcome harm: population protection is not per-event
protection. No empirical latency claims, production changes or publication.

