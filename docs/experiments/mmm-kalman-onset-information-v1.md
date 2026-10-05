# Kalman delayed-onset information audit v1

This is a diagnostic for Goal 2, not a candidate model, confirmation cohort,
or serving change. It tests whether the existing delayed-feedback generator
contains an observable cue that can identify the changed outcome rule before
the first post-change label arrives.

## Matched worlds

`TestKalmanDelayedOnsetInformation` uses 32 streams of 512 clocks. Each pair
shares the same nine-bit contexts, baseline, 5% outcome-noise draws, 25%
outcome-blind audit decisions, 25% missingness decisions, and 16-clock label
delay. Both worlds have the same pre-change parity rule. At clock 256, one
world retains that rule and the other switches to bit 2. The realized world
and its change parameter are hidden from the filters. Forecasts are emitted
before current outcomes are generated; delayed packets update a filter only
after delivery, and only audited packets update these filters. Seed base:
`2026100101`; stream coordinates and roles use `Seed` in
`internal/observationlearners/experiment.go`.

This matched pair is deliberately **not** the original `stable05` versus
`delayed_missing` pair: those scenarios have different observation schedules,
which could themselves reveal the scenario. Equal schedules isolate the
information in the outcome rule.

## Result

The focused test passed. All 512 forecasts over the first 16 post-change
clocks (32 streams x 16 clocks) were bit-identical across worlds. Among those
contexts, 257 had different noiseless outcome rules. No changed label could
be delivered before clock 272, and no changed audited label arrived before
clock 272. Once distinguishing audited evidence arrived, the filters later
diverged in all 32 streams. This is a paired-path behavior check, not a score
improvement or proof of calibrated uncertainty.

## Information bound

The identical-observation claim is independent of the Kalman equations. A
deterministic learner with the same initial state and matched visible packets
has the same state and forecast by induction until a differing packet is
delivered. A randomized learner with matched internal randomness has the same
coupling. Contexts alone do not announce the switch in this generator: they
are uniform and have the same law in both worlds.

For an equal prior over the two worlds, consider a context where the rules
disagree. Its outcome probabilities are 0.05 and 0.95. Before any
distinguishing evidence, the equal-world expected Brier loss of a forecast
`q` is minimized at `q = 0.5` and equals 0.25. A world-aware oracle has
expected Brier loss `0.05 x 0.95 = 0.0475`; the irreducible excess is
`(0.95 - 0.05)^2 / 4 = 0.2025` on that context. The rules disagree on half
of uniform nine-bit contexts, so the population first-16 excess is at least
0.10125 per forecast under this two-world prior. The observed 257/512
discordance corresponds to 0.1016455 on this fixed context tape. These are
information bounds relative to a world-aware oracle, not measured Kalman
losses or general lower bounds under a different prior.

## Consequence

Increasing process noise, adding a context-only onset detector, or switching
between context-only experts cannot identify the changed rule during this
blind interval. A declared change-hazard prior can hedge the two worlds and
reduce worst-case confidence, but it cannot know which rule occurred. A
candidate promising earlier identification needs a genuinely observed
pre-label signal or earlier label availability; otherwise the useful target
is recovery after evidence arrives, with stationary protection and nonlinear
controls on fresh frozen cohorts. The existing [v1](mmm-kalman-window-v1-results.md)
and [v2](mmm-kalman-early-v2-results.md) Kalman failures remain unchanged.

Reproduce with:

```sh
go test ./internal/observationlearners -run '^TestKalmanDelayedOnsetInformation$' -count=1 -v
```
