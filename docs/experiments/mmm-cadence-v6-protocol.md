# Cadence-matched learner v6 protocol

Frozen 2026-09-12 before evaluation. Items 1 and 4; this does not replace v5.
Reuse v5's ten scenarios, nine-coordinate EventReader, independent .25 audit
selection, missing/delayed-label schedule and preserved forecast mixer. No MMM
partial-view or sharing integration is claimed in this isolation experiment.

Four arms: fixed count mixture; immediate forest mixture; batched forest mixture;
rolling forest mixture. Every bundle is [incumbent,challenger,long256,uniform],
with .7/.1/.1/.1 initial weights and .002 fixed share. Fixed challenger is latest64
count model. Count models first fit at32 admitted audits, then every16.

Immediate forest is v5's five-tree depth4 model, updated at each available audit.
Batched forest is EXACTLY the same algorithm and random seed, but does not see
the first32 audits until that count, then processes batches of16. Thus states
must agree exactly at batch boundaries. Both persist their accumulated evidence.
Rolling forest rebuilds from latest64 available audits at the same boundaries,
with a fixed tree RNG seed per trajectory reused for each rebuild. This is a
separate forgetting hypothesis, not a pure cadence ablation. Before first fit,
batched/rolling forecast .5. All arms have the same labels and input budget.

Three independent fitting seeds per case, eight streams per fit per split:
480 streams of512 steps. Fit base2026092601, design2026092602 and confirmation
2026092603. Fit seed=base*1000000+scenario*1000+fit. Evaluation seed uses v5's
base*1000000+scenario*10000+fit*1000+stream*10+role scheme. Roles0/1/2/3 are
generator/audit/missing/forest; rolling uses role4. No tuning between splits.

Report full/post Brier, accuracy and log loss; delivery counts, tree updates,
tree node counts and separately timed refit/update work. Candidate acceptance
versus fixed: BOTH shift128 and shift256 Brier gain >=.005 with paired z=3.6
lower>0, both stable full harm upper<=.01, and no >.01 mean harm on any window.
Each of three fit-group primary mean gains must be positive. Preserve ALL
120 candidate/fixed comparisons (3arms*10cases*2windows*2splits).
Intervals remain approximate conditional on the three fitting models.

Also report batched-minus-immediate Brier on every case, without declaring
equivalence from a nonsignificant difference. Any forest success still requires
partial-view integration and untouched replication. Rolling failure does not
exhaust adaptive-window alternatives.

Verify replay, source hashes, prediction-before-feedback, invalid/missing label
exclusion, matched fixed/immediate behavior against v5 on a shared unit seed,
and exact forest state agreement at batch boundaries before interpreting results.
