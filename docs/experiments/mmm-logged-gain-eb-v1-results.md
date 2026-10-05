# Logged gain EB v1: narrower, still inconclusive

**PASS measurement comparison; no policy promotion.** The empirical-Bernstein
mixture substantially narrows the earlier range-based intervals while preserving
the target, logging assignments, regressions, point estimates and coverage level.
It still cannot resolve the small policy differences in this data.

[Protocol](mmm-logged-gain-eb-v1-protocol.md),
[method and source mapping](../../research/logged-gain-eb-method.md),
[artifact](mmm-logged-gain-eb-v1.json),
[independent audit](mmm-logged-gain-eb-v1-audit.json).

## Results

Same consumed2688-record projection,672 phase0 training episodes,672 phase1
evaluation episodes,64 masked logging assignments and1344 paid queries per
assignment including training. The64 assignments are not64 new datasets.
Each comparison/method retains alpha=.05/4 with two-sided internal allocation.

| Candidate versus entropy | Estimator | Previous mean CS width | EB mean CS width | Positive lower bound at any time |
| --- | --- | ---: | ---: | ---: |
| Random | IPS | 0.732357 | 0.198922 | 0/64 |
| Random | DR | 0.615299 | 0.173644 | 0/64 |
| Joint8 | IPS | 0.676489 | 0.196087 | 0/64 |
| Joint8 | DR | 0.562832 | 0.172096 | 0/64 |

DR interval widths fall about69-72%. Point estimates and their RMSE are exactly
unchanged, so this is not better forecasting or reduced estimator error.
Realized full-table gains remain negative: random-0.000733161 and
joint8-0.000857510. Widths around0.17 remain far too large to resolve them.
No assignment violates running-average coverage in any comparison/method;
this finite observation is not a proof of coverage or a prospective guarantee.

## Verification

-Exact original assignments, training counts/regressions, point estimates,
 full-table truths, per-case results and final point errors match.
-3456 conditional MGF checks,4096 adaptive tree leaves and positive-signal,
 zero-residual/non-singleton, range, sequence and atomicity controls pass.
-Independent audit reconstructs172032 increments and prior-center variance
 updates,172032 prefix decisions and256 final interval inversions. Its psi
 evaluation uses a separate power series, not the implementation's log1p.
-Full experiment replay is byte-identical. Original projection isolation and
 replay remain available separately; no source or future outcome was changed.

This is a fixed-grid adaptation of the source martingale, not its continuous
gamma-mixture implementation. The whole predictable gain support must lie
within[-3,3]; the fixed scale4 and pre-outcome center are load-bearing. No
end-to-end runtime or production-throughput claim is made.

## Decision

Retain the tighter bounded measurement primitive, but do not tune its grid or
coverage level on these outcomes. Confidence tightening cannot create a
positive acquisition-policy gain. Future work should compare the acquisition
cost and variance of paired audits against single-action logging, or improve
the underlying policy using separately gathered evidence. A paired audit must
charge both distinct queries and keep branch learners isolated; existing
full-information simulator tables are not free prospective counterfactuals.
All seven whole goals remain open. No production, whitepaper, dependencies,
commits or pushes changed.
