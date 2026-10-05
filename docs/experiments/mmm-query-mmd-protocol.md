# MMD representativeness acquisition test

R-I-inspired adaptation of Tang/Sloman/Kaski2026 equation10, not R-IDeA and
not an inherited theorem. Freeze choices before evaluating consumed histories.

Represent observed design history by the selector's latest63 retained observed
inputs (coverage.Origins), including initial inputs. Preserve multiplicity.
Each candidate appends its visible input, matching the selector's64-label paid
conditional support. No unknown label is used by this representation feature.
The target distribution is the empirical multiset of all161 visible inputs0..160.
Do not use future161..191 inputs, generator identities or teacher probabilities.

Kernel: k(x,z)=exp(-Hamming(x xor z)/2) on nine binary coordinates, the RBF
kernel with bandwidth1. Use the biased empirical MMD-squared V-statistic,
including diagonal terms; MMD is its nonnegative square root. Numerically tiny
negative values above-1e-12 clamp to0; larger negatives are errors. This is not
an unbiased U-statistic and duplicates remain weighted observations.

Factor=1-.5*MMD(history+candidate,target)/MMD(history,target).
Keep the signed factor, with no silent positive clipping. If baseline MMD^2
<=1e-12, use factor1 for all candidates (explicit denominator extension).
No bandwidth, factor coefficient, target count or epsilon tuning afterward.

Multiply original eight-probe Brier information gain by this factor; use the
existing lower-origin/1e-10 tie rule. A separate factor-only arm (equivalent to
minimum post-addition MMD when denominator>0) diagnoses representativeness alone.
Controls: noquery, random, entropy, original joint8. Every nonempty pool buys
one query. Publish through the unchanged actual C branch laws; MMD does not
alter the scored forecast itself or its confidence.

Use all2688 records and84cells; report both phases separately. Primary metric is
actual-answer sampled Brier, supplementary is teacher-weighted population Brier.
For each candidate arm apply existing phase1 lower-bound nonharm>=-.001 in all
21cells and positive gains on cases19,20 versus random+entropy. Bounds use
mean +/-3.5SE across32 trajectories; not fresh or sequential confirmation.

Test incremental/direct MMD equality, duplicates, permutation and bit-permutation
invariance, zero denominator, negative factors and malformed inputs. Verify
as-of support, no-future/teacher feature access, old policy equality, independent
scoring/aggregation and exact replay. Measure whole MMD selector setup plus
candidate costs with precomputed forecasts, excluding model fits and serving.
No production, whitepaper, dependency, commit or push changes.
