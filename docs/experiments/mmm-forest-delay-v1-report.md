# Delayed forest learning: no rescue

**FAIL on incremental improvement.** All24 non-harm screens pass, but0/8
delayed changed-case gain screens pass. Both learners recover poorly under
delayed/missing feedback. Passing a relative harm allowance is not evidence
of adequate absolute accuracy. All seven research goals remain open.

## Evidence

- [Frozen contract](mmm-forest-delay-v1-contract.md)
- [Raw frame-level results and captured sources](mmm-forest-delay-v1.jsonl)
- [Independent reconstruction](mmm-forest-delay-v1-summary.json)
- [Byte-identical summary replay](mmm-forest-delay-v1-summary-replay.json)
- [Evaluator](../../research/forest-delay-summary.mjs)
- [Consumed-data diagnostic](mmm-forest-delay-v1-diagnostic.json)

192 independent trajectories, each under immediate and delayed schedules:
384 schedule-runs, each with uniform-input and forest arms. Schedules and arms
are paired, not additional independent trajectories. Each run scores512 frames,
including labels deliberately hidden from learning. Delay0..31 and missingness.2
are independent of inputs/outcomes. All delayed changed-case intervals include0.

SHA-256: `27f61eae58001bbf2de2546d05ddd21fb31d117b1a2ce14c1920cbf46ef5f423`.
All836 captured source snapshots match. Independent reconstruction verifies
all scores, cost totals for full-input audits, fit-origin availability, fit
cadence, generation-scoped journal accounting and shared latent tapes. Complete
Go replay reproduces all384 runs, including forecasts and fit-origin lists.
The summary regenerates byte for byte. Collection41.02s; full replay41.06s.

Compatibility tests match earlier immediate Full/Post metrics and split times.
Race-enabled contracts passed in3.877s before collection; package vet passed.
The replay/benchmark source was added after the collection snapshot and is not
represented by those836 hashes; it does not change the runner or recorded data.

## Second-cohort results

Post-change mean Brier (lower is better):

| Case | Immediate control | Immediate forest | Delayed control | Delayed forest |
| --- | ---: | ---: | ---: | ---: |
| Copied bit | .203749 | .202992 | .276639 | .276993 |
| Copied XOR2 | .200579 | .195471 | .285225 | .285110 |
| Noisy-copy bit | .202365 | .201322 | .277939 | .277762 |
| Reversing XOR2 | .179342 | .179483 | .271899 | .272174 |
| Stable | .052689 | .052691 | .052696 | .052696 |
| Null | .252321 | .252306 | .253019 | .252979 |

Only the second-cohort immediate copied-XOR2 comparison meets the .005 gain
floor: gain.005108, interval[.002355,.007861]. It does not establish success
across both cohorts or under delay. Delayed copied-XOR2 gain is.000116 with
interval[-.000487,.000718]. Frozen intervals are paired mean +/-3.5SE over16
trajectories: exploratory screens, not confidence sequences or universal bounds.

The constant-.5 forecast has Brier exactly.25 for every binary outcome. All four
delayed changed-case means exceed it. This arithmetic diagnostic was not an
extra frozen adoption gate and cannot be used to claim a tested abstention rescue.

## Why the next step is not another forest tweak

The shared delay machinery, not merely the input estimator, needs scrutiny.
In the second-cohort changed cases, only282-286 of512 forecast losses are applied
to the current generation,122-129 are stale, and101-107 are censored on average.
Stale labels can still train later models; they are not silently converted into
negative evidence. Gate authority remains separate from stale selector advice.

The [read-only diagnostic](../../research/forest-delay-diagnostic.mjs) separates
four64-frame post-change windows using the actual issued expert forecasts.
For copied XOR2, the first delayed window has base/short/long Brier
.44591/.36038/.35252; the served mixture scores.39506 versus neutral.25.
In the final window, short improves to.19044 while the served mixture is.23681.
This suggests distinguishing early protection from later selection, not assuming
that every phase lacks a useful fitted expert. Pure-expert averages do not bound
all convex mixtures, and these consumed outcomes do not authorize oracle routing.

Before a rescue, inspect earlier neutrality/selection experiments and calculate
paired temporal diagnostics. Do not repeat unconditional stale-loss carry,
already rejected in v104/v106, or tune forest thresholds on these consumed data.
Any new policy needs separate frozen evaluation on fresh trajectories.

## Cost and limitations

Each arm spends about4-6 foreground coordinates/frame, but shared monitoring
adds8-12 and full-input audits about4.45-4.74 on these second-cohort cases.
These costs must not disappear behind a six-coordinate foreground claim.
Models fit only arrived audited labels; delayed runs perform fewer fits.

Apple M4, darwin/arm64, three one-iteration complete two-arm512-frame fixtures:

| Schedule | Time | Allocated bytes | Allocations |
| --- | --- | --- | --- |
| Immediate | 124.814-129.074ms | 72.189-72.195MB | 48179-48190 |
| Delayed | 94.059-94.213ms | 56.612-56.613MB | 45655-45662 |

Command: `go test ./internal/observationgate -run '^$' -bench '^BenchmarkForestDelayFixture$' -benchtime=1x -count=3`.
Fewer usable labels and fits can make delayed execution faster while predictions
get worse. These are whole-fixture costs, not request latencies. The driver uses
a bounded512-frame scheduling scan, not a production queue implementation.
The study does not validate persistence, real-agent tasks, or equal-total-cost
superiority over random/uncertainty observation.

No production changes, whitepaper edits, commits or pushes were made.
