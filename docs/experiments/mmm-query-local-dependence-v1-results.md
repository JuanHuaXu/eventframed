# Four-cell dependence shrinkage: no change, no rescue

Follow the [frozen protocol](mmm-query-local-dependence-protocol.md): split only
on as-of query uncertainty and exact target-input membership in retained observed
support. Fit four conditional-log-loss mixing weights on phase0 actual labels.
No case IDs, teacher probabilities or phase1 labels enter training or the gate.

## Result

Every cell selects lambda1, retaining original dependence:

| Query uncertainty | Observed input match | Training pairs | Weight | Derivative at1 |
| --- | --- | ---: | ---: | ---: |
| Low | No | 40134 | 192.170737 | -0.466817 |
| Low | Yes | 5746 | 27.517358 | -0.067624 |
| High | No | 86464 | 399.345392 | -1.283414 |
| High | Yes | 11496 | 52.966513 | -0.206014 |

All four right-boundary derivatives are negative. Convexity proves the constrained
optimum at1, independently of optimizer implementation. The fitted law is
therefore identical to the original everywhere; none of672 phase1 query choices
changes. Both forecast screens and both selection screens fail.

Phase1 actual-publication selection, actual-answer sampled Brier:

| Selector | Brier | Paid queries |
| --- | ---: | ---: |
| Random | 0.167334 | 672 |
| Entropy | 0.166630 | 672 |
| Original joint8 | 0.167366 | 672 |
| Four-cell calibrated joint8 | 0.167366 | 672 |

The forecast diagnostic also exactly reproduces dependence-v1: joint log loss
0.988835 and target-expected Brier0.168410, versus independence0.993932 and
0.170226. Those means do not satisfy all-case improvement requirements.

## Evidence and Cost

All2688 records and84 cells are retained, including complete histories with no
candidate pairs. There are287587 pairs total; training has143840 pairs from672
histories. They are not independent samples. The synthetic archive's missing
query labels remain declared offline supervision, not free online feedback.

Unit tests cover known four-cell optima, exact as-of support, duplicate-input
invariance, future/teacher access traps, ownership, fallback and invalid input.
Independent audit reconstructs cell assignments and convex endpoint derivatives,
9277 acquisition scores,8064 reference forecast values,1092 aggregate means and
504 paired bounds. All pass; full replay is byte-identical.

Selection with precomputed laws and the observed-input set,2016 calls on Node
v26.8.1: median0.003125ms, p99 0.006167ms, maximum0.704875ms. This excludes fits,
retrieval, observed-set construction, persistence and serving; it is not an
end-to-end latency result and brings no quality gain.

Artifacts: `mmm-query-local-dependence-v1.json`, `-replay.json`, `-audit.json`,
`-benchmark.json`, and `-replay-benchmark.json`. Source and implementation hashes
are recorded. Data remain consumed, not fresh confirmation. No production or
whitepaper promotion.

## Next Lead

Both global and four-cell shrinkage prefer the largest allowed strength. This
supports testing the other direction, not assuming it will succeed: a low-
capacity, phase0-trained amplification constrained by the valid Bernoulli joint
probability region. Preserve both marginals and strict positive probabilities;
derive the admissible bound before fitting. Do not extrapolate the current
mixture beyond1 without those checks, add outcome-specific cells, or claim that
negative training derivatives establish evaluation gains. Keep unchanged
actual-publication query-selection evaluation separate from modified-forecast
diagnostics. All seven goals remain open.
