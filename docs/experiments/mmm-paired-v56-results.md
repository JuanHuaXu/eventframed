# V56 cache-only rescue: exact execution, incomplete cost rescue

## Verdict

The strict execution gate PASSES: all 840 arms match V54 bitwise on every
non-cost field, including the selected observation stream. The full computational
rescue FAILS: six cells exceed the unchanged 400 ms limit. No threshold relaxation,
fresh confirmation, adoption, or whole-goal completion follows from this result.

[Protocol](mmm-paired-v56-protocol.md),
[outputs](../../research/paired-v56-diagnostic/diagnostic.jsonl),
[independent readback](../../research/paired-v56-diagnostic/readback.json),
[bitwise comparison](../../research/paired-v56-diagnostic/equivalence.json),
[source freeze](../../research/paired-v56-diagnostic/freeze.json).

## Scope And Fidelity

This reuses all 40 consumed V54 worlds: 20 regimes, two geometries, three delay
schedules, and seven arms, totaling 840 arms and 96,000 underlying outcomes.
Four policies request 2,800 measurements per arm; Full, Adaptive, and no-pair
have 2,400 and remain resource ablations. Missing and correlated-source stresses
are retained. These are computational comparisons, not independent scientific
confirmation or an equal-total-cost observation experiment.

The separate comparison checked 34,655,560 numeric values, with zero bitwise
differences, zero discrete differences, and zero changed decision batches or
choice arms. All population records match. Only cost fields are excluded.
Unlike V55 smoothing, V56 memoizes the original two hypothetical integrations
without changing their arithmetic, applying current global weights afresh.

## Cost Result

| Arm | Maximum full loop (ms) | Total across 120 cells (s) | Cells over 400 ms |
| --- | ---: | ---: | ---: |
| Full | 61.414 | 6.167 | 0 |
| Adaptive | 253.720 | 24.718 | 0 |
| No-pair | 197.613 | 19.493 | 0 |
| Random | 309.791 | 25.499 | 0 |
| Uncertainty | 487.110 | 42.294 | 4 |
| Information | 441.378 | 42.746 | 1 |
| Falsification | 400.611 | 42.243 | 1 |

The six failures are recorded individually in the readback, including the
falsification cell only 0.611 ms above the boundary; it still fails. Constructor
allocation is 3,625,744 bytes, below 8 MiB; this is not a resident-memory bound.

Cross-run total time falls 7.14%, 7.26%, and 7.73% for uncertainty, information,
and falsification relative to V54. However, proposal time does not fall: it
increases 1.00%, 0.72%, and 0.027%, respectively. Repeated query reuse saves work
at request time, but most selection queries have a newly issued member history
and miss the cache. Aggregate elapsed differences also include shared-host
variation. Do not attribute every timing change to memoization or infer a
population speed guarantee from fixed-order single-run timings.

Falsification still costs 65.66% more than random at the same request budget.
Its tiny risk advantage does not establish Goal 7's equal-total-cost superiority.
These offline loops are not loaded serving latency or a 100 ms freshness result.

## Scientific Result And Audit

All original scientific failures remain unchanged: each selection policy has
17 cells with more than 0.01 Adaptive harm and misses the 0.01 Full-relative
mean-gain criterion. Falsification improves recovery versus no-pair in one of
60 shift cells but remains 0.716667 rounds slower than Adaptive on average.
Its issued Brier is 0.217772260109, terminal Brier 0.188171962460, and top-10
utility 0.809809663696. Exact preservation is not a quality improvement.

Eight model roots and three fixture roots pass under race. Original path and
independent-update checks, 15 future forks, 38 corruption rejections, and
3,840 seed / 23,040 channel separation checks are retained. The full independent
auditor reconstructs all 840 trajectories; separate readback verifies metrics,
cost accounting, and source closure. All seven serialized commands terminate 0.
65 compiler/test inputs plus 15 support files were frozen and copied before
execution. All 14 pre-existing tracked edits remain hash-identical.

Production, private data, untouched task cohorts, the whitepaper, and publication
are untouched. No commit, push, installation, or deployment. All seven whole
goals remain OPEN and the persistent goal remains ACTIVE.

## Next Research Lead

[V57 predictive observation value](../../research/paired-v57-predictive-value-direction.md)
is a recommendation, not a confirmed defect in the concentration criterion.
Finite joint-table identities demonstrate that global-class concentration can
be settled while an additional measurement still improves model-predicted
future Brier risk. Controlled implementation, external-loss evaluation, delay
protection, and total-cost measurement must test whether that matters here.
This does not erase V54/V55 failures or authorize opening sealed confirmation.
