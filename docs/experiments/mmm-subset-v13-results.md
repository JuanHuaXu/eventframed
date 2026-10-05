# Bayesian subset challenger v13: pilot PASSED

Both adaptive and static retained subset challengers pass every frozen mean
screen across576 fresh streams (294,912 frames). This is a bounded synthetic
pilot, not proof of population non-inferiority, real-task transfer or production
readiness. No prior or threshold was tuned between design and confirmation.

Confirmation shift128 post-change Brier, adaptive retention:

| Inputs | Fixed counts | Retained forest | Retained subset average |
| --- | --- | --- | --- |
| Fair | 0.217602 | 0.202284 | 0.165494 |
| Biased | 0.187375 | 0.176203 | 0.160849 |
| Clustered | 0.140733 | 0.138810 | 0.114863 |

The replacement averages conditional label models instead of selecting one
greedily fitted tree. It retains uncertainty about which subset matters, with
a fixed complexity prior and Beta-smoothed conditional cells. The same64 audit
labels, six-coordinate limit and incumbent protections remain in place.

Passing does not mean improvement everywhere. The largest retained-arm harm
versus fixed is .004399 (static, clustered interaction confirmation post).
The largest harm versus the corresponding forest is .005456 (adaptive, biased
delayed/missing design post). Both are below the predeclared .01 allowance but
still matter for subsequent testing. Null-outcome forecasts remain near .25
Brier; that is not a demonstrated calibration guarantee.

## Verification

- Sequential evidence agrees with closed-form Beta integrals for every subset;
  no binomial count factor is added to the labeled-sequence likelihood.
- Normalization, sample-order invariance, input validation, conditional
  integration and immutable snapshots pass targeted tests.
- Original forest controls match exactly; fixed-count forecasts, audit decisions
  and delayed delivery remain matched across modes.
- All576 records replay exactly excluding fitting-time instrumentation.
- Raw summary reconstructs full/post Brier and accuracy, checks source hashes,
  observation budgets, acquired values, pairing and label timing.
- Targeted race tests and command vet pass. Previous artifacts remain unchanged.

The artifact field `uniform` in the summary denotes the paired original forest
mode's metric; the candidate mode is `subset`. Legacy tree counters are zero in
subset mode and must not be interpreted as its model size. The model explicitly
contains512 subsets and uses at most3^9 conditional states.

[Standalone benchmarks](mmm-subset-v13-benchmark.txt): a complete64-label fit
and table build takes5.90-5.95ms and allocates651,265 bytes on Apple M4. This is
not a hot-path or loaded-worker measurement, and no comparison with table-only
construction should be read as a full algorithm speed ratio.

## Remaining Work

Use more independent fitted incumbents and new label families before treating
this as robust evidence. The nine-bit representation still does not solve real
text feature collisions. Integrate only behind research controls after real-task
validation, and measure actual slow-worker costs under load. Adaptive windows,
Anti-Pigeon gate integration, provenance uncertainty and agent outcomes remain
separate open directions. Do not promote this result into those claims.

Artifacts: [protocol](mmm-subset-v13-protocol.md), [raw](mmm-subset-v13.json.gz),
[summary](mmm-subset-v13-summary.json).

```sh
go run ./cmd/eventframe-observation-subset NEW.json.gz
python3 research/subset_summary.py NEW.json.gz NEW-summary.json
EVENTFRAME_SUBSET_ARTIFACT=NEW.json.gz go test ./internal/observationlearners -run '^TestSubsetArtifactReplay$' -count=1
```
