# V49 normal hybrid: complete negative result

2026-10-04. Both original V48 normal cohorts executed with the exact V49
constructor optimization; no runtime forecast/update change or cohort fitting.
[Protocol](mmm-memo-v49-normal-protocol.md).

`research/memo-v49-normal/` preserves 126 frozen sources/copies, 11 terminal
code-zero commands, 28 hashed execution artifacts, all four raw JSONL tapes,
independent source/artifact audit, metric readback, and descriptive summary.
Race, actual seed separation, vet, complete original control audit, and unchanged
dense candidate replay pass. Parallel reference checks are excluded from
scientific timing; controls and candidate collection remain serialized.

Design base 2026104809 and untouched confirmation base 2026104811 each contain
448 worlds: two geometries, 14 regimes, 16 worlds per geometry/regime. Each
world supplies three evidence schedules and three fixed candidate policies.
Per cohort: 4,032 candidate arms, 67,200 snapshots, 1,075,200 distinct labels.
Across cohorts: 896 worlds, 8,064 arms, 134,400 snapshots, 2,150,400 labels.
Labels are counted once per world, not once per reused control/policy arm.

## Unchanged acceptance outcome

Original per-cell paired mean +/- 3.5 SE screens and all cost gates remain.
One SAME fixed policy must pass all 84 cells in BOTH cohorts. No resampling,
policy switching, margin adjustment, or post-design optimization was performed.
This predeclared finite study screen is not population-wide certification.

| Cohort | Policy | Whole failed cells / 84 | Quality-only failed cells / 84 |
| --- | --- | ---: | ---: |
| Design | static | 29 | 27 |
| Design | slow | 28 | 26 |
| Design | round | 28 | 26 |
| Confirmation | static | 30 | 21 |
| Confirmation | slow | 29 | 20 |
| Confirmation | round | 29 | 19 |

All policies fail adoption. Quality-only removes Work/Allocation checks, not
the original score, usefulness, priority, or recovery requirements. Every policy
fails recovery in 17 cells in each cohort. Other failures include the required
shifted score gain and protection against Adaptive or final-packet harm.
Per-check failure counts overlap; they must not be added as disjoint failures.
The raw `readback.json` retains every cell's interval and pass/fail outcome.

## Descriptive gains

Equal-cell mean issued Brier gains; positive means improvement. These are
absolute loss differences, not accuracy percentages or simultaneous intervals.

| Cohort | Policy | vs Full | vs Adaptive | Shifted cells vs Full |
| --- | --- | ---: | ---: | ---: |
| Design | static | 0.011713 | -0.000764 | 0.018052 |
| Design | slow | 0.011751 | -0.000726 | 0.018104 |
| Design | round | 0.011889 | -0.000588 | 0.018301 |
| Confirmation | static | 0.011881 | -0.000715 | 0.018328 |
| Confirmation | slow | 0.011886 | -0.000710 | 0.018330 |
| Confirmation | round | 0.012011 | -0.000585 | 0.018506 |

Mean final usefulness versus Adaptive is negative for every policy/cohort:
design -0.002072/-0.001703/-0.001534; confirmation
-0.001579/-0.001511/-0.001520 (static/slow/round). Broad Full gains do not
erase these failures or establish untouched agent-task benefit.

## Cost outcome

The repeated nine-loop preliminary paired cost check still favors the exact
memo constructor: original maximum 407.523625 ms; memo 372.115584 ms.
Memo constructor allocation is 7,886,136 bytes, below the unchanged 8 MiB cap.
However full scientific collection maxima are 432.736167 ms (design) and
426.834250 ms (confirmation), above the unchanged 400 ms complete-loop screen.
The cell-wide Work gate fails 5 design and 13 confirmation cells. It is shared
across the three policies, exactly as in the original readback.

These are complete 2,400-observation loop times, NOT single-request latency,
loaded serving tails, RSS bounds, or a 100 ms robotics/agent guarantee. The
[previous exact constructor saving](mmm-memo-v49-results.md) remains measured,
but its diagnostic screen pass did not provide broad worst-case headroom.

## Next Lead

Preserve this negative result. Loss/objective alignment is a hypothesis worth
testing with a NEW bounded Brier-aligned integration, not a proven root cause.
The [isolated primitive](mmm-brier-v50-preflight-results.md) has correctness
checks, but no broad quality result. All seven whole goals remain open/active;
production, private corpora, whitepaper, commits and publishing were untouched.
