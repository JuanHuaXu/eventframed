# Frozen spike-and-slab pilot: complete, mixed quality

All84 planned fits and2,688 forecasts completed after the numerical log-odds
repair. No prior, data subset, coordinate equation, compute cap or tolerance
was retuned. All84 met the numerical bound-and-state stopping criterion in
5-42 iterations. Initial complete collection and replay each took13.85s;
their full JSONL artifacts are byte-identical. This is an already-consumed
one-index exploratory pilot, not untouched confirmation or a completed goal.

## Quality

Expected Brier (lower is better),21 trajectories/672 forecasts per cell:

| Phase / schedule | Mixture | Mean plug-in | Generic64 | Boolean64 | Markov |
| --- | ---: | ---: | ---: | ---: | ---: |
| 0 / immediate | .187476 | .188096 | .184342 | .189741 | .178898 |
| 0 / delayed/missing | .188413 | .189092 | .182553 | .189745 | .179315 |
| 1 / immediate | .182602 | .183422 | .195279 | .183052 | .176164 |
| 1 / delayed/missing | .185928 | .186948 | .189211 | .187908 | .181535 |

Realized Brier, same arm order:

| Phase / schedule | Mixture | Mean plug-in | Generic64 | Boolean64 | Markov |
| --- | ---: | ---: | ---: | ---: | ---: |
| 0 / immediate | .194913 | .195831 | .193626 | .196354 | .188473 |
| 0 / delayed/missing | .192612 | .193442 | .188852 | .194298 | .185871 |
| 1 / immediate | .181261 | .182080 | .201704 | .189415 | .188092 |
| 1 / delayed/missing | .182955 | .183941 | .186026 | .189997 | .184482 |

Interpretation: numerical mixture integration consistently improves on the
mean plug-in in these cell averages. The model improves over generic64 in
phase1, not phase0, and trails Markov in expected Brier in every cell. It is
substantially better than the earlier Gamma convergence diagnostic (~.258-.273
expected Brier), but that comparison changes the prior/model, not just code.
Do not attribute the quality improvement to fixing probability saturation.
Markov is a system control with within-window adaptation, not a matched-cadence
ablation. No full trajectory non-harm/recovery gates or confidence intervals
are established by this one-index pilot.

## Verification

The independent Node auditor checks every source origin, evaluator field,
saved logit mean/variance, factor dimensions/masses, bound monotonicity,
stopping condition and quadrature accounting. Maximum reconstructed moment
errors: mean1.7764e-15, variance2.2205e-16. Max2,337 integrand evaluations;
zero range clamps. This reconstructs moments, not the entire255-component
predictive integral independently. Small-mixture/normal-domain quadrature
comparisons provide the separate numerical component tests, not a posterior
calibration certificate.

The full spike race suite passes (final9.780s), including a regression that
failed before repair at finite log-odds118.99846, dense/fast comparison,
extreme-tail entropy/variance/characteristic atom tests, ownership, poisoned
future/evaluator inputs, and admitted-label positive control. The completed
pilot contains two inclusion masses rounded to1; their exclusion tails and
finite log-odds remain represented. Maximum fitted log-odds37.37449.

## Performance repair audit

Stable odds initially raised the old pi=.1 fixture fit to474.8-484.2ms from
the historical74.2-75.5ms. Inspection found repeated scalar conversions in
sample-by-feature loops. Hoisting probabilities and new-state moments into
bounded cycle-local arrays removes repeated conversions; no cross-cycle cache
or new lifecycle is introduced.

The first hoist changed arithmetic grouping: forecast differences<=4.45e-16,
trace differences<=4.27e-14. Preserve this intermediate artifact as a failed
bit-identity check. Hoisting probabilities rather than multiplying means
early preserves the original operation order. Final full collection12.55s
is byte-identical to BOTH complete pre-hoist runs, including all factors,
traces, quadrature accounting and forecasts. The new regression verifies
exact predictor and prepared-moment arithmetic across successive states.

Final Apple M4 benchmarks, three repetitions of three operations:

| Operation | ns/op repetitions | B/op | allocations/op |
| --- | --- | ---: | ---: |
| Full fit, original pi=.1 fixture | 117050569 / 117013820 / 117103056 | ~58911100 | 6718-6720 |
| Mixture prediction, same fixture | 3170278 / 3146722 / 3121972 | 8192 | 4 |
| Full fit, frozen pilot pi=1/255 on same fixture | 9517639 / 9568389 / 9499722 | ~4932230 | 571 |

Command: `go test ./internal/observationlearners -run '^$' -bench '^BenchmarkSpike(SlabFit|MixturePrediction|PilotPriorFit)$' -benchtime=3x -count=3 -benchmem`.
These are research operations, not serving tail latency; allocations are not
peak RSS. Different priors change convergence cost, so9.5ms is not a pure
code speedup over117ms. Same-model hoisting cuts475-484ms to117ms (~75%),
but remains slower than the historical probability-only implementation.
Prediction timing excludes fitting and is not measured under serving load.

## Artifacts and next decision

Complete equal artifacts: `mmm-spike-slab-v1-pilot-stable.jsonl`,
`mmm-spike-slab-v1-pilot-stable-replay.jsonl`,
`mmm-spike-slab-v1-pilot-hoisted-order.jsonl`.
SHA256: `088d44d139090a80e2b0525834bcbaebc92a0536dd0d94b2ac3c6827db0fc167`.
Audit: `mmm-spike-slab-v1-pilot-stable-audit.json`, SHA256
`dec39cf2e4283d236f383d45712c88edc617b46e61b6d39bf55f1db18ad50169`.
Audit script: `research/spike-slab-v1-pilot-audit.mjs`.
Keep the two earlier aborted artifacts and the non-bit-identical
`mmm-spike-slab-v1-pilot-hoisted.jsonl`; none are the final complete result.

Next: evaluate the frozen model across more trajectories/publication clocks
before any prior tuning or promotion. Separate stationary harm, adaptation
and equal-cadence controls. The strong Markov comparator remains unbeaten in
expected cell averages. Goal4 has a better candidate, not a completed success;
all seven goals remain OPEN. No production, whitepaper or publication changes.
