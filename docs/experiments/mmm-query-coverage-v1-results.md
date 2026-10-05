# Coverage and predictive-information results

## Verdict

FAIL: both empirical-coverage Brier rules and all three EPIG-style rules fail
their frozen primary screens. All15 primary/supplementary screens fail. None
establishes a positive lower gain bound against both random and uncertainty
sampling in any phase1 delayed cell. Wider target coverage and a change from
Brier to entropy acquisition are not sufficient rescues on this evaluation.

Phase1 delayed results,672 records and672 paid queries per non-baseline arm:

| Rule | Actual-answer sampled Brier | Population expected Brier |
| --- | ---: | ---: |
| No query | .168800035 | .168728804 |
| Random | .167334488 | .167203079 |
| Query entropy | .166629983 | .166794264 |
| Brier gain, original8 | .167365847 | .166847695 |
| Brier gain, disjoint8 | .167744461 | .167051515 |
| Brier gain, visible161 | .168057846 | .166943178 |
| Brier gain, recent32 | .167742195 | .167045581 |
| EPIG-style, original8 | .167510756 | .166877261 |
| EPIG-style, visible161 | .167938631 | .166979032 |
| EPIG-style, recent32 | .167478832 | .167042345 |

The visible161 and recent32 Brier rules change267 and250 of672 phase1 delayed
selections relative to original8. This is an exercised change, not an inactive
feature. Publication, paid labels, learner and risk evaluation remain unchanged.
EPIG defines an entropy criterion; our primary Brier screen deliberately tests
downstream usefulness rather than reasserting its defining mathematical identity.

## Verification

- [Coverage protocol](mmm-query-coverage-protocol.md) and
  [EPIG protocol](mmm-query-epig-protocol.md) were frozen before efficacy review.
- [Full response artifact](mmm-query-coverage-v1.jsonl):2688 records,19898 fits.
- [Race contracts](mmm-query-coverage-v1-contracts.txt): six fixtures,11616
  conditional checks, max discrepancy2.56e-13; ownership, future-input/label
  poisoning, teacher/identity isolation and histogram multiplicity pass.
- [Coverage summary](mmm-query-coverage-v1-summary.json) verifies4749824 full
  marginal identities and296864 original/disjoint conditional entries.
  [Replay](mmm-query-coverage-v1-summary-replay.json) is byte-identical.
- [EPIG summary](mmm-query-epig-v1.json) checks4749824 entropy/KL identities;
  its [replay](mmm-query-epig-v1-replay.json) is byte-identical. Its
  [unit contracts](mmm-query-epig-v1-contracts.txt) cover125 entropy/KL,
  relabeling and upper-bound fixtures plus weighting and invalid inputs.
- Independent outcome/interval/screen audits check56448 coverage loss values
  and80640 EPIG loss values: [coverage](mmm-query-coverage-v1-audit.json),
  [EPIG](mmm-query-epig-v1-audit.json). All84 cells are retained per experiment.
- `go vet ./internal/observationlearners` passes. No whole-goal validation,
  fresh-confirmation claim or production/paper promotion follows.

## Performance

[Reference collection](mmm-query-coverage-v1-run.txt):234.68s wall,932.66s user,
3.10s system with four workers. This is offline collection, not request latency.

[Standalone benchmark](mmm-query-coverage-v1-benchmark.txt), Go1.27.1 on Apple M4,
one fixed phase1/case0/index0/delayed fixture, three repeats of three operations:

| Component | ns/op converted to ms/op | Allocated bytes/op |
| --- | ---: | ---: |
| Existing original8 batch | 45.04-45.44 | about365270 |
| Full response via explicit refits | 621.10-624.66 | 7325312 |
| Two histograms, responses precomputed | .01174-.01396 | 0 |

These are repeated benchmark means, not p95/p99 or loaded serving measurements.
The reference implementation is expensive and has no validated quality gain;
there is no reason to promote it or claim it satisfies a100ms latency target.
[Supplemental hashes](mmm-query-coverage-v1-benchmark-hashes.json) bind the
benchmark source and output to the reference implementation.

[EPIG component timing](mmm-query-epig-v1-benchmark.json), all three targets
together on2016 calls with precomputed conditional arrays: median .27983ms,
p99 .35483ms. This excludes posterior fitting, independent KL checks, retrieval,
I/O and serving. It cannot offset the full-refit cost or establish a speedup.

## Next Distinction

[Support accounting](mmm-query-support-v1-results.md) separates reserved-slot
restoration from newly eligible publication evidence. The selection score's
63-label hypothetical conditional update is not always the actual64-label
publication update. That is an explicit existing contract boundary, not proof
of a code defect or of the cause of the failed scores. An action-aligned utility
test must preserve the real baseline rather than weakening it to make gains
appear. See also the [primary-source notes](mmm-query-coverage-literature.md).

All seven research goals remain OPEN. No production, paper, commit or push changes.
