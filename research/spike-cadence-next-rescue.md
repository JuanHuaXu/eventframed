# Next rescue: protect the incumbent without hiding feedback delay

Arrival-triggered refitting improves every pooled pilot cell over frozen32,
but five of84 cases worsen by>.01 Brier. It is not a safe unconditional
replacement. Distinguish early update quality from permissions to publish
those updates. Do not tune a new prior or clip this negative evidence away.

Primary source read: Joulani, Gyorgy and Szepesvari (2013),
[Online Learning under Delayed Feedback](https://proceedings.mlr.press/v28/joulani13.pdf),
Algorithm1 and Theorem1. BOLD uses separate learner instances while earlier
feedback is pending; its regret transfer depends on outstanding feedback and
assumptions about delays. This is not proof that a single delayed score-weight
update inherits the ordinary no-delay bound. Permanent missing feedback also
prevents blindly applying finite-expected-delay results or an unbounded-copy
implementation. No BOLD implementation or guarantee is claimed here.

Next bounded experiment: keep the strong incumbent (Markov) and the new
arrival-updated forecast as explicit experts. Record both probabilities at
issue time; update a two-expert aggregator only from subsequently received
outcomes for those exact forecasts, once per origin. No re-scoring old labels
with a newly trained candidate. Missing outcomes must not become zero losses
or false confirmations. Freeze initialization, loss, weighting and expiration
before running, keep all fit cost, and label delayed/missing-feedback handling
honestly. Check non-harm by scenario, not just aggregate gain.

This is a proposed next experiment, not a guarantee or an implemented rescue.
The cadence study establishes that fresher evidence can help this model;
it does not establish that this particular aggregation scheme will protect
stationary behavior or satisfy the full seven-goal objective.
