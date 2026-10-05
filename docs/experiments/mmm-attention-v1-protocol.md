# MMM observation attention pilot v1

Frozen before outcome inspection, 2026-09-12. Isolated Go research command; no
serving, OpenClaw, database, raw private text, or production publication changes.
Conceptual credit: Surnex, MMM time/depth framing, https://youtu.be/3Ey8XZYPCsE.
This is our operational adaptation, not a validated psychological model.

## Tested surface

Three supplied scopes (local, episode, process), each with two real model.Event
envelopes from a synthetic simulator. An inspection prefix exposes (1) current
What, (2) current How, (3) preceding What. These stand for surface, mechanism-field,
and temporal-context inspection. Why is an inferred distractor and cannot supply
an observed bit. No text-to-ontology extraction, scope discovery, actual causal
explanation, estimation/delegation skill, fuzzing, or sheaf discovery is claimed.

Nine independent binary observed coordinates yield 512 possible states. The
learner sees only 4096 fitting trajectories and later outcomes per task, not the
generator's function. It builds frozen conditional count tables with Beta(1,1)
smoothing. All policies use exactly this predictor, initialization, and fitting
data. A model can reuse seen feature combinations in new trajectories: this is
NOT unseen-combination generalization or grokking. No inference-time fitting.

MMM greedily chooses a scope/depth prefix with highest estimated conditional
entropy reduction per newly inspected coordinate, computed from FITTING counts
and previously inspected values only. It cannot read candidate hidden values
before selection. Ties use scope/depth order. This is a learned-information-value
heuristic, not an optimal policy. Full prefix choices permit within-scope joint
signals; greedy allocation may still miss cross-scope synergy.

All policies first inspect local What, charged one unit. Budget is six coordinate
attempts, including unavailable fields. Stop at probability <=0.1 or >=0.9 except
the exhaustive control. Reinspection never adds posterior support. Controls:

- fixed: no further scope or depth change (same-view repetition adds no evidence).
- scope: inspect only the surface coordinate of additional scopes.
- depth: inspect only deeper local fields.
- breadth: fixed breadth-first sequence across scopes then depths.
- depth_first: fixed local, episode, process depth-first sequence.
- random: uniformly select an affordable unvisited prefix.
- mmm: joint adaptive scope/depth selection.
- exhaustive: all nine coordinates, nine-unit budget, no early stopping; an
  unequal-budget information control, NOT an oracle or matched competitor.

## Families and negative controls

Targets below are visible to the GENERATOR and scorer only. All bits are uniform
independent draws. Five percent independent outcome flips except missing_evidence.

| Family | Simulator target |
| --- | --- |
| local_detail | local third coordinate |
| episode_relation | XOR of episode first two coordinates |
| process_relation | XOR of all process coordinates |
| mixed_scope | majority of local first, episode second, process third |
| irrelevant_novelty | local first; inferred Why and raw text are irrelevant |
| missing_evidence | independent outcome; no informative observed coordinate |
| regime_shift | process XOR during fitting and first half of evaluation; local third during second half |

Seven families, 256 evaluation trajectories per family per split. Fitting seed
2026091301; design seed 2026091302; untouched confirmation seed 2026091303. Both
evaluation splits run without tuning in between; different seeds alone do not
establish robustness to a different generator. Each outcome is scored after
prediction and is absent from reader/job interfaces. Availability/epoch tests
also inject future, inferred and stale records. Missing evidence is not a zero.

## Criteria and reporting

Primary: confirmation paired Brier gain of MMM over breadth and depth_first pooled
across the five stationary informative families, with simultaneous lower bounds >0 and
gain >=0.02. Normal paired intervals use z=3.4 across the 14 family/control
comparisons (z=3.4 also covers the additional aggregate/shift comparisons, at most
36 reported intervals across both splits), with all per-trajectory rows retained; these are approximate,
fixed-sample intervals, not confidence sequences or adaptive certification.
Report every family including missing_evidence and regime_shift, and flag a
family/control mean harm >0.01 regardless of aggregate benefit. No blanket
success if negative controls fail. Also report log loss, accuracy, attempted and
observed coordinates, confidence errors, stopping, complete inspection traces,
model support invariance, and serialized source/protocol hashes.

Microbenchmarks run after tests and experiments, serially with -cpu=1. They measure
the in-memory controller and fixture adapter only, not real retrieval cost,
worker queues, model fitting, or serving latency. Table size is fixed at 512x512
count pairs (~2 MiB/model); this deliberately finite reference is not a generic
scalable belief model. Jobs use immutable model state and versioned inputs.

Commands:

```sh
go test ./internal/observation ./internal/observationexperiment
go test -race ./internal/observation ./internal/observationexperiment
go run ./cmd/eventframe-observation-experiment -output docs/experiments/mmm-attention-v1.json.gz -summary docs/experiments/mmm-attention-v1-summary.json
go test ./internal/observationexperiment -run '^$' -bench . -benchmem -cpu=1 -count=3
```
