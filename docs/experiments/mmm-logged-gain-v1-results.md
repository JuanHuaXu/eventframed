# Logged gain v1: valid reference, insufficient precision

The evidence-measurement component passes its algebra, isolation and replay
checks. It does not rescue acquisition quality or provide a useful promotion
certificate at this sample size. All seven research goals remain open.

[Protocol](mmm-logged-gain-v1-protocol.md),
[method and source review](../../research/logged-gain-method.md),
[experiment](mmm-logged-gain-v1.json),
[independent audit](mmm-logged-gain-v1-audit.json).

## What changed

The projection reconstructs realized future Brier from source Y and recorded
branch forecasts for all2688 records. Earlier actual-answer sampled risk used
the realized query answer but averaged future label loss under teacher Q.
Those earlier metrics remain correctly named in their protocols; they are not
equivalent to observable logged future loss.

Three unchanged paid acquisition strategies (random, entropy, joint8) are masked
to one logged action per episode. Exact same-origin strategies collapse to a
single action with known multiplicity/3 propensity. This simulation reuses
consumed fixed potential outcomes; it is not new agent evidence. Sixty-four
logging assignments are not64 independent datasets.

Each replicate uses672 phase0 delayed episodes to fit constant loss regressions
from only logged outcomes, then672 phase1 episodes for evaluation. There are
223 phase1 episodes with identical entropy/joint8 actions. Each replicate costs
1344 paid queries including regression training; acquisition fitting and future
outcome collection are not free. No serving-latency result is claimed.

## Results

Realized full-information phase1 mean Brier:

| Strategy | Brier |
| --- | ---: |
| Random | 0.165777231 |
| Entropy | 0.165044070 |
| Joint8 | 0.165901580 |

Positive gain means entropy loss minus candidate loss. The full-information
gains are-0.000733161 (random) and-0.000857510 (joint8). These are descriptive
consumed-data differences, not population certificates or all-case gate passes.

| Candidate | Logged estimator | RMSE against fixed-table gain | Mean final CS width | Positive final lower bounds |
| --- | --- | ---: | ---: | ---: |
| Random | Importance weighting | 0.013928 | 0.732357 | 0/64 |
| Random | Doubly robust | 0.006717 | 0.615299 | 0/64 |
| Joint8 | Importance weighting | 0.014914 | 0.676489 | 0/64 |
| Joint8 | Doubly robust | 0.007090 | 0.562832 | 0/64 |

The simple regression roughly halves estimator RMSE, but this is improved
measurement, not improved predictions. The bounded Hoeffding-mixture intervals
are far too wide to resolve these small gains. No run produces a positive lower
bound, even at intermediate times. All64 assignments cover every running
fixed-table target in each comparison/method; zero observed coverage violations
does not prove coverage or sharpness. The mathematical justification, bounded
logging assumptions and limits are in the method note.

## Checks

-576 exact DR expectation identities and4608 conditional Hoeffding MGF checks.
-4096 adaptive six-step action/outcome tree leaves, ownership, positivity,
 sequence/atomicity and numeric-extreme regression tests pass.
-249984 forecast/outcome terms and4837 exact action-identity checks pass.
-Throwing accessors on teacher fields, non-outcome source fields and integrated
 branch risks reproduce the entire projection exactly:2688 records.
-Independent audit reconstructs172032 estimator increments and256 final
 intervals, including phase0-only fitted constants and logged action decoding.
 Per-prefix coverage is replayed, not independently reconstructed by that audit.
-Full projection and experiment replays are byte-identical; source hashes kept.

During development, a deliberately extreme betting-rate input caused overflow
and false precision. Supported numerical rates are now bounded, threshold logs
avoid alpha-division overflow, and tiny nonzero ranges round conservatively
upward. The frozen default rate grid did not change; all results above use the
corrected implementation.

## Next lead

The source paper's variance-adaptive time-varying off-policy construction is a
distinct next test; our current reference uses worst-case conditional ranges.
Do not import a constant-mean betting interval into these varying contexts or
weaken the coverage target to obtain a decision. A tighter interval also cannot
turn the observed negative policy gain into a positive one. Policy-learning
improvement and actual prospective logging remain separate unresolved work.
No production, whitepaper, dependency, commit or push changes were made.
