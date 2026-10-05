# Continuous-noise risk diagnostic for a frozen policy

Freeze the budget policy and hashes from risk-budget-evaluation.json. Construct
population final and pre-query-average Brier polynomials in the common noise
parameter on [.05,.35], for each16 masks and policies risk_budget, random,
entropy, tie_entropy. Forecasts and action decisions remain fixed.

Each ordered-history joint class mass is a polynomial of degree at most10:
four root likelihoods and at most six fresh-report likelihoods. Copied reports
are indicators, not independent factors. Use the existing Bernstein likelihood
builder, degree-elevate to10 by multiplication by1, and sum squared forecast
loss weighted by policy path multiplicity. Do not insert multinomial factors.

Gain polynomials are control minus candidate. Restrict the Bernstein curve by
de Casteljau subdivision to [.1,.3], [.05,.1], [.3,.35]. Its minimum coefficient
is a lower bound on the exact polynomial and its maximum an upper bound. Use
ten levels of midpoint subdivision to tighten global bounds, without selecting
evaluation points based on an observed failure.

This implementation uses ordinary floating-point arithmetic, not outward-rounded
intervals. Its bounds are numerical diagnostics, not formally certified signs
for the exact real-arithmetic model. Report values within1e-12 of zero as
numerically unresolved/tied, never as strict learning gains. No success criteria
change or promotion follows. The finite mathematical envelope is exact in real
arithmetic, but arithmetic validation remains a separate obligation.

Verify degree elevation/subdivision on known polynomials, compare all archived
on-grid and off-grid risk scores against polynomial evaluation, and replay
byte-for-byte. The archived scorers use direct likelihood multiplication rather
than the Bernstein construction. All previous failed screens remain recorded.
