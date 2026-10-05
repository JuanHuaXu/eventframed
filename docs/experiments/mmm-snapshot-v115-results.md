# Snapshot-specialist v115 results

Status: FAIL. No promotion. The frozen eleven-arm comparison completed1,536
schedule-runs over768 latent trajectories. See the
[protocol](mmm-snapshot-v115-protocol.md) and
[complete machine-readable summary](mmm-snapshot-v115-summary.json).

## Required gates

- 754/1,018 pass overall.
- 739/960 non-harm gates pass;221 fail.
- 15/58 gain gates pass;43 fail.

Failures are not confined to one recovery direction: parity3, parity4,
complement4, dependent4, null, and both switch families have failed gates.
All results remain in the summary; no favorable-cell substitution or new
threshold is used. Each gate uses the frozen paired mean +/-3.5SE across32
trajectories. These are approximate fixed-sample screens, not confidence
sequences or guarantees across the adaptive research history.

## Confirmation, delayed/missing, late segment

Lower Brier is better. These illustrative cells do not replace the full gates.

| Case | Generic | Arrival log | Fixed Markov | Matched12800 | Snapshot |
| --- | ---: | ---: | ---: | ---: | ---: |
| Parity4 |.074745|.053540|.053064|.053064|.066214|
| Majority to parity |.257427|.220530|.216170|.216170|.245765|
| Parity to majority |.218482|.219478|.215305|.215305|.228753|

Against the matched-budget control, snapshot harm is .029594 with interval
[.015343,.043845] for majority-to-parity, and .013448 [.004594,.022302] for
parity-to-majority. Snapshot expected accuracy is94.27%,62.96%,68.37% in these
three cells; matched control94.45%,68.96%,72.24%. These are synthetic expected
accuracies, not general agent performance or a claim to retain94.7% everywhere.

Across all reported phase/case/schedule/segment means, the largest absolute
matched-budget versus original-Markov Brier difference is .0006023 (confirmation
immediate reverse-switch late segment). The control weakens an explanation based
solely on the changed threshold. It does not isolate every interaction of bank
size, alternative population, retained rejection and admission/retirement.

The implemented bounded snapshot strategy is rejected under this protocol. This
does not refute the growing-expert theorem or prove all immutable-model methods
fail. Component correctness, stale-ID repairs and quality are separate claims.

## Audit

- Original-budget matched control passes immediate/delayed path, probability
  (within1e-12) and accounting parity against archived Markov.
- Source compatibility preserves all nine v113 arms on consumed test streams.
- Compatibility, seed and control race suite PASS,8.537s; vet PASS.
- Generation PASS,258.44s.
- All58 frozen source hashes verified.
- Independent evaluator reconstructs expected/realized scores, accuracy,
  acquisition costs, as-of fitting origins, selector clocks, delayed/censored
  accounting and paired latent streams. Summary recomputation is byte-identical.
- Artifact SHA256:
  `7a54a527874fd133a9c1fff93bbb1c3eb5cbedfbec361e51f9876d49c889e3d2`.

Full deterministic replay PASS,264.90s. The exclusive artifact is472,783,664
bytes with0600 permissions.

[Whole-fixture benchmarks](mmm-snapshot-v115-benchmarks.txt) retain all three
repeats per schedule:154.21-176.86ms immediate and194.09-195.39ms delayed/missing,
about47.8MB allocated per operation. One operation includes256 frames, fits and
all eleven policies on a previously consumed v113 design trajectory. These are
not per-query serving latencies or loaded p99 measurements. The earlier
candidate-only journal cost remains recorded separately; a complete-policy
quality failure cannot be offset by passing a component runtime test.

## Next action

Preserve this failure and the unchanged controls. Do not sweep incoming mass,
retirement periods or rejection persistence against these confirmation outcomes.
Prioritize [independent-generator transfer](../../research/independent-generator-transfer-proposal.md)
before another weighting variant. That directly addresses the missing scope of
direction1 rather than counting another seed block as independent construction.
All seven roadmap directions remain open; this result does not authorize a
production, private-data, whitepaper or GitHub change.
