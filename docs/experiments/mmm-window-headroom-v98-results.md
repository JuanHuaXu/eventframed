# v98: available forecast headroom after a regime change

## Status: diagnostic, not confirmation

All128 selected v97 switch-stream records reproduce exactly. No original
forecast, score, training window or weight update changed. The other640 v97
records are outside this targeted diagnosis; full768-record replay remains
the separately documented v97 check.

The [protocol](mmm-window-headroom-v98-protocol.md),
[raw pre-outcome forecasts with returned outcomes](mmm-window-headroom-v98.json),
[summary](mmm-window-headroom-v98-summary.json) and
[independent evaluator](../../research/window-headroom-v98-summary.mjs) expose
eight32-step blocks per stream and both views. Simulator truth is used only
after the unchanged stream is scored, for retrospective evaluation.

## The useful window is brief

Confirmation-phase, full-view parity-to-majority, means across32 consumed
streams:

| Steps | Generic64 Brier | Generic32 | Bank | Best fixed pair | Oracle hull |
|---|---:|---:|---:|---:|---:|
| 128-159 | 0.370809 | 0.298873 | 0.396044 | 0.298142 | 0.244501 |
| 160-191 | 0.197247 | 0.089953 | 0.187493 | 0.089099 | 0.076441 |
| 192-223 | 0.054976 | 0.108610 | 0.066670 | 0.054973 | 0.053657 |
| 224-255 | 0.059084 | 0.116395 | 0.061580 | 0.059058 | 0.056799 |

During128-159 BOTH training windows still precede the change; the short model
is not using new-regime evidence yet. It is less harmful, not newly informed.
At160 the short window becomes entirely post-change. By192 the long window
also does, and retaining less evidence becomes a disadvantage.

Mean retrospectively optimal short-model weight is97.47%,95.84%,0.58%,2.19%
over these four blocks. Actual mean bank weight is0.23%,1.59%,3.78%,3.16%.
These oracle weights are fitted separately for each stream/block using hidden
simulator probabilities. They are NOT one deployable schedule or a validated
rule, and using the known switch location would leak future knowledge.

There is substantial available improvement in the first two blocks without
adding new forecast families. Bank gain against generic64 is-0.025235 then
0.009754; the retrospective fixed-pair gains are0.072667 and0.108148. Later,
the oracle hull gains are only0.001319 and0.002285. Demanding a0.005 gain in
each of those later blocks from reweighting these forecasts would exceed
their observed mean attainable headroom. This does not change the existing
whole-late-half gate or reclassify v97's failures.

The other direction corroborates the timing pattern: in majority-to-parity,
generic32 helps in128-159 and160-191 but hurts after192. The complete summary
includes both phases, both directions, both views and all early blocks.
Partial-view oracle estimates may exploit simulator information unavailable
to the predictor and must remain optimistic diagnostic bounds.

## Consequence for the next rescue

The evidence argues against a permanent boost for short memory. It supports
testing whether forecasts can be rejected when their own returned outcomes
contradict them, with re-entry tied to a new model version. See the
[version-scoped falsification proposal](../../research/forecast-falsification-proposal.md).
This is the next hypothesis, not a demonstrated causal explanation or a
successful rescue. Existing quality criteria and failed records remain intact.

## Verification

- Race math and unchanged-parent smoke tests PASS:2.049s package time.
- Generation PASS:20.382s; full128-stream diagnostic replay PASS:20.412s.
- Parent artifact and all18 source hashes checked.
- Independent evaluator rechecks simulator probabilities, every recorded
  forecast's aggregate expected and realized scores, bank convex composition,
  publication weights and all oracle/block calculations.
- Endpoint, equal-forecast and interior optimum cases pass. Hull loss is no
  greater than bank or fixed-pair loss; fixed-pair loss is no greater than
  either endpoint within numerical tolerance.
- Vet, exact summary reproduction and new-file whitespace checks PASS.
- Artifact SHA256:e01f24234e9826ea981778ab2b5b79a2d678903c9d73ccfce9c70e4e8016fd0f.

No new production algorithm was introduced, so there is no serving-performance
claim or new latency benchmark. Diagnostic execution times above include
reconstructing the existing models and writing/reading traces. No whole
research direction is complete or exhausted by this result.
