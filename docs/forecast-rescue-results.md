# Complete-forecast rescue results

Frozen protocol: forecast-rescue-protocol.md. Design 2026090611, confirmation
2026090612; 2048 trajectories per split (32 scenario/baseline combinations,
64 trajectories each, 1000 predict-before-update steps). No hyperparameter search.
Both complete raw JSONs are retained. Approximate 95% simultaneous intervals
use trajectory means and 3.6 SE over 128 comparisons, not exact confidence
sequences. The protocol is locally frozen, not externally preregistered.

## Target failures: baseline .5

| Scenario | Old composed Brier | Grid composed Brier | Direct Beta | Rescue Brier | Gain over grid, simultaneous interval | Result |
| --- | --- | --- | --- | --- | --- | --- |
| Stationary .20 | 0.233243 | 0.234878 | 0.160166 | 0.161909 | 0.072969 [0.069875, 0.076064] | Validated in fixture |
| Stationary .80 | 0.233334 | 0.235088 | 0.161168 | 0.162907 | 0.072181 [0.069098, 0.075263] | Validated in fixture |
| Gradual .1 to .9 | 0.239799 | 0.241203 | 0.250839 | 0.205036 | 0.036167 [0.033769, 0.038564] | Validated in fixture |

This repairs the *served-law composition* failures, not the grid's raw stationary
estimation noise. The old fixed blend could not emit probabilities near .2/.8
from baseline .5; a direct predictive expert removes that artificial limitation.
Direct Beta still beats the rescue in the two stationary rows. This is model
selection cost, not suppressed evidence.

Against old composition: 30/32 simultaneous Brier intervals show improvement.
Two regressions occur when the baseline already equals the stationary rate:
at .2, excess Brier 0.000127, simultaneous upper 0.000156; at .8, excess
0.000150, upper 0.000198. All 32 meet the predeclared .003 mean-excess ceiling.
Universal superiority is still false. Relative to grid composition, the
baseline .5/stationary .5 case also regresses by 0.000130 (upper 0.000206).
These losses are small, not zero, and remain in the evidence.

## Runtime semantics and boundary

The selector is implemented, not just an evaluator: --forecast-rescue requires
--grid-belief --working-belief --evidence-trust-file /path/to/keys.json
--residual-mode disabled, and rejects contextual/hierarchical scoring. It
aggregates complete probability laws rather than changing rank deltas or
asserting a raw belief should compensate for a fixed scoring bias. The .25
belief-blend cap applies only to the blend expert, not the new final mixture.
This is an explicit change of composition contract, not a claim to preserve
the old output envelope. Defaults are unchanged.

The first rescue deliberately disables residual application so the law being
optimized is the final properly scored law. It has NOT yet validated coexistence
with active residual correction, improved ranking/agent answers, or a production
rollout. The model-level experiment omits runtime changepoint resets; separate
integration checks test actual reset, policy, journal, and persistence behavior.

Expert probabilities are committed as value arrays in the query journal, then
read back after authenticated outcomes. Updates occur in the existing atomic
outcome transaction and reuse its replay ledger. Delayed outcomes use original
forecasts in arrival order, not forecasts reconstructed with future evidence.
No immediate-feedback Bayesian interpretation is claimed for delayed order.
Epoch/policy mismatch cannot train the selector; changing the composition policy
invalidates served selector weights. Reset discards selector history and does
not score old-regime experts against the revealing outcome in the new regime.
Split children start with prior selector weights, not pooled performance.

Two review rounds caught and fixed a stale-policy reuse gap: rejecting old
journals alone did not invalidate already stored selector weights. The state
now carries its policy version and both serving and transaction paths enforce
it. Version checks inside the store lock/transaction also cover an epoch change
between service admission and commit. Value-copy, replay, restart, exact served
mixture, historical-forecast learning, reset and stale-policy tests pass. Full
go test, focused race tests, go vet, and go build passed.

The new model still depends on meaningful outcome labels, source independence,
and representative admission. It does not authenticate factual truth or rescue
unseen outcomes outside retrieval coverage. All data in this experiment is
synthetic; production and private conversations were not accessed.

## Performance

Apple M4, Go 1.27.0, GOMAXPROCS 10. Isolated 50-event, 32-dimensional hash
fixture, recall/pack 50/10. Three runs x500 requests per cell. Mixed requests
are 75% recall and 25% signed outcome writes. All arms disable residual
application and retain authentication; setup and external signing are excluded.
There is one learned posterior. Rescue may additionally serve explicit priors
for other eligible candidates, while controls skip absent posterior records;
therefore this is complete-mode cost, not an isolated arithmetic-overhead estimate.

| Workload | Two-hypothesis ms/op | Grid ms/op | Rescue ms/op | Rescue p99 range ms |
| --- | --- | --- | --- | --- |
| Serial recall | 7.242 | 7.275 | 7.463 | 9.02-12.62 |
| Serial mixed | 6.760 | 7.109 | 7.322 | 9.08-12.03 |
| Four-worker recall | 2.684 | 2.683 | 2.724 | 13.17-21.34 |
| Four-worker mixed | 3.707 | 3.726 | 3.736 | 33.04-33.98 |

Concurrent ms/op is inverse throughput. Four-worker mixed rescue adds a
descriptive 0.26% over grid; its maximum observed request is 59.96 ms. Small
nonrandomized runs do not establish causal overhead, a hard deadline, or repair
the previous 16-worker tail failure. Primitive predict+update median is 35.81 ns
with zero allocations (three runs x100,000; result retained in a benchmark sink).
This excludes grid/Beta computation, authentication, journaling and storage.

Reproduce:

```sh
go run ./cmd/forecast-rescue-experiment -seed 2026090611 > docs/forecast-rescue-design.json
go run ./cmd/forecast-rescue-experiment -seed 2026090612 > docs/forecast-rescue-confirmation.json
go test ./internal/service -run '^$' -bench '^BenchmarkEvidenceInternalRequests$/^mode=(two-nr|grid-nr|rescue)$' -benchmem -benchtime=500x -count=3 > docs/forecast-rescue-benchmark.txt
go test ./internal/bayes -run '^$' -bench '^BenchmarkForecastMixture$' -benchmem -benchtime=100000x -count=3 > docs/forecast-rescue-primitive-benchmark.txt
```
