# Inner arrival-credit isolation

Reuse192 consumed coupled-arrival-v1 trajectories/both schedules. Add one full
arrival arm to original, outer-only coupled, and fixed-view endpoints. Preserve
those three endpoints exactly, including masks, forecasts, labels and fit sets.

Full arm updates outer and available inner selectors from journaled pre-outcome
advice when the label arrives. Missing labels never update. An external split
caps the outer long slot and rejects outstanding pre-split credits; publication
alone does not discard experimental role losses. The fixed-share likelihood
primitive and all observation/audit/gate budgets are unchanged. Original inner
selector remains copied only in the outer-only arm.

Record all four arms, scores/costs and exact endpoint checks. No fresh confirmation,
new thresholds or seed replacement. Primary diagnosis: full minus outer-only.
Retain stable/null controls and missing outcomes in scoring. This does not
resolve the older cross-generator failures or establish posterior calibration.
