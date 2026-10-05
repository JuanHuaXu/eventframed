# Predictable model-based bets v76

Frozen before fresh outcomes. Research-only follow-up to v74/v75. Keep v74's
query proposal, mean correction, eta, eight starts and threshold100. Replace
the fixed .25 bet with a predictable, sign-specific rate chosen before querying.

From the last32 ternary observations per channel form probabilities for D=-1,0,1
with prior counts(.5,1,.5). These reproduce v74's first/second moment estimates.
For each sign build the twelve-outcome predictive law q_i*P_i(D), with
X=s*(w_i D+eta*c_i)-.15. Compute min X over ALL channels and D endpoints,
even when predicted probability is tiny. Rate cap=min(.8,.92/(-min X)) if
min X<0, else.8. Maximize predicted expected log(1+lambda X) on[0,cap],
using derivative endpoints and24 bisection steps when needed. Zero bets are
allowed. Apply the same pre-query rate to each currently active start.

Actual propensities preserve the target mean even if the predictive law is
wrong. Rates are predictable and factors>=.08, so the same conditional-null
supermartingale argument holds. Forecast misspecification can still harm power.
No true generator probabilities or current outcomes enter rate selection.
Source framework: [Waudby-Smith & Ramdas](https://arxiv.org/abs/2010.09686).
This specific ternary optimizer is our adaptation, not a claimed reproduction.

Four paired arms: uniform, old IPW, fixed-rate augmented(v74), predictive-rate
augmented. Last two share query policy and observations. Retain all ten v72
scenarios,512 steps,512 streams per cell, two phases. Fresh bases2026117601/02,
seed=base*1e6+scenario*1000+stream. No tuning between phases.

Primary candidate predictive only. Preserve v74's10% restricted-mean gain
requirement for sparse128/256 with paired z3.3 lower>0 and no extra premature
alarms; each null Wilson95 upper<=.02; other alternatives at most10 steps mean
delay harm. Six alternatives require excess misses<=.01 and v73 paired upper
<=.02, alpha=.05/12 for six alternatives/two phases. Fixed samples, not repeated
testing until success. Keep weak/late failures, even if relative guards pass.

Record source/evaluator hashes, first alarms, query counts, rate/model/outcome
tape hashes and zero-rate counts. Full replay, exact old-control/query parity,
factor bounds including wrong-model endpoints, predictive optimizer checks,
immutable pre-query snapshots and invalid-input isolation are required.
Measure policy/gate cost separately. No production/agent/diameter guarantee or
completed research direction follows from a finite gate-level pass alone.
