# MMM observation attention: useful allocation, failed shift robustness

Executed 2026-09-12 under the frozen [protocol](mmm-attention-v1-protocol.md).
Credit: Surnex's MMM temporal-scope/depth framing motivated the controller. The
finite information-value heuristic and this experiment are EventFrame adaptations.

## What was actually run

28,672 predictions: two evaluation splits, seven task families, 256 independent
simulated trajectories per family, eight policies. This means 3,584 distinct
evaluation trajectories, NOT 28,672 independent observations. Every policy uses
the same frozen predictor fitted on 4096 earlier simulator trajectories per task.
Design and confirmation used different frozen seeds, without intervening tuning.

The new Go controller chooses among supplied local, episode and process views,
and among surface, mechanism-field and preceding-event inspections. Actual
model.Event envelopes carry the 5W1H fields and availability times. Hidden values
are returned only AFTER a view is selected. Selection uses conditional entropy
reduction per observation cost from fitting data, not the later outcome.

This is a finite observation-allocation experiment, not a neural-grokking trial,
an explanation engine, or a complete implementation of MMM's five depth levels.
Scopes and binary observation schemas are supplied. The controller does not
discover them from text, and known task families have separate fitted models.

## Confirmation results

Pooled over five stationary informative families (1280 trajectories per policy):

| Policy | Accuracy | Brier loss (lower is better) | Mean inspected coordinates |
| --- | ---: | ---: | ---: |
| Fixed surface view | 63.91% | 0.20026 | 1.00 |
| Scope only | 63.36% | 0.19999 | 2.60 |
| Depth only | 72.19% | 0.16390 | 2.60 |
| Fixed breadth-first | 72.97% | 0.14732 | 4.71 |
| Fixed depth-first | 81.95% | 0.11176 | 4.11 |
| Random affordable view | 76.17% | 0.12906 | 4.50 |
| MMM joint scope/depth | **94.69%** | **0.05061** | **3.11** |
| Exhaustive, unequal nine-unit budget | 94.69% | 0.06459 | 9.00 |

The stationary primary criterion passed. Paired Brier gain over breadth-first
was 0.09671 (approximate simultaneous interval 0.07892--0.11449); over depth-first
it was 0.06115 (0.04712--0.07518). Both clear the frozen 0.02 minimum gain.
Design gains were 0.10061 and 0.06203 respectively. All controls retain the same
scoring model; the change is which fields they inspect under the six-unit cap.
Nominally deeper inspection conditions on more fitted variables, so the
exhaustive predictor has sparser counts and is NOT an oracle. Its worse Brier
does not imply that more information is inherently harmful or useless.

The gain is not universal across families. MMM ties depth-first on local detail,
ties the other policies when the first coordinate is already sufficient, and
has no resolved paired advantage on the episode task. Its main wins are reaching
process-level information and combining informative scopes without wasting the
budget. Richer descriptions by themselves are not scored as success.

## Important failure

After the declared regime change (128 confirmation trajectories), the frozen
controller continued to attend to the OLD predictive relationship:

| Policy | Post-shift accuracy | Brier | Confidently wrong / 128 |
| --- | ---: | ---: | ---: |
| Breadth-first | 50.00% | 0.25650 | 0 |
| Depth-first | 47.66% | 0.25824 | 0 |
| MMM | 53.91% | **0.41235** | **59** |

The better-looking accuracy does not rescue the probability forecast. MMM was
much more confident in wrong outcomes. Its Brier excess over breadth-first was
0.15585 (0.02119--0.29052); over depth-first, 0.15411
(0.02217--0.28605). Both exceed the 0.01 harm ceiling. The same failure appeared
in design. Pooling pre-change and post-change cases would mask this failure.

The missing-informative-evidence control stayed uncertain: MMM Brier 0.24419,
zero confident errors, and all six inspection units spent. Its 57.42% sample
accuracy is not evidence of predicting independent random outcomes; paired
intervals versus the fixed-order controls include zero. It also demonstrates
that this controller has no demonstrated ability to avoid wasting effort when
all observations are uninformative.

Verdict: **stationary allocation supported in this fixture; shift robustness
failed; overall not ready for serving.** The existing daemon's changepoint,
Anti-Pigeon, and admission mechanisms were not connected to this isolated learner.
The failure does not show that those mechanisms fail, nor that simply attaching
them would repair attention. Model validity and the attention policy's validity
must both be tested after change.

## Performance

[Three serial microbenchmark repetitions](mmm-attention-v1-benchmark.txt), Apple
M4, Go 1.27, one CPU thread, fixed mixed-scope fixture:

| Controller | Mean time per invocation across benchmark repetitions |
| --- | ---: |
| Fixed | 7.46--7.48 microseconds |
| Breadth-first | 9.77--9.96 microseconds |
| Depth-first | 10.06--10.10 microseconds |
| MMM | **8.45--8.62 microseconds** |
| Exhaustive | 12.93--12.94 microseconds |

These are ns/op-derived means, NOT p99 or production latency. MMM allocates 5712
bytes in seven allocations for this fixture; even deterministic policies
currently instantiate a small RNG. Fewer inspections offset its selection cost
here. No assertion about retrieval, queue delay, fitting, real observation costs,
large feature spaces, or performance under serving load follows. The frozen
conditional table is approximately 2 MiB per task model and is exponentially
sized in the nine-variable fixture; it is not a scalable generic backend.

## Evidence and checks

- [Compact results, seeds and source hashes](mmm-attention-v1-summary.json).
- [All per-trajectory predictions and inspection traces](mmm-attention-v1.json.gz).
- Source: internal/observation, internal/observationexperiment, and
  cmd/eventframe-observation-experiment. No production package imports the new
  controller and no existing daemon behavior changed.
- Unit, integration and race checks cover learned selection before hidden-value
  access, masks, cost caps, immutable count tables, future/inferred evidence,
  tenant boundaries, reversed timelines, duplicate sources, and epoch changes.
- The evidence test checks source/protocol SHA256, recalculates every score and
  summary, and regenerates/replays all 28,672 predictions for exact equality.

All records are synthetic, with no private sessions or actual personal facts.
This is common-observation replay with charged field inspections, not an active
sensor-acquisition experiment. It does not change EventFrame's scored serving law
or validate the full X3 integration, publication, or persistent-learning claims.

## Next rescue to test, not implemented here

Freeze an explicit attention-model validity rule: after independently admitted
evidence contradicts the currently attended relationship, invalidate or soften
that relationship and allocate a protected budget to alternative scopes. Compare
the rescue with broad fallback, recalibration alone, and model reset alone on NEW
trajectories, including false-shift alarms and recovery costs. Confidence should
not authorize ignoring observations that could falsify the attention policy.
