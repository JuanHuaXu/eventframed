# Independent replication status

Superseded by [completed replication report](mmm-spike-independent-v1-report.md).
The preparation and commands below are retained as the historical audit trail;
fresh fitting, scoring, and all eight window audits are now complete.

## Completed

The frozen serial source collection completed successfully in 545.32 seconds:
672 trajectories, 172,032 issued prediction steps, all 15 v120 baseline arms.
Source: `mmm-spike-independent-v1-source.jsonl`.
SHA256: `a7aba805e2d1e089f95a9f4ed298dff4e4c5985c4d325ea61fec18acaf54846a`.
This is offline collection time, not serving latency.

All source rows passed an audit of recorded code hashes, generator metadata,
initial observations, outcomes, probabilities, delays, missingness, forecast
ranges, ordering, and as-of fit origins. Normal test: 1.133 seconds including
test overhead; race test: 11.333 seconds. This audit did not replay baseline
fits or independently recompute their forecast values.

A new no-cache challenger collector is implemented and compiles, with an
identity-order unit test. It records raw fits for all eight windows, retains
caps, checks unchanged boundary fits, and writes a source/code-hash manifest.
It has not yet been run end-to-end.

The new scorer implements seven paired arms: share/local, reset-both,
Markov, static/global, static/local, share/global, and raw challenger.
Self-tests cover unavailable evidence and local realized budgets. On consumed
data, all 172,032 primary guarded forecasts and all 672 whole-stream expected
and realized scores for the first three arms exactly reproduce the prior
implementation. This is regression evidence, not fresh quality evidence.

## Next

From the repository root, run serially:

```sh
EVENTFRAME_SPIKE_INDEPENDENT_FIT_SOURCE=../../docs/experiments/mmm-spike-independent-v1-source.jsonl EVENTFRAME_SPIKE_INDEPENDENT_FIT_PREFIX=../../docs/experiments/mmm-spike-independent-v1-fits go test ./internal/observationlearners -run '^TestSpikeIndependentFits$' -count=1 -timeout=90m -v
```

Then audit every raw window with `research/spike-cadence-v1-audit.mjs`, using
`-` for no frozen comparator, `eight` mode, and the corresponding clock.
The existing auditor's consumed-data wording is historical; label this cohort
explicitly in the final report. Its moment checks do not independently audit
the full posterior-predictive integral.

Score the compact artifact using:

```sh
node research/spike-independent-score.mjs docs/experiments/mmm-spike-independent-v1-source.jsonl docs/experiments/mmm-spike-independent-v1-fits.jsonl docs/experiments/mmm-spike-independent-v1-results.json
```

Paired uncertainty, scenario/window harms, fit caps, and measured compute must
still be reported. No independent challenger quality conclusion is available.
All seven research goals remain open. Production and the whitepaper unchanged.
