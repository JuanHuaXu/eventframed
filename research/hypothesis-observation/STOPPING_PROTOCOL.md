# Cost-sensitive stopping v10

Freeze before execution. This is a NEW cost/accuracy question, not replacement
criteria for v9's failed independent-stream/misleading-signal rescue.

Retain v9's exact receding two-step acquisition and reliability-averaged scorer.
Compare full16 reports with stop_one and stop_two. A report costs lambda=0.005
multiclass Brier units; eight initial signal checks are charged the same amount.
This is a declared synthetic utility conversion, not a measured service price.

At each prefix compute g1=max one-step Gini gain. stop_one stops if g1<=lambda.
stop_two stops only if g1<=lambda AND g2=max two-step Gini gain<=2*lambda. When
one report remains, both use the one-step condition. Before stopping, both follow
the SAME v9 action rule, not a newly cost-optimized acquisition policy. This is
a bounded-horizon stopping heuristic, not globally optimal stopping. A useful
three-or-more-report chain could be missed.

At stop, preserve the current forecast for every remaining evaluation tick; no
negative evidence, future report, or hidden target enters the stopped forecast.
Evaluate curve Brier over the same16 pre-report ticks and final Brier. Recorded
null pairs after stopping incur zero report cost. Counterfactual full-policy
simulation may continue solely to evaluate the matched control; stopped arms
must be unchanged by that future suffix.

Seven unchanged generator cases, two splits,128 episodes per case, fresh seed
base202609220137 plus existing split/case/episode offsets. Compare the two stop
arms primarily against full two-step. Keep v7 controls in raw artifacts for audit.
Total cost =8+reports; penalized final risk = final Brier + lambda*total cost.

Frozen confirmation screen for stop_two:

- Mean curve AND final Brier harm <=0.01 in ALL seven cases against full two-step.
- In copied20 and matched05, mean report savings >=20% of16 (not total cost).
- In those same two cases, penalized final-risk gain has paired z3.3 lower >0.

Report stop_one as diagnostic without selecting whichever arm passes afterward.
All intervals are descriptive approximations, not simultaneous or sequential
coverage. A pass does not establish the original benefit claims or real-agent
success. Preserve every failed condition and uncertainty caveat.

Verify full-policy parity with v9 on shared test seeds, stopping inequalities and
last-report horizon, frozen forecasts/null outcomes after stopping, suffix
independence, unique paid slots, separate cost accounting, source hashes and
complete replay. Exclusive-create artifacts, no deployed behavior change.
