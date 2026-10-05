# Augmented evidence v74: closer, speed target still failed

**FAIL the full frozen adoption screen.** Primary detection improves7.21% and
6.81%, with all confirmation non-harm/reliability checks passing. Both gains
remain below10%. No roadmap direction is complete and no serving change is made.

## Fresh paired results

10240 fresh streams across two phases/ten scenarios,512 steps each. All arms
receive one paid query per step. New IPW/augmented arms share selected outcomes,
isolating augmentation from the proposal change. Uniform and old-IPW controls
match the existing algorithms at every step on these fresh tapes.

Confirmation restricted mean delay (misses/premature alarms cost the whole
remaining horizon; lower is better):

| Scenario | Uniform | Old IPW | Variance proposal/IPW | Augmented |
| --- | ---: | ---: | ---: | ---: |
| Homogeneous128 | 151.47 | 151.52 | 151.78 | 152.59 |
| Sparse128 | 140.17 | 134.00 | 135.01 | 130.06 |
| Sparse256 | 136.85 | 131.26 | 132.59 | 127.54 |
| Negative256 | 139.00 | 132.28 | 133.30 | 129.05 |
| Sparse384 | 122.16 | 121.57 | 122.16 | 121.58 |
| Weak256 | 256.00 | 256.00 | 256.00 | 256.00 |

Primary gains are10.11 and9.31 steps, with paired z=3.3 lower gain bounds7.00
and6.24. Proposal-only changes slightly worsen old IPW; the control variate
accounts for the improvement in the matched-observation ablation. That is not
a guarantee that an arbitrary learned model reduces variance or stopping time.

Augmented has zero alarms in every confirmation null cell and no premature
alarms in alternatives. Zero/512 has Wilson95 upper0.745%, not zero population
uncertainty. Late sparse misses are329/512 uniform versus279/512 augmented,
with42 harmful and92 beneficial discordances. Net excess misses=-9.766
percentage points; the newly predeclared paired upper is-2.315 points, passing
the2% ceiling on fresh data. Other alternative upper bounds are1.277%.
**All arms still miss every one of512 weak changes.** Relative non-harm does
not establish adequate absolute sensitivity.

## Mathematical boundary

Z=w_I D_I+eta*(mbar-w_I m_I) has the target conditional mean when propensities
are correct, even with wrong historical predictions. eta is bounded from all
possible channels before the outcome, preserving positive wealth factors. It
does not clip the revealing outcome. eta clipping did not activate in these
streams, so it cannot explain the observed speed shortfall. Boundary tests
exercise the clipped case separately.

Source motivation: [Dudik, Langford & Li (2011)](https://icml.cc/2011/papers/554_icmlpaper.pdf).
Our bounded betting adaptation is not their policy optimizer. No robustness
against wrong propensity metadata is claimed. The null is population mean,
not Anti-Pigeon's full context-wise diameter certificate or causal evidence.

## Verification and artifacts

- Protocol: `mmm-augmented-evidence-v74-protocol.md`, frozen before fresh data,
  including the v73 paired bound and unchanged10% requirement.
- Raw: `mmm-augmented-evidence-v74.jsonl`, SHA256
  `17e043d2a607e3e672870bae9decdf9a5c54011d354fa0ee32e6d0f9106153ef`.
-17 source/protocol/dependency/evaluator hashes verified.
- Full10240-stream replay passed in5.89s, matching q/m/eta/action/outcome tape
  hashes, first alarms, clipping counts and query budgets.
- Focused race tests passed in1.440s: wrong-model unbiasedness, factor bounds,
  clipping, zero-model IPW parity, invalid state isolation, predictable proposals,
  old-control parity and adjacent evidence checks. Vet passed.
- Summaries: `node research/augmented-v74-summary.mjs`.
- Replay: `EVENTFRAME_AUGMENTED_REPLAY=<artifact> go test ./internal/observationgate -run '^TestAugmentedV74Replay$' -count=1 -v`.

The experiment took9.78s including old-control recomputation, not serving time.
Isolated Apple M4, Go1.27.1 darwin/arm64, GOMAXPROCS10, three500ms replicates:

```text
BenchmarkAugmentedObserve-10 1757481 334.8 ns/op 0 B/op 0 allocs/op
BenchmarkAugmentedObserve-10 1785938 334.2 ns/op 0 B/op 0 allocs/op
BenchmarkAugmentedObserve-10 1795255 334.3 ns/op 0 B/op 0 allocs/op
```

About0.335us for bounded policy/gate computation, excluding evidence acquisition,
random sampling, persistence, queues and agent generation. This is not an
end-to-end sub100ms claim or measured production throughput improvement.

Next investigate stopping-time-aware predictable bets using declared pre-query
bounds without lowering the error threshold. Fresh generators, delayed/missing
feedback, warning-trigger integration, provenance and real agent tasks remain.
Delayed feedback must retain original q/m/eta snapshots; recomputing from later
history needs a new validity analysis. No default, whitepaper, commit, remote
or production system changed.
