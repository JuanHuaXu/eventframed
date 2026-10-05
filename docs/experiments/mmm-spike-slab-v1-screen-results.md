# Multi-clock screen: not a replacement for the controls

Frozen contract: `mmm-spike-slab-v1-screen-contract.md`. All2,016 fits and
64,512 forecasts completed in276.39s. All fits met numerical convergence,
5-123 iterations. No model settings changed from the84-fit pilot.
This is the consumed v120 tape, indices0-7 at clocks0/128/224; it is not
untouched confirmation or a full trajectory/adaptation-delay evaluation.

## Quality diagnostics

Mean expected Brier difference (mixture minus control; positive is worse):

| Clock | Mean plug-in | Generic64 | Boolean64 | Markov |
| ---: | ---: | ---: | ---: | ---: |
| 0 | +.002179 | +.000842 | +.027099 | +.020884 |
| 128 | -.000593 | -.000661 | -.006137 | +.006603 |
| 224 | +.000259 | +.007289 | -.008603 | +.011411 |

Scenario/phase/schedule cells with mean harm >.01, out of84 at each clock:

| Clock | Mean plug-in | Generic64 | Boolean64 | Markov |
| ---: | ---: | ---: | ---: | ---: |
| 0 | 10 | 30 | 38 | 42 |
| 128 | 0 | 18 | 9 | 20 |
| 224 | 0 | 25 | 5 | 21 |

These are descriptive eight-trajectory means, not simultaneous confidence
bounds. Raw per-index paired differences and standard errors are preserved
in `mmm-spike-slab-v1-screen-summary.json`; all phase/schedule/clock expected
and realized scores are in `mmm-spike-slab-v1-screen-audit.json`.

Large examples: initial parity1 harm versus generic64 reaches+.12713;
midstream mux3 reaches+.09952; late parity-to-majority reaches+.08958.
The Markov comparison also shows large mux3 losses. Initial frames use only
16 admitted observations; later publications use up to64. Integration of
the variational law helps on average midstream but hurts early and slightly
late relative to the mean plug-in. Do not generalize the first pilot's
integration benefit to all times. Markov updates within the publication
window and remains a system comparator rather than a matched-cadence ablation.

## Audit and decision

All source origins, evaluator fields, factor masses, convergence conditions,
and saved query moments were independently checked. Maximum reconstruction
errors: mean1.7764e-15, variance8.8818e-16. Maximum2,974 integration evaluations,
zero range clamps. All84 overlapping old pilot records match exactly.
As-of poisoning, admitted-label complement and invalid-clock tests pass
under race (11.388s package). Independent full255-component integration is
still not claimed; see the separate small-mixture numerical checks.
The final full spike race suite, including the new original-likelihood
diagnostic, passes in21.309s. Generalizing the independent auditor preserves
all numeric outputs for the old84-fit pilot (only its limitations text was
clarified); compatibility artifact is retained.

Full replay took276.64s and is byte-identical. Source/pilot overlap checks
cover all84 shared records, not just forecasts. Final source SHA256 remains
`5f576bfee45a3354e9fe164d1436e78591a70417d5ac71f0c9dbc4b4c646e24f`.
Screen and replay SHA256:
`e7fac54246a0823916af9df11539572a36cf464eb418bbc589786dc77e2eff82`.
Audit SHA256:
`79882143802ea6a6a8b08023add5ba1399dff53b35f3b6c32c921db5141f4a67`.
Summary SHA256:
`792174e716a93c922ce49f43ee74d9ff823c06d5477f16ebba43641a9ed3f150`.

Collection command: `EVENTFRAME_SPIKE_SCREEN_SOURCE=../../docs/experiments/mmm-soft-learners-v120.jsonl EVENTFRAME_SPIKE_SCREEN_OUTPUT=../../docs/experiments/mmm-spike-slab-v1-screen.jsonl go test ./internal/observationlearners -run '^TestSpikeScreenCollect$' -count=1 -timeout=20m -v`.
Output paths are exclusive; use a new path for another replay. Audit command:
`node research/spike-slab-v1-pilot-audit.mjs docs/experiments/mmm-soft-learners-v120.jsonl docs/experiments/mmm-spike-slab-v1-screen.jsonl NEW_AUDIT.json screen`.
Summary: `node research/spike-slab-v1-screen-summary.mjs NEW_AUDIT.json NEW_SUMMARY.json`.

Decision: do not promote this fixed-prior candidate or spend a full32-index,
all-clock confirmation run on it unchanged. The screen contradicts a blanket
improvement claim. Preserve the numerically correct component and all negative
cells. Next isolate fixed-prior inference error with original-likelihood
quadrature before choosing among prior adaptation, structured alternatives
and cadence changes. New primary-source leads and caveats are in
`research/spike-slab-rescue-source-note.md`. No goal is complete and production
is untouched.
