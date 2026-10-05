# Distribution-aware retained-subset pilot

Freeze before collection: ten v83 cases, two cohorts seeds2026092151/52,
16 independently fitted4096-label bases per case/cohort,512 live frames each.
Total320 streams. Retain current selector, priors, clipping, gate authorization,
split behavior,64-label subset training window, audit schedule and fit cadence.

Uniform retained-subset control versus empirical-input weights using the same
past64 fully audited samples and one total uniform pseudocount. No current
query, future labels, oracle weights or extra observations enter fitting.
One candidate uses the control's issued mask exactly; another runs its own
observer with the same six-coordinate cap. Both learn from identical labels.
The fixed-observer arm isolates prediction changes; the coupled arm tests the
actual observation-policy interaction. Report exact costs, fits and split times.

Primary coupled versus control: changed-case mean post Brier gain>=.005 and
mean-minus3.5SE>0; all cases full/post non-harm lower gain bound>=-.01. Do not
change criteria or sweep pseudocounts after results. Fixed-arm contrast is
diagnostic, not a substitute for primary failure. Finite exploratory pilot,
not rare-error, delayed-feedback or real-agent validation. Earlier empirical
forest integration failed; the fixed-mask diagnostic is motivation only.

Verify unchanged-control tapes and metrics against original runner, exact
fixed-arm acquisition, shared fit/split counts, before-label prediction order,
and deterministic artifact replay. Measure component fit/foreground costs
separately if quality merits follow-up. No production or remote changes.
