# Member fitting-sample breadth v1 results

**PASS declared finite false-revocation coverage and speed; FAIL full pilot.**
Fresh fitting samples do not remove the post-split forecast-quality failure.
This is robustness evidence for the gate component, not a new learning rescue.

[Protocol](mmm-member-fit-breadth-v1-protocol.md),
[raw records](mmm-member-fit-breadth-v1.jsonl),
[summary](mmm-member-fit-breadth-v1-summary.json),
[audit](mmm-member-fit-breadth-v1-audit.json).

## Design

5,120 independent fitted bases, each trained on4,096 samples before its own
512-frame stream. Five scenarios,512 trajectories per scenario per phase, with
phase seed bases2026091511 and2026091512. All five original paired arms and
learning/gate settings are unchanged. Each record contains the fit seed and
training-sample SHA256. All25,600 fitting/live/reference/audit/random seeds are
distinct. The paired arms share their trajectory's fitted base and stream.

The legacy-seed helper exactly reproduces both original null/non-null fitted
models; changing the fitting seed changes their samples and model state.
This varies fitting samples, not fitting-set size, noise level, shift time or
the nine-bit representation. It is not actual agent or production evidence.

## Findings

Every one of the twelve predeclared false-revocation cases has0/512 events.
The exact one-sided Clopper-Pearson upper bounds at alpha=.05/12 are1.0647285%,
below2%, with simultaneous95% Bonferroni coverage. Unlike the fixed-base study,
these bound the combined fitted-sample/stream population for the declared
generators. They do not certify every possible trained model, an anytime
monitor, or an external target-law diameter.

| Confirmation scenario | Old restricted delay | Mixture restricted delay | Old post Brier | Mixture post Brier | Old / mixture splits |
| --- | ---: | ---: | ---: | ---: | ---: |
| Stable | No splits | No splits | 0.047248 | 0.047248 | 0 / 0 |
| Member shift | 116.4492 | 74.8906 | 0.239833 | 0.239864 | 512 / 512 |
| Common shift | No splits | No splits | 0.239610 | 0.239610 | 0 / 0 |
| Recurring | 175.6523 | 89.2266 | 0.193396 | 0.193334 | 486 / 512 |
| Null | No splits | No splits | 0.251676 | 0.251676 | 0 / 0 |

Restricted delay charges absent/premature splits the remaining horizon.
Recurring old detected-only delay is164.5062, excluding26 misses; it must not
replace the restricted mean. Member-shift arms have no misses or premature
splits. No-change numerical delay sentinels are not observed detection times.

Member-shift delay improves35.6882% in confirmation and37.3160% in design,
passing the10% speed gate. Confirmation post-Brier gain is-0.0000310083 with
paired mean +/-3.5SE interval[-0.000298382,0.000236365]. Design gain is
-0.000122470 with interval[-0.000345384,0.000100444]. Both fail the unchanged
mean gain>=.005 and positive-lower-bound conjunction. These are the frozen pilot
intervals, not confidence sequences. Confirmation member-shift accuracy is
58.3504% in both arms; stationary accuracy remains95.0615%.

## Work and validation

Confirmation member-shift foreground cost is4.72153 old versus4.72314 mixture
coordinates/frame. Separate monitoring averages8.00295, auditing4.50124 and
post-initial fitting19.58789 models/trajectory. The extra initial fit per
trajectory is charged to total experiment time. Monitoring is not exactly8
under variable fitted bases; actual counts are retained.

- Focused race tests pass: fresh-base parity/diversity, role-seed uniqueness,
  paired investigator, scalar rounding and original integration parity.
- Collection completes in233.56 seconds wall time, including fresh fitting.
  This is a multi-arm experiment time, not serving latency.
- Independent audit passes250 metric aggregations,20 paired intervals,12
  probability checks,5,120 unique fit digests and25,600 distinct role seeds.
- Summary replay is byte-identical. Full Go replay matches all source hashes,
  training hashes, prediction/decision tapes and metrics in235.77 seconds wall
  time: [replay log](mmm-member-fit-breadth-v1-replay.txt).
- Raw SHA256: `15c32323e87f256d75af69b318e53b1f97873e5373e42c078e8e49bca290d8ff`.

## Decision

The earlier fixed-base result was not reversed by this fitting-sample test.
Do not continue gate-only replication in search of a forecast gain, or weaken
the forecast gate. Reconnect future work to the latest learner/observation
failures, retaining v82/v83's narrower positive results and the later delayed,
dependent and full-case failures. All seven goals remain open. No production,
whitepaper, dependency, commit or push changes were made.
