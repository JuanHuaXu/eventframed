# Segment mixing without a pointwise restriction

Frozen before this screen. Consumed independent-v1 cohort only, not fresh
confirmation. This is an ablation of the optional guard, not a relaxation of
the original statistical acceptance requirements.

Use existing P10 segment64, P13 static64, and P12 Markov forecasts unchanged.
Four arms: segment/Markov delayed Fixed Share, static/Markov delayed Fixed
Share, segment/Markov fixed half, static/Markov fixed half. Delayed Fixed Share
keeps the existing uniform prior, eta=1, alpha_j=1/(j+1), with alpha_0=0.
No new fit, parameter search, source fields, labels, or scenario-specific route.
Store each forecast and score expected/realized Brier, whole and terminal64.

For each arm retain the v120 comparison structure: 672 non-harm comparisons
against generic64, Boolean64, Markov and static64; 96 recovery gains against
generic64, Markov and static64. Non-harm upper <=.01, gain mean >=.005 and
lower >0. Report mean +/-3.5SE over the eight existing indices, explicitly NOT
the original 32-index confirmation or simultaneous coverage. Report stationary
and changing pooled means, all individual gates, and 32-frame harm counts.
Neither an overall mean win nor a zero-harm-window count substitutes for gates.

Current/future/unavailable-label poisoning and equal-forecast controls precede
scoring. Verify raw controls against the archived source metrics and guarded
screen. Full numerical replay must match. Do not infer deployment safety from
an empirical screen. This ablation may fail and all failures must remain.

Time the mixing work separately from source I/O and scoring; fit cost remains
the original 545.32s source collection. No loaded-serving claim. Keep production,
Go runtime, private data, whitepaper and remote repositories untouched.
