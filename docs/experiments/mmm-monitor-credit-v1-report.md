# Shared-read credit gate

## Result

All frozen exploratory shadow screens pass: reverse split counts improve in
both cohorts, forward counts are retained, no stable split is added, and every
acquisition prefix stays within its old coordinate allowance. This justifies
closed-loop testing, NOT adoption or completion of a research goal.

Delayed schedules,16 trajectories per cell:

| Cohort / direction | Original splits | Credit splits | Credit splits by511 | Mean detected clock |
| --- | ---: | ---: | ---: | ---: |
| 1 / majority to parity | 16 | 16 | 16 | 379.94 |
| 1 / parity to majority | 5 | 15 | 14 | 416.67 |
| 2 / majority to parity | 16 | 16 | 16 | 383.50 |
| 2 / parity to majority | 1 | 14 | 11 | 456.14 |

Detected-clock means exclude misses and include detections during final flush.
Four reverse detections occur after511 and cannot improve predictions within
the scored window. Cohort2 forward mean clock is slightly later than original
382.06; retaining counts is not the same as uniformly faster detection.

All stable cases retain zero splits under both feedback schedules. These are
consumed exploratory trajectories, not fresh false-positive-rate confirmation.
The outcome law, selection policy and external gate assumptions remain limited
to the declared synthetic setting.

## Implemented Invariant

The isolated test harness preserves each original bounded forecast. An
epoch-scoped cache avoids charging a known coordinate again during full audit
completion. For bounded pair cost b, audit flag a, and prior credit c:

- Audit: charge18 total and set credit to c+b.
- Non-audit with c>=18-b: complete both views, charge18, subtract18-b credit.
- Otherwise: charge b and retain bounded views and credit.

At EVERY prefix, credit equals old allowance minus candidate charge and is
nonnegative. Old allowance is bounded charge plus18 per audit. Selection never
uses outcome, missing status, future audit occurrence, or gate results. Missing
outcomes do not refund acquisition. Learning audits and their arrival schedule
are unchanged. One released event supplies one correctness pair to one gate.

The cache checks view bounds, overlap consistency, and epochs before and after
reads. Stale cached data cannot silently cross epochs. This is test-only
coordinate reuse, not proof of equivalent backend byte/request savings.

Cohort2 reverse monitoring completes full views on72.38% of origins, versus
25% expected random audit nomination. Mean coordinate charge is16.3425 versus
original16.3901 per frame. The goal is to spend the same budget more usefully,
not claim a large cost reduction. Maximum total cost ratio over all trajectories
is1; the independent evaluator also checks every intermediate prefix.

## Certificate Scope

The original evidence multiplier and .15 correctness-gap tolerance are unchanged.
View selection is symmetric in the two sides. Under identical independent
reference/live joint laws, the frozen predictor and common audit decision preserve
a zero-mean signed comparison. This does not transfer every marginal .15
tolerance through adaptive view selection, nor certify full target-law diameter.
The stated bounded conditional-mean null must hold for the new score. No OR of
uncorrected permissions or weakening of thresholds is used.

## Checks And Performance

- Contract: `mmm-monitor-credit-v1-contract.md`.
- Raw: `mmm-monitor-credit-v1.json`, SHA-256
  `4e0c10e9d5fa9632794fe572b6d413785c15e6d61ce3c86bdb276e9374457777`.
- Summary: `mmm-monitor-credit-v1-summary.json`.
- Race unit tests1.300s: exhaustive credit/bounded/audit boundary combinations,
  invalid/overflow inputs,10000-step prefix accounting, all512 reader values,
  duplicate-read charging, and pre/mid-read epoch changes.
- Collection4.53s; all256 race replay Records match exactly in33.54s. Replay
  metadata includes a subsequently added benchmark file, not a whole-file match.
- All875 captured source hashes verified; independent summary regenerated
  byte-identically. Every observed mask/value, charge, prefix and released origin
  checked. Non-full correctness values match the original issued forecasts.
- Package vet passes. Existing production changes were left untouched.

Apple M4, three500ms acquisition/forecast benchmark repetitions per arm:

| Arm | Time per pair | Allocated bytes | Allocations |
| --- | --- | --- | --- |
| Original | 23.737 /22.786 /21.997 us | about22419 | 72 |
| Credit | 22.421 /22.411 /22.421 us | about22516 | 74 |

These close timings do not establish a speedup. The candidate adds two cache
allocations. The fixture uses synthetic readers and excludes gate updates,
model fitting, queueing, retrieval, storage and networking; it is not a serving
latency guarantee. Benchmark source is separate from the collected snapshot.

## Next Required Test

Wire the credit-controlled signal into the actual sequential learner in isolation.
Reproduce these gate clocks and budgets, then compare the corrected forecast
law against the same original arrival-learning control. Freeze gain/nonharm
criteria before collection; retain both immediate and delayed schedules, both
cohorts, stable controls, and absolute Brier versus .25. Earlier splits can favor
poor replacement forecasts, so better detection alone is insufficient.

Fresh confirmation, broader generator/noise/shift coverage, target-linked error
control, agent utility and integrated serving/freshness evidence remain open.
No production, whitepaper, remote repository, or private-data changes were made.
