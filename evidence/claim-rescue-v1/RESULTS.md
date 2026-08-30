# Claim Rescue and Replacement v1 Results

## Protocol and provenance

- Final protocol: `eventframe.claims-experiment.protocol.v7-rescue-replacement`
- Failed first design: seed `86028121`, retained as `design-v1-failed.json`
- Final design: seed `961748941`, retained as `design-v2.json`
- Untouched confirmation: seed `104729951`, bound to design seed `961748941`, retained as `confirmation.json`
- The original falsified propositions remain falsified. These results evaluate the replacement propositions in `docs/proposals/claim-rescue-and-replacement-v1.md`.

## Result register

| Target | Result | Evidence |
| --- | --- | --- |
| Bounded frontier-all plus selective deep work | **Validated in synthetic mechanism fixture** | Its Brier, priority-weighted Brier, and Recall@10 outputs exactly matched frontier-all. The old 5%-cheap-update policy remained materially worse. |
| Bayesian update repairs post-shift Recall@10 | **Still falsified as written** | Frontier-all and the replacement both remained at `0.3000` post-shift Recall@10. The valid replacement is posterior-conditioned rank correction inside the nominated frontier, which has separate earlier evidence. |
| Safe residual expert | **Replacement validated in confirmation fixture** | Mean Brier gain `0.0137795`, trajectory bootstrap 95% interval `[0.0090405, 0.0186310]`; worst-trajectory excess loss `0.0090879` below the frozen `0.02` budget. |
| Five-percent individually harmful residual reuse | **Still falsified; not used as the replacement** | `256/861`, or `29.73%`, of applied confirmation corrections were individually harmful. Abstention bounded trajectory loss but cannot foresee the first unannounced reversal. |
| Practical-equivalence group recognition | **Replacement validated in confirmation fixture** | Compatible groups: `61/64` share recommendations; moderate divergent: `60/64` split; strong divergent: `64/64` split; zero wrong terminal decisions. All share recommendations still required external Anti-Pigeon authority. |
| Omitted-influence estimator coverage | **Validated only in the declared synthetic finite-population fixture** | `256/256` upper bounds covered exact finite-population influence; Wilson 95% lower bound `0.9852`. Runtime certificates are query-journal and omitted-population bound. Mean UCB was `0.21999` for mean true influence `0.00526`, so practical certificate power at this population and sampling rate remains weak. |
| Predictive graph affects runtime output | **Mechanism validated** | Integration proved publication changes graph features and post-contract rank deltas only for nominated candidates; rollback removes both. Untouched forecast improvement remains untested. |
| Priority deployment constraint | **Mechanism validated; empirical claim inconclusive** | The evaluator blocked a policy with better aggregate Brier but additional high-priority misses and refuses evaluation without high-priority observations. No new independent real trajectories were available. |
| Matched structured-frame ablation | **Mechanism validated; empirical claim inconclusive** | The evaluator rejects unmatched source/model/budget/ranking contracts and refuses readiness below three trajectories or without blinded ratings. No new qualifying holdout was available. |
| Revised changepoint monitor | **Mixed; current confirmation missed one frozen criterion** | Five scenarios passed. Gradual drift detected `59/64` changes, but `0.203125` unmatched alarms per trajectory narrowly exceeded the `0.20` ceiling. No threshold was changed after confirmation. |

## Performance

On Apple M4, bounded 200-node/400-edge graph propagation measured
`84.9-92.1 us/op` across five runs. The in-memory 1,000-event recall benchmark
measured approximately `1.52-1.53 ms` p50 and `3.13-3.36 ms` p99. These exclude
external embedding, LibraVDB RPC, production concurrency, and cold-start costs.

## Verification

- `go test ./...`: pass
- `go test -race ./...`: pass
- plugin `npm test`: 18/18 pass
- graph/audit integration controls: pass; machine-readable output in `integration-tests.jsonl`
- priority/representation evaluation controls: pass; machine-readable output in `evaluation-contract-tests.jsonl`
- benchmark records: `benchmarks.txt` and `graph-benchmark.txt`

## Interpretation

The rescue supports cheap frontier-all updates plus bounded specialists. It does
not support fixed-percentage suppression of cheap updates. Residuals are now
cumulative-harm-bounded experts rather than per-application safety guarantees.
Practical equivalence improves statistical borrowing but never grants sharing
authority. Graph snaps now have an output path, but outcome benefit still
requires chronological confirmation. Structured frames and priority remain
evidence gaps rather than implementation gaps.
