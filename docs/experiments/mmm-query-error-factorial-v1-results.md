# Error-state by model-form factorial: all screens fail

The [frozen factorial](mmm-query-error-factorial-protocol.md) tests whether
historical error information or bounded nonlinear interactions rescue the
query-value critic. Each model trains on phase0 only and evaluates all phase1
cases without tuning. Both phases were previously consumed.

## Policy results

Phase1 delayed actual-answer expected Brier, lower is better:

| Features / model | Forced critic | Gated critic | Same-mask entropy | Gated queries |
| --- | ---: | ---: | ---: | ---: |
| Original / linear | 0.167148793 | 0.167068293 | 0.166443884 | 588 |
| Original / quadratic | 0.166980725 | 0.166650785 | 0.166315713 | 593 |
| Error-informed / linear | 0.167430259 | 0.167303530 | 0.166390921 | 572 |
| Error-informed / quadratic | 0.166668430 | 0.166808596 | 0.166191594 | 560 |

Ungated entropy has Brier0.166629983 at672 queries; ungated random0.167334488.
All forced models spend672 queries. Every gated comparison uses the exact same
trajectory mask for critic, random and entropy, not merely the same total cost.

All four forced and all four gated advancement screens FAIL. No forced model
establishes the two-control non-inferiority lower bound in any of21 cells.
Gated models establish it in2/1/2/0 cells respectively. None establishes the
required switching-case gains or positive lower bounds against both matched
controls in any cell. Failure of these confidence screens does not prove every
model harms every case, nor does near-equality of pooled means prove equivalence.

The error-informed quadratic model's descriptive forced mean comes closest to
entropy, but its outcome-averaged mean0.167252511 remains worse than entropy
0.167019776. It is not a validated rescue. Its gated mode loses to matched
entropy by0.000617003 in actual-answer Brier despite saving112/672 queries.

## Generalization diagnostic

Within-pool value R-squared, training versus evaluation:

| Features / model | Phase0 | Phase1 |
| --- | ---: | ---: |
| Original / linear | 0.01390 | 0.00875 |
| Original / quadratic | 0.03880 | 0.00579 |
| Error-informed / linear | 0.01442 | 0.00866 |
| Error-informed / quadratic | 0.08029 | -0.03991 |

The largest model improves training fit but loses that advantage in evaluation.
This is a warning against increasing capacity or tuning thresholds on these
same examples. Between-pool evaluation R-squared remains positive, but does
not establish reliable ordering of candidate queries within a pool.

## Verification and cost

All four variants retain2688 records and84 phase/case/schedule cells. Base-linear
exactly reproduces the old centered model and decisions. Old controls remain
unchanged. Phase1-target poisoning leaves models and choices unchanged.
All main summaries and variance diagnostics replay byte-for-byte.

Polynomial basis dimensions10/55/20/210 are checked; an independent pivoted
solver checks an expanded55-column fixture. Full-fit normal-equation residuals
are below5.3e-18. Ownership, target-access traps and detached replay pass.
The error projection retains its previously verified as-of boundary.

| Features / model | Fit range, ms | Selection median / p99, ms |
| --- | ---: | ---: |
| Original / linear | 8.09-12.39 | 0.02263 / 0.04954 |
| Original / quadratic | 58.76-117.57 | 0.04125 / 0.11446 |
| Error-informed / linear | 14.21-27.35 | 0.02804 / 0.06829 |
| Error-informed / quadratic | 421.56-512.06 | 0.10979 / 0.18012 |

Fit ranges cover three repetitions; selection covers2016 warmed calls each.
These timings exclude creating Bayesian bundles and prequential projections,
retrieval, parsing, persistence, queues and contention. Fast component selection
does not establish end-to-end latency or justify deploying a failed policy.

## Next distinction

The oracle targets still condition on the particular31 future inputs and on
natural arrivals at publication. Earlier counterfactual averaging removed paid
answer luck, not this remaining future-information advantage. Before another
critic search, measure opportunity averaged over the generator's full input law
instead of its particular future input sample. This is an offline oracle
diagnostic, not a way to supply future truth to an online policy. It can help
separate noisy/clairvoyant targets from missing usable state information.

Artifacts use prefix `mmm-query-error-factorial-v1-` with variants0..3, their
`-replay`, `-variance`, `-variance-replay`, and `-benchmark` JSON files. All seven
whole goals remain open. No production, paper, commit or push changes.
