# Timely-label acquisition result

Status: FAIL added-value criteria.504/504 non-harm checks pass;0/48 required
improvement checks pass. All2688 consumed trajectories retained. Not a fresh
confirmation or production change.

Phase1, delayed schedule, terminal64 expected Brier:

| Case | Natural only | Random | Entropy | Disagreement |
| --- | ---: | ---: | ---: | ---: |
| Additive stationary | .221747 | .221572 | .221646 | .221398 |
| Parity4 | .048364 | .048320 | .048299 | .048295 |
| Null | .258712 | .258441 | .258582 | .258330 |
| Majority to parity | .057088 | .056392 | .056608 | .056476 |
| Parity to majority | .098247 | .094506 | .096658 | .095853 |

Each paid policy uses31 unit-cost queries in these cells, versus0 for natural.
Paid counts match per trajectory; immediate complete delivery produces0 queries.
Current predictions precede acquisition, and new labels affect forecasts only
from the next clock.1008 prefix poisoning checks pass. Natural control matches
the previous baseline-heavy switch mixture to1e-12. Underlying expert forecasts
are fixed; extra labels do NOT retrain them. A perfect label service is assumed.

Earlier labels modestly help some means, but this screen does not establish that
the tested disagreement policy spends the same acquisition budget better than
random or entropy. Do not promote it or erase the earlier v4 disagreement failure.

## Source and limitation

[Houlsby et al. (2011), equation2](https://mlg.eng.cam.ac.uk/pub/pdf/HouHusGha11a.pdf)
expresses parameter information gain as predictive entropy minus expected
conditional entropy. Our weighted Jensen-Shannon heuristic has that algebraic
form, but applies CURRENT expert weights to OLDER issued forecasts. In a
switching-state model this is not automatically the information that the old
label supplies about the CURRENT state. The frozen protocol calls it predictive
disagreement, not an exact BALD implementation. No paper guarantee transfers.

One distinct next lead: compute the joint distribution of an outstanding label
and the current latent expert through the switching model, then acquire the
label that maximizes expected CURRENT-state entropy reduction. Compare directly
with this heuristic at identical times/costs. Marginalize past states rather than
treating their identity as the current state's identity. This can be calculated
by exact forward/backward messages or two hypothetical evidence refilters per
candidate, without consulting hidden Q/Y. The added compute belongs in the slow
path. It remains a working-model hypothesis, not an established rescue.

Artifacts: [protocol](mmm-expedite-v120-protocol.md),
[results](mmm-expedite-v120.json), [replay](mmm-expedite-v120-replay.json),
[replay timing](mmm-expedite-v120-replay-timing.txt). Replay is byte-identical
(`cmp` exit0); measured replay wall time27.65s. This is a complete offline research
job runtime, not per-request latency or a loaded serving benchmark.
