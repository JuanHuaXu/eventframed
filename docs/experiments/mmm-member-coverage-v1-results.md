# Fresh member coverage v1 results

**PASS finite false-revocation coverage; FAIL full integration pilot.**
The unchanged mixture gate splits sooner but does not meaningfully improve
post-change Brier. This closes the v80 sample-size gap for the declared
fixed-model synthetic generators, not goal3 as a whole.

[Protocol](mmm-member-coverage-v1-protocol.md),
[records](mmm-member-coverage-v1.jsonl),
[summary](mmm-member-coverage-v1-summary.json),
[audit](mmm-member-coverage-v1-audit.json),
[replay](mmm-member-coverage-v1-replay.txt).

## Coverage

512 trajectories per scenario per phase, five scenarios, two fresh seed bases:
5,120 trajectories and 2,621,440 frames. The five paired arms, 512-step horizon,
fixed4096-label base, acquisition budgets and fitting rules are unchanged.
Seeds2026091507 and2026091508 generate fresh streams, not fresh fitting samples.

All twelve false-revocation cases (old and mixture gates, stable, common shift
and null, both phases) have **0/512** revocations. Each one-sided exact
Clopper-Pearson upper bound at alpha=.05/12 is **1.0647285%**, below2%.
Bonferroni gives simultaneous95% coverage under independent trajectory sampling
within each generator, without requiring independence between paired arms.
This is a finite-horizon empirical rate bound, not an anytime guarantee or
external target-law diameter certificate.

## Confirmation

| Scenario | Old restricted delay | Mixture restricted delay | Old post Brier | Mixture post Brier | Old / mixture splits |
| --- | ---: | ---: | ---: | ---: | ---: |
| Stable | No splits | No splits | 0.047556 | 0.047556 | 0 / 0 |
| Member shift | 115.9141 | 74.1836 | 0.239424 | 0.239446 | 512 / 512 |
| Common shift | No splits | No splits | 0.239286 | 0.239286 | 0 / 0 |
| Recurring | 177.5293 | 90.6680 | 0.192848 | 0.192705 | 499 / 512 |
| Null | No splits | No splits | 0.251495 | 0.251495 | 0 / 0 |

Restricted delay penalizes missing or premature splits by the remaining horizon.
Recurring old-gate detected-only delay is172.1503, excluding13 missed splits.
Member-shift arms have no misses or premature splits. No-change delay sentinels
in the machine summary are not observed detection times.

Member-shift delay improves36.0012% in confirmation and36.5277% in design,
passing the10% speed gate. Confirmation post-Brier gain (old minus mixture) is
**-0.00002166**, paired mean +/-3.5SE interval[-0.00029437,0.00025106].
Design gain is-0.00003009, interval[-0.00027742,0.00021723]. Both fail the
unchanged gain>=.005 and positive-lower-bound requirements. These are frozen
pilot intervals, not confidence sequences. Confirmation member-shift accuracy
changes58.5709% to58.5495%. Faster isolation is not forecast improvement.

## Verification and costs

Confirmation member-shift foreground cost is4.70915 versus4.70984 coordinates
per frame, within the six-coordinate cap. Monitoring adds8 coordinates/frame;
auditing adds4.51579 on average and fitting averages19.69336 models/trajectory.
These are additional work, not hidden inside foreground cost.

- Focused race contracts pass, including scalar-rounding and integration parity.
- Collection takes226.65 seconds wall time; full replay takes228.24 seconds.
  These are multi-arm experiment times, not serving-latency benchmarks.
- Full replay matches source hashes, decision/outcome tapes and all metrics.
  Independently regenerated summary is byte-identical.
- Independent audit passes250 metric aggregations,20 paired intervals and12
  probability checks. Bound tests cover28 inversions and40 exact coverage cells.
- Raw SHA256: `fc76e4e75451d2a9acfb8266a88b86b8ff23411899be9f629bd09eb16d471245`.

## Remaining work

Retain coverage and speed findings without weakening the failed quality gate.
Post-split learning and observation quality remain unresolved; v81-v120 already
investigate several such paths and must not be repeated as new rescues. Fresh
fitting samples, delayed evidence, prospective agent outcomes and durable serving
remain separate requirements. All seven goals remain open. No production,
whitepaper, dependency, commit or deployment changes were made.
