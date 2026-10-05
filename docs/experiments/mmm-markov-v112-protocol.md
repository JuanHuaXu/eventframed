# v112: delayed Markov advice, frozen quality protocol

Freeze before generation. Preserve v111's complete scenario, sample and
evaluation design: 12 cases, two phase-disjoint rule pools,32 trajectories per
cell, paired immediate and jitter0..31/missing.2 schedules. There are768 latent
trajectories and1,536 runs; initial16 labels,256 scored frames, changes at128,
as-of long64/short32 fits every32, six-coordinate acquisition and separate full
training audits. No interim quality inspection or parameter tuning.

Fresh seed base2144111200 + phase*1000000 + case*10000 + index*10, with roles
0..4 for rule/input/outcome/delay/missingness. Audit effective seeds against
archived learner and null allocations through v111. No optional enlargement.

## Eight policies

Controls0..6 remain generic, conservative, full carry, arrival Brier, Brier
neutral, log/no-neutral and log/neutral. They must reproduce frozen v109 exactly
on consumed compatibility trajectories. Candidate7 is the delayed Markov filter
with the unchanged four-role prior and alpha=.001. Failed v111 handoff variants
remain recorded failures, not replacements for these seven controls.

The candidate assigns one fixed-share transition per issued event, not per
arrival. Missing labels have unit emission. On delivery, immutable issue-time
likelihoods update the original event and messages propagate through the retained
suffix. Settled/censored prefixes become checkpoints. No invented transitions
during wall-clock flushes; no current-model rescoring, duplicate admission or
new-version gate evidence. See markov_advice.go and its independently checked
path/dense-filter tests for the exact finite calculation. This does not assume
the latent-state model is true or inherit a delayed Brier theorem.

Advance clock, release earlier labels, expire unresolved age>=32 entries, then
publish as-of fits. Acquire all eight forecasts before sampling current outcome;
zero-delay delivery follows. Flush256..287. Brier is primary; log loss and
accuracy are secondary. Model/gate/acquisition budgets remain unchanged.

## All718 gates must pass

Retain the same seven-control gate set as v111:672 non-harm gates across every
phase/case/schedule/full-or-late segment (upper Brier harm<=.01),20 generic gains
(full parity3/4/complement4 and late switches, both phases/schedules),6 conservative
gains (delayed late parity4 and both switches, both phases), and4 gains against
each control2..6 (delayed late switches, both phases). Every gain requires
mean>=.005 and lower>0. Mean +/-3.5SE intervals over32 trajectories are approximate
fixed-sample screens, not confidence sequences or history-wide coverage.
Schedules share a latent trajectory and are not independent replicates.

No favorable-cell selection, replacement of failed Brier by accuracy/log loss,
or alpha tuning after outcomes. Preserve every predecessor failure. Require
exact control compatibility, seed/race checks, full artifact replay, source
hashes and independent scores/costs/clocks/as-of-origin/accounting reconstruction.
The existing isolated component benchmark accounts for O(D*M) refiltering;
additional whole-fixture timing must wait until all other research processes
stop and must not be described as serving latency.

No production access, private-data expansion, commits, pushes or whitepaper
promotion. Even a complete finite pass does not close all seven research
directions or establish real-agent utility.
