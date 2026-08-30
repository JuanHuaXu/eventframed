# Claims experiment v5 protocol

Schema `eventframe.claims-experiment.v5` applies only to runs after the v4
evidence review. It does not rewrite or retroactively grade v1-v4 artifacts.
Its frozen protocol identifier is
`eventframe.claims-experiment.protocol.v5-post-v4`.

## Evidence semantics

- Terminal proportions carry two-sided 95% Wilson score intervals.
- These intervals describe a fixed terminal sample. They are not sequential
  confidence sequences and do not justify repeated optional inspection.
- Changepoint triggers are matched chronologically, one-to-one, to the first
  unmatched trigger from the declared change through the inclusive window end.
- Remaining triggers are unmatched alarms.
- Mean delay is computed over matched detected changes only. Misses are not
  inserted as zero-delay observations.
- Reports include total triggers, unmatched alarms per trajectory, unmatched
  trigger fraction, and the delay sample count.

## Frozen post-v4 criteria

These are synthetic mechanism criteria selected after reviewing v4. They are
not production thresholds and cannot be used as independent tests of v4.

| Group scenario | Minimum expected decision | Maximum wrong-rate Wilson upper |
| --- | ---: | ---: |
| Compatible 0.8 / 0.8 | 80% share | 6% |
| Moderate 0.65 / 0.35 | 80% split | 6% |
| Strong 0.9 / 0.1 | 95% split | 6% |

| Changepoint scenario | Minimum detection | Maximum unmatched alarms/trajectory | Maximum mean detected-change delay |
| --- | ---: | ---: | ---: |
| Stable | n/a | 0.02 | n/a |
| Abrupt, noiseless | 99% | 0 | 0 |
| Abrupt, noisy | 80% | 0.30 | 15 |
| Gradual | 90% | 0.20 | 50 |
| Recurring, noiseless | 99% | 0 | 0 |
| Recurring, noisy | 75% | 0.35 | 15 |

Every report evaluates these criteria and records all violations. The compatible
group criterion intentionally preserves the known v4 failure instead of treating
conservative uncertainty as successful positive sharing.

## Run provenance

Exploratory and design runs record their role and seed base. Confirmation runs
must name a different design seed base; the CLI rejects missing or reused design
seeds. The protocol is compiled into the report before any scenario runs.

```sh
go run ./cmd/eventframe-claims-experiment \
  --run-role design \
  --seed-base 100000001 \
  --output docs/experiments/example-v5-design.json
```

After code, policy, criteria, and analysis are frozen, use a distinct seed:

```sh
go run ./cmd/eventframe-claims-experiment \
  --run-role confirmation \
  --seed-base 200000003 \
  --design-seed-base 100000001 \
  --output docs/experiments/example-v5-confirmation.json
```

The report also executes and records a deterministic integration control. It
must reach `split` for the divergent group, `share` for the all-useful compatible
group, require Anti-Pigeon certification, and leave the snapshot and posterior
key authority unchanged.
