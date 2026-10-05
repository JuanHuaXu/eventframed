# Restarted audit-evidence screen

2026-10-01. This is a consumed-data diagnostic for research goal 3, not a
new confirmation sample or a change to served forecasts. The frozen protocol
is [mmm-audit-restart-v1-contract.md](mmm-audit-restart-v1-contract.md).

## Result

The fixed eight-start, 0.15-margin audit wealth nominated 72/128 changed
schedules by release clock 511, versus 16/128 for the earlier single-start
0.15-margin process. It nominated 82/128 by frame 543. No nomination occurred
before the true change at frame 256 or in the 128 stable schedules. These
schedule counts include paired immediate/delayed versions of the same
trajectories and are not 256 independent replications. The existing external
gate split 123/128 changed schedules by flush, and its evidence path is not
the finalized audit stream used here.

| Changed cell | Immediate by 511 | Delayed by 511 |
| --- | ---: | ---: |
| Cohort 1, majority to parity | 11/16 | 9/16 |
| Cohort 1, parity to majority | 11/16 | 4/16 |
| Cohort 2, majority to parity | 12/16 | 5/16 |
| Cohort 2, parity to majority | 12/16 | 8/16 |

The delayed reverse-shift cell is particularly weak. Across detected changed
cells, mean nomination clocks range from about 415 to 501. Thus the restart
bank recovers evidence drained by a long pre-change prefix, but it neither
matches the existing split frequency nor demonstrates earlier useful recovery.
It is **not adopted** as split authority or as a predictive rescue.

## Verification and cost

The output is [mmm-audit-restart-v1.json](mmm-audit-restart-v1.json), generated
from [mmm-audit-localization-stream-v1.json](mmm-audit-localization-stream-v1.json)
by `node research/audit-restart-v1.mjs` with the frozen tolerance result as the
second input. Re-running against the pinned replay stream produced a byte-exact
output. A separate suffix-product calculation in
`research/audit-restart-v1-check.mjs` verified all 29,305 released audit scores,
every first crossing, stable/pre-change counts, and release ordering. The
reconstructed full-audit forecasts also passed 32,562 immediate and 26,048
delayed score checks against the archived generator, with zero discrepancy.

On this machine, the eight-product arithmetic alone took a median 14.24 ns per
released score across five timed trials, range 14.06-16.66 ns. The loop omits
retrieval, full-audit acquisition, persistence, fitting, synchronization, and
serving; it is not a request-latency measurement. The method performs O(8)
arithmetic per finalized audit, not per corpus event.

## Boundary

The fixed average of eight nonnegative products is an e-process only under
the declared conditional scalar null E[W | finalized audit history] <= 0.15,
where W is reference correctness minus live correctness. Nominal Ville alpha
0.05 then applies to that scalar null. It does not establish Anti-Pigeon's
target-law diameter or simultaneous bucket coverage. Random audits, bounded
delay, frozen score reconstruction, and origin-finalized release are properties
of this simulator. The actual external split gate also sees non-audits and
buffered audit labels, so stopping-time coverage cannot be transferred to it.
No onset confidence set, point forecast benefit, stationary population bound,
or real-agent result follows from this screen. Next work should test a
separately frozen investigation-to-prediction policy on genuinely new data,
or leave this as a negative split-authority result.

Reproduction from the repository root:

```sh
node research/audit-restart-v1-check.mjs \
  docs/experiments/mmm-audit-localization-stream-v1.json \
  docs/experiments/mmm-audit-restart-v1.json
node research/audit-restart-v1.mjs \
  docs/experiments/mmm-audit-localization-stream-v1-replay.json \
  docs/experiments/mmm-audit-tolerance-v1.json \
  docs/experiments/mmm-audit-restart-v1-replay.json
cmp docs/experiments/mmm-audit-restart-v1.json \
  docs/experiments/mmm-audit-restart-v1-replay.json
```
