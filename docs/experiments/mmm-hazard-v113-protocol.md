# v113: uncertain switching rate, frozen quality protocol

Freeze before fresh generation. Preserve v112's768 latent trajectories and
1,536 paired schedule-runs:12 cases, two phase-disjoint rule pools,32 indices,
initial16 labels,256 scored frames, change128, as-of long64/short32 fits every32,
six-coordinate acquisition and separate full training audits. Schedules are
immediate or independent delay0..31/missing.2. No interim tuning.

Fresh base2150111300 + phase*1000000 + case*10000 + index*10, roles0..4 for
rules/input/outcome/delay/missingness. Audit effective seeds modulo2147483647
against all earlier learner/null allocations through v112. No replacement
seeds, optional sample enlargement or favorable-window selection.

## Nine policies

Controls0..7 are generic, conservative, full carry, arrival Brier, Brier neutral,
log/no-neutral, log/neutral and fixed-rate Markov. They must reproduce v112
exactly on consumed compatibility trajectories. Candidate8 is the joint
hazard mixture: rates{0,.001} union{2^-k:k=0..8}, prior.5 on.001 and.05 on each
other rate. No rate or prior was chosen through consumed quality comparisons.

Keep joint rate/role likelihood normalization. Rate is fixed along each latent
path; it is uncertain, not itself a switching state. Marginalize rates into
the same four role weights before the unchanged gate and coherent acquisition.
Use immutable issue-time emissions, unit emission for missing labels, delayed
refiltering, settled-prefix checkpoints and the unchanged expiry contract.
Neither the finite prior nor exact filtering establishes that this model is true.

Advance clock, release earlier labels, expire unresolved age>=32 records, then
publish as-of fits. Acquire all nine forecasts before current outcome generation;
deliver zero-delay labels afterward. Flush256..287 without extra issued-state
transitions. Preserve all predecessor failures and source kernels.

## All818 gates required

Retain the original718 gates against controls0..6 and add96 non-harm plus4
delayed-switch gains against fixed Markov. Totals:768 non-harm and50 gain gates.
Non-harm spans every phase/case/schedule/full-or-late segment against controls0..7,
requiring upper Brier harm<=.01. Gains are20 generic (full parity3/4/complement4
and late switches, both phases/schedules),6 conservative (delayed late parity4
and both switches, both phases), and4 against each control2..7 (delayed late
switches, both phases). Every gain needs mean>=.005 and lower>0.

Intervals are paired mean +/-3.5SE over32 trajectories: approximate fixed-sample
screens, not confidence sequences or history-wide coverage. Paired schedules
are not independent latent replicates. Brier remains primary; accuracy and log
loss cannot override failed gates. No rate sweep or posterior-best-rate selection
after inspection; the scored candidate is the mixture itself.

Require control/seed/race checks, full replay, source hashes and independent
score/cost/clock/as-of-origin/accounting reconstruction. The component's
equivalent arithmetic optimization and tests predate quality generation.
Record performance with all other research processes stopped; report the
O(D*H*M) cost honestly. Microbenchmarks do not certify serving latency.

No production calls, private-data expansion, commits, pushes or whitepaper
promotion. Even a finite pass does not complete all seven research directions.
