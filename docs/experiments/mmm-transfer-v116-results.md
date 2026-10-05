# Independent-generator transfer v116: FAIL

Both unchanged weighting policies fail the frozen transfer requirements. All
288 non-harm screens pass, but only2/60 recovery-gain screens pass. This provides
bounded synthetic evidence of protection under the declared .01 tolerance, not
reliable cross-family improvement or a general continuous-learning result.

| Policy | Non-harm | Recovery gain | Total | Verdict |
| --- | --- | --- | --- | --- |
| Arrival log/no-neutral | 144/144 | 1/24 | 145/168 | FAIL |
| Fixed-rate delayed Markov | 144/144 | 1/36 | 145/180 | FAIL |
| Combined protocol | 288/288 | 2/60 | 290/348 | FAIL |

All58 failed gain screens also have mean gain below the predeclared .005 floor.
Of these,21 arrival-log and34 Markov gain intervals also have lower bound <=0.
This is not merely a sample-precision failure; do not rescue it by widening the
sample or relaxing thresholds after inspection. Conversely, failure to establish
gain is not proof of harmful performance. The largest non-harm upper bound is
.007488, below the fixed .01 tolerance; zero harm was not required or established.

## Complete comparison

576 independent latent trajectories span three new response families, three
dynamics, two phases and32 indices per cell. Each has immediate and delayed/
missing feedback:1152 schedule runs,294912 scored steps,884736 forecasts.
All arms retain the old fitting, observation and gate parameters. No candidate
was retuned between phases. These are synthetic transfer tasks, not actual
agent conversations, prospective answer quality or production traffic.

Confirmation, delayed/missing feedback, terminal ticks192..255:

| Family/dynamics | Generic Brier | Arrival log | Markov | Full-input Bayes floor |
| --- | ---: | ---: | ---: | ---: |
| Additive/stationary | .213827 | .214757 | .214697 | .169995 |
| Additive/abrupt | .232139 | .231851 | .231788 | .170536 |
| Additive/gradual | .241089 | .239130 | .238922 | .172573 |
| Hierarchy/stationary | .237638 | .237002 | .236933 | .197520 |
| Hierarchy/abrupt | .249938 | .247923 | .247772 | .195686 |
| Hierarchy/gradual | .257528 | .253495 | .253503 | .196194 |
| Local table/stationary | .246564 | .245790 | .245531 | .192572 |
| Local table/abrupt | .250626 | .249332 | .249120 | .196421 |
| Local table/gradual | .256658 | .250042 | .249973 | .194868 |

Lower is better. A constant probability .5 has Brier .25; several changed cases
remain near that reference. The Bayes floor uses the evaluator's complete input
and true law; it is a diagnostic lower bound, not an achievable policy result
under finite training and partial observations. This table alone cannot tell
whether model estimation or observation selection causes the gap.

Only confirmation/delayed/local-table/gradual passes gain against generic:
arrival-log .006616 with paired interval [.000578,.012654]; Markov .006685 with
[.000784,.012585]. Neither policy passes the full two-phase recovery criteria.
Markov passes none of its12 additional delayed gain checks against arrival log.
Intervals are mean +/-3.5SE over32 trajectories, approximate fixed-sample screens,
not anytime or research-history-wide guarantees. Complete cells and all bounds
are preserved in [the summary](mmm-transfer-v116-summary.json).

The teacher accuracy ceilings in this table are about70%-76%, not95%. Current
Markov expected accuracy is about55%-67% depending on the cell. Neither number
can be compared directly with the earlier94.7% near-deterministic Boolean result.
Forecast acquisition averages close to the six-read cap. Full fitting audits
remain separately charged at nine coordinates per arrived packet plus144 initial
reads; this is declared simulator accounting, not measured storage/I/O cost.

## Verification and artifacts

- [Frozen protocol](mmm-transfer-v116-protocol.md), [raw artifact](mmm-transfer-v116.json),
  [independent summary](mmm-transfer-v116-summary.json).
- Generator and runner QA preceded quality generation on a separate consumed base.
  Effective quality seed separation passed under the race detector in1.351s.
- Generation completed in128.46s; raw JSON is108261008 bytes, mode0600.
- Complete deterministic replay passed in132.76s, comparing every record exactly.
- All154 frozen source hashes match. The independent JavaScript verifier rebuilt
  teacher probabilities, expected/realized scores, acquisition bounds, as-of
  fits, feedback clocks, journal accounting, paired streams and all348 gates.
- Independent summary reproduced byte-for-byte after full replay. No outcome-led
  source correction or threshold change was needed. No new serving benchmark
  was conducted; generation/replay times are not per-request latency.

Artifact SHA-256:
`cad9f6f36879280044f0148895910f9fc10e3d0df902da95cc8276980f079d1e`.

## Interpretation and next action

This closes the absence of an independently implemented response-family test,
but its negative outcome does not close research direction1. Directions2/4 still
lack robust shifted improvement, and directions3/5/6/7 retain their own gaps.
No result is promoted to the whitepaper or daemon.

Next use this consumed artifact for diagnosis, not fresh confirmation: rebuild
the as-of fitted models and compare their full-input, issued-view and common-view
risks. That separates a weak fitted forecast from loss due to partial observation
and weighting. Do not assume another weight transition can repair an inadequate
model family. Any subsequent learner replacement needs a separately frozen
experiment on fresh seeds while retaining these failed outcomes.
