# Sample-count-matched regime diagnostic results

Status: **mechanism evidence, not an online rescue**. The known boundary128 is
hindsight information. None of the seven full research directions is closed.

The [frozen protocol](mmm-regime-match-protocol.md) was evaluated on384 selected
schedule trajectories (192 paired latent trajectories), with three snapshots
each. All1152 snapshots and ten arms are retained in
[raw data](mmm-regime-match-v1.jsonl) and the
[independent summary](mmm-regime-match-v1-summary.json).

## Result

Recent-regime evidence beats eight averaged, equally sized random subsets in
all12 mixed-regime cells: both transition directions, both phases, immediate
clock160 and delayed clocks160/192. Every paired mean-minus3.5SE gain is positive.
This controls sample count, not every covariate/recency difference. The32 units
per cell are trajectories, not the eight random draws or the32 forecast frames.

Phase1, delayed/missing feedback, expected Brier (lower is better):

| Case / clock | Full64 | Current-regime only | Equal-count random | Random minus current interval |
| --- | ---: | ---: | ---: | --- |
| Stationary /160 | .217734 | .242037 | .238477 | -.003561 [-.013194,.006072] |
| Majority to parity /160 | .338282 | .203094 | .277497 | .074403 [.028049,.120757] |
| Majority to parity /192 | .144384 | .052183 | .194568 | .142386 [.103030,.181742] |
| Parity to majority /160 | .347784 | .194155 | .254105 | .059951 [.039674,.080227] |
| Parity to majority /192 | .159083 | .073285 | .188740 | .115456 [.088920,.141991] |

At clock160, delayed full windows contain about50 pre-boundary labels and14
post-boundary labels. At192 they still contain23-25 pre-boundary labels. At224
the contamination is about0.5-1.6 labels; current-versus-random intervals cross
zero. With immediate feedback at192/224 all retained labels are current and
all ten arms agree exactly (summary averaging can leave rounding residuals).

The stationary placebo matters: at delayed clock160, discarding old evidence
increases Brier by.018009 in phase0 and.024303 in phase1, with positive paired
harm intervals in both. There is no justification for unconditional truncation.
The remaining current-only error at14 labels also shows that clean evidence is
not equivalent to adequate evidence.

## Verification

- Race contracts: PASS,41.375s package time. Empty priors, all-current equality,
  six as-of/count-matched snapshots, input ownership and future-label/Q poisoning.
- Collection: PASS,26.89s wall,99.31s user, four workers. Not serving latency.
- Independent direct Beta integrals:368640 probabilities checked, including
  every window's family evidence and weight; maximum recorded marginal/weight
  error7.11e-14. Full-window forecasts match the earlier family artifact.
- Reference identities:25 predictive-ratio checks and five likelihood checks.
- Scoring replay: byte-identical. The summary pins raw/source/scorer hashes.

Intervals are exploratory fixed-sample mean +/-3.5SE, not simultaneous or
anytime certificates. Both phases are now consumed development data. The test
does not isolate all possible effects of recency and does not identify a real
world causal change. No prior tuning, production changes or publication.

## Next lead

Test a bounded one-change posterior against both the no-change model and the
existing full segmentation model. Infer the cut from available evidence; never
supply128. A strong no-change prior and minimum segment support are an explicit
complexity restriction, not a validated safety guarantee. Earlier segmentation
and age-challenger failures remain relevant; do not call this a new invention
or replace them with selected successful cells.
