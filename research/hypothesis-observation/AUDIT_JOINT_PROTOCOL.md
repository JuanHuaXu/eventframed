# Joint optimization with the mandatory noisy audit

Keep the entire noisy-audit observation contract and its controls fixed: four
roots, one mandatory2-credit audit, five2-credit renewals, total16 credits.
Candidate and random/entropy controls have the same channel and budget. The
audit sensitivity/specificity remains assumed.8, not measured or implemented.

Optimize both forecasts and actions with the previous128-round, rate32 joint
risk game. Use the same160 final rows (allowance.01) and150 strict area rows.
Do not import the old no-audit certificate into this changed model.

Represent each root pattern as one pre-audit state followed by two audit-result
subgraphs. The pre-audit state has unconditioned joint masses and a single
forecast; its only transition is the mandatory audit, summing both outcomes.
Post-audit states multiply their joint mass by audit likelihood, and have five
remaining renewals. Their level is1+renewal count, terminal at6. Six pre-action
losses are charged, so the audit does not create a free observation.

Reuse the linear joint oracle by giving the forced node four identical action
aliases, all leading to the same pair of audit children. Verify this identity;
an alias must never select an audit outcome. There is no action that can skip
the audit or condition the pre-audit forecast on its result.

Report numerical primal/dual bounds and original pass counts, not assumed
convergence. Replay, graph accounting, audit marginalization and a separate
recursive optimized-risk calculation are required. This is consumed-model
research; no actual provenance accuracy or production performance is claimed.
