# Transfer diagnosis v117: fitted forecasts are the main remaining error term

Status: completed consumed-data diagnosis, NOT a validated rescue. V116 remains
FAIL and no policy changed. All1152 parent runs were reconstructed, including both
phases, schedules and all families/modes. No favorable subset was selected.

In all36 terminal64 cells, the mean generic64 full-input excess over the true-law
noise floor exceeds the absolute observation and routing terms combined. Its
diagnostic lower interval bound is positive in every cell. Mean full-input excess
ranges from .037505 to .072952. All four raw models still have appreciable error:
even the best per-cell mean full-input model is at least .037505 above the floor.

This narrows the next hypothesis: better weights or more observed coordinates
alone are unlikely to remove most of the measured gap. It does NOT prove which
part of fitting is responsible. The full-input term combines model structure,
finite-sample estimation and stale/mixed-regime training; it is not exclusively
representation bias. Known regimes/relevance are never supplied to a policy.

## Exact decomposition

For the fixed-Markov issued forecast, the identity is:

served risk - true-law noise floor = full-input generic fit excess
+ generic view-versus-full difference + served-versus-generic matched-view difference.

The final term includes model selection and weighting, not just a filter update.
Terms are evaluated on the same inputs/outcomes and their sum is checked at every
step. These are arithmetic comparisons, not identified causal effects.

Confirmation/delayed, terminal ticks192..255:

| Family/dynamics | Full-input fit excess | View difference | Routing difference |
| --- | ---: | ---: | ---: |
| Additive/abrupt | .063278 | -.001412 | -.000614 |
| Additive/gradual | .070841 | -.002703 | -.001790 |
| Hierarchy/abrupt | .055053 | -.000395 | -.002572 |
| Hierarchy/gradual | .062410 | -.001177 | -.003925 |
| Local table/abrupt | .056201 | -.001724 | -.001778 |
| Local table/gradual | .065488 | -.004343 | -.006040 |

For example, local-table/gradual has .065488-.004343-.006040 = .055105 excess,
plus .194868 intrinsic noise, giving the observed .249973 Brier. The fit term's
paired diagnostic interval is [.050136,.080839], compared with view
[-.007876,-.000810] and routing[-.011614,-.000465]. Intervals use mean +/-3.5SE
over32 trajectories and are post-result descriptions, not fresh confirmation or
simultaneous/adaptive guarantees.

Negative view differences are possible because marginalizing an imperfect model
can reduce overfitting. They do not mean missing information improves the true
optimal forecast. The fixed mask63, which knows where these generators place
relevant coordinates, also improves several model scores; it is not a deployable
general observation policy or permission to hard-code those coordinates.

Pointwise convex-hull oracles show additional headroom but use the true q to
choose a mixture independently at each input. They cannot establish that a
coherent learned weighting policy can obtain it. Full raw models, fixed-view
models, all three issued views and both oracle bounds remain in the complete
[diagnostic summary](mmm-transfer-diagnosis-v117-summary.json), including all32-tick
blocks and both phases, not just the table above.

## Audit trail

- [Protocol](mmm-transfer-diagnosis-v117-protocol.md) and
  [raw artifact](mmm-transfer-diagnosis-v117.json) pin the unchanged v116 parent.
- Scoped race smoke tests cover all nine family/mode combinations under both
  schedules, plus rejection of changed labels and future fit origins. Passed:
  test10.49s/package11.744s. Scoped vet also passed.
- Generation completed in121.54s; full exact replay in122.49s. These are research
  reconstruction durations, not per-request or serving-performance benchmarks.
- Artifact contains9216 fitted profiles,294912 step rows, and157 matching source
  hashes. All154 parent hashes are preserved, with three diagnosis files added.
- Independent table-completion checks pass for3538944 partial probabilities.
  All1769472 decomposition identities pass. Reconstructed generic-own forecasts
  match the parent; fit origins and generator packets reconstruct exactly.
- Independent summary reproduced byte-for-byte after replay.
- Raw artifact size442915112 bytes, written exclusively with mode0600.

Artifact SHA-256:
`7f671f61df1af605d6f4a4f1cabacac947c4e275a0153b1109e69b332f07e94c`.

## Next lead

The [soft-response challenger proposal](../../research/soft-response-challenger-proposal.md)
uses primary research to motivate a bounded regularized logistic control, with
a separately labeled variational Bayesian extension and the existing context-tree
learner as a structural comparison. First isolate learner-level error at equal
evidence; then test a budget-matched composite through partial observation and
delayed feedback. Retain generic/Boolean specialists and original-regime protection
tests. No logistic-only additive success may stand in for cross-family success.

Neither proposal nor this diagnosis promotes a result to the daemon/whitepaper.
All seven directions remain open, including real tasks and loaded serving.
