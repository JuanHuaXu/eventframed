# MMM + Anti-Pigeon rescue v2 results

## Verdict

**Failed the predeclared overall criterion.** Post-shift proper-score gains
passed; stationary protection failed badly. The experiment does not validate
safe MMM + Anti-Pigeon integration, nor show that Anti-Pigeon is unnecessary.
It tests one composition of existing primitives, not the full service/certificate
pipeline. Nothing in this experiment is enabled in production.

Authoritative artifacts: `mmm-antipigeon-v2.json.gz` (all forecast/action traces),
`mmm-antipigeon-v2-summary.json` (per-stream metrics, comparisons, hashes), and
`mmm-antipigeon-v2-protocol.md`. The initial seed-overlap run is explicitly
superseded; see that protocol's correction history. No thresholds or acceptance
criteria changed after either run. Both design and confirmation remain reported.

320 streams, 512 live steps each, six paired arms: 163,840 live observations and
983,040 scored forecasts, not 983,040 independent observations. There are 32
streams per split/scenario. Reference observations are additional and charged.

## Confirmation results

Brier is lower-is-better. Stable/null use all 512 steps; single shifts use the
256 post-change steps; recurring uses steps 128--511 (384 steps). All recovery
warm-up predictions count. The 128-step late window is separately in the JSON.

| Scenario / scored window | Frozen MMM | Always audit-learn | CP + audit | AP + audit | AP fallback only |
| --- | ---: | ---: | ---: | ---: | ---: |
| Stable / full | 0.04561 | 0.13736 | 0.15371 | 0.15371 | 0.17937 |
| Member shift / post | 0.44673 | 0.25140 | 0.23027 | 0.23027 | 0.25203 |
| Common shift / post | 0.45691 | 0.26334 | 0.23610 | 0.23610 | 0.25297 |
| Recurring / post | 0.31151 | 0.24785 | 0.25438 | 0.24769 | 0.25633 |
| Null / full | 0.25834 | 0.27028 | 0.25859 | 0.25906 | 0.25729 |

AP + audit's post-shift gain over frozen MMM was 0.21646
[0.19813, 0.23479] for member shift and 0.22081 [0.19937, 0.24225]
for common shift. Its full-stream stationary **harm** was 0.10811
[0.09545, 0.12076], well beyond the allowed 0.01.
Intervals use the protocol's approximate z=3.5 paired normal construction over
32 streams, conditional on the fitting model; not sequential coverage guarantees.

The combined and CP-only forecasts were identical in stable/member/common
scenarios. Recurring AP-vs-CP gain was 0.00668 [-0.00063, 0.01400], unresolved.
Thus a unique Anti-Pigeon advantage is **not established**. Null full-stream harm
was about 0.00072, below the 0.01 mean-harm flag; no useful signal was learned.

Relearning improved over AP fallback by 0.02177 [0.01252, 0.03101] and
0.01687 [0.00596, 0.02778] on the single shifts. Most of the gain over frozen
MMM nevertheless came from stopping overconfident use of the old relationship.
Post-change accuracy of AP + audit was only 59.64% / 57.31%, not restored
old-regime performance. Confident errors were 81/8192 and 97/8192 versus
4085/8192 and 4175/8192 for frozen MMM. These are prediction-level descriptive
counts, not independent trajectory confidence intervals.

## Invalidation audit

Stable false invalidation: 32/32 streams, Wilson 95% [89.28%, 100%], versus
the allowed point rate of 5%. In member/common shift, 28/32 and 24/32 streams
were already invalidated BEFORE the true change. First post-change invalidation
was therefore observed in only 4/32 and 8/32 streams; delay means 12.75 and
10.875 steps apply only to those detected subsets. Early invalidations count as
misses of the first-change timing criterion, not successful detections. Recurring
had 13 early invalidations and 19 post-change detections (mean 19.74 steps).
Null falsely invalidated 9/32 streams, Wilson 95% [15.56%, 45.37%]. The experiment
does not measure repeated detection after every recurring change: revocation of
the original dependency is latched once, while local relearning continues.

Full deterministic replay found that all 32 stable first invalidations crossed
the instantaneous Bayesian changepoint probability threshold (.30). The existing
revision primitive then returned a reset even without independently supported
member divergence. A rare failure under a highly successful reference can cross
that threshold; it is not an anytime false-alarm guarantee. Early reset plus
discarding a 4096-example predictor for a 32--128-example working replacement
explains the demonstrated stationary damage. This is a transfer/composition
failure of this policy, not a reason to tune against the confirmation traces.

## Cost and scope

Steady in-memory Predict + feedback/monitor benchmark on Apple M4, one CPU,
three serial repeats: combined 11.04--11.62 microseconds; instrumented frozen
control 10.95--11.50 microseconds. Both execute the diagnostic monitor, so this
does not measure the full incremental cost over uninstrumented MMM. No audits,
reference reads, I/O, queue contention, or refits occur inside this steady case.
128-audit fitting took 0.336--0.339 ms and allocated 2 MiB per fit. Means are
not p99 latency. Source: `mmm-antipigeon-v2-benchmark.txt`.

Combined informative scenarios averaged 5.28--5.43 foreground coordinate reads
per step, plus 8 diagnostic reference/frozen-live reads and about 2.2--2.3 audit
reads. The six-coordinate cap is FOREGROUND ONLY, not an equal-total-cost win.
Refitting is synchronous in this isolated sequential replay to define the next
step deterministically; production integration would require the existing async
queue, freshness checks, held-out publication gate, and measured queue latency.

## Next hypothesis, not tested here

Separate a surprise-triggered investigation request from authority to revoke a
good attention model. Require independently supported degradation/divergence or
a calibrated sustained-change rule for revocation. Retain the strong incumbent
while a shadow candidate accumulates evidence; compare complete candidate laws
prospectively before replacement. For common shifts, use absolute proper-score
deterioration as well as relative member disagreement. Repeated testing needs
calibrated error control, and a future experiment needs fresh seeds and both
stationary/noise and shifted controls. These are proposals, not rescued results.
