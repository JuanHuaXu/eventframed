# Audit-only onset nomination, consumed-data screen

Freeze before scoring. Use all 128 archived trajectories, immediate and delayed
schedules, both stationary controls, and both shift directions. Read only the
already-paid, independently nominated full audits that are nonmissing. The
per-audit bounded score is reference correctness minus live correctness, in
{-1,0,1}. The frozen model score provenance must be reconstructed first.

Use the origin-finalized stream from `mmm-audit-localization-stream-v1.json`.
At wall clock t the newest allowed delayed origin is t-31. An immediate audit
may enter at its origin. Do not consult the simulator's rule, change time,
unreleased observations, or other arm's data when forming the signal.

Start E=1 and multiply by (1+0.5*W) once per finalized audit in origin order.
Nominate on the first E>=20 (nominal alpha=.05 under a zero-mean martingale
assumption). No reset, parameter sweep, or split authorization. This narrower
zero-gap null is not the Anti-Pigeon tolerance null and cannot take its place.
Report detection clock and origin, number of audit updates, pre-change/stable
nominations, detection by511/543, onset-to-detection delay, and comparison to
the existing external gate. Record null/alternative strengths; do not infer a
finite-sample real-world guarantee from a finite synthetic run.

Assess incremental CPU/memory for the bounded update separately from data
preparation. A successful nomination would still require a stopping-aware
onset confidence set and independent forward predictive validation before
any model reset. All cases remain in the report. This is consumed exploration,
not fresh confirmation.
