# Population-target critic results

## Verdict

FAIL: all eight primary advancement screens (four frozen models, forced/gated)
fail. All sixteen supplementary sample-expectation/population screens also fail.
Removing sampled-input target noise alone is not a rescue on this consumed
evaluation. No variant establishes positive lower gain bounds against both
matched controls in any phase1 delayed case. Do not promote or tune on phase1.

All2688 records are evaluated per variant. Training uses only672 phase0 delayed
pools,4640 candidate rows, weight672. Phase1 contains672 delayed records. Actual
answer sampled-input Brier is reported below; lower is better.

| Frozen model | Forced | Gated | Entropy, same gate | Paid gated queries |
| --- | ---: | ---: | ---: | ---: |
| Original10, linear | .167362682 | .167220581 | .166436638 | 593 |
| Original10, quadratic | .166847690 | .166611719 | .166313645 | 594 |
| Augmented20, linear | .167245607 | .167146662 | .166468041 | 584 |
| Augmented20, quadratic | .166725865 | .166771714 | .166200005 | 556 |

Forced entropy is .166629983 at672 queries. The quadratic10 mean improves over
its previous sample-target result, but neither mode passes the frozen screen.
The apparently favorable gated-versus-forced mean is not a matched-cost result;
the same-gate entropy column is the relevant comparator.

Population-risk forced means are .166909611/.166818886/.166929407/.167041264,
versus entropy .166794264. Their gated means are .166938036/.166693532/
.166994828/.166955568 versus respective same-gate entropy .166830267/.166647320/
.166872890/.166653885. Even the cleaner evaluation target does not pass.

## Diagnosis

Population-target within-pool R-squared, training -> evaluation:

| Model | Train | Evaluation |
| --- | ---: | ---: |
| Original10, linear | .01812 | .02534 |
| Original10, quadratic | .04669 | .03518 |
| Augmented20, linear | .01902 | .02520 |
| Augmented20, quadratic | .08268 | -.00388 |

The target-only change improves some diagnostics but the models still explain
little candidate-ranking variation. More capacity is not established as the
answer. Next isolate the contribution of query-answer probability error by
reweighting the same counterfactual population branches with strictly as-of
posterior masses instead of teacher probabilities. This remains an oracle
diagnostic because branch risks still use teacher outcomes and future arrivals.

## Verification

[Frozen protocol](mmm-population-critic-protocol.md) and
[experiment](../../research/population-critic-experiment.mjs) preserve features,
normalization, ridge.01, pool centering, tie rule and abstention threshold.
Retraining with old targets reproduces every old model and decision exactly.
Phase1-target poisoning, selector oracle-access traps, complete-delivery identity
and all unchanged controls pass. The original factorial unit suite also passes.

Per-variant JSON artifacts are `mmm-population-critic-v1-{0,1,2,3}.json`, with
`-replay.json`, `-benchmark.json`, `-replay-benchmark.json` and `-audit.json`
companions in this directory. All four summary replays are byte-identical.
The [independent audit](../../research/population-critic-audit.mjs) checks88704
saved loss values and all84 cells per model, including interval arithmetic,
gate masks, selection and screen decisions. Integrated oracle risks themselves
rely on the separately tested Go kernel, not an independent full reconstruction.

Three-fit component timing ranges, milliseconds:9.85-14.78,57.10-65.53,
13.05-16.85,560.52-939.71. Selection p99 on2016 calls with precomputed feature
vectors is .00658/.01638/.00625/.11654ms. These timings exclude Bayesian and
prequential feature production, retrieval, I/O, contention and serving. They
are not comparable to earlier feature-inclusive selection timings as speedups.

An initial aggregate read encountered a still-running fourth output and failed;
the existing process was awaited, then all outputs were read and replayed. No
experiment was restarted because of that premature read.

All seven whole research goals remain OPEN. No paper, production, commit or push
changes. Goal5's read-only model check still exposes only an embedding model;
prospective agent evaluation requires a separately approved completion endpoint.
