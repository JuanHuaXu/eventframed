# Recurrent discovery pilot v1: results

Executed 2026-09-12. **Negative pilot, not a deployable rescue.** All 16 runs
completed under the frozen [protocol](PROTOCOL.md). The complete
[machine-readable evidence](../../docs/experiments/recurrent-discovery-v1.json)
includes unsuccessful runs, intermediate curves, hashes and timings.

## What happened

The 31,871-parameter networks learned training examples but did not learn the
hidden addition rule. Seven of eight structured-task runs ended at 100% training
accuracy; the remaining flat run ended at 85.62% after previously fitting. None
reached the held-out threshold or the delayed-generalization signature.

Final structured-task results, with no best-checkpoint selection:

| Arm | Seed 1201 test accuracy | Seed 1202 test accuracy | Seed 1201 test log loss | Seed 1202 test log loss |
| --- | ---: | ---: | ---: | ---: |
| Single pass | 2.06% | 1.30% | 14.029 | 14.531 |
| Flat recurrence | 0.51% | 0.26% | 8.319 | 7.121 |
| Nested recurrence | 1.29% | 0.00% | 8.648 | 8.577 |
| Single pass, three times the updates | 3.60% | 3.12% | 12.575 | 13.103 |
| Uniform fallback, analytical reference | 3.23% expected | 3.23% expected | 3.434 | 3.434 |

Seed labels abbreviate 2026091201/2026091202. Accuracy is over 389/384 held-out
ordered pairs respectively, grouped by unordered pair to avoid reversal leakage.
Groups, not orientations or checkpoints, are the independent data units. No
population confidence claim is supported by these two optimization seeds.

Nested minus flat accuracy was +0.77 and -0.26 percentage points, with WORSE log
loss on both seeds. Nested minus extended-single accuracy was -2.31 and -3.12
points, with better log loss, but still much worse log loss than uniform. The
predeclared promising-signal rule failed. Random-label final accuracy ranged from
2.06% to 4.11%, below the predeclared 10% negative-control ceiling.

All four flat/nested structured runs became worse in log loss when their inference
schedule was doubled at frozen weights (8.319 -> 16.317, 8.648 -> 13.796,
7.121 -> 21.741, 8.577 -> 19.376). The random-label recurrent runs also worsened.
These models were trained for fixed depths; extrapolating their depth is not
guaranteed to help. Looping does not constitute additional observed evidence.

## Interpretation

- **This configuration did not induce grokking.** Zero of 16 runs met the
  operational delayed-generalization flag. This does not falsify grokking or HRM.
- **Persistent learning occurred, useful generalization did not.** All saved
  weights reproduced predictions bit-for-bit on reload. That demonstrates
  persistence only, not discovery or beneficial continuous learning.
- **The models were confidently wrong out of sample.** Every structured final
  log loss was worse than a uniform forecast. No publication into EventFrame is
  justified; training fit or internal agreement would be unsafe promotion tests.
- **No circular representation or integrated fuzz/snap claim was tested.** No
  Fourier ablation, learned restriction map, or causal-chain transfer measurement
  was performed. X1/X2 remain unvalidated as EventFrame capabilities.

## Runtime and audit

CPU-only PyTorch 2.8.0, Python 3.12.14, one thread, 70.94 seconds summed training
time across all runs (excluding evaluation, initialization and artifact writes).
Median single-example inference was 22.5-23.7 microseconds for single-pass arms and
42.4-43.4 microseconds for recurrent arms. Across the individual runs, p95 ranged
23.6-26.6 and 43.5-47.8 microseconds respectively, based on 200 warm measurements
per run. These tiny-network timings are NOT daemon end-to-end latency, queue
latency, a concurrent-load benchmark, or evidence of useful inference.

Core-call matching does not equal total-compute matching: the extended single
arm spends more on heads, embeddings, and optimizer updates. Nested versus flat
is the closer matched comparison. Effective computational capacity also differs
with the unroll despite identical allocated parameter counts.

Five test cases cover partition/reversal isolation, parameter/state immutability
during inference, gradient flow, honest delayed-signature classification, and
rejection of incomplete/duplicate/modified evidence. The independent verifier
recomputes all partitions, signatures and paired comparisons, checks all 16 runs,
and matches source/protocol hashes. The full CLI run also verified every saved
checkpoint round-trip. These checks reduce implementation errors; they do not
prove an absence of all statistical or modeling defects.

No production service or database was touched. The Go daemon, its dependencies,
and the OpenClaw contract are unchanged. No checkpoint or private data is needed
for the published aggregate artifact. The optional Python environment and model
checkpoints live under /tmp and are not release dependencies.

## Next experiment, not yet run

First reproduce a published grokking-capable positive control, with its documented
architecture and a sufficiently long frozen training budget. This pilot lacks
that positive control, so it cannot isolate why discovery failed. Then introduce
recurrence as an ablation under both measured-compute and update-count controls.
Only after a learner generalizes should we test Fourier/mechanistic ablations,
fuzz-derived transformations, and composition against that functioning baseline.
Do not tune repeatedly against this pilot's held-out pairs and call the next
result confirmation; freeze a revised protocol and use new seeds/groups.

Recheck the artifact with:

```sh
python research/recurrent-discovery/verify_evidence.py docs/experiments/recurrent-discovery-v1.json
```
