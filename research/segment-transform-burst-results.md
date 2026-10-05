# Isolated burst results

**FAIL the predeclared performance screen.** Forecast parity and deadline
non-harm passed; no worker configuration reached10% median makespan improvement.
This does not overturn numerical equivalence or prove the transform useless.
It prevents promoting the component speedup into a backlog-rescue claim.

18 bursts /216 fits completed, each on the same272-frame fixture with label
cap64. Apple M4, Go1.27.1, GOMAXPROCS4. Three paired trials per worker count,
alternating arm order. No network, database, real serving or production access.

| Workers | Direct median burst ms | Transform median burst ms | Reduction | Fits by100ms, either arm |
|---|---:|---:|---:|---:|
| 1 | 774.609 | 726.498 | 6.21% | 1/12 |
| 2 | 396.832 | 369.733 | 6.83% | 2/12 |
| 4 | 223.884 | 204.257 | 8.77% | 4/12 |

At500ms, one-worker direct completed7/12 in every trial; transform completed
8/12,7/12,8/12. Two/four workers completed12/12 in both arms in every trial.
No completed fit differed from the reference by more than1e-12. These are
descriptive counts over three bursts, not statistical confidence guarantees.

The immutable `segment-transform-burst-results.json` records all216 individual
queue/compute/completion durations plus source hashes. Run
`node research/verify-segment-transform-burst.mjs` to verify hashes, accounting,
cell coverage and the declared decision. The opt-in test passing means
measurement integrity/parity, not passage of the performance criterion.

## Interpretation

Only the first worker wave met100ms, even with the transform. The result
supports keeping this full fitter off the immediate path and treating queue
admission, stale work and useful outcomes as separate requirements. Increasing
workers trades CPU availability for burst completion; foreground interference
was not measured and no recommendation to increase production workers follows.

The next informative experiment should use the actual background scheduler
and serving diagnostic, retaining outcome/staleness accounting. Repeatedly
tuning the transform to this tiny benchmark would not establish direction6.
The unchanged quadratic interval recurrence also remains a separate compute
lead. All seven research directions remain open; nothing is deployed.
