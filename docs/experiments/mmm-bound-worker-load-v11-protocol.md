# Bound worker v11: loaded service latency and label freshness

Status: frozen before implementation, 2026-10-01. Research-only persistent
LibraVDB and durable-lineage wrapper; production unchanged. This follows the
[v10 fitted restart screen](mmm-bound-worker-fit-exit-v10-results.md).

Run three paired trials, alternating arm order to reduce simple warm-order
bias. Each fresh store has one visible synthetic source event, one empty durable
source log, and an opt-in motion-bound worker. In both arms:

- An independent four-worker client dispatches 192 full `Service.Recall`
  calls at 1 ms offered cadence, with a bounded 192-job queue and no research
  tap, while another client admits 64 witnessed frontier forecasts and offers
  64 externally assigned labels.
- Measure each Recall from offer to return with a monotonic clock, including
  queue delay; report call-entry-to-return and queue delay separately. Measure
  each label from immediately before its Feedback offer until the background
  worker reports that ID applied; record admission and Feedback durations
  separately. Do not replace offered age with after-return age.
- Require all 64 labels to complete, none to fail, and a complete source-bound
  durable replay after the workload.

The writer arm additionally commits 256 distinct events available one hour
after the largest query as-of time, paced at 1 ms. Require all writes to
complete and at least one write while the learner workload is active. They
must not appear in the as-of frontier. The quiet arm has no writer. Pool the
three trials by arm and report nearest-rank p50/p95/p99/max and per-trial
counts. Use ordinary builds for latency: writer-arm Recall p99 must be below
100 ms and offered-label-to-applied p99 below 250 ms. Report quiet-arm values
and writer/quiet ratios without post-hoc thresholds. Focused `-race` runs test
correctness only; instrumented timing is not judged against these gates.

Record hardware, command, request cardinalities, errors, and store version
motion. This finite isolated service screen is not an OpenClaw request trace,
population tail guarantee, hardware power-loss test, or answer-quality result.
