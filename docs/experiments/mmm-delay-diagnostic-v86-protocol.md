# Delay diagnostic v86

Post-hoc diagnosis of v85, not fresh confirmation or a rescue. Preserve all
2,560 original records, including complete prediction/feedback tape hashes.
Instrument a mechanically copied driver with read-only hooks only. No new
observations, labels, fits, random draws, publication or selector transitions.

For each 64-step window record the candidate's emitted mixture Brier, outer
expert Briers/weights, available inner count/subset Briers/weights, guide use,
model age since publication and age of the newest/oldest short-training audit.
Record the fraction of its short-training audits in the current simulator
regime; this is an offline oracle diagnostic, never a runtime feature.
Count applied/stale feedback and its age by arrival window. Compare all four
feedback schedules on identical latent streams, including stable/null controls.

Competing hypotheses: old-regime training influence, refit delay, or selector
lag/stale feedback. These measurements may identify bottlenecks but cannot
alone establish a causal rescue. Future interventions require fresh frozen
protocols and data; v85 failure remains recorded regardless of diagnosis.

Require exact v85 parity, source-manifest verification, deterministic diagnostic
replay, finite bounded counters and a focused race/unit test. Instrumentation
does not count as a serving implementation or performance result.
