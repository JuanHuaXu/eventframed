# V67 Rate-Dynamics Boundary Results

Diagnostic FAILED as a rescue; parameter tuning alone is not supported.
All seven whole goals remain OPEN. See the frozen
[protocol](mmm-joint-rate-v67-protocol.md).

40 consumed worlds x three delays x three hazards x two policies = 720 arms;
96,000 underlying Y shared with V66, not a new confirmation cohort.

| Hazard | Policy | Issued Brier | Terminal Brier | Top-10 usefulness |
| --- | --- | ---: | ---: | ---: |
| 0 (static) | no_pair | .272288314 | .231985665 | .703989898 |
| 0 (static) | uncertainty | .268112023 | .222714668 | .719964412 |
| 1/16 (nominal) | no_pair | .269705599 | .217094257 | .787380705 |
| 1/16 (nominal) | uncertainty | .264831503 | .208304871 | .792780334 |
| 1 (memoryless) | no_pair | .348212362 | .348220407 | .509863671 |
| 1 (memoryless) | uncertainty | .348212362 | .348220407 | .509863671 |

Removing resets worsens uncertainty issued risk .003280520 and terminal risk
.014409797; wins 52/loses 68 cells versus nominal. It slightly lowers stationary
mean harm (.042942112 versus .044586141), still far above the original threshold.
Both static and nominal have 114 Adaptive-harm cells; memoryless has 120.
All six combinations fail improvement, non-harm and recovery gates. Complete
core loops pass 400ms (largest 17.488ms); not loaded serving or equal TOTAL cost.

Memoryless next clean forecasts stay at baseline as required: past measurements
cannot influence a wholly reset rate. W2 can still inform static noise, but not
move that clean forecast. Uncertainty's additional observations leave clean
quality identical to no_pair. This is an intentional negative control.

Independent reconstruction of ALL 720 arms passes: 1,728,000 issued packets,
3,456,000 clean/W1 scalar comparisons, requests/queries/missingness/choices,
snapshots and metrics. All 240 nominal arms are bitwise V66-equivalent excluding
timing; all 40 populations unchanged. Race/vet pass. 100 compiler inputs, seven
support files and two generated mains frozen/copied; five commands terminal 0.
Artifacts: `research/joint-rate-v67-diagnostic`, including raw JSONL/readback.

This falsifies "turn off resetting to rescue this learner" on the consumed
cohort. It does NOT identify independent-member sampling, prior shape or noise
misspecification as the unique cause. Those need separate tests, not a renamed
window heuristic or retrospectively optimized hazard. Next: a coherently shared
context model with nonlocal observation values, separate non-sharing controls,
capacity/memory/runtime audits and external Anti-Pigeon authority unchanged.
No adoption, new seeds, private/sealed outcomes, production or paper edits.
