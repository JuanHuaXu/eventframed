# Window competition with observation feedback: partial gain, FAIL overall

## Result

**3/4 reverse recovery screens pass;14/16 nonharm screens pass. Not adopted.**
The fourth reverse interval crosses zero. Both immediate forward-change cases
fail nonharm. This does not complete any of the seven research goals.

All128 consumed trajectories and both schedules use original random-audit fits
and an independently monitored split clock. The fixed arm reconstructs the prior
window-competition candidate per frame. The coupled arm chooses its own proxy
observations and updates private old-inner/window-inner/outer weights from its
own issued advice. Each arm gets a private read cache seeded only with as-of paid
monitoring coordinates. No other arm's acquisition enters that cache.

Post-change Brier (lower is better):

| Cohort / schedule / change | Original | Fixed competition | Coupled |
| --- | ---: | ---: | ---: |
| 1 / immediate / parity to majority | .231338 | .224145 | .218184 |
| 2 / immediate / parity to majority | .218989 | .208538 | .207545 |
| 1 / delayed / parity to majority | .259748 | .254950 | .250972 |
| 2 / delayed / parity to majority | .252457 | .248096 | .246148 |
| 1 / immediate / majority to parity | .243420 | .245103 | .251267 |
| 2 / immediate / majority to parity | .240403 | .242393 | .249508 |

Delayed reverse gains versus original are.008776 interval[.003833,.013720] and
.006309 interval[-.000776,.013394]. The second does not pass despite its positive
mean. Delayed reverse accuracy improves53.13%->58.72% and55.03%->58.74%.
All coupling-versus-fixed reverse gain intervals still cross zero; do not claim
that this study independently establishes the acquisition change as beneficial.
Intervals are consumed exploratory mean+-3.5SE over16 trajectories, not anytime
confidence sequences or independent-frame intervals.

## Observation cost and regression diagnosis

Total unique-coordinate cost includes live/reference monitoring and audits plus
new foreground coordinates beyond the paid live monitor mask. Compared with the
fixed arm, coupled cost is greater in104 schedules, equal in121, and lower in31.
Thus this is NOT an equal-cost win. Per-frame mean totals in immediate forward
cases rise12.351->12.565 and13.261->13.352. Their post Brier worsens despite more
observations. Reverse totals also increase modestly. Complete tables are retained.

The retrospective [coverage diagnostic](mmm-window-coupling-v1-coverage.json)
uses the simulator's rule mask ONLY after the run. In late immediate forward
windows, full parity-rule coverage falls:

- Cohort1:36.91% fixed ->29.35% coupled; Brier.22311->.23557.
- Cohort2:49.80% fixed ->35.30% coupled; Brier.22171->.23555.

The event-count guide is selected50.73%/50.34% of these late windows. Early
coverage instead rises, so this is not simply failure to observe more at all.
Subgroup averages do not establish causal attribution. Complete rule coverage
is not a guarantee of correct prediction, and partial majority coverage can
still be informative. No simulator rule enters runtime guide choice.

Next test a guide-only ablation: when event-count wins, use the available
event-subset model as its observation proxy while retaining the count predictor
in scored advice. This separates the suspected acquisition failure from the
observed count-model forecasting value. Preserve every stable/forward/reverse
case, fixed baseline, missing/delay schedule and frozen gate; no data-dependent
rule-name exception or threshold tuning. Inspect both cost and loss. The test
may fail: incoherence alone does not prove that the alternative proxy is better.

## Verification

- Race contracts PASS (1.434s package), covering private cache isolation,
  malformed seed rejection, missing-label isolation, future-label isolation
  and same-clock publication ordering. Vet PASS.
- Collection21.69s; complete native replay21.65s. Raw files are byte-identical.
  The full collection was NOT run with the race detector.
- Every fixed candidate forecast and advice vector reconstructs the prior
  artifact to1e-12; all original fit supports/publication clocks remain unchanged.
- Independent JavaScript checks5,111,808 values, maximum difference1.11e-15,
  including all weight updates, pending-origin rejection after split, masks,
  component probabilities, fit origin lists and incremental coordinate costs.
- All892 captured source/contract hashes match disk; summary replay is identical.
  This independently reconstructs arithmetic/timing, not every model fitter.

## Performance scope

Apple M4, single existing reverse trajectory, entire two-arm512-frame fixture:

- Immediate:93.71/92.09/92.73ms, about73.09MB allocation,18,419-18,435 objects.
- Delayed:67.61/67.18/67.55ms, about54.70MB allocation,20,024-20,026 objects.

These include fits, both arms, simulated observations, feedback and output
records; initial base fitting and file decoding are excluded. They are NOT
per-request tail latency or a production benchmark. Three one-iteration runs
are descriptive, not confidence intervals. Delayed runs fit fewer times.

Five count tables (base,label,event,local,pooled) plus two subset tables retain
11,115,616bytes and can be shared between arms. That excludes journals, headers,
transient publication overlap, runtime overhead and process RSS. No peak-RSS
or concurrent service-freshness claim is made.

## Artifacts

Raw `mmm-window-coupling-v1.jsonl` and native replay
`mmm-window-coupling-v1-replay.jsonl` SHA256:
`ab3756fc535db2a1cc47d3c97629d5e2e291452284dd1238702e7cebf840dda5`.
Summary `mmm-window-coupling-v1-summary.json`; captured evaluator
`research/window-coupling-summary.mjs`.
Post-hoc coverage script SHA256:
`db6cb4a0abde4327695768a8826cd194fb473ac483eb287dd9cbe9b3c1ee961e`;
coverage output SHA256:
`28015ee32ce84bf5fb2f2364064c4bb676f2d5febbc220625b8f412fbee8e804`.

Production, whitepaper, remotes and preexisting tracked modifications untouched.
No fresh confirmation or real-agent outcome evidence collected in this run.
