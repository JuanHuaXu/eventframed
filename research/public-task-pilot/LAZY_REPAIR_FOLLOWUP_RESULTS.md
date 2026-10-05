# Fixed lazy-repair follow-up results

PASS for all8 size/repeat comparisons in the four predeclared additional pairs.
All eight test arms completed in the frozen alternating order before evaluation.
Every initial/final graph hash matched its paired control. Thresholds were not
changed: delete ratio<=0.90; seed/insert ratios<=1.05.

| Pair | N | Seed ratio | Insert ratio | Delete ratio |
| --- | ---: | ---: | ---: | ---: |
| 2 | 800 | 0.9841 | 1.0452 | 0.1264 |
| 2 | 6400 | 1.0050 | 0.9524 | 0.2647 |
| 3 | 800 | 1.0077 | 1.0188 | 0.1350 |
| 3 | 6400 | 1.0064 | 0.9799 | 0.2696 |
| 4 | 800 | 1.0003 | 1.0200 | 0.1108 |
| 4 | 6400 | 1.0059 | 0.9675 | 0.2567 |
| 5 | 800 | 1.0141 | 1.0115 | 0.1251 |
| 5 | 6400 | 1.0047 | 0.9893 | 0.3039 |

Ratios are candidate/control timed totals, not population effect estimates.
Observed deletion reduction is69.6-88.9%. Maximum seed regression1.41%; maximum
insertion regression4.52%. All timing and correctness details remain in the raw
pair2-5 files under `lazy-repair-cost/`; `check-lazy-repair-followup.mjs` regenerates
`followup-comparison.json` from those files.

## Interpretation and boundary

The original two-pair screen remains FAILED because one seed observation exceeded
its ceiling. This follow-up is additional evidence on the same designed workload,
not independent unseen task data and not permission to erase the failure. Six
pairs in total still provide neither a tail-latency guarantee nor whole-system
non-inferiority across platforms/workloads.

Together with exact-state mutation checks and dense/sparse/mixed branch controls,
the follow-up supports retaining lazy matrix preparation in the research private
update design. It removes unused neighbor-repair computation, not the requirement
to implement private graph edits, coherent metadata, durability, byte budgets,
and the original sustained offered-load validation. Production and standard
dependencies remain unchanged. All seven whole research goals remain open.
