# Retained challenger v8 protocol

Frozen 2026-09-12 before evaluation. Rescue hypothesis from v7: replacing the
short-count challenger may discard useful recurring/interaction capability.
Retain both; do not loosen the .01 harm guard or alter old evidence.

Same ten generators, full-field independent post-forecast audits, six-field live
budget, frozen incumbent, long256 model and rolling64 tree as v7. Four arms:
fixed-count MMM; replacement-tree MMM (v7 control); adaptive retained MMM;
static retained MMM. First two must match v7 on a shared unit seed.

Retained outer bundle=[incumbent,inner(short,tree),long256,uniform]. Adaptive
inner uses existing ForecastMix with [short,tree,tree,tree]; its prior therefore
aggregates to .7 short/.3 tree and its fixed share is .002. Static inner is
.5 short+.5 tree. Both predict from the SAME observed mask, never hidden bits.
If outer guide selects the inner slot, adaptive selects tree when aggregated
inner tree weight exceeds short weight (otherwise count); static uses count as
the equal-weight tie rule. Outer guide otherwise follows v7. Adaptive inner is
updated on available outcomes using its journaled pre-outcome forecasts,
including delayed labels; not recalculated predictions. Mixtures are online
predictive weights, not claims about posterior truth or causal mechanisms.

All fits, audit decisions and observation budget rules remain unchanged. Both
inner forecasts are .5 until first supported fit. Reference tree replacement
remains scored even if it loses. Forest missing-field marginalization continues
to rely explicitly on independent fair input bits, not a real-text assumption.

Three fits*eight streams*ten cases*two splits=480 streams of512 steps. Fit base
2026093001; design2026093002; confirmation2026093003. Same role encoding as v7.
No tuning between splits. Both retained candidates must meet the same v7
primary gains (>=.005 and paired z=3.6 lower>0 at shifts128/256), stable harm
upper<=.01, no >.01 mean harm on any case/window, and all three fit-group
primary gains positive. The 120 comparisons versus fixed are retained. Also
show paired raw means versus replacement without inferring equivalence.

Verify matched controls, label ordering, journaled inner updates, bounded view
budgets and full replay. Passing synthetic means another stage, not completed
real-task, AP-sharing or production shadow validation.
