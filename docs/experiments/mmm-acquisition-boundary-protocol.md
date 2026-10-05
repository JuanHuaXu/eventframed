# Align boundary acquisition with publication

Frozen before implementation/quality outcomes. Consumed original v120 data.
Hypothesis:7 queries per delayed trajectory currently miss the immediately
preceding fit by one clock; moving those queries one clock earlier increases
their chance to enter training before natural delivery. No automatic efficacy claim.

Use original4-expert/Markov acquisition-training reference unchanged except timing.
For m=1..7, move query32m to32m-1 AFTER issuing that frame's prediction. Candidate
origin block remains [32m-8,32m), so its final origin is now the current frame.
Input and all expert forecasts for that frame are visible; Y is visible only if
ordinary immediate delivery has occurred. Exclude already known labels, reveal
paid answer at32m before fitting. This is not prediction using the current label.
All other query clocks stay8,16,24,40,...248. Blocks remain disjoint; no duplicate
evidence. Check actual count equality with all original paid controls per run.

Retain natural, random, entropy and disagreement policies. Compare aligned
disagreement served law to aligned natural/random/entropy and original timing
disagreement. Retain individual expert metrics and effective training lead.
All168 cells need lower gain>=-.01; delayed terminal changes1,2,4,5,7,8,19,20
need mean gain>=.005 and lower>0 against all4 comparators. Paired mean +/-3.5SE
over32 trajectories, exploratory. No case selection; all2688 consumed runs.

Before quality collection, verify original mode exact compatibility; boundary
query uses no unseen label; paid current-origin label affects only future
predictions; no double delivery; immediate controls unchanged; paid label appears
at the intended next fit. Reuse race testing and independent scorer. Archive
source needed to reproduce original artifact before modifying a hashed helper.
Do not call this a six-expert switch experiment or fresh confirmation.
