# Acquired labels train the experts

Freeze before quality evaluation. Stage1 is consumed v120 replay. Fresh seeds
require a separate later protocol; do not relabel existing phase1 as untouched.

Restore the original four-expert Go reference: generic64, Boolean64, generic32,
Boolean32 and Markov mixer alpha.001/prior .95,.05/3,.05/3,.05/3. This is NOT the
six-expert switch mixture of recent fixed-tape experiments. Natural control must
match original v120 arms0,1,2,3,12 and training-origin lists before quality claims.

Fit on each32-frame clock using the latest64/32 revealed labels by origin;
initial16 samples always eligible before capping. No current label during fit.
Natural and paid reveals update issued-forecast mixer once; acquired labels
also enter all four expert fits at the next scheduled publication. Do not
retroactively replace issued forecasts. Keep original delayed-journal expiry.

Policies: natural only, random, entropy, weighted disagreement. Paid pools and
unit costs match prior expedite experiment: unavailable origins[t-8,t) at clocks
8..248 step8, one query per nonempty pool, reveal t+1 after current prediction.
Use current mixer weights and original issued expert forecasts for acquisition;
not a claim of exact temporal information gain. Perfect label service remains
assumed. No Q, unrevealed Y or future inputs in selection/fitting.

Retain all2688 consumed runs and both schedules; no cherry-picked quality subset.
Compare served mixer and each expert separately. Main candidate is disagreement
served mixer; comparators natural, random, entropy served mixers. Nonharm requires
paired lower gain>=-.01 in all168 cells. Delayed terminal changing cases
1,2,4,5,7,8,19,20 need mean gain>=.005 and lower>0 against all three comparators.
Intervals mean +/-3.5SE over32 trajectories, exploratory not simultaneous. Equal
paid query counts per trajectory; immediate schedule0 queries and identical
forecasts. Report fits, acquired evidence, forecast metrics, source/data hashes
and complete offline wall time. No serving or production claim.

Stage1 component gate: twelve reference fixtures span6 cases and2 schedules.
Verify all5 forecasts/origin lists, evidence reveal rules, no-op complete delivery,
nontrivial paid changes in delayed fits/forecasts, future/Q poisoning of both
prediction and query prefixes. Race check before full quality collection. These
fixtures establish integration, not efficacy; all seven whole goals remain open.
