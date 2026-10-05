# v111 compatibility handoff: failed quality rescue

Both candidates FAIL the frozen complete quality screen. The implementation
and complete replay pass; this is a negative research result, not a numerical
or lifecycle error. Neither candidate is promoted.

| Candidate | Non-harm gates | Gain gates | Complete gates | Result |
| --- | ---: | ---: | ---: | --- |
| Publication handoff | 667/672 | 38/46 | 705/718 | FAIL |
| Publication plus pending transfer | 662/672 | 29/46 | 691/718 | FAIL |

All failures concern switch scenarios. Each candidate passes all560 stationary
non-harm gates and all480 non-harm comparisons against controls0..4 across the
whole experiment. This does not rescue the result: the previous log policies
are stronger controls, and some original generic/conservative recovery gains
also remain unproven. Do not compare raw pass counts with v109 as though the
two studies used the same number of controls or gates.

The [protocol](mmm-compatibility-v111-protocol.md) uses768 fresh latent
trajectories, two paired schedules (1,536 runs), nine policies and256 scored
frames per run. Both phases were completed without tuning or interim outcome
inspection. Mean +/-3.5SE intervals are approximate fixed-sample screens over
32 trajectories, not confidence sequences or history-wide guarantees.

## Confirmation, delayed late segment

Brier is lower-is-better. The late segment covers frames128-255, including the
entire declared recovery window rather than a favorable final block.

| Case | Generic | Conservative | Arrival Brier | Log/no-neutral | Publication | Plus pending |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Stable parity4 | .072719 | .069735 | .061315 | .052553 | .052406 | .051334 |
| Majority to parity | .260505 | .256075 | .256813 | .215067 | .221338 | .230097 |
| Parity to majority | .220525 | .220912 | .224389 | .216708 | .215773 | .219483 |

Stable parity expected accuracy is94.58% for publication handoff and94.67%
with pending transfer, versus94.45% for original log and94.30% for arrival
Brier on these particular fresh trajectories. This remains finite synthetic
accuracy, not general chatbot accuracy or a universal94.7% guarantee.

Publication-only gives just .000935 mean Brier improvement over original log
in reverse recovery, with interval[-.007864,.009734]. It is below the .005
gain floor and its interval includes zero. Against conservative, mean .005139
has interval[-.001198,.011475], also failing the declared improvement test.

In majority-to-parity, publication-only mean harm versus original log is
.006270, interval[-.004962,.017503]: non-harm is not established, though this
interval does not establish positive harm either. Pending transfer produces
.015030 mean harm, interval[.003660,.026399], a clearer adverse result. Accuracy
also falls there: original log69.58%, publication68.76%, pending67.78%.

## Pending transfer is not the rescue

A descriptive paired ablation compares arm8 directly with arm7. These intervals
are post-hoc diagnostics, not additional predeclared gates or an excuse to
change the original verdict. Positive values mean pending transfer is worse.

| Phase | Delayed late case | Pending minus publication Brier | Approximate interval |
| --- | --- | ---: | --- |
| Design | Majority to parity | .008178 | [.003267,.013089] |
| Design | Parity to majority | .006178 | [.002362,.009994] |
| Confirmation | Majority to parity | .008759 | [.003880,.013638] |
| Confirmation | Parity to majority | .003710 | [.000274,.007147] |

All four intervals are positive. This paired policy comparison includes any
downstream observation-path changes; it does not isolate a single label's
causal influence. Nevertheless it argues against retaining this pending-loss
attenuation as a successful rescue. Forecast movement need not mean the old
evidence should be weakened: movement may reflect useful learning itself.
That interpretation is a hypothesis, not established solely by these means.

## Integrity

- Seven frozen v109 controls match on all eight consumed compatibility runs
  (four scenarios, both schedules), including predictions, fit origins, scores
  and accounting. Compatibility plus effective-seed audits pass under race in
  6.491s. Component race/lifecycle tests were recorded separately.
- Full generation completes in207.76s; every one of the1,536 records replays
  exactly in210.68s. No artifact or source was altered to repair a failing gate.
- The independent evaluator reconstructs every issued Brier/log score,
  accuracy, cost, label clock, as-of fitting origin and journal count. It checks
  37 source hashes and reproduces the summary byte-for-byte. Package vet passes.
- All768 latent trajectories are unique and their schedules share the same
  rule/input/outcome stream. Delays and missingness do not create a second
  independent replicate of the latent trajectory. Both candidates retain the
  unchanged model/gate budgets and use no future labels.

Raw artifact389,041,677 bytes, created0600. SHA256:
`5b7807bdd41a93222c70c9435b55cfeacc8c37f72f16438f58db0eca602de017`.
Artifacts: [records](mmm-compatibility-v111.json),
[independent summary and every gate](mmm-compatibility-v111-summary.json).

## Performance and next lead

The previously recorded [paired component measurements](../../research/compatibility-handoff-component-results.md)
remain valid: extra work did not buy a complete quality improvement.
After all quality/replay/verification processes stopped, the complete nine-arm
fixture was benchmarked on consumed design parity4 with all repeats retained:

```sh
go test ./internal/observationlearners -run '^$' -bench '^BenchmarkCompatV111WholeFixture$' -benchmem -benchtime=1x -count=3
```

Apple M4, darwin/arm64, Go1.27.1, suffix10. Immediate runs take131.45-158.65ms,
delayed/missing runs128.95-130.20ms; each operation allocates about27.0MB.
This includes fitted tables,256 frames and all nine policies. It is not one
forecast, a database query, or a serving-tail benchmark. The first/cold repeat
is retained. [Raw benchmark output](mmm-compatibility-v111-benchmarks.txt).
Benchmark source SHA256:
`7c1fc5a6bff5811b4761fceb37b50d87741922a7257bfb0a6b84fbc23d4af894`.

Next is [explicit delayed Markov advice](../../research/delayed-markov-advice-proposal.md):
condition on past evidence at its origin and propagate state-transition
messages forward, rather than attenuating it according to forecast distance.
The first requirement is equality with literal state-path enumeration and the
unchanged immediate log/share policy. This is a new untested mechanism, not a
claim that the declared arrival-log implementation was buggy.

All seven research directions remain open. No production/OpenClaw changes,
new private data, commits, pushes or whitepaper promotion.
