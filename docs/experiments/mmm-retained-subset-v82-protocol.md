# Retained subset challenger integration v82

Frozen before fresh streams. Research-only response to v81's overconditioning
diagnostic. Reuse NewSubsetConditional with its existing512-subset Bayesian
label model and frozen Bernoulli1/3 inclusion prior, not a supplied feature mask.
Input weights are uniform, explicitly matching this generator, not a claim
about arbitrary correlated or shifted input distributions.

Keep the base, short count model, local/pooled models and v80 split gate. Fit
the subset challenger from exactly the short model's latest64 audit labels at
the same minimum32/cadence16 schedule. No extra labels or foreground budget.
Retain count and subset in an inner ForecastMix: [count,subset,subset,subset],
whose [.7,.1,.1,.1] prior means70% count and30% subset, not three independent
pieces of evidence. The outer four-expert ForecastMix remains unchanged. Both
mixes update only with journaled pre-label predictions. No current-label refit.

If the outer guide selects the short slot, use the subset observer only when
its combined inner weight exceeds the count weight; otherwise retain the count
observer. All experts forecast from that one recorded observation mask. The
subset observer uses its declared joint input/outcome model, maximum6 coordinate
cost, and existing confidence stopping. No generator coordinates enter runtime.

Fresh bases2026118201/02, same existing role-separated Seed function.64 streams
per five scenarios per phase,640 trajectories,512 steps. Keep all five v80
controls, add the retained subset arm, and verify the unchanged control workflow.
Candidate split authorization is identical to v80 mixture-gate + nomination.
Record full/post Brier, accuracy/log loss, realized observation costs, shared
audit/monitor costs, subset fit count and prediction hash. Same labels does
not mean the additional fitting CPU/memory is free; benchmark it separately.

Primary member/common-shift post-Brier gain>=.005 against v80 mixture gate,
paired z3.5 lower>0. Other scenario full/post Brier harm upper<=.01. Both phases
reported, no tuning between phases. Keep stable and null controls and recurring
results even if they fail. This is a finite matched-input pilot; false-split
coverage and the previous broader subset-model failures remain unresolved.
Run control-disabled parity, forecasting-before-feedback, duplicate isolation,
immutable base, source-hashed full replay, focused race/vet and separate fit/
foreground benchmarks. A pilot pass does not complete any roadmap direction.
