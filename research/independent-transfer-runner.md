# Independent transfer runner: integration verification

Status: implemented in `internal/observationlearners/transfer_runner_test.go`.
No untouched quality experiment or learner promotion has occurred.

The runner uses three unchanged policies: generic64, arrival-log/no-neutral,
and fixed-rate delayed Markov. It mechanically retains the v115 fitting,
observation and delayed journal lifecycle, replacing only its truth/data source
and selecting those existing policy constructors. Historical source files and
quality failures are preserved. This is research test code, not daemon wiring.

There are 16 initial fully audited packets, 256 scored forecasts, publication
every 32 ticks, as-of 64/32-example fits, and a six-coordinate forecast acquisition
cap. Full fitting audits are a separate shared cost, not included in that cap.
The complete four-model publication is still fitted for all arms, so this runner
does not measure the optimized standalone cost of generic-only deployment.

Feedback is released before publication when its origin is earlier than the
current clock and its delay has elapsed. Zero-delay current feedback is released
only after all forecasts. Missing feedback never reaches fitting or journals;
expiry and the terminal flush remain unchanged. Every issued probability, mask,
cost, fit-origin list, feedback count and journal accounting state is retained.

The two score segments are all 256 ticks and terminal ticks 192 through 255.
The latter begins after the latest possible gradual transition finishes at tick
190. This differs intentionally from the old fixed-switch fixture's late128
window; it must not be retroactively compared as the same recovery metric.

## Boundary checks

The scoped race test runs all nine family/mode combinations on consumed QA data.
For each it executes five trajectories: immediate, delayed/missing, a modified
future packet, a modified current label, and an alternative evaluator oracle.

- Independently reconstructs every fitting window from packet arrival times.
- Counts all observed feedback, checks terminal journal settlement, and verifies
  input, outcome and oracle pairing across the two schedules.
- Changes the future packet at tick160, including its input, label and
  missingness; all earlier complete step records must remain identical.
- Flips only the tick160 outcome; every issued forecast/view/cost through tick160
  must remain identical, including the forecast before that label is revealed.
- Substitutes a different teacher while keeping evidence fixed; forecasts,
  acquisition, selectors and journal state must remain identical for all ticks.
  The evaluator's scores are deliberately allowed to change in this negative
  control. This modified pairing is not valid quality data.

These checks test actual runner behavior beyond the generator packet's shape.
They are not a formal noninterference proof for every possible future adapter.
No accuracy/Brier outcome from QA is used to choose policy parameters or to claim
independent confirmation. The full three-family quality comparison remains next.

## Next frozen comparison

Final verification: `go test -race ./internal/observationlearners -run
'^TestTransferRunnerAsOf$' -count=1 -v` passed (test13.91s, package15.165s).
`go vet ./internal/observationlearners ./internal/transfergenerator` passed.
Runner SHA-256: `3a661407062aa0260898094596babde70c0d9070ec41438ed55d62652c4f7e85`.
These are correctness-test timings, not forecast performance benchmarks.

Prepare fresh disjoint quality seeds, the complete exclusive artifact/replay
writer, and an independent metric verifier before generation. Preserve the
original .01 upper paired Brier-harm tolerance and .005 positive-lower-bound
gain requirement; do not relax either after seeing transfer outcomes. Include
both phases, all nine family/mode cells, both feedback schedules, and every
changed family in the predeclared recovery comparisons. State the total gate
counts and fixed-sample interval limitations in the final protocol.

Report each teacher's intrinsic Brier floor and accuracy ceiling separately.
The old 94.7% result on near-deterministic Boolean tasks is not a target or a
comparable accuracy ceiling for these noisy response families. A transfer pass
would not erase earlier switch failures or close the other research directions.
