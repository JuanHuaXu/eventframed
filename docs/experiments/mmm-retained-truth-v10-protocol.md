# Retained challenger: independent outcome-family stress

Frozen 2026-10-01 before scoring. This study probes research goals 1 and 4
after the v9 input-generator study. It is not a rescue tuned to v9's failed
adaptive cell. No production or original v8 source changes are allowed.
The original `internal/observationlearners/retained.go` SHA-256 is
`ecbc832eb2ea3d71957fe0ecd0da226325c60c5b4e247cc7ab245ddcdda86370`.

Use an exact Go build overlay, replacing only two input draws, two outcome
calls, the base-fit seed expression, and the evaluation-seed expression.
All four v8 arms, six-coordinate MMM controller, audit probability 0.25,
scenario schedules/noise/missingness/delay, 4096-label base size, 64/256
audit windows, tree cadence, scoring windows, and `SummarizeRetained` gates
remain frozen. The source diff and generated hash are archived.

Two input distributions are fixed:

1. `uniform`: independent fair bits 0-8.
2. `latent`: draw fair Z; bits 0, 1, 6 and 7 independently copy Z with
   probability 0.85; bits 2, 3, 4, 5 and 8 are independent fair bits.

The new outcome family uses the same scenario change schedule and noise
probabilities but not v8's parity-to-single-bit truth function. Before a
change, the noiseless outcome is majority(bits 0, 3, 6). After a change, it
is majority(bits 2, 5, 8). In the `interaction` case only, the post-change
outcome is `(bit0 AND bit1) OR (bit2 AND bit3)`. `gradual` interpolates by
choosing the post mechanism with probability clamp((t-128)/256,0,1);
`recurring` uses the post mechanism in alternating 128-frame blocks. The
`null` case remains an independent fair outcome. Finally flip a non-null
outcome with the scenario's frozen noise probability. No label or future
feedback is consulted when selecting an input or making a forecast.

For each input mode, run all ten original scenarios, three independent
4096-label base fits per phase, eight independent 512-frame streams per fit,
and separate design/confirmation phases: 480 streams per mode, 960 total.
Base-fit seed is `(2026100122+phase)*1000000 + scenario*1000 + fit`;
evaluation seed passed to `RunRetainedStream` is `2026100124+phase` and is
further partitioned by scenario/fit/stream/role by the existing `Seed`
function. Phase is 0 for design and 1 for confirmation. Therefore no base
fit or evaluation stream is reused between phases. The same stream is paired
across all four arms.

Reuse the unchanged v8 gate separately in each input mode. A strong
synthetic outcome-family replication requires at least one retained arm to
pass every confirmation gate in both modes: primary shift gains >=0.005 with
positive paired 3.6-SE lower bounds and positive fit-group signs, stable
protection, and no scenario/window mean harm below -0.01. Report every
failed gate and both positive and negative results. Do not choose thresholds,
arms or new families after seeing the output.

Audit the overlay diff, source hashes, pre-outcome forecasts, availability
chronology, per-arm Brier/log/accuracy recomputation, independent summary
arithmetic and all 120 paired comparisons per mode. Measure wall time,
count/tree fit time, update count and current tree nodes; these are offline
costs, not daemon or loaded-service latency. This synthetic study cannot
establish real-agent utility, causal discovery, Anti-Pigeon coverage or
production qualification. Subsequent agent-task evidence must be untouched.
