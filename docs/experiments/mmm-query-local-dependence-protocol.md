# Evidence-conditioned dependence: four frozen cells

Follow the marginal-preserving Bernoulli mixture in the dependence-v1 protocol.
Before scoring, freeze four calibration cells: query uncertainty
4*p*(1-p)>=.5 versus below.5, crossed with whether the target's encoded input
exactly matches any retained decision-time observed input. Cell index is
2*highUncertainty+hasObservedMatch. No generator/case identity, hidden label,
future arrival, teacher probability or outcome enters cell assignment.

Build the observed-input set from coverage.Origins, which must be eligible
strictly before the decision's target observation and available by clock160.
An initial negative origin uses its initial encoded input. Duplicate support
inputs create one membership value, not extra evidence weight. Future target
inputs are used only for evaluating their forecasts; acquisition targets remain
the already visible original eight inputs153..160.

Fit one lambda in[0,1] per cell with the existing convex conditional-log-loss
optimizer. Train on phase0 only, all candidate/31target pairs, preserving equal
history/candidate/target weighting. An empty training cell uses global lambda1,
the previous frozen result, not a fitted phase1 fallback. Retain counts/weights.
These are consumed data; phase0 archived labels include unavailable historical
query answers and are explicitly offline supervision. No claim of free feedback.

Two distinct tests, neither interchangeable with the other:
1. Forecast diagnostic: compare independent, original, and four-cell pair laws
   on all31 future target inputs. Both marginals remain fixed; pairwise law
   validity does not establish a full multivariate process.
2. Selection test: score each candidate using the average lambda(cell)^2 times
   its original Brier-information gain on the original eight visible probes.
   Preserve original lower-origin/tolerance tie behavior. Publish the selected
   query through unchanged actual publication C, using stored original branch
   risks. Calibrated selection does NOT silently alter the scored prediction
   model. Compare with random, entropy, and original joint8 at the same paid
   query count. Empty pools abstain.

No feature, threshold, cell count, loss, or pass criterion may be retuned after
evaluation. Preserve all2688 records/84cells. Use existing phase1 all-cell
nonharm bounds>=-.001 and positive gains on cases19,20 versus both controls:
forecast joint log loss/target-expected Brier versus independence+original;
selection actual-answer sampled Brier versus random+entropy. Supplementary
selection teacher-weighted population risk gets the same descriptive screen.
Intervals are mean +/-3.5SE over32 trajectories, not independent pair samples.

Check eligibility, duplicate invariance, future/teacher/phase1 isolation, cell
normalization and mixture validity, fitted optima, original-policy replay,
independent scoring, exact replay, and cost of selection with precomputed laws.
All development remains research-only; no production or whitepaper promotion.
