# Consumed interval-surrogate protocol

Freeze before quality outcomes. Use every2688 v120 record, both phases and
schedules, all21 cases,256 forecasts and both all256/terminal64 summaries.
This is consumed-data diagnosis, not untouched confirmation.

Two arms: original four experts; eight including variational64/32 and
segment64/32. Expert prior generic64=.95, remainder uniform. Rates
{.5,.25,.125,.0625}, uniform prior. Uniform prior over one-based geometric
covering intervals ending by256, plus a persistent full-horizon interval.
No interval chosen from teacher changes; prefix tests retain the declared256
horizon. No post-outcome prior, interval or rate sweep.

Each active interval normalizes its rate/expert mass before the meta mixture.
Meta weights use relative cumulative surrogate losses, exactly removing the
common inactive increment. Each issued snapshot stores its probabilities,
expert weights, active intervals, interval distributions and meta weights.
Late observed evidence updates from that snapshot. Current outcome follows
prediction. Missing packets expire after the existing31-frame delay window,
without becoming negative labels. Duplicate/conflicting deliveries are tested.

The explicit normalized mixture and surrogate updates are inspired by
[Neuteboom and van Erven, Sections4-5](https://arxiv.org/pdf/2209.06826).
We do not claim identity with every displayed shortcut in that source or its
theorem for delayed updates. The full-horizon anchor and delayed bookkeeping
are declared research choices. Ordinary Bayes/calibration is not implied.

Component validation: exhaustive256 binary eight-step sequences under one
interval and the covering family,4096 predictions against an independent
absolute-loss/multiplicative-mass reference. An intentionally unnormalized
negative control must differ. Test late-batch commutativity, duplicate and
conflict handling, expiry, invalid inputs and snapshot ownership.

Quality audit:1,376,256 forecast evaluations,672 poisoned-prefix checks.
Compare both candidates to archived Markov12 using paired mean +/-3.5SE over32
trajectories per cell. Non-harm permits upper loss increase<=.01; meaningful
gain requires mean>=.005 and lower>0. Preserve all168 cells per candidate,
not only switching cases. These are finite approximate screens, not anytime
or simultaneous population guarantees. Preserve every failure.

Require byte-exact output replay and paired-comparison replay. The component
retains bounded research state for at most256 frames; do not infer serving
costs from this reference implementation or run wall time. No Go/production,
OpenClaw, whitepaper or publishing changes.
