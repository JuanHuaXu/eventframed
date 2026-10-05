# Fixed-rate results: workable low-rate point, overload failure

**10 requests/s passes all three finite cells;40 requests/s fails all three.**
This characterizes two operating points, not an exact capacity boundary. It
does not replace earlier saturated-workload failures or complete direction6.

| Rate | Trial | Completed | Stale | Drops | Conservative on-time returns | Scheduled read p99 on/off ms |
|---|---:|---:|---:|---:|---:|---:|
|10/s|0|64/64|0|0|64|14.094/13.983|
|10/s|1|64/64|0|0|64|11.903/12.953|
|10/s|2|64/64|0|0|64|12.809/15.957|
|40/s|0|9/64|55|0|8|13.092/22.580|
|40/s|1|11/64|53|0|10|12.023/14.724|
|40/s|2|7/64|57|0|5|12.988/22.358|

No request errors; all16 writes overlap reader activity in each arm. All64
distinct fresh fits are offered, without caching or prepared prefixes. Requests
retain absolute scheduled times even when dispatch is late. The four readers
issue strided IDs, not completion-paced independent loops. Scheduled read
latency includes time through timing-record insertion, not just the Recall call.

Conservative on-time returns subtract all successful-but-late processor returns
from completed count. Final scheduler validation time is not individually
traced, so these are not certificates of scheduled-offer-to-final-publication
latency. The scheduler independently enforces100ms since nomination. All returned
forecasts match their own reference. Real task usefulness remains untested.

## Meaning

The exact table-based fitter can finish this low-rate workload without harming
the measured foreground tail; previously saturated failures are not evidence
that it cannot be scheduled at any rate. At40/s, foreground calls still finish
but background learning expires. An empty/drop-free queue alone is therefore
not proof the learning workload is healthy. Neither10/s nor40/s is a general
hardware guarantee or a learned production rate limit.

Writes occur near the start rather than continuously throughout each run.
Three64-request trials are not population tail bounds. Actual semantic outcomes,
continuous writes, longer runs and final-publication timing remain gaps.

Evidence: `segment-paced-fit-results.json` with12 arms,768 request samples,
192 writes, traces and hashes. Verify with
`node research/verify-segment-paced-fit.mjs`. The verifier explicitly handles
Go's null empty trace slice only for disabled arms and still requires zero
started/checked processors there. Artifact unchanged after this verifier fix.

The next step should connect a declared acceptable arrival/load envelope to
real learning outcomes, not continually relax the overload completion screen.
Production remains unchanged; all seven research directions remain open.
