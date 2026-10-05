# Normalized interval-surrogate result

**FAIL as a broad rescue.** This is consumed-data research on every2688 v120
schedule run, not untouched confirmation, a production implementation or a
completed research direction. No acceptance threshold was changed.

## Artifacts

- [Frozen protocol](mmm-interval-v120-protocol.md)
- [Component checks and hashes](mmm-interval-v120-component-checks.json)
- [All record scores and groups](mmm-segment-v120-interval-diagnostic.json)
- [Paired Markov comparisons](mmm-segment-v120-interval-comparison.json)
- [Research component](../../research/interval-surrogate.mjs)
- [Independent component reference](../../research/interval-surrogate-test.mjs)

## Mathematical and lifecycle checks

[Neuteboom and van Erven, Sections4-5](https://arxiv.org/pdf/2209.06826) motivate
the interval surrogate construction. We implement the explicit mixture of
normalized interval distributions, not a shortcut that drops their distinct
partition functions. Expert weights are obtained by rate-weighted
marginalization. An absolute-loss reference verifies the equivalent update
that subtracts the common inactive meta loss from every interval's meta loss.

The component agrees on4096 forecasts over256 complete binary eight-step
sequences and two interval families. The intentionally unnormalized comparison
differs by up to .00230957 in expert weight, confirming the normalization test
is nondegenerate. This does not assert that every expression in the source has
the same interpretation as our prototype.

Late updates use the ORIGINAL issued probabilities, weights, interval membership
and distributions. Tests cover order-independent acceptance of one issued batch
up to floating-point error, idempotent duplicates, conflict rejection without
mutation, expiry, invalid horizons/intervals and detached snapshots. Every
immediate reference update has nonnegative surrogate mix loss within tolerance.
The delayed adaptation is not claimed to inherit the source regret theorem.

The full diagnostic evaluates1,376,256 forecasts and passes672 as-of poisoned
prefix checks. All prefixes keep the declared256 horizon, avoiding accidental
changes to the interval family when checking future-label independence.

## Quality

| Candidate | Non-harm against Markov | Meaningful-gain cells | Decision |
| --- | ---: | ---: | --- |
| Four experts | 117/168 | 2/168 | Reject |
| Eight experts | 116/168 | 2/168 | Reject |

The four-expert set is generic64/32 and Boolean64/32. Eight experts additionally
include variational64/32 and segment64/32. The candidate has uniform priors over
its declared intervals and four rates, with generic64 expert prior.95 and the
remainder uniform. A persistent full-horizon interval is included alongside
geometric intervals. This is one frozen design, not a rate or interval sweep.

Confirmation delayed terminal64 expected Brier:

| Case | Markov | Interval four | Interval eight |
| --- | ---: | ---: | ---: |
| Additive stationary | .221755 | .221283 | .220856 |
| Additive gradual | .239065 | .238119 | .237820 |
| Parity4 stationary | .049066 | .065490 | .067736 |
| Majority to parity | .060262 | .121308 | .122050 |
| Parity to majority | .102272 | .108395 | .106859 |

For eight experts, majority-to-parity loss increases .061787 with paired
interval [.041915,.081659]; stationary parity4 increases .018671,
[.013079,.024263]. Both are clear regressions under the declared comparisons.
Reverse-switch uncertainty also fails the .01 non-harm ceiling even though
its mean increase is smaller. Mean +/-3.5SE over32 trajectories is an approximate
fixed-sample interval, not simultaneous or anytime coverage.

## Interpretation

Correct normalization and orderly delayed evidence do not make this surrogate
weighting effective on the workload. Plain Squint had also failed; this result
does not justify tuning interval priors until the same consumed cases pass.
The earlier [v96 interval study](mmm-online-interval-v96-results.md) already
exposed a related protection/recovery tradeoff with a different learner and
older input/view regime. It remains failed evidence, not an independent
replication of this result. The new forecast family did not by itself rescue
this branch.

Global time-only expert weighting is now a weak next bet without a more
specific diagnosis. An alternative lead is input-conditional expert assignment:
ask whether different contexts should select different experts rather than
moving one global weight vector. That needs a declared probabilistic model,
an identical evidence budget, an input-independent control and fresh validation
if a consumed-data diagnostic succeeds. It must not use the oracle hull choice,
teacher case identifiers or simulator change times as routing inputs.

The reference implementation is bounded to256 frames and at most eight experts.
It precomputes active sets, retains pending issued snapshots and stores interval
statistics. No serving benchmark, throughput improvement, calibrated-law claim
or production integration follows from this experiment.

## Reproduction

The diagnostic and paired comparison record input/component/source hashes and
both reproduce byte-for-byte in separate executions. Component
tests were rerun before quality execution. The v120 Go sources and original
artifacts remain unchanged; no private data or production system was accessed.
