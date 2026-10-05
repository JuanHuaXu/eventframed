# Signal-reliability averaging v6: partial rescue, FAILED full screen

Completed1,792 fresh episodes and5,376 scored trajectories. All arms use fixed
Joint-Gini acquisition and consume the same16 outcomes plus eight signals.
`RELIABILITY_PROTOCOL.md` froze the model and gates; raw traces, model weights,
calibration and source hashes are in `reliability-v6.json.gz`. The summary JSON
retains all contrasts and the failed condition. No deployed behavior changed.

## Confirmation results

Lower multiclass final Brier is better; each row has128 paired episodes.

| Case | Fixed | Reliable only | Model average |
| --- | ---: | ---: | ---: |
| Independent20 | 0.181594 | 0.170596 | 0.174651 |
| Copied20 | 0.484818 | 0.498400 | 0.486876 |
| Mixed20 | 0.338002 | 0.323570 | 0.330216 |
| Matched05 | 0.038342 | 0.050374 | 0.040425 |
| Matched20 | 0.332378 | 0.327732 | 0.327592 |
| Matched random signal20 | 0.323490 | 0.344185 | 0.330274 |
| Matched misleading signal20 | 0.321288 | 0.407851 | 0.298984 |

The misleading-signal rescue passed its frozen component gate: gain0.108867
against reliable-only, paired z3.3 descriptive interval[0.015026,0.202708].
Accuracy was78.125% versus74.219%, confidently-wrong outcomes2/128 versus8/128.
Against fixed, mean final gain0.022304 had interval[-0.004138,0.048745]; this
does not establish improvement over the fixed control.

All seven confirmation cases met the <=0.01 mean curve/final harm ceilings
against fixed. This is screening, not confidence-certified non-harm. The full
screen FAILED: independent20 curve Brier was0.331798 for averaging versus0.318537
reliable-only, harm0.013261 above the0.01 ceiling. Its interval[0.004394,0.022128]
supports a learning-speed tradeoff. Independent20 final harm0.004055 met that
separate mean ceiling.

The first split also showed independent20 curve harm0.010438 and final
harm0.020735 versus reliable-only. Its misleading-signal final gain0.046201 had
an interval spanning zero. Replication is needed; the confirmation's large gain
should not be treated as a stable effect-size estimate.

## What the model learned

Mean final weights on independent20 were[0.592,0.345,0.063] for
reliable/uninformative/reversed; misleading signals gave[0.094,0.318,0.588].
Random signals retained broad uncertainty[0.316,0.395,0.289]. These are weights
under the finite model, not authenticated probabilities of source truth.

With no repeated reports from a group, independent and copied report likelihoods
coincide. The test confirms that reliability weights do not move after one
observation per group. Reliable-mode certainty cannot arrive for free before
discriminating evidence. Marginalizing shared reliability also couples source
modes; independently averaged priors would lose that coupling.

## Verification and limits

`python3 -m unittest test_reliability_v6 -v` passed three tests, including full
replay of all1,792 episodes, hashes and identical acquired observations. Direct
enumeration over12,288 joint states agrees with reliability and hypothesis
marginals initially and after five observations, including disagreements.
Forecast/mode-weight normalization and unique source slots pass.

Exactness is conditional on the finite family and frozen calibration likelihoods.
Calibration uncertainty, within-episode changes in signal reliability, arbitrary
collusion and missing hypotheses remain outside it. Model averaging is described
by [Hoeting et al. (1999)](https://sites.stat.washington.edu/www/research/online/hoeting1999.pdf);
this provenance application does not inherit empirical validation from that
paper. No runtime or full-request latency claim was tested.

## Next lead

Test whether joint-model acquisition obtains discriminating reports earlier at
the same budget, with fixed acquisition as control. Start with exact one-step
target-Gini acquisition under the mixture; consider bounded lookahead only if
diagnostics show missed delayed value. Preserve independent20 speed protection
and misleading-signal rescue gates. A longer horizon may diagnose information
shortage but cannot substitute for an equal-budget pass. No adoption yet.
