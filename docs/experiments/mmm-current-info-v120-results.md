# Current-state information result

Status: FAIL.672/672 non-harm checks pass;0/64 required improvement checks pass.
All2688 consumed runs retained. No production promotion or fresh confirmation.

Phase1, delayed schedule, terminal64 expected Brier:

| Case | Natural | Random | Entropy | Old heuristic | Current-state information |
| --- | ---: | ---: | ---: | ---: | ---: |
| Additive stationary | .221747 | .221572 | .221646 | .221398 | .221389 |
| Parity4 | .048364 | .048320 | .048299 | .048295 | .048304 |
| Null | .258712 | .258441 | .258582 | .258330 | .258261 |
| Majority to parity | .057088 | .056392 | .056608 | .056476 | .056522 |
| Parity to majority | .098247 | .094506 | .096658 | .095853 | .095908 |

Query budget identical to paid controls,31 per listed trajectory. Across the
whole artifact,37460/41664 acquisitions (89.91%) choose the same origin as the
older heuristic. The changed choices do not produce the required recovery gain.
Largest absolute per-trajectory Brier difference from that heuristic is.00982231;
this is a diagnostic maximum, not an uncertainty bound.

Exact joint path tests:64 comparisons PASS.252 prefix poisoning checks PASS;
outcome probability normalization and conditional-to-marginal mixture identities
checked for every query candidate. Immediate-complete schedule has0 acquisitions
and unchanged predictions. Source/input/protocol hashes retained. Initial full
job9.93s wall; one candidate only, unlike the four-policy27.65s previous job, so
these wall times are NOT a per-policy speed comparison or serving benchmark.

## Interpretation

The temporal information calculation is now exact within the declared model,
yet benefit remains small. That rules out this particular approximation as the
main missing rescue on these tapes. It does not prove that information acquisition
is useless or that the switching model is calibrated to actual agent behavior.
Queries use a perfect label service; experts are still not retrained.

[Houlsby et al. (2011), equations2 and6](https://mlg.eng.cam.ac.uk/pub/pdf/HouHusGha11a.pdf)
motivates expected entropy reduction and marginalizing nuisance parameters.
Here the target is the current expert identity, with active/stopped and old
states marginalized. The finite delayed-state implementation is our adaptation;
their empirical results and guarantees do not transfer automatically.

Next distinct question is whether reducing uncertainty about expert identity is
the wrong objective: several experts can make essentially the same useful
prediction. Test decision-relevant expected predictive-risk reduction, not another
entropy-score tuning. Also distinguish mixer-only interventions from actual
expert retraining; no fixed-tape experiment can establish that feedback improves
the underlying representations. Preserve both equal-cost and no-acquisition
controls. All seven full directions remain open.

Artifacts: [protocol](mmm-current-info-v120-protocol.md),
[results](mmm-current-info-v120.json), [replay](mmm-current-info-v120-replay.json),
[initial timing](mmm-current-info-v120-timing.txt). Full replay is byte-identical
(`cmp` exit0).
