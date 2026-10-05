# Independent-data confirmation, v1

Frozen before collection. This is synthetic replication, not evidence of real
agent-task usefulness or full completion of any research goal.

Use transfer seed base 3000011000 and Boolean seed base 3004011000,
two phases, 21 cases, eight indices, and both existing feedback schedules:
672 trajectories with 256 issued predictions each. Reuse the v120 algorithms,
not their fitted states, forecasts, or outcome tapes. Run serially.
Transfer change times come from the new teacher draws. Boolean shift cases
19/20 use role seed 5 to choose an odd change time uniformly from 65 through
191; roles 0..4 retain their existing meanings. Stationary cases are unchanged
in definition, but use new draws. Record change times and their seed.

Freeze the challenger prior, inference stopping rules, and quadrature settings
from the previous spike experiment. Freeze two-expert delayed Fixed Share with
uniform prior, eta=1, alpha_t=1/t, continuous expert weights, and independent
32-step loss ledgers with allowance .01 per issued prediction. Retain Markov,
reset-both-at-32, static/global, static/local, and share/global controls.
No tuning based on this cohort and no retries of capped fits.

Primary paired comparisons: expected Brier over all 256 predictions and the
terminal 64, versus Markov and reset-both. Report every 32-step window,
individual harms above .01, stationary/shift scenario means, realized Brier,
fit counts, caps, and elapsed compute. Report eight-index cluster bootstrap
intervals as pointwise, exploratory uncertainty, not simultaneous coverage.
Do not equate improved pooled averages with a uniform non-harm guarantee.
Any remaining individual harm above .01 contradicts that uniform claim.

Generator tests must verify reproducibility, recorded shift/outcome agreement,
and seed uniqueness modulo 2147483647 against known prior quality and null
cohorts. As-of tests must perturb oracle truth and unavailable outcomes without
changing earlier forecasts. Preserve partial artifacts on failure; collectors
must refuse existing output paths. Algorithm version and cohort identity are
separate metadata fields.
