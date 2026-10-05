# Closed-loop uncertain-renewal pilot

Freeze before outcomes. Use the exact v12 prior/model, now selecting its own
actions with the existing receding two-credit planner. No adjusted freshness
thresholds, priors or costs. Maintain ordinary cost1, renewal cost2, budget16,
eight types, four ordinary slots/type and eight renewal slots/type.

Six arms: original regular-only planner, original certain-fresh mixed planner,
uncertain regular-only planner, uncertain mixed planner, uncertain random and
uncertain entropy/credit. The last two use the same uncertain model and action
set as the candidate. Original planners retain their actual posteriors and
choices, not forecasts copied from the candidate. Every arm pays16 credits.

Same five environments as v11: independent20, copied20, mixed20, copied05 and
false_renewal20. Conditional measurement tapes are shared across arms within an
episode; seeds are new:2027111801*1000000 + split*100000 + case*1000 + index,
two splits,64 episodes/case/split. This directory had no matching seed base.
All640 episodes retained. Freshness still means new conditional draws or copies
of the ordinary root, not arbitrary adversarial correlation or real validation.

Primary candidate is uncertain mixed. In EACH split require:
- Genuine copied20/mixed20 Brier gain vs original regular >=.02 and lower>0.
- Gain vs uncertain random and uncertain entropy on those cases with lower>0.
- Nonharm lower gain>=-.01 vs original regular AND original mixed on all four
  genuine-renewal environments.
- False-renewal Brier gain vs original mixed >=.02 and lower>0, plus nonharm
  lower gain>=-.01 vs original regular.
All32 gates must pass (16 per split). Confidence intervals are paired mean
+/-3.3 SE over64 trajectories, descriptive not simultaneous. This retains the
conditional efficacy goal and adds robust false-freshness protection. Do not
replace failure with confidently-wrong improvement alone.

Record final/credit-area Brier, accuracy, confidently-wrong fraction, full
forecasts/actions/outcomes, credits, renewals and final source-mode probabilities.
Verify copy-on-write branch ownership, original-control parity, deterministic
full replay, as-of action reconstruction, costs and normalized posteriors. The
finite model should agree with the independent v12 batch joint likelihood.
Report computational cost separately from synthetic acquisition credits; no
serving-latency inference from this run. Passing a finite pilot is not completion
of the real-data, unknown-hypothesis or production-shadow research directions.
