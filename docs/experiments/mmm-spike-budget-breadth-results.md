# Eight-index breadth screen: gains persist, residual harm remains

Frozen protocol: `mmm-spike-budget-breadth-contract.md`. Previously consumed
source indices 0-7, all 21 scenarios, both phases and schedules, clocks128-159.
672 records, 21,504 forecasts, 16,847 fresh challenger fits. No hyperparameter
changes. This remains exploratory, not untouched confirmation or deployment.

| Method | Expected Brier | Realized Brier |
| --- | ---: | ---: |
| Markov incumbent | .177166295 | .176743516 |
| Uncombined arrival-refitted challenger | .179502874 | see cadence audit |
| Unguarded feedback mixture | .171984547 | .172261379 |
| Budget-guarded mixture | .172369670 | .172252353 |
| Fixed half-mixture | .173348854 | .173586341 |

Guarded mean expected gain versus Markov is .004796625, with exploratory
trajectory-cluster percentile interval [.004247531, .005330104]. Resampling
uses eight index clusters across all scenarios and schedules, not independent
frames. With only eight consumed clusters, this is not simultaneous assurance
or a substitute for prospective validation. Indices1-7 alone gain .004610945.

The guard limits 5,999/21,504 proposals (27.90%) and improves361/672 records.
It retains about93% of the unguarded average expected gain. Four records,
versus47 unguarded, still worsen expected Brier by more than .01:

| Phase | Scenario | Index | Schedule | Expected harm |
| ---: | --- | ---: | --- | ---: |
| 0 | majority3 | 4 | immediate | .010041923 |
| 1 | mux3 | 2 | immediate | .010578996 |
| 1 | mux3 | 2 | delayed | .011252015 |
| 1 | parity-to-majority | 2 | immediate | .011811517 |

No eight-trajectory scenario mean exceeds .01, but majority/mux mean harms
around .005 have pointwise exploratory intervals above zero. This is evidence
of structured residual harm, not merely numerical noise. The original pilot's
zero-large-regressions observation does not generalize. Do not redefine that
screen as passed by averaging the four failures away.

## Verification and scope

All16,847 fits converge by bound-and-state, maximum409 iterations. Independent
source audit checks every arrival-filtered origin, publication decision,
initial frozen32 fit, final moments and score fields. Maximum moment errors
are1.78e-15(mean) and3.33e-16(variance). Maximum integrand evaluations2536.
Full255-factor predictive integration is not independently reconstructed;
small-mixture component tests remain the separate numerical evidence.

The budget ledger passes every prefix, including retrospectively evaluated
labels unavailable to the algorithm. Largest floating-point boundary excess
is1.32e-16; terminal excess reaches .32 within rounding. This supports the
realized-loss invariant, not a conditional expected-score bound. Independent
batch expert weights agree within3.33e-16. All84 index0 output records exactly
match the original guarded pilot, including forecasts and guard decisions.

Collection took164.23s. Cost includes16,847 fits rather than the672 required
for frozen32. This ideal immediate-publication study is not an equal-compute
comparison or a loaded serving benchmark. Full replay took165.94s and is
byte-identical. Budget replay reproduces all results exactly except elapsed
time. Full spike race suite PASS44.718s. No running sessions remain.

SHA256:

- Cadence raw and replay: `03e7a131e1b9ab0183dfedfb9aabc8382580a4b2b1a74c70bc0a4570f46eab7a`
- Cadence audit: `b249ed5d6e23300a0924867035bb583b6ee071757bc4368f63729f9020367f4b`
- Guarded result: `e20579b9fa258abe4a5c71c386adbfd7518c94aedec64e05fd2065c67ce8d6d2`
- Bootstrap summary: `2db3fb55e6aa82744c22dcf005fad76df2695c6196a2a88fc75b0e7be8f009f4`

Next: evaluate the unchanged mechanism at early/late forecast times and then
continuous operation, where startup and budget accumulation differ. Preserve
the four counterexamples. Do not tune the budget to erase them on consumed
data. All seven goals remain OPEN; this is a stronger partial result for the
challenger/robustness work, not completion of those goals.
