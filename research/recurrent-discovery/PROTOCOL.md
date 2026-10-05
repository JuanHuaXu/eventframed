# Recurrent discovery pilot v1

Status: frozen before the first training run, 2026-09-12. This is an isolated
research harness, not a serving feature, HRM reproduction, or EventFrame claim
validation. No daemon, database, private corpus, network inference, or publication
gate participates. Python/PyTorch is an optional research dependency only.

## Question and controls

Does persistent neural learning generalize on unseen opaque operation pairs?
Does a nested recurrent schedule improve it relative to a flat recurrent schedule
or additional ordinary training? This tests prerequisites for X1/X2, not learned
sheaf restrictions, fuzzing, real conversation transfer, or autonomous discovery.

Data: addition modulo 31, with a seeded permutation hiding all numerical token
identities. Only the generator knows the operation. Split 60% of unordered pair
groups for training; both orientations stay together. All remaining groups are
held out. A random-label control assigns an independent uniform label to every
unordered pair group. No reverse-pair lookup, true Fourier coordinates, supplied
transformation maps, or held-out target access during optimization.

Two paired seeds: 2026091201 and 2026091202. Both are pilot repetitions, not an
independent confirmatory study after selection. All arms share data and initial
parameters within seed/task. Full-batch AdamW, learning rate 0.001, weight decay
1.0, betas (0.9, 0.98), 32-dimensional embeddings, 64-dimensional hidden state,
tanh updates, cross-entropy loss. Full differentiation through every recurrent
step; zero hidden-state initialization for every example/forward pass. No state
carried across examples. No early stopping, hyperparameter search, or test-driven
checkpoint choice. This intentionally differs from HRM's transformer modules,
one-step gradient approximation, deep supervision, and adaptive halting.

All arms have the same trainable parameters: input embedding/projection, equal-size
L/H update modules and output head. H updates consume the current L state.

| Arm | Forward schedule | Optimizer updates | Core module calls |
| --- | --- | ---: | ---: |
| single | LH | 3000 | 6000 |
| flat | LHLHLH | 3000 | 18000 |
| nested | LLHLLH | 3000 | 18000 |
| single_extended | LH | 9000 | 18000 |

Core-call matching is NOT exact FLOP or wall-time matching: embedding/head and
optimizer work differs. Report measured training wall time and single-item
inference timings. No claim of equal total compute based only on this proxy.

Record train/test accuracy and log loss at initialization and every 100 updates,
plus parameter norm and cumulative core calls. Held-out curves are diagnostic
output only; never inputs to the optimizer or a model-selection rule. Exemplar
lookup reports test accuracy 1/31 under a uniform fallback in expectation and
log loss log(31); it has no held-out groups to retrieve.

Pilot delayed-generalization signature: training accuracy >= 0.99 for three
consecutive recorded checkpoints, then at least 500 optimizer updates later test
accuracy >= 0.90 for three consecutive checkpoints. This operational flag is not
proof of Fourier circuits or the grokking mechanism; record censored/nonfitting
runs too. Do not equate a publication step with this signature.

Report final paired differences for nested minus flat and nested minus extended
single on BOTH seeds, including log loss, not best checkpoints. Two seeds are too
few for a robust population guarantee. A promising signal requires >=5 percentage
points final test accuracy improvement on both structured runs, lower test log
loss on both, and no random-label accuracy above 10%. Failure only rejects this
configuration/budget as a rescue, not recurrence or grokking in general.

At final frozen weights, evaluate twice the inference schedule, without gradient
updates; record it separately from learned improvements. Save/reload a checkpoint
and require bit-identical predictions to verify persistence without retraining.
Training replay increments optimizer steps, never independent evidence/support.

## Running and evidence

Use an isolated virtual environment with requirements.txt. Run unit tests first:

```sh
python -m unittest discover -s research/recurrent-discovery -p 'test_*.py'
python research/recurrent-discovery/experiment.py --output docs/experiments/recurrent-discovery-v1.json --checkpoints /tmp/eventframe-recurrent-checkpoints
```

The output includes the complete config, source/protocol SHA256, runtime versions,
split hashes/counts, all learning curves, all final metrics, parameter hashes, and
timings. Checkpoints are local only; synthetic tables are reproducible from seeds.
Budget: CPU only, one torch thread, sequential runs. No MPS/CUDA or production
services. This is a bounded slow-path feasibility experiment, not a hot-path
latency benchmark. Results cannot justify shipping a neural dependency.

## Research basis

- [Power et al., Grokking](https://arxiv.org/abs/2201.02177): delayed generalization
  on small algorithmic datasets; the phenomenon is measured, not guaranteed.
- [Nanda et al., Progress measures](https://arxiv.org/abs/2301.05217): modular
  addition circuits and ablation evidence. This pilot does not yet reproduce
  Fourier circuit diagnostics or supply those circuits to the learner.
- [Wang et al., HRM](https://arxiv.org/abs/2506.21734v3): separate fast/slow recurrent
  states motivate the schedule comparison only; no fidelity claim.
