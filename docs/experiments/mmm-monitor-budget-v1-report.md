# Monitoring budget reallocation

## Result: Not Adoptable

The fixed alternating/full-audit policy improves reverse detection substantially,
but fails forward retention and per-trajectory cost. It is not an adopted rescue.
All seven research goals remain open.

Delayed schedules,16 trajectories per cell, all candidate detections by511:

| Cohort / direction | Original splits | Candidate splits | Mean candidate detected clock |
| --- | ---: | ---: | ---: |
| 1 / majority to parity | 16 | 15 | 434.67 |
| 1 / parity to majority | 5 | 15 | 429.27 |
| 2 / majority to parity | 16 | 15 | 430.93 |
| 2 / parity to majority | 1 | 16 | 433.13 |

Means exclude undetected trajectories. Original forward detected clocks average
379.94/382.06: the candidate is slower and misses one in each cohort. Neither
immediate nor delayed stable cells add a split in this consumed sample. Do not
infer a general false-positive guarantee from that observation.

## Cost Is Lower On Average, Not Always

The candidate acquires a full reference/live pair on even-numbered origins or
existing random audits. It removes old bounded-monitor calls entirely and keeps
all learning audits. Each selected pair costs18 coordinates, including pairs
whose outcome is missing. Reader masks/values and all costs are checked.

Cohort2 mean monitor-plus-audit coordinates per frame (foreground excluded):

| Case | Original | Candidate |
| --- | ---: | ---: |
| Stable majority | 13.3855 | 11.3071 |
| Stable parity | 16.4264 | 11.3599 |
| Majority to parity | 12.8231 | 11.2522 |
| Parity to majority | 16.3901 | 11.1995 |

Nevertheless32/256 schedules, representing16/128 paired trajectories, exceed
the old actual budget. Maximum ratio is1.18248, an18.25% overrun. A lower mean
cannot substitute for the frozen per-trajectory requirement. The schedule was
not filtered or adjusted after those failures.

## What This Establishes

The current frozen model and gate can detect most reverse changes when given
more informative observations. This supports the signal-availability diagnosis,
but neither shows better downstream forecasts nor proves that any new policy
is safe, cost-compliant, or suitable for adoption. Periodic nomination is tested
only on iid input streams; periodic real data could alias with the schedule.

The next candidate is described in
`research/monitor-budget-repair-proposal.md`: preserve bounded monitoring, reuse
its reads on learning audits, and spend only accumulated measured savings on
later full-view completion. Its prefix credit invariant avoids speculative
borrowing. That proposal is not yet implemented or validated.

## Evidence And Checks

- Frozen contract: `mmm-monitor-budget-v1-contract.md`.
- Raw: `mmm-monitor-budget-v1.json`.
- SHA-256: `4d0e4bf4e1eeca1105dd087c5f55416b2dddbd4b052c5812c4de3b7c0046af8b`.
- Summary: `mmm-monitor-budget-v1-summary.json`.
- Actual reader enumeration on all512 inputs and nomination-boundary tests
  pass under race testing,1.304s.
- All256 consumed schedules replay under race testing with whole-artifact
  byte equality. Initial diagnostic1.82s, race23.41s (package times excluded).
- All873 captured source hashes and independent cost/mask/origin calculations
  verify; summary regenerates byte-identically; package vet passes.
- Missing-label acquisitions remain charged; only selected arrived pairs
  update the gate; original released-origin order is verified exactly.

No production code, thresholds, whitepaper, remote repositories, or private
data were changed. This was a shadow diagnostic; no new end-to-end performance
benchmark or closed-loop learning-quality test was run.
