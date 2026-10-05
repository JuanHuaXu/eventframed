# Onset localization: distinct from detection and prediction

Status: research proposal, no adoption. Latest clean-fit and split-retention
diagnostics motivate this work; they do not validate an onset estimator.

## Primary-Source Check

[Saha and Ramdas, 2026, Distribution-free changepoint localization after
sequential change detection](https://arxiv.org/html/2606.01256v1), sections1-2,
construct post-detection confidence sets using separately exchangeable segments.
Their test inversion accounts for data-dependent stopping; ordinary offline
localization at an alarm does not automatically retain coverage. The
construction needs calibrated null-survival information independent of the
localization p-values. Section2.4 gives a distribution-free detector route;
the current EventFrame detector has not been shown to satisfy that route.
This is not a theorem for arbitrary adaptively selected observation streams.

[Saha and Ramdas, Post-detection inference for sequential changepoint
localization](https://arxiv.org/html/2502.06096v2), sections2-4, provide a related
simulation-based approach with specified pre/post distribution classes. The
coverage distinction between false-alarm-probability control and finite average
run length matters. Treating an arbitrary detector's alarm as a fixed sample
size does not supply the needed guarantee. A synthetic implementation could
test a declared model, but would not certify unknown real-world distributions.

These are candidate tools for localization, not replacement Anti-Pigeon
authorities or forecasting upgrades. No claims are imported without an explicit
mapping of the data stream, filtration, missingness, and stopping rule.

## Avoid Repeating Failed Work

- v97's short-window bank failed reverse recovery despite useful standalone
  experts; another unconditional short-history preference is not a new rescue.
- Guarded segmentation restricted potential recovery too strongly; removing the
  pointwise guard also failed broad recovery/nonharm screens. A different weight
  search on those same heads is not sufficient evidence of progress.
- Fresh local priors failed with both coupled and fixed observers.
- Actual detection-time origin resets discard most clean support in delayed
  reverse cases. Detection clock and estimated onset must not be conflated.

## Boundary Contract

Use zero-based origin T for the first post-change event. Suppose a nonempty
confidence set C satisfies a justified coverage statement for T. Define
L=min(C), U=max(C). Conditional on T belonging to C:

```text
i < L       => pre-change
i >= U      => post-change
L <= i < U  => unresolved
```

This implication is deterministic; it does NOT establish coverage of C.
The post-change fit may include only already-arrived, audited, nonmissing
origins at or after U. Using L instead would retain potentially pre-change
observations. Do not substitute the point estimate for U and call the result
certified. If C is empty or absent, do not produce a certified subset; if U is
too late, support may still be insufficient. Conservative bounds do not
magically recover the oracle's sample efficiency.

Retain unresolved evidence in immutable storage, tagged with origin and epoch.
It may support a separately validated shadow candidate but cannot silently
become certified post-change training. Support below an existing declared
minimum leaves that candidate unpromoted. Do not tune the minimum on these
diagnostic results.

## Feasibility Before Implementation

1. Determine the actual independent evidence stream. Raw outcome frequency is
   uninformative for balanced majority/parity shifts; a fixed pre-change
   predictor's error or joint input/outcome statistic is more relevant.
2. Audit whether the current adaptive monitoring masks preserve the chosen
   theorem's assumptions. Prefer already-paid random full audits for an initial
   bounded feasibility test; do not claim independence from selection by name.
3. Keep audit origin, arrival clock, localization sample index, and gate stopping
   time separate. Map a sample-index bound back to an origin interval without
   assigning missing labels or peeking at later arrivals.
4. Calibrate stopping-aware uncertainty before claiming clean membership. A
   Bayesian run-length point estimate may nominate a shadow model, not certify
   retained evidence. Keep external split authorization unchanged.
5. Freeze one model/score/calibration contract and compare actual-origin,
   strict-detection, and localized fits on both shift directions plus stable
   controls. Preserve data used for development as consumed. Measure localization
   coverage, false alarms, retained clean/stale counts, fitting cost, and later
   forecast quality separately. Only then test integration and fresh cohorts.

All work belongs in the asynchronous research path. A bounded snapshot may
eventually be published, but enumeration/simulation/localization costs must not
be hidden in the serving hot path. No prototype performance or efficacy claim
is made by this proposal.
