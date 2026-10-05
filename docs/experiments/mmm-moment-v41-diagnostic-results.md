# Declared-Mean And Breadth V41 Diagnostic Results

2026-10-03 local date / 2026-10-04 UTC. Research component only.
Previous goal research turn V40 made verified progress (completed normal
collection/audits); this continuation also makes progress: new isolated model,
independent reference, frozen protocol, race/leakage/corruption checks,
benchmarks, diagnostic collection and independent readback all completed.
No whole goal completed. DESIGN and CONFIRMATION remain unconsumed.

## What Changed

[Protocol](mmm-moment-v41-protocol.md) separates narrow/rich calibration,
density-grid/exact-declared-mean priors, and strength2/4 in eight shared-family
candidates. Rate atoms21; family count3/28; local rate hazard1/16.
Exact mean uses a finite minimum-KL exponential tilt, not an outcome-fitted
center or a continuous Beta pseudocount. Matching first moments still does
not isolate variance: higher moments can change. Rich means change initial
mixture laws, and global calibration family remains static.

The new learner retains latest rows plus a bounded64-trial ledger; an arrival
reintegrates the whole member history then replaces that member's old evidence.
Memory O(MHQ+ML), Predict/Issue O(HQ), arrival O(HQL); M<=200,Q21,H3/28,L64.
These are research packages, absent from the daemon's437 production deps.
No production/private corpus/whitepaper/branch/commit/push changes by this work.
Other concurrent worktree edits were observed and preserved, not reverted or
included in this study. Frozen V39/V40 sources remain unchanged.

## Verification

All nine model race tests PASS: mean/KL stationarity and feasible perturbation,
equal means/different higher moments, private issued-law lifecycle, failed
normalization atomicity, contract/caps, independent prior/full-history
integration, interleaved censoring, frozen-V40 narrow-density equivalence,
and explicit21^3 three-position path enumeration.

Integration race PASS:13 nonidentity semantic raw corruptions plus missing
allocation rejection, future-prefix invariance for all eight candidates and
three schedules, exact seed separation against V39/V40. Vet PASS.
Independent reference constructs priors with safeguarded Newton (candidate
uses bisection), integrates unnormalized full histories, recomputes global
evidence afresh. It checks every issued forecast, original receipt, as-of
snapshot, drain, metrics and recovery. Auxiliary exact rerun is distinct from
this independent mathematical verification.

Raw diagnostic has28worlds/924arms/15400snapshots/67200 distinct labels across
14regimes/two geometries/three schedules. Model/schedule copies are correlated
and NOT extra observations. Independent Go audit PASS; separate JS readback
PASS hashes, statistics, contrasts, coverage and two corrupted-report controls.
63 sources frozen before outputs, all unchanged. Each cell has ONE world:
NO confidence intervals, adoption gates, population certification or tuning.

## Findings, Not Confirmation

Descriptive paired issued-Brier contrasts across84 correlated cells:

| Contrast | Positive / Negative | Minimum | Maximum | Unweighted Cell Mean |
| --- | --- | --- | --- | --- |
| Moment minus density |22/62|-.00143461|.00045890|-.00026728|
| Rich minus narrow |63/21|-.00203202|.01456569|.00076716|
| Strength2 minus4 |64/20|-.01222776|.00498080|.00095294|
| Moment-only strength2 minus4 |62/22|-.01240079|.00497128|.00086725|

Positive means lower expected loss. These signs/counts are descriptive, not
independent votes or significance claims. The mean-preserving constructor
fixes the V40 interpretation confound but does not automatically improve
predictions. Breadth is promising in some cells, not a broad quality win.

Wide/aligned/immediate stationary example (n1): Full issued Brier .19645285,
narrow density2 .21147976, rich moment4 .20377203. Final top10 usefulness
.87583893/.84416107/.84953020 respectively. Rich moment4 reduces the harm
relative to narrow density2 but does not erase it or beat Full.

Wide/partial/immediate (n1): Full issued Brier .22685660, recovery9 (miss
penalty); Adaptive .20386435/recovery5; narrow density2 .20827106/recovery5;
rich moment2 .20829615/recovery5. Exact mean is not an automatic rescue.
This example is not the normal-cohort adoption test.

## Performance

Apple M4,10logical CPUs, three benchmark repetitions:
- Rich moment150 constructor83.43-130.82ms,1,707,200-1,707,216B,7-8allocs.
- Predict all150members75.56-79.67us,0allocs.
- Fully-observed64 replay93.73-104.09us,0allocs; checkpoint restore excluded.

Diagnostic maximum learner-loop milliseconds per2400labels:

| Arm | Maximum ms |
| --- | --- |
| Full |65.91|
| Adaptive |250.11|
| V40 raw2 shared |5.66|
| Narrow density2 / density4 |10.58 /5.84|
| Narrow moment2 / moment4 |22.95 /35.82|
| Rich density2 / density4 |64.05 /44.25|
| Rich moment2 / moment4 |143.27 /197.10|

These diagnostic costs satisfy400ms/8MiB component limits, not confirmation
or loaded100/250ms service/freshness requirements. New elapsed timers include
initial output arrays; old controls preserve their earlier exclusion. Neither
includes acquisition, scoring/serialization/audit, durability or serving.
Cold construction exceeds100ms in one repetition; reset peak/multi-model
deployment and concurrent serving remain unmeasured. No sub100ms deployment
promise follows from warm microsecond prediction.

Collector52.58s including test overhead; offline independent audit98.45s.
RawSHA256:5152b368abebf2028ea96154f987e411afc3b774588e27de742154dd46124809.
Artifacts: research/moment-v41-diagnostic/{freeze.json,completed.json,
diagnostic.jsonl,diagnostic-audit.json,benchmarks.log,readback.json} and command
logs/manifests. The first readback invocation omitted weekly-usage env and
failed at its final assertion; READBACK_RETRY.md preserves it. Same artifacts
and unchanged checker passed with observed usage13%. No resampling.

## Next Stage And All Seven Goals

Run `node research/moment-v41-run.mjs normal` with exactly the diagnostic
freeze: design2026104103 and confirmation2026104104,448worlds each,16/cell.
Do not tune from diagnostic signs, select a single passing slice, overwrite
artifacts or weaken the stationary/Adaptive/recovery/cost screens. Then run
`EVENTFRAME_WEEKLY_USAGE=<observed> node research/moment-v41-readback.mjs`.

All seven whole goals OPEN/ACTIVE: robustness, safe adaptive recovery, useful
certified AP splitting, sample-efficient cost-bounded challenger, untouched
agent-task answers/retrieval, loaded background freshness/latency, and equal
TOTAL-cost falsification-driven observation. V41 only advances model and
evaluation components. Public multi-domain retrieval/fusion and eager immutable
head publication remain viable independent leads; rate-shape breadth and
switching calibration are separate future ablations. Goal7 full observation
cost and Goals3/5 downstream usefulness are not established by these forecasts.
