# Bounded one-change snapshot screen

Frozen before candidate results. Research-only prototype, not a serving change.
Motivation: the count-matched regime diagnostic supports contamination while
the stationary placebo rejects unconditional forgetting. Existing full
segmentation already handles unknown boundaries but failed broad screens.
This tests a simpler prior, not a claim that changepoint inference was missing.

## Model

At any requested clock, take the latest64 naturally arrived labels with origin
strictly before the clock, ordered by origin. Initial16 labels are available.
The model function does not receive case, phase, change time, unavailable labels,
or teacher probabilities. Refit from scratch; never multiply likelihoods from
overlapping fits. Let S have n observations and M(S) be the existing.95 generic,
.05 Boolean family marginal with mask inclusion1/3 and Jeffreys Beta cells.

H0 has one family/parameter draw for all S. H1 draws independent family/parameter
values before and after one boundary. Conditional on the observation inputs,
candidate boundary indices j are8..n-8 inclusive, with uniform prior. Set
prior P(H0)=.9 and P(H1)=.1. If n<16, only H0 exists. This prior is over eligible
label order, NOT a per-calendar-frame hazard or an anytime testing guarantee.

Unnormalized weights are w0=.9*M(S), wj=.1*M(S[:j])*M(S[j:])/(n-15).
Normalize over H0 and every j. Forecast is their posterior mixture of the full
and corresponding tail predictive laws. Every observation is used once in each
hypothesis. No hard reset, max-boundary selection, empty tail, hazard smoothing,
threshold sweep, or oracle boundary. A split label remains an explanatory
posterior hypothesis, not SCM causal evidence.

## First screen

Use the same consumed384 schedule trajectories (cases0,19,20; both phases and
schedules;32 indices) and snapshots160,192,224. Fit without future inputs and
then evaluate frozen forecasts on the next32 actual inputs, matching original
publication windows. Retain all1152 records, origins, weights, interval marginal
logs, and predictions. Four workers, deterministic output order, exclusive files.
The known change is used only to describe evaluation cells, never the fitter.

Controls: existing no-change64 and full segment64 at cadence32, verified against
their original v120/family artifacts. All36 cells must satisfy upper paired
Brier harm<=.01 against BOTH controls (72 non-harm checks). The12 contaminated
cells (two transitions, two phases, immediate160 and delayed160/192) must each
gain mean>=.005 and lower>0 against BOTH controls (24 gain checks). Report all
other cells and never substitute a count of partial passes for whole success.
Intervals are mean +/-3.5SE over32 trajectories, exploratory not simultaneous.
Even a pass only warrants a full21-case stream test with unchanged broad gates.

## Contracts and cost

Reuse the audited batched interval likelihood builder. Independently reconstruct
all candidate weights and forecasts from direct Beta integrals in JavaScript.
Test tiny evidence enumeration, prior normalization, absent-cut support, boundary
inclusion, input immutability, unavailable/current/future-label poisoning, Q
poisoning, and serial/detached concurrent equality under race. Snapshot adapters
must never read unknown Y into fitting even if the raw test tape contains it.
Record collection wall/CPU and component fit timing separately from serving
latency. Interval building remains bounded quadratic in retained-label count;
this experiment makes no speedup claim.

Falsifier: failure against the existing full segmentation model, stationary harm,
or no meaningful contaminated-cell improvement rejects this as a rescue at the
screened scope. Do not tune the.1 prior,8-label support, or64 cap to the results.
No production, OpenClaw, dependency, paper, commit or push changes.
