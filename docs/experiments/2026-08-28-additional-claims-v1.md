# Additional synthetic claims experiments, 2026-08-28

> This report is the preserved pre-upgrade baseline. The revised detector and
> untouched v4 confirmation are documented in
> `2026-08-28-bayesian-upgrade-v4.md`.

This is a design-stage synthetic experiment, not real-world evidence. It runs
the production EventFrame service and Bayesian updater with frozen parameters
and fixed random seeds. The raw machine-readable report is
`2026-08-28-additional-claims-v1.json`.

## Frozen protocol

- Residual utility: 60 chronological training outcomes followed by 40 untouched
  evaluation outcomes for one repeatedly retrieved but never useful event. The
  comparison disables versus enables the production residual gate.
- Anti-Pigeon granularity: four observationally identical events, two always
  useful and two never useful, over 30 training and 20 evaluation turns. The
  comparison uses separate posteriors, one deliberately invalid broad shared
  certificate, and two oracle group certificates.
- Changepoints: 64 fixed-seed Bernoulli trajectories per scenario. The
  production bounded updater uses hazard 0.05, threshold 0.30, maximum run
  length 32, and a 15-observation matching window. No parameter was retuned
  after seeing these results.

## Results

### Residual utility

The baseline Brier loss was 0.8550 and residual-corrected loss was 0.6001, an
absolute gain of 0.2549 and a 29.81% relative reduction. The learned residual
was accepted in all 40 evaluation cases.

This supports the narrow mechanism claim that a residual cache can correct a
stable, repeated baseline bias after forward training. It does not establish
benefit on heterogeneous workloads, and this fixture does not measure cache
maintenance cost or false residual reuse.

### Anti-Pigeon granularity

| Variant | Brier loss | False merge rate | Posterior keys |
| --- | ---: | ---: | ---: |
| Deliberately broad shared bucket | 0.3638 | 1.00 | 1 |
| Oracle Anti-Pigeon split | 0.2568 | 0.00 | 2 |
| One posterior per event | 0.2591 | 0.00 | 4 |

The oracle split reduced loss by 29.41% relative to the broad bucket while
using two posterior keys instead of four. This supports the mechanism claim
that preserving downstream distinctions can prevent harmful posterior sharing
without requiring one posterior per event.

The experiment does not validate the empirical certificate procedure. The
broad certificate is an intentional negative control whose claimed diameter is
false, while the two-group certificate is supplied by the known fixture labels.
Target-law diameter estimation, audit coverage, and false splits remain open.

### Changepoint adaptation

| Scenario | Expected | Detected | Miss rate | False alarms | Mean delay |
| --- | ---: | ---: | ---: | ---: | ---: |
| Stable 0.8 | 0 | 0 | n/a | 0 | n/a |
| Abrupt, noiseless | 64 | 64 | 0.00% | 0 | 0.00 |
| Abrupt, 0.9 to 0.1 | 64 | 2 | 96.88% | 6 | 0.00 |
| Gradual, 0.9 to 0.1 midpoint | 64 | 0 | 100.00% | 0 | n/a |
| Recurring, noiseless | 128 | 128 | 0.00% | 0 | 0.00 |
| Recurring, 0.9/0.1 | 128 | 8 | 93.75% | 6 | 0.13 |

The noiseless controls show that the reset path and bounded run-length state can
detect an unambiguous reversal. The current frozen policy does not support a
general claim of robust drift adaptation: it misses nearly all noisy shifts and
all gradual midpoint crossings in this fixture. The low delay among detected
cases is not reassuring because it is conditioned on the very small detected
subset.

This is a confirmed specification-to-implementation gap, not a reason to tune
against this report. A revised detector or declared policy family needs a design
dataset, then a separately seeded confirmation run with delay, miss, and false
alarm targets frozen in advance.

## Reproduction

```sh
go run ./cmd/eventframe-claims-experiment \
  -output docs/experiments/2026-08-28-additional-claims-v1.json
```

The command emits schema `eventframe.claims-experiment.v1`. Its experiment
tests verify bookkeeping and expected synthetic separability; they do not turn
the observed metrics into assertions, so an unfavorable result remains visible.
