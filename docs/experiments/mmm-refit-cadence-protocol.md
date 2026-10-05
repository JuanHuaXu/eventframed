# Four-expert refit cadence diagnostic

Frozen before cadence results. This is a diagnostic on the consumed v120
trajectories, not a new confirmation or equal-compute rescue. All seven full
directions remain open. Previous cadence v6 used forest learners and does not
answer this current four-expert question.

## Hypotheses

The query-burst experiment improves label training lead but not reliable quality.
Credible remaining explanations include infrequent fit publication, inappropriate
training examples/windows, and model or selector limitations. This diagnostic
changes publication cadence only. A missing improvement weakens the hypothesis
that publication lag is a sufficient rescue; it does not establish which of
the other explanations is correct.

## Intervention

Use all2688 original v120 records: both phases,21 cases,32 trajectories, and
both matched feedback schedules. No paid acquisition or new selection policy.
Keep the same initial16 samples, four expert algorithms, latest64/32 eligible
training samples, issued-forecast mixer, journal expiry, and natural deliveries.
Fit at clocks0,8,...248 instead of0,32,...224. Every current-frame forecast
precedes its outcome, including zero-delay feedback. No future label or Q enters
training, scheduling or selection. Natural evidence received before a fit can
enter that fit; a current zero-delay outcome cannot.

The existing32-frame forecasts are the paired control, verified against raw
v120 before using them. Fit count rises from8 to32 bundles per trajectory,
each bundle containing four experts. This is four times as many fit invocations,
not necessarily four times CPU or memory. It also permits publication later
in the terminal window; that is part of the cadence intervention. Do not claim
the extra computation is free or a serving-path latency result.

## Analysis

Recompute full256 and terminal64 expected Brier from recorded predictions and
external Q. Report all five forecast components, not only the best case/arm.
Paired gain is32-frame loss minus8-frame loss. Use mean +/-3.5SE across32
trajectories, explicitly exploratory and not simultaneous/anytime coverage.
For diagnostic screening report positive lower bounds, upper bounds below zero,
and intervals crossing zero across all cells. For the served mixer also report
the earlier0.01 non-harm tolerance and changing cases1,2,4,5,7,8,19,20 delayed
terminal mean gain>=0.005 with positive lower bound. Even a pass does not meet
the resource-capped research goals without a subsequent bounded implementation.

Compare cadence effects under complete/immediate versus delayed/missing feedback
on the existing paired trajectories. Do not interpret an interaction from point
estimates alone as identified without its paired uncertainty.

## Verification

Before full collection: preserve the32-frame control, check all fit origins
against as-of delivery, check forecast-before-current-label behavior, poison
unarrived labels/Q at selected clocks, and verify input immutability under race.
The independent scorer must check every origin list, baseline forecast, finite
probability and served mixture. Save raw predictions and source hashes. Keep
failed records and account for incomplete collection. Replay scoring separately.
Report whole experiment wall/CPU time and fit counts. No production changes.
