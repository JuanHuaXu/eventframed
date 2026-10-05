# Switching-complexity diagnostic

Status: hindsight diagnostic complete; no online rescue established.
All 2688 previously consumed trajectories retained. Exact dynamic programming
over six issued experts, with budgets 0,1,2,4,8. Independent exhaustive test:
7776 paths,160 comparisons PASS. Original and replay JSON are byte-identical.

Selected terminal64 cells from phase1/delayed schedule (32 trajectories each):

| Case | Issued Markov Brier | Best fixed oracle | One-switch oracle | Unrestricted oracle |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .221755 | .217051 | .213225 | .194114 |
| Parity4 | .049066 | .047984 | .047808 | .047781 |
| Null | .258525 | .251901 | .251426 | .250707 |
| Majority to parity | .060262 | .049799 | .049299 | .049067 |
| Parity to majority | .102272 | .075819 | .069921 | .063038 |

For parity-to-majority, a single hindsight switch captures 82.46% of the gain
available from unrestricted switching relative to Markov. Its absolute Brier
gain is .032351, exploratory paired interval [.017769,.046934]. This suggests
rapid switching is unnecessary on this tape. It does NOT show that available
delayed labels identify the right expert or switch time. Majority-to-parity's
one-switch gain interval still crosses zero [-.000114,.022041].

Both initial expert and switch times are chosen using hidden Q and future loss;
the best fixed expert may differ by trajectory. Terminal64 optimization restarts
at its boundary, unlike the real mixer. These are not admissible forecasts,
generalization claims or fresh confirmation. All168 cells are in the JSON.

Next test: a finite switch-distribution working mixture, contrasted with matched
no-switch priors and the issued Markov baseline. Preserve previous failed mixers.

Artifacts: [protocol](mmm-switch-budget-v120-protocol.md),
[results](mmm-switch-budget-v120.json), [replay](mmm-switch-budget-v120-replay.json).
