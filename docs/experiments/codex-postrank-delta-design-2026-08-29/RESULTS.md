# Post-Contract Rank-Delta Retest

## Scope

This development-only retest repeats the frozen 146-case slice from three
independent Codex-session clusters used by the earlier LibraVDB contract retest.
It uses the same four-session cap, one 100-turn segment per session, embedding
model, calibration parameters, and chronological design assignment. There are
no confirmation cases in this bounded slice.

Every query used LibraVDB 1.9.36-beta.5 through the public
`SearchTextCollections` and `RankCandidates` contracts. Unlike the previous
implementation, LibraVDB received its nomination/base score, and EventFrame
applied its bounded Bayesian/residual rank delta only after `RankCandidates`
returned and before packing.

## Results

| Variant | Brier | Recall@10 | Recall@50 | MRR | Packed recall | Result |
|---|---:|---:|---:|---:|---:|---|
| LibraVDB contract baseline | 0.23397 | 41.97% | 91.72% | 0.7711 | 28.93% | Control |
| Current EventFrame | 0.18555 | 51.93% | 96.05% | 0.8846 | 38.96% | Supported on design slice |
| Contextual composition | 0.18056 | 51.98% | 96.16% | 0.8863 | 39.28% | Best point estimate; not promoted without confirmation |
| Full upgrade | 0.18392 | 50.82% | 95.91% | 0.8858 | 37.71% | Mixed; below current EventFrame retrieval |
| Weak hierarchy | 0.19050 | 50.83% | 95.68% | 0.8829 | 37.52% | Mixed; below current EventFrame |
| Shuffled feedback | 0.25099 | 42.40% | 91.89% | 0.7997 | 29.15% | Negative control; no comparable lift |

Relative to the LibraVDB baseline, current EventFrame gained:

- `+0.09958` Recall@10, cluster-bootstrap 95% interval
  `[0.04241, 0.16151]`;
- `+0.10031` packed recall, cluster-bootstrap 95% interval
  `[0.09363, 0.11039]`;
- `+0.04336` Recall@50;
- `+0.11350` mean reciprocal rank; and
- `+0.04842` Brier improvement, cluster-bootstrap 95% interval
  `[0.03761, 0.05765]`.

The shuffled-feedback control remained close to the unlearned retrieval output
and worsened Brier. The corrected retrieval improvement therefore depends on
aligned historical outcomes rather than merely adding a deterministic score
perturbation.

## Interpretation

The earlier contract test correctly showed probability calibration improvement
but incorrectly concluded that the tested EventFrame layer did not improve
retrieval output. The implementation had passed the adjusted score into
LibraVDB and then overwritten it with LibraVDB's returned score. This retest
supports the intended architecture: LibraVDB nominates and base-ranks;
EventFrame applies a separate learned delta; packing sees the final order.

This is strong design evidence, not confirmation. Only three independent
clusters contribute to the bootstrap, no untouched confirmation block is
present, and the contextual variant's slightly better point estimate must not be
promoted from this run alone. The external LibraVDB daemon also emitted deferred
micro-temporal stale-generation maintenance warnings during the replay; the
synchronous search and ranking contracts continued to complete.

## Reproducibility

The frozen parameters and model identities are in `protocol.json`. Per-case
candidate laws, rankings, and packing outputs are in `design-dataset.json`.
Aggregate metrics and cluster-bootstrap comparisons are in
`design-report.json`; the negative-control comparison is in
`design-control-report.json`.
