# Timely-label acquisition screen

Frozen before outcomes. Consumed v120 fixed-forecast replay only; not prospective
real-data validation. Prior disagreement failure (mmm-falsification-v4-results.md)
remains. This tests timely acquisition, not automatically generated hypotheses.

Use all2688 runs, six issued experts [0,1,2,3,10,11], basic switch distribution
with prior [.95,.01,.01,.01,.01,.01]. No expert refitting. Four policies: natural
arrival only; uniformly random unavailable packet; maximum predictive entropy;
maximum weighted expert Jensen-Shannon disagreement. Score all256/terminal64.

At clocks8,16,...248 AFTER issuing the current forecast, nominate unavailable
origins in [t-8,t). Pools do not overlap. Naturally arrived labels excluded.
Acquire at most1 label per nonempty pool, available from t+1. Query returns the
true recorded label even if natural delivery would be missing. This assumes a
perfect fixed-cost label service: a mechanism isolation, not a real capability.
No use of Q, hidden Y, or eventual delivery time in selection. All three paid
policies see identical pools/costs. Random uses a separate SHA256 draw keyed by
record identity and clock. Ties select earliest origin. Entropy and disagreement
use the as-of posterior weights and originally issued expert forecasts, not
predictions refitted after observing the selected label.

Natural missingness/arrival enter only through whether a label has arrived.
Acquired evidence is not duplicated when it later arrives naturally. Acquired
labels from earlier batches may change posterior weights and later selections.
This is a conditional working model, not a general selection-bias theorem.

Candidate disagreement must protect all168 cells against each of natural,
random,uncertainty: paired mean +/-3.5SE gain lower>=-.01. Terminal delayed
changing cases1,2,4,5,7,8,19,20 require mean>=.005 and lower>0 against all three.
Keep both phases separate, all cases, all costs. Intervals exploratory, not
simultaneous confidence sequences. A failure is not repaired by case selection.

Tests: natural control exactly matches previous switch artifact; no paid queries
under immediate complete delivery; identical paid counts; queried origins unique;
future/unavailable label poisoning leaves prefix choices/forecasts unchanged
until label natural or paid reveal. Hash inputs/components/protocol. Full replay.
Report research wall time, not serving latency. Prefix replay O(K*T^2) per policy;
bounded pool scoring O(K*8), budgets<=31 queries per trajectory. No claim of
end-to-end acquisition cost equivalence beyond the declared one-unit oracle.
