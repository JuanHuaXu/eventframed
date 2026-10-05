# Fit-time age challenger v87 results

Verdict: **PASS the frozen finite pilot in both phases.** All 40
phase/scenario/schedule cells meet their declared rules. This is a successful
matched research rescue of the v85 combined-delay failure, not completion of
general delayed learning, calibration, real-agent validation or production work.

## Artifacts and design

- [Protocol](mmm-age-challenger-v87-protocol.md),
  [records](mmm-age-challenger-v87.jsonl),
  [summary](mmm-age-challenger-v87-summary.json),
  [benchmark](mmm-age-challenger-v87-benchmark.txt).
- SHA-256:
  `5344aa5dc6d1bbe934a2020268f948f6d3dccf06b0af8060e5751ac9b7e7dbaf`.
  The manifest pins 126 source/protocol/evaluator files.
- Fresh stream bases 2026118701/02. There are 640 underlying trajectories,
  reused across four feedback schedules for 2,560 schedule-runs. Each cell has
  64 trajectories with 512 forecasts per arm. These are not 2,560 independent
  underlying streams, nor unseen real-world tasks.
- Control is the retained-subset learner, not the weaker count-only control.
  Candidate keeps it and adds a fit-time age-bounded challenger. Both use the
  same received evidence, audits, refit opportunities and split/epoch rules.
  The extra model fit is charged; no additional labels or observation budget.
- At publication, the new model uses at most64 received audits from the last128
  event steps, requiring16 samples. Otherwise its declared strategy falls back
  to the retained model. The age bound does not stop the snapshot from aging
  between publications. No true change-point or regime label enters prediction.
- Two identical inner slots aggregate the new algorithm's weight, not independent
  evidence. The selector tracks adaptive algorithm experts, not probabilities
  that immutable models are true. Only one observer runs per forecast.

## Primary confirmation

Post-Brier gain is retained-control score minus candidate score. Both combined
stress cases require mean gain >=0.005 and positive paired z3.5 lower bound.
Intervals are the predeclared approximate trajectory screen, not exact coverage
or confidence sequences. Post windows are 256..511 except recurring 128..511.

| Combined jitter0..31 + missing20% | Retained Brier | Candidate Brier | Gain interval | Retained accuracy | Candidate accuracy |
| --- | ---: | ---: | --- | ---: | ---: |
| Member shift | 0.263533 | 0.208380 | 0.055153 [0.043962, 0.066344] | 56.49% | 70.06% |
| Common shift | 0.263517 | 0.207693 | 0.055824 [0.044894, 0.066753] | 56.12% | 70.36% |

Design-phase combined gains also pass: member 0.052455
[0.043231, 0.061679], common 0.049048 [0.038938, 0.059157].
Both confirmation candidate scores are below the 0.25 constant0.5 baseline,
unlike the stressed retained control. This does not establish calibration.

| Confirmation member shift | Retained Brier | Candidate Brier | Retained accuracy | Candidate accuracy |
| --- | ---: | ---: | ---: | ---: |
| Immediate | 0.203395 | 0.155188 | 67.36% | 77.28% |
| Fixed delay16 | 0.250396 | 0.201118 | 60.82% | 71.56% |
| Missing20% | 0.234094 | 0.169922 | 60.88% | 74.76% |
| Jitter + missing | 0.263533 | 0.208380 | 56.49% | 70.06% |

The new candidate still suffers delay: combined member Brier is 0.053193
[0.041141, 0.065244] worse than its own immediate run. The rescue improves
the learned model; it does not erase missing information or delayed evidence.

## Protection and remaining regressions

Stationary confirmation accuracy is 95.13% for both arms in all four schedules.
That level differs slightly from v85 because these are fresh streams; only the
paired within-v87 comparison establishes preservation. Null combined Brier is
0.252664 to0.252596, with near-chance accuracy. There is no claimed null skill.

Recurring changes do not uniformly improve. Fixed-delay recurring Brier worsens
0.224845 to0.228082 (gain -0.003237, interval [-0.007276, 0.000802]); combined
recurring worsens0.221383 to0.223663 (gain -0.002281, interval
[-0.007038, 0.002477]). Both satisfy the frozen0.01 non-harm tolerance, but they
remain regressions in the observed means, not success stories. All cells and
their full/post comparisons remain in the summary.

## Cost and implementation boundaries

Combined member runs average4.56 extra age-model fits per512 forecasts, using
about23.96 samples per executed fit. The count/retained fits still occur at the
same opportunities; the candidate does more total fitting. Filtering is bounded
by256 received audits and selection by64 samples. Acquisition stays capped at
six foreground coordinates, plus the separately charged shared monitor8 and
audit average~4.5. Mean combined member foreground cost falls4.555 to4.444;
this is acquisition count, not a serving-time benchmark.

On Go1.27.1, Apple M4, darwin/arm64 (three repeats):

- Extra subset fitting: 16 samples5.562-5.593ms; 64 samples6.252-6.323ms;
  about651KB allocated per fit, three allocations.
- Forced retained-guide fixture:0.633-0.715us,224B/two allocations.
- Forced age-guide fixture:3.163-3.170us,416B/four allocations.

Thus the age fixture costs roughly2.5us more and is several times slower in
that particular microbenchmark. The models may choose different views/stopping
depths; this is not an apples-to-apples same-mask speed claim or end-to-end tail
latency. The extra compiled conditional table has19,683 cells, about315KB of
cell payload, and exact subset fitting is restricted to the9-bit research
domain. No arbitrary-dimensional scaling result follows.

The experiment fits synchronously between logical steps. Moving this candidate
to a background publisher with bounded load, cancellation and durable lineage
still requires the separate integration work in direction6. Nothing is enabled
in the production daemon or OpenClaw.

## Verification and next work

- Disabled-age prediction/update parity and control integration parity across
  all four schedules: PASS under race instrumentation,3.311 seconds.
- Active age forecast/observer wiring, invalid/future/duplicate origins,
  age boundary, support fallback and stale publication rejection: PASS.
- Fresh generation: PASS,247.774 seconds; evaluator verifies the frozen
  manifest, paired latent streams, all metrics and policy/accounting equality.
- `go vet`: PASS. Microbenchmarks complete separately from the experiment.
- Full deterministic replay: PASS,249.770 seconds. All2,560 records, prediction
  tapes, paired latent streams and accounting reproduce against the126-file
  frozen manifest. `git diff --check` also passes.

The supported intervention is adding this declared recency-sensitive algorithm
to the retained mixture. It does not prove that every shorter window is better,
that selector delay is solved, or that the benefit transfers outside these
generators. Next freeze a breadth test with interactions, dependent inputs and
different fitting samples before altering delayed-selector behavior. Keep the
recurring regressions visible and preserve the unchanged retained control.
All seven directions remain open. No commits, pushes or production changes.
