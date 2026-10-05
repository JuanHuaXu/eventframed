# Same-window family update: all variants failed

Full2688-trajectory consumed-data re-evaluation completed. No prior was tuned
after the outcome. No production adoption; all seven full goals remain open.

| Window / cadence | Non-harm passes | Required gain passes | Status |
| --- | ---: | ---: | --- |
| 64 / 32 (primary) | 149/168 | 0/16 | FAIL |
| 32 / 32 | 48/168 | 0/16 | FAIL |
| 64 / 8 | 157/168 | 1/16 | FAIL |
| 32 / 8 | 61/168 | 0/16 | FAIL |

Each candidate is compared with the served Markov control at the SAME cadence.
These are exploratory per-cell paired bounds, not simultaneous guarantees.
The component's mathematical agreement did not establish predictive benefit.

## Important examples

Phase1 delayed terminal64 Brier:

| Case | Markov32 | Family64/cadence32 | Markov8 | Family64/cadence8 |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .221755 | .221954 | .221658 | .222136 |
| Additive gradual | .239065 | .238771 | .234161 | .234364 |
| Parity4 | .049066 | .048143 | .050053 | .048138 |
| Null | .258525 | .263054 | .257049 | .261535 |
| Majority to parity | .060262 | .096959 | .053252 | .068969 |
| Parity to majority | .102272 | .109095 | .098235 | .084479 |

The primary majority-to-parity gain is-.036697 with interval
[-.062559,-.010834], an observed regression with its whole interval below zero.
The same comparison regresses in phase0. This is stronger negative evidence than
merely failing to certify non-harm. The primary also has smaller null-case harm.

The faster64-window reverse-transition point improvement.013756 has interval
[-.000885,.028397], so it does not meet the positive-lower-bound criterion.
Do not turn this favorable secondary point estimate into a rescue claim.
The32-window variants have widespread protection failures and are not adopted.

## What this establishes

Recomputing family odds from training-window marginal likelihood is not a
replacement for demonstrated forward predictive performance. The static
single-family window assumption can be inappropriate in a changing stream.
That explanation is a hypothesis to investigate, not yet an isolated cause:
window size, mixed regimes, model misspecification and delayed evidence can
interact. No additional prior tuning is warranted by this result alone.

The genericMass field in the summary is an average over the whole256-frame
trajectory, NOT terminal-window confidence; do not use it to explain a specific
terminal error without inspecting the recorded per-fit masses and origins.
Earlier V91/V92 static-data failures and the failed full-segmentation rescues
remain part of the evidence rather than being replaced by this experiment.

## Verification and cost

The independent scorer recomputes215040 window-fit marginal likelihoods using
closed beta integrals over sufficient counts, rather than the Go implementation's
ordered predictive products. Largest likelihood/weight discrepancy is
7.11e-14. It checks every full origin list,2,752,512 candidate probabilities,
and their exact family composition from the already-audited original/faster
expert forecasts. Both underlying source hashes and identities match.
Scoring replay is byte-identical (`cmp` exit0); this is not a second collection.

Independent-reference tests cover24 single-observation identities, four repeated
input/order cases and five exhaustive-mask factorial references. Component and
stream race checks passed previously, including future-data poisoning and
agreement with the existing parameterized-prior implementation.

Collection:416.92s wall,1671.52s user CPU,6.00s system CPU, four workers.
This is whole offline experimental cost, not daemon latency. The validated
aggregation primitive remains useful research infrastructure despite failure
of this forecast-replacement policy.

Artifacts: [protocol](mmm-family-evidence-protocol.md),
[raw](mmm-family-evidence-v1.jsonl), [summary](mmm-family-evidence-v1-summary.json),
[replay](mmm-family-evidence-v1-summary-replay.json), [run](mmm-family-evidence-v1-run.txt),
[component verification](mmm-family-evidence-component-results.md).

## Next lead

Inspect the failed transition's per-fit evidence and use a sample-count-matched
diagnostic to distinguish mixed-regime contamination from simply too few labels.
Known generator change times may be used ONLY as an explicitly hindsight
diagnostic, never supplied to a deployable policy. This can test whether better
temporal separation has relevant headroom before repeating segmentation or
forgetting experiments. No new rescue is declared here.
