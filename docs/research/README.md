# Research Track

This track is experimental and separate from the daemon's deployed defaults.
The local research archive includes unsuccessful approaches; their protocols
and findings remain part of the record. Curated publication omits private or
unreviewed source text, machine-local build records, copied backend overlays,
and duplicated archives. An excluded artifact is not silently treated as a
passing experiment.

## Current Baselines

The prospective accuracy reference is **Adaptive**, selected from the complete
consumed native cohort under its original **400-ms whole-core budget**. **Full**
remains the speed control. This is constrained empirical selection, not a claim
that one model is both globally fastest and most accurate.

| Control | Expected Issued Brier | Fresh Mean Core ms | Fresh Worst Core ms |
| --- | ---: | ---: | ---: |
| Adaptive | 0.214747855 | 197.731 | 246.690 |
| Full | 0.224699096 | 49.285 | 73.614 |

The replication covers 40 consumed development worlds, three delay schedules,
240 control arms and 576,000 independently recomputed issued losses. The risks
reproduce the prior controls within 1e-12. Every fresh cell stays under 400 ms.
These measurements cover serial synthetic loops of 2,400 events, not loaded
100-ms agent serving, independent confirmation, or general real-task success.

The policy is executable and scoped:

```sh
go run ./cmd/eventframe-research-baseline
go test ./internal/researchbaseline -count=1
```

The [registry](../../research/baselines.json),
[frozen selection protocol](../../research/publication-2026-10-05/BASELINE-PROTOCOL.md),
and [comparison evidence](../../research/publication-2026-10-05/selection-results.json)
record the criterion and scope. Future experiments should use this primary
reference while retaining Full and matched acquisition controls. Historical
frozen protocols retain their original baselines.

V83 is the latest numerically repaired regime research implementation. Its
prepared-state query performance is promising, but it has no equivalent broad
quality result and is not promoted over Adaptive. All seven whole goals remain
open; no production deployment or automatic discovery claim follows.

## Older Findings

- [Early challenger and retention experiments](../experiments/mmm-retained-v8-results.md).
- [Anti-Pigeon coverage and downstream limitations](../experiments/mmm-member-fit-breadth-v1-results.md).
- [Joint-model quality failures](../experiments/mmm-mean-anchor-v75-cohort-results.md).
- [Equivalent computation with measured cost](../experiments/mmm-dynvarcache-v72-results.md).
- [Protected-support development comparison](../experiments/mmm-regime-protected-v81-screen-results.md).
- [Long-history numerical repair](../experiments/mmm-regime-log-v82-results.md).

The publication audit documents omissions and limitations rather than claiming
that signature scanning proves the absence of every possible privacy leak.
