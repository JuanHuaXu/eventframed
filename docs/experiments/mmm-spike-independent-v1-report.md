# Independent replication: partial success

The policy was frozen before new fitting/outcome draws and randomized Boolean
shift timing. All 672 trajectories and 172,032 forecasts were collected; no
policy parameters were retuned. This cohort is now consumed and must not be
represented as untouched evidence for a subsequent rescue.

## Scores and failures

Lower Brier is better. Harms count paired expected-score regressions above
.01 versus Markov, not observed-label loss-budget violations.

| Policy | Whole expected Brier | Terminal-64 harms | 32-step harms |
|---|---:|---:|---:|
| Fixed Share, local ledger | .155144958 | 1 | 34 |
| Reset weights and ledger every32 | .155162845 | 5 | 58 |
| Markov | .157516545 | reference | reference |
| Static weights, global ledger | .155544500 | 23 | 170 |
| Static weights, local ledger | .155442733 | 2 | 34 |
| Fixed Share, global ledger | .154993853 | 22 | 179 |
| Raw challenger | .164432730 | 190 | 1544 |

The frozen combined policy improves whole-stream expected Brier by .002371587
(1.5056%) versus Markov. Its eight-index cluster bootstrap pointwise95% interval
for this paired difference is [-.002716930,-.002130823]. Realized Brier also
improves, .156815822 to .154546025.

Versus reset32, the whole-stream difference is -.000017887 with interval
[-.000080482,+.000049982]: no clear whole-stream win. Terminal64 improves
by .000317147 versus reset32, interval[-.000387355,-.000248122], and by
.000505167 versus Markov, interval[-.000647900,-.000354365]. These intervals
are pointwise, not simultaneous safety guarantees.

Regime stratification prevents an adaptation overclaim. On changing scenarios,
the combined policy is worse than reset32 over the whole stream by .000305957,
interval[.000181469,.000435767]. Its changing-scenario terminal difference is
+ .000102048, interval[-.000022087,.000209098]. The aggregate terminal gain
over reset32 is driven by stationary cases, not demonstrated faster recovery.

No combined-policy whole-trajectory expected harm exceeds .01; maximum .007403243.
But34 of5376 fixed windows exceed .01, worst .026819805. One of672 terminal64
records exceeds .01: phase0/case20/index6/immediate, parity-to-majority shift
at171, terminal excess .010601298. This trajectory has no capped fit. No
scenario-window mean across eight indices exceeds .01. These observations
do not establish uniform or population non-harm.

Local ledgers substantially reduce large regressions relative to global
ledgers, but switching does not improve the short-window failure count over
static/local in this cohort. The unguarded challenger remains worse than
Markov; the result is evidence for the guarded combination, not a superior
standalone learner. All seven research goals remain open.

## Audits and cost

- Fresh fitting:132,772 fits in1305.30s, no old cache reuse. Source baseline
  collection was a separate545.32s. These are offline reference costs.
- Two retained1024-iteration caps, one forecast each: phase0/case19/index5/
  immediate at29 (motion8.41e-6), and phase0/case14/index1/immediate at189
  (motion1.14e-6). Small bound movement did not count as state convergence.
- All eight raw-window audits passed as-of origins, convergence declarations,
  source matching, moment reconstruction, and quadrature-budget checks.
  Maximum reconstructed mean discrepancy1.78e-15. The full predictive
  integral is not independently recomputed by these moment audits.
- Scoring replay reproduced all results and summary values exactly except
  elapsed time. Baseline and challenger fitting were not fully replayed.
  All283 fit-manifest code hashes match current source.
- Summary fixture checks passed known constant deltas, cluster counts,
  pointwise intervals, and incomplete-cohort rejection.
- Warm filter-plus-local-guard replay:1.4095--1.4161 microseconds per forecast
  amortized, across three full672-trajectory rounds. Per256-step trajectory
  p95 .435--.438ms. This excludes model fits, retrieval, persistence, network,
  and loaded serving. Reference filtering remains O(T^2). It is not online
  tail latency or evidence that goal6 is complete.

## Reproducibility

Raw fits: `mmm-spike-independent-v1-fits-clock{0,32,64,96,128,160,192,224}.jsonl`.
Audits: `mmm-spike-independent-v1-audit-clock*.json`.
Protocol and source provenance remain in the earlier status document.

| Artifact | SHA256 |
|---|---|
| fits.jsonl | f5c92edac9045eeb790b0101f6cea6a29d55895ca08a2558ac0b227e7af060cb |
| fits-manifest.json | 4de683d6bb6938a05ffb71265e7b458f9c72783c495c816b7a4e8006a8e32156 |
| results.json | 2271d42d0387edbef26982c076fa00335bf2f7f49c15f4201d80d81e231dc038 |
| summary.json | 570b80e1849268fddd1e0ab04426995c44f04631225388a5255d1d91da5f9ff1 |
| benchmark.json | 6e63a5d0a3a856047c0595822f59241a094ec2f025dd3904e3cf8900a29ce8a5 |

All artifact names above use prefix `mmm-spike-independent-v1-`.

Next: diagnose the now-consumed conditional failures by separating raw learner
error, expert allocation, and ledger allowance. Do not raise iteration limits
or tune switching/budget constants to erase the uncapped counterexample.
Any revised policy needs a new frozen evaluation. Production, whitepaper,
commits, and publication remain untouched.
