# Rank-Certainty Modulation: Synthetic Regression

Date: 2026-08-29

This regression reruns the frozen 50-candidate/10-packed synthetic mechanism
suite after replacing calibrated-usefulness modulation with rank-boundary
certainty multiplied by independent correction reliability.

## Confirmation results

| Criterion | Result |
| --- | ---: |
| Promote useful item from ranks 11/15/25/40/50 | 200/200 (100%; lower 95% CI 98.12%) |
| Demote harmful item from ranks 1/3/5/7/10 | 200/200 (100%; lower 95% CI 98.12%) |
| Joint promotion and demotion | 200/200 (100%; lower 95% CI 98.12%) |
| Retain useful controls | 90/90 (100%; lower 95% CI 95.91%) |
| Cross deliberately wide correction envelope | 0/50 (upper 95% CI 7.13%) |
| Mean unnecessary packet churn | 0 |
| In-process active recall p99 | 0.891 ms |

The mechanism remains validated on its constructed corpus: uncertainty-sensitive
scaling did not weaken bidirectional correction, evict known-useful controls, or
bypass the hard correction envelope. This test does not establish performance on
organic queries; that evidence is in the paired Codex replay.

The design block passed every point criterion. Its aggregate `overall_passed`
field is false only because the predeclared retention lower-bound threshold is
not reachable with the smaller 30-case design sample; the untouched 90-case
confirmation block passes it.

Artifacts: `dataset.json` and `report.json`.
