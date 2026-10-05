# Mixed-outcome scored-law v22: frozen service component

Date: 2026-10-02. This research-only Goal 6 screen extends the consumed
v18/v19 geometry without editing its archived runner. It asks whether
actual `Service.Recall` forecasts learn from mixed feedback before the
next visible write, and whether the existing epoch guard correctly
withholds that old posterior after a changed frontier. It does **not**
test a relaxed posterior certificate or establish external-law coverage.

## Worlds and timing

- Run four independent synthetic worlds in each of design and
  confirmation. Use distinct frozen label RNG seeds, fresh temporary
  LibraVDB/SQLite instances, and the same declared 256-dimensional
  cosine fixture: two angle-zero rows, 198 eligible rows and 16
  future-only rows. Query once to create a durable pre-feedback journal.
- Before seeing any training labels, select the first 16 distinct
  activated nominees from that journal. Assign them alternating hidden
  Bernoulli probabilities 0.8 and 0.2 in that pre-feedback order.
  Generate one training outcome per selected event from the world seed.
  No model receives the hidden probability. Use `OutcomeFullStream`
  with inclusion probability 1; no invented or future outcome may
  enter a forecast.
- Issue all 16 outcomes against that journal and obtain a later Recall.
  Score each selected event's actual journaled corrected law against its
  hidden Bernoulli distribution using expected Brier, not the observed
  training label. Compare with the same response's base law. Report how
  many selected events reach the packed response; a nominated journal
  forecast is not necessarily visible to the agent. Keep all selected
  events in the denominator; a missing nominee is a failure.
- Append 16 visible backfilled events at the v19 angles. Refresh the
  test-only synthetic selection/omitted certificates using the v19
  publication gate, then Recall at exactly the same as-of time as the
  learned Recall. Require the same 16 selected event IDs to remain
  nominated. Record the scored law, posterior epoch alignment and
  whether any old belief law is applied. Do not retag or bypass the
  service's epoch guard.

## Predeclared reading

Report per-world positive/negative training counts, selected survival,
base/learned/postwrite expected Brier, paired world differences, and
which forecasts actually have `BeliefLaw`. The engineering component
passes only if every world completes journaled feedback without
future-data/identity failures, the learned Recall uses at least one
updated belief, and after the visible writes the old belief is not
served despite refreshed synthetic certificates. Improvement in
expected Brier is reported, **not** required to pass this guard
component: this small per-event synthetic law is not a real-agent
population test. Preserve regressions and any non-learning result.

Measure sequential Recall and outcome durations as diagnostics only;
this is not the v19 loaded p99 screen. Synthetic certificates assume
coverage and confer none. The selected event's hidden probability is
fixed across its training and hypothetical future Bernoulli draws, but
the experiment does not sample a second label; expected Brier uses the
declared probability solely for offline evaluation. Design and
confirmation are frozen before inspecting their outcomes.
