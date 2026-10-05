# Direct mixture screen: no quality rescue

Read the primary Vovk algorithm and declared the adaptation differences in
`mmm-direct-mixture-v1-contract.md`. Tested four fixed scalar proposals on the
consumed independent-v1 cohort. No new fitted forecasts, seed draws, or
parameter sweep. The original paper's guarantee is not claimed.

| Guarded proposal | Whole expected Brier | Terminal64 | >.01 harmful32-windows |
|---|---:|---:|---:|
| Ridge all-history | .156891775 | .145333413 | 0 |
| Ridge last64 | .156838173 | .145279261 | 0 |
| Current-covariance all-history | .156891388 | .145333167 | 0 |
| Current-covariance last64 | .156841153 | .145286929 | 0 |
| Fixed Share pointwise control | .156847006 | .145296192 | 0 |
| Fixed Share local-ledger control | .155144958 | .145014962 | 34 |
| Markov | .157516545 | .145520129 | reference |

Ridge64's whole difference versus Fixed Share/pointwise is -.000008833,
pointwise eight-index bootstrap95% interval[-.000031607,+.000011000]. Its
terminal difference -.000016931 also crosses zero. These are exploratory
consumed-data intervals, not a fresh or selection-adjusted result.

All four variants are worse on changing scenarios over the whole stream.
Ridge64's difference there is+.000078040, interval[.000032600,.000118638].
Its small pooled advantage is driven by stationary scenarios. Do not call
this faster adaptation. Unguarded direct mixtures score better on average
(.154180571 for ridge64) but have293--302 harmful windows across the variants.
The safety/utility tradeoff remains.

## Verification and runtime

38784 direct objective comparisons passed, alongside current/future/missing
outcome invariance, exact retained-set completeness, boundary clipping,
empty evidence, equal predictions, and a hand-calculated ridge/current-input
distinction. Strengthened a retention test that previously only checked
validity of included origins, not completeness. No behavioral bug found.
The full672-record screen replayed byte-identically. All15972 pointwise
guard checks still pass.

Warm proposal+guard replay per forecast, amortized: ridgeAll .671--.672us,
ridge64 .480--.492us, currentCovAll .705--.745us, currentCov64 .547--.565us.
These reference implementations scan histories and materialize origin lists:
O(T^2), not O(1) online code. Fits, I/O and serving are excluded. Runtime
alone does not justify promotion of a worse changing-regime predictor.

## Constraint saturation and remaining headroom

Ridge64 changes proposed weights by .11051 on average relative to Fixed
Share, but changes final forecast probabilities by only .002548 on average.
Its86149 clipped forecasts include82915 whose final probability matches a
simultaneously clipped Fixed Share forecast (within1e-12). The fixed bound
erases many differences between mixing algorithms.

An evaluator-only conditional minimizer uses the inaccessible synthetic q to
choose the best weight within the fixed safe segment. Whole Brier .155316261
is worse than the old local-ledger score .155144958. Thus no weight-selection
algorithm restricted to these same forecast segments and this bound can
fully recover that old score on this cohort. However, the safe oracle is
substantially better than current safe methods (~.15684). Terminal oracle
Brier .143966692 also leaves headroom. This is an infeasible ceiling, not a
new predictor, proof that context is learnable, or empirical validation.

Next: a small context-dependent mixture, using only current observable frame
features and arrived outcomes, could test whether some headroom is learnable.
Freeze feature scaling, regularization, retention, and controls before scoring;
keep the bound unchanged and preserve a matched global-mixture control.
Do not relabel this consumed cohort as new confirmation. Do not prioritize
additional scalar hyperparameter sweeps or deploy these failed rescues.

## Artifacts

`mmm-direct-mixture-v1-screen.json`, `-summary.json`, `-benchmark.json`,
`-constraint-audit.json`, and `mmm-pointwise-oracle-headroom-v1.json`.

Core SHA256: ebdeb03d6f1ceb99ded5c06603a9b413bb1ef298599959b683ec425426c65902.
Screen SHA256: d58606dff72a9dd4a177afd8bcf04d8dc6a5c7f38491f2acb24c37449f0bca63.
Summary SHA256:808471c60365c8ae38823162887ea28b8154636632dc14ebae315b542a16ee41.

All seven goals remain open. Research-only JavaScript; no Go/runtime,
production, whitepaper, commit, or push changes.
