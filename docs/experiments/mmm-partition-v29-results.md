# Partition V29: nonlinear gain, overall rescue fails

Report completed 2026-10-03 from the preserved 2026-10-02 artifacts.
Both independent splits FAIL the frozen overall screen. Design fails 14 of
18 geometry/regime groups; confirmation fails 16. All seven whole goals
remain OPEN. Production and the whitepaper are untouched.

## Evidence and Audit

The [protocol](mmm-partition-v29-protocol.md) preceded collection. Each
split contains 576 worlds and 11,520 arms, across two geometries and nine
regimes. The [summary](mmm-partition-v29-summary.json) records every paired
gate and its uncertainty interval, not just favorable averages.

- Design SHA256: `4be1a0fe4dd85c9299d18f561cf5c5675529af991f78ef0292e43e27cdd99869`.
- Confirmation SHA256: `12759648758d2b63c7254dbaf8edc645c39a10e039db265173e3c1effb7c93d3`.
- Executed audit SHA256: `d80cc71f4e4d6358fad6f558b5faba02505102f9fa17874d68a1ae6143342cb8`.
- Frozen checker SHA256: `96ae77ae89f8fc255677d536dc439be22df66d943e9733b3b5947909f65708c5`.

The original checker rejected a null baseline trace serialized by Go. The
separate `research/partition-v29-audit.mjs` treats that representation as
an empty trace. The original checker and tapes remain unchanged; no model,
score, or gate was repaired after seeing outcomes. The audit reconstructs
likelihoods, posterior weights, pre-label forecasts, deterministic policy
choices, costs, final laws, packets and metrics for all 23,040 arms. Random
policy draws and Gamma/Beta latent draws are not independently regenerated
by the JavaScript audit. Source hashes are checked against each manifest.

## Confirmation Findings

Primary comparisons use 32 distinct labels and the frozen deterministic
stratified policy. Lower Brier is better; usefulness is the known mean
future success probability of the packed ten members, not agent accuracy.

| Geometry / pattern | Old affine Brier | Partition Brier | Old usefulness | Partition usefulness | Partition bias upper | Group verdict |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Tight / curved | .251156 | .206899 | .654006 | .781323 | .063146 | PASS |
| Wide / curved | .251073 | .215653 | .633896 | .728273 | .231148 | FAIL |
| Tight / reversed | .206731 | .221035 | .845889 | .736326 | .101788 | FAIL |
| Wide / reversed | .208140 | .224358 | .848993 | .741242 | .131429 | FAIL |

The bias upper is `abs(mean packet bias) + 3.5 SE` over 32 worlds; the
frozen ceiling is .10. These intervals are not confidence sequences or
external calibration certificates. Tight-curved Brier gain over affine is
.044257 with paired lower endpoint .036733; usefulness gain is .127317
with lower endpoint .068589. Those gains do not override the other failures.
Only tight-curved and tight-calibrated pass complete confirmation groups.

## Observation Aliasing Diagnostic

Direct reconstruction of both raw tapes finds that BOTH old and partition
fixed32 stratified arms nominate 32 even indices and zero odd indices in
every world: 576/576 design and 576/576 confirmation. The bit-reversal
schedule spans the rank range but does not cover the alternating pattern.
In that regime its sampled true success mean is .8, versus .5 over the
frontier. This is a limitation of the declared deterministic design, not
future-label leakage or a violation of its implementation contract.

Model adequacy is a separate limitation: even with identical evidence, the
coarse partitions lose to affine models on linear patterns. A new random
within-stratum design can address zero coverage, but cannot automatically
repair an inadequate predictive family or establish Anti-Pigeon authority.

## Limits and Next Lead

The model has 26 bounded tree alternatives plus its baseline branch. The
coordinates are benchmark rank proxies, not semantic or causal variables.
This is offline, single-owner research; no full-serving, agent-answer,
source-authentication, or loaded durability claim follows. Modeled cost
sensitivity is not a passed Goal 7 equal-total-cost screen.

Next test coherent averaging of baseline, affine, and partition families,
with randomized within-stratum observation and matched controls. Preserve
these original tapes and thresholds. Any successor needs newly frozen
criteria and untouched outcomes; it must not promote a nonlinear gain
while ignoring calibrated, linear, and coordinate-mismatch protection.

Reproduce the independent audit from `<LOCAL_ROOT>`:

```sh
node research/partition-v29-audit.mjs
```
