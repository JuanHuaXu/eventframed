# Evidence availability diagnostic

Status: diagnostic complete; no new predictor validated. All2688 consumed runs,
756 separate grouped cells,192 short-product identities and756 sampled as-of
checks. Immediate and delayed schedules are paired, not independent replicates.

At clock192, phase1 delayed schedule, comparing the future64 hindsight-best
fixed expert to generic64:

| Case | Prior window | Opportunities | Mean available labels in opportunities | Actual LLR<=0 | Actual LLR>log(19) |
| --- | ---: | ---: | ---: | ---: | ---: |
| Parity to majority | 32 | 25 | 13.32 | 0 | 10 |
| Parity to majority | 64 | 25 | 38.56 | 8 | 9 |
| Parity to majority | 128 | 25 | 90.60 | 6 | 12 |
| Majority to parity | 32 | 32 | 14.97 | 3 | 20 |
| Majority to parity | 64 | 32 | 40.63 | 1 | 31 |
| Majority to parity | 128 | 32 | 91.41 | 14 | 12 |
| Additive stationary | 32 | 10 | 13.70 | 8 | 1 |
| Additive stationary | 64 | 10 | 39.10 | 8 | 0 |
| Additive stationary | 128 | 10 | 90.20 | 7 | 0 |

Opportunity means future expected Brier gain>=.005, selected with hidden Q.
This is NOT an implementable opportunity detector. LLR thresholds are descriptive,
not sequential tests, posterior odds or valid confidence after hindsight selection.

For parity-to-majority opportunities, recent32 expected LLR is2.966 on arrived
labels versus8.199 with all window outcomes integrated under the teacher law.
Missingness/delay remove useful expected evidence, but expanding to64 lowers the
available expected LLR to1.981 and leaves10/25 nonpositive. More historical data
does not necessarily mean more relevant evidence. All25 recent32 actual ratios
are positive, so the difficulty is not solely an absence of any distinguishing
evidence. A baseline-heavy prior can also impose a substantial initial odds cost.

For majority-to-parity,64-frame evidence is considerably stronger than32;128
pulls in conflicting older evidence. Therefore the diagnostic does not justify
choosing one universally shorter window. Stationary opportunities often have
unhelpful prior evidence: hindsight variation between fitted experts cannot be
assumed predictable. Future expected log-score gains are positive in all listed
opportunities, so those particular comparisons are not a Brier/log-score conflict.

Next lead: test a prospective observation policy focused on RECENT disagreements
between declared candidates, with equal acquisition budget, against random and
uncertainty selection. Separate acquiring labels sooner from acquiring more
labels. Retain stationary cases and negative results from prior disagreement
policies. This diagnostic supports testing that interaction, not assuming success.
Do not tune windows on these consumed cases or relabel phase1 as fresh confirmation.

Artifacts: [protocol](mmm-evidence-availability-v120-protocol.md),
[results](mmm-evidence-availability-v120.json),
[replay](mmm-evidence-availability-v120-replay.json).
