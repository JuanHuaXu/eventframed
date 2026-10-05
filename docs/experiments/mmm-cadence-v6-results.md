# Cadence-matched learner v6 results

2026-09-12. [Protocol](mmm-cadence-v6-protocol.md),
[summary](mmm-cadence-v6-summary.json), [raw journal](mmm-cadence-v6.json.gz).

480 streams, ten scenarios, three fitting seeds and four arms. No tuning between
design and confirmation. Reuses the v5 synthetic generators with fresh fitting
and evaluation seeds; not independent generator or real-task replication.

## Frozen verdict

Rolling forest PASSED. Immediate and batched persistent forests FAILED the
complete frozen acceptance rule. All passed stationary protection and the
no-large-mean-harm guard. This is learner-level success only, not a completed
MMM observation/sharing integration or production recommendation.

Confirmation Brier (24 streams per case; stable/null full, others post-change):

| Scenario | Fixed count | Immediate forest | Batched forest | Rolling forest |
| --- | ---: | ---: | ---: | ---: |
| Stable05 | 0.062227 | 0.062228 | 0.062228 | 0.062228 |
| Stable20 | 0.176843 | 0.176840 | 0.176845 | 0.176840 |
| Shift128 | 0.250585 | 0.197275 | 0.209918 | 0.189535 |
| Shift256 | 0.255678 | 0.249622 | 0.252580 | 0.238742 |
| Shift384 | 0.265457 | 0.264946 | 0.265484 | 0.265168 |
| Gradual | 0.207044 | 0.195627 | 0.199266 | 0.194785 |
| Recurring | 0.199374 | 0.192412 | 0.194902 | 0.191882 |
| Delayed/missing | 0.267993 | 0.260917 | 0.264606 | 0.260905 |
| Interaction | 0.255683 | 0.256867 | 0.256812 | 0.256938 |
| Null | 0.252089 | 0.252146 | 0.252131 | 0.252114 |

Rolling gains on shift128/256 are 0.061050 and 0.016936. The frozen approximate
paired intervals are [0.041239,0.080861] and [0.002757,0.031114]. All three
fitting-group means were positive for both primary shifts. These intervals are
conditional on the three fitting models, not unconditional generalization bounds.

## What the ablation establishes

Batched-minus-immediate Brier on shift128 is +0.012643; on shift256 +0.002958.
The only changed training behavior between those two is update timing; tests
verify exact same tree state at batch boundaries. More frequent updates account
for part of the earlier gain. Nevertheless the matched-cadence tree still improves
the early shift substantially over counts, so cadence is not the whole story.

The rolling candidate restores stronger gains at matched cadence, but it is a
policy bundle: discard old samples and rebuild a tree with a separate declared
RNG seed. It does not isolate forgetting from all tree-construction effects.
The result supports testing that bundle in MMM, not a universal claim that
forgetting is always better or that adaptive windows are exhausted.

Stable05 accuracy remains about 95.05% for rolling. Shift128 accuracy is 71.23%
versus fixed 52.66%; shift256 58.50% versus 51.89%. Probability improvement is
not restoration of stationary accuracy. Late shifts and interaction-only
transitions remain near chance. These weaknesses must remain in later gates.

## Cost and verification

Confirmation mean accumulated tree work per 512-step stream was approximately
41us immediate, 33us batched and 147us rolling, separately from count refits.
These are instrumented in-process totals, not per-request latency bounds or
tail benchmarks. Rolling intentionally reprocesses samples, using more total
training operations despite equal acquired-label volume. Maximum confirmation
node counts were 17,17,15, under each forest's 155-node cap.

The fixed-count and immediate-forest predictions match v5 exactly on shared
unit seeds across stationary, shift, missing/delayed and interaction scenarios.
Unit race checks passed; source hashes, delivery accounting, all 480 traces and
summaries replayed exactly except wall-clock timings. Vet passed. Previous v5
sources and evidence were not edited. Unused trace slots4..6 are not arms.
The full learner package, including preserved v5 and new v6 artifact replays,
also passed `go test -race ./internal/observationlearners
./cmd/eventframe-observation-cadence -count=1` (445.927s). That instrumented
verification duration is not a runtime performance measurement.

Next: partial-view observation integration on a frozen fresh matrix. A forest
must forecast using only observed fields; consulting unseen event coordinates
would invalidate the experiment. Uniform marginalization may be used only under
an explicitly declared generator assumption, not silently on real text inputs.
