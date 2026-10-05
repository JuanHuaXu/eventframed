# Usage-limit checkpoint: audit localization stream

Research stopped when weekly Codex usage reached81%, exceeding the user's80%
limit. All seven goals remain open. No production or whitepaper changes, commits
or pushes. No live experiment remains.

## Work Completed Before Stop

Read credit_learning_run_test.go. The synthetic fixture uses independent RNG
roles for25% audit nomination and missingness/delay; full audits are charged
and use a frozen baseline. These design assumptions support investigating a
random-full-audit score stream, not claiming independence in production.

New research/audit-localization-stream.mjs constructs an origin-ordered audit
stream using the fixture's31-step delay bound. All256 schedules pass prefix
extension, observed-label availability, and paid-full-read checks. Four small
controls cover the watermark, missing evidence and out-of-order arrivals.
The script emits mmm-audit-localization-stream-v1.json, including source/input
hashes, per-audit scores, checkpoint counts, and gate-time metadata.

At delayed changed-case alarms, available audit labels beyond this finalized
prefix exist in16/16,15/15,16/16 and13/14 split trajectories across the two
directions/cohorts. Availability is not proof those specific labels caused an
alarm, but it defeats assuming the alarm is automatically a stopping time for
the narrower finalized audit filtration. The gate also observes non-audits.
No localization confidence guarantee is transferred.

At clock480, the finalized stream contains about36-39 post256 audit observations
in delayed changed cells. This exceeds clean-fit training counts in some cases
because training publication lags collection. It is not automatically a usable
clean set: the true boundary is evaluator-only, and localization is unimplemented.

## Not Yet Verified or Implemented

- No independent/repeated replay of the new stream artifact yet.
- No full frozen-score reconstruction from base-model parameters yet; this
  script reads existing audited Baseline/Reference scores from the pinned tape.
- No statistical independence test, stopping-aware calibration, onset estimator,
  confidence coverage, candidate fitting, predictive validation or benchmark.
- The31-step delay bound belongs only to this simulator.

## Resume After Budget Allows

First replay the new script to a fresh output and compare bytes; verify source
and input provenance. Then decide whether localization needs its own audit-only
stopping rule or a valid enlarged-filtration/anytime construction. Preserve the
external Anti-Pigeon authorization gate. Do not quietly substitute an audit
sample count for wall-clock stopping time or expose buffered/future labels.
Only after that mapping is explicit should a localization candidate be built.

The previous verified checkpoints remain authoritative for clean-fit benefit,
strict-reset evidence loss, failed proxy rescue, and all earlier failures.
