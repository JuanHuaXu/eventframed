# Publication timing isolation: no rescue

## Scope

Follow the [frozen protocol](mmm-query-publication-protocol.md) on all 2688
already-consumed histories. This is a mechanism diagnostic, not untouched
confirmation. Ten previously frozen policies select exactly the same queries.

- A: latest 63 decision-time labels, plus the paid answer.
- B: latest 64 decision-time labels, retaining 64 after a paid answer.
- C: unchanged actual publication, including newly eligible evidence.

For all 9277 paid candidates, A and B have identical conditioning supports.
Restoring the reserved slot therefore changes only the no-query baseline, not
the relative performance of forced-paid policies. A is not an acceptable weaker
baseline for claiming a rescue.

## Results

Phase-1 delayed evaluation: 672 histories. Sampled-input expected Brier uses
the actual acquired answer in all three states; lower is better.

| Policy | A | B | C |
| --- | ---: | ---: | ---: |
| No query | 0.170226 | 0.170187 | 0.168800 |
| Random | 0.168323 | 0.168323 | 0.167334 |
| Entropy | 0.167730 | 0.167730 | 0.166630 |
| Joint 8 | 0.168616 | 0.168616 | 0.167366 |
| Disjoint 8 | 0.168778 | 0.168778 | 0.167744 |
| Visible 161 | 0.169075 | 0.169075 | 0.168058 |
| Recent 32 | 0.168781 | 0.168781 | 0.167742 |
| EPIG 8 | 0.168758 | 0.168758 | 0.167511 |
| EPIG 161 | 0.168966 | 0.168966 | 0.167939 |
| EPIG 32 | 0.168534 | 0.168534 | 0.167479 |

All seven candidate rules fail the frozen nonharm/transition screen in each
state. This holds for actual-answer sampled risk, actual-answer population risk,
and both answer-marginalized supplementary risk summaries: 84 failed screens
across three states and four risk definitions. Screens use all 84 cells, with
the pre-existing mean +/- 3.5 SE descriptive bounds, not new confidence claims.

Natural publication evidence improves the aggregate values here, but excluding
it does not make joint or EPIG selection beat entropy. Publication timing is
therefore not a sufficient explanation or rescue. The answer-marginalized
summaries condition on different natural evidence across A/B/C and must not be
read as one common causal estimand.

## Verification and Cost

- Focused race tests: six fixtures, 13824 full forecast comparisons, maximum
  difference 2e-15; support, explicit answers, future/oracle isolation,
  cancellation, sentinel handling and ownership checks pass.
- Collector: all 2688 records, 4032 new fits, no failed records, source hashes
  unchanged before/after collection. Previously verified paid arrays are reused.
- Source scorer: 23930 independently calculated sampled losses and 23930 support
  checks. The 595 identical no-query supports and 4182 identical paid supports
  agree with unchanged actual-publication risks; redundant natural labels are
  compared using their actual answer only.
- Independent aggregation audit: 10200 means, 14112 paired bounds and 84 screens
  pass. A second summary run is byte-identical.
- Offline collection: 49.58 seconds wall, 198.09 user, 1.46 system, four workers
  on this Mac Mini. This is collection cost, not serving latency or a percentile.

Artifacts: `mmm-query-publication-v1.jsonl`, `-run.txt`, `-final-contracts.txt`,
`-summary.json`, `-summary-replay.json`, and `-audit.json` in this directory.
The summary records input/scorer hashes; the collector records implementation
and protocol hashes. Population integration relies on the previously tested
population kernel; the independent JS check reconstructs sampled losses, not
that whole population integration independently.

## Next Discriminator

Hold the frozen publication support fixed and compare model-predicted acquisition
value with the evaluator's teacher-weighted and model-weighted oracle query
values on that same support. This requires no new posterior fits. It can isolate
conditional-response error from answer-probability error without mixing in
newly arriving labels. Oracle access remains diagnostic only. Do not promote
these methods, change production, or report a goal as completed from this run.
