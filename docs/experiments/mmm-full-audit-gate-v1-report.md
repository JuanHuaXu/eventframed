# Audit signal and available-view rescue

## Outcome

Both frozen shadow comparisons fail their exploratory screens. Full-view
correctness restores the signal, but audit-only evidence is too sparse. Combining
full audits with bounded non-audit observations retains forward detection but
does not increase reverse split counts. Neither is promoted to closed-loop use.

Delayed split counts, including final flush (16 trajectories per cell):

| Cohort / direction | Original | Bounded audit only | Full audit only | Available view |
| --- | ---: | ---: | ---: | ---: |
| 1 / majority to parity | 16 | 0 | 0 | 16 |
| 1 / parity to majority | 5 | 0 | 1 | 5 |
| 2 / majority to parity | 16 | 4 | 4 | 16 |
| 2 / parity to majority | 1 | 0 | 3 | 1 |

All four cohort2 forward audit-only detections happen after the scored horizon
(mean clock529). Only one of three cohort2 reverse full-audit detections occurs
by511. Available-view's single reverse detection is earlier, clock370 versus488,
but one detected trajectory cannot establish a general recovery rescue.

Neither candidate adds stable splits in these consumed trajectories. That is
not an empirical certificate for deployment, nor a guarantee that false alarms
are zero. Forecast output was not changed: this is a gate-only shadow replay.

## Signal Versus Support

In cohort2 reverse cases, frozen full-input expected accuracy changes from
.949121 to .500220, recovering a .448901 contrast. The original bounded-view
gap is .126123. However, only102.69 audit pairs arrive on average across the
entire512-frame trajectory; roughly half precede the change. Full-audit gating
discards three quarters of potential observations and loses substantial power.

Available-view uses exactly one pair per released event: full-input correctness
on a preselected random audit, bounded correctness otherwise. It uses no extra
coordinates and does not OR separate certificates. With a common independent
audit flag and equal reference/live joint laws, equality of expected correctness
is preserved conditional on view. This does not cover arbitrary covariate or
selection shifts, and it is not a universal conditional-law diameter test.

## Important Certificate Qualification

Source inspection of `predictive_bet.go` shows the actual evidence multiplier
is `1 + rate * (sign * z - .15)`. The external gate tolerates an absolute scalar
mean difference up to .15; it is not a test against exactly zero difference.
The earlier attenuation diagnosis should be read with this qualification.

In reverse cases, bounded expected gaps are inside that tolerance in3/16 first
cohort and10/16 second cohort trajectories. The available-view population gap
is `.75 * bounded_gap + .25 * full_gap`;2/16 and4/16 respectively remain inside
the tolerance. On those cases the monitor's selected scalar remains within its
null even though the full conditional law changed by TV .45. This is a mismatch
between a proxy and the desired target, not a justification to lower a threshold.

The other cases can still lack power within512 frames. The present experiments
do not attribute every miss exclusively to tolerance or sample size.

## Reproducibility

- Full-audit contract: `mmm-full-audit-gate-v1-contract.md`.
- Rescue contract: `mmm-available-view-gate-v1-contract.md`.
- Both use all256 consumed schedules from `mmm-arrival-switch-transfer-v1.jsonl`.
- Full-audit raw SHA-256:
  `4fa55842386feaed4bfbefeab975512a54dd9dfb595a0d22ab1bd757b34449f7`.
- Available-view raw SHA-256:
  `7896594729680c606530a9593f5d87ea086ff9397366e71d4255991bdba1d001`.
- Matching `-summary.json` artifacts reconstruct first permissions/nominations,
  split clocks, arrived audit support, and exact original-arm split times.
- All870/871 captured source hashes checked; independent summaries regenerate
  byte-identically. New benchmark files added later are outside these snapshots.
- All256 records reproduce exactly under race testing for each diagnostic;
  all three shared control arms are exactly equal between diagnostics.
- Race artifact metadata contains subsequently added source files, so compare
  Records rather than claiming whole-file byte equality across those snapshots.
- Initial full-audit/available-view collection2.52/3.17s; race24.19/28.05s.
- Package vet passes. No source changes were made to the actual gate or monitor.

## Computational Cost

Apple M4, three500ms repetitions of `BenchmarkFullAuditForecastPair`: two
full-mask lookups on already-acquired inputs take1.466/1.469/1.469ns with0
allocations. This is a tiny cache-hot fixed-table microbenchmark, not retrieval,
gate-update, learning, queue, storage, or serving latency. It does not estimate
the cost of obtaining a missing full frame or of general high-dimensional models.

## Next Research Boundary

Consider reallocating the EXISTING monitoring budget rather than adding full
reads or suppressing most updates. Parent cohort2 mean monitor-plus-audit costs
are about13.39,16.43,12.82,16.39 coordinates/frame for stable majority, stable
parity, forward switch, reverse switch respectively, excluding foreground
acquisition. Full reference/live observation costs18 coordinates, so observing
it on every frame is NOT automatically budget-neutral.

Any proposal must freeze a predictable cost-accounted observation schedule,
preserve learning-audit availability, and compare null control and detection
before closing the learning loop. Current misses also motivate checking whether
the certified scalar tolerance corresponds to the actual target-law distinction.
No prior failures are erased. All seven research goals remain open; production,
whitepaper, remote repositories, and private data are untouched.
