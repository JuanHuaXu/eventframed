# Input-population query opportunity

Freeze before collection. Prior oracle risk uses the particular31 future input
draws. Retain the same2688 consumed source histories, candidate pool, natural
publication evidence, paid-label counterfactuals, model and horizon161..191.
Replace only scoring over those draws with expectation over the generator's
full input law. Transfer inputs are uniform9-bit values; Boolean inputs are the
pushforward of uniform9-bit values through the generator's declared input map.
Future teacher probabilities remain an offline oracle, never learner inputs.

At publication let p_x be the fixed fitted forecast and a_i=.99^i. Then
p_{x,i}=.5+a_i(p_x-.5). Expected Brier against q_{x,161+i} is
p_{x,i}^2-2*p_{x,i}*q+q. Enumerate all512 raw input values with weight1/512 at
each of31 times. Equivalently accumulate a quadratic A_x*p_x^2+B_x*p_x+C_x,
with b_i=.5*(1-a_i), coefficients a_i^2,2*a_i*(b_i-q),b_i^2+q*(1-2*b_i).
Mapped duplicate inputs contribute their probability mass; do not deduplicate
them into an incorrectly uniform distribution over distinct outputs.

First verify the quadratic reduction against direct enumeration on all21 cases,
both phases, and both delivery schedules, including nonconstant forecast vectors.
Regenerate/check teacher metadata and verify stored source Q against the
reconstructed law. Reject malformed identities; test duplicate-input weighting.
No prediction-policy efficacy claim follows from this kernel component alone.

For full collection, fit each no-query/paid-label branch using only its permitted
publication evidence. Record sampled and population risks for both possible
nonredundant purchased answers. Naturally redundant branches remain the observed
no-query forecast; do not flip natural evidence. Compare sampled forecasts and
risks exactly/tolerantly with the existing counterfactual artifacts, and average
the two population risks using the query's generator probability offline.
Report all84 cells and unchanged random/entropy/joint policies, best paid oracle,
and abstaining oracle. Conditional knowledge of the teacher and future natural
arrivals remains: population averaging removes input-draw luck only.

Preserve source/code hashes, failures, numerical tolerances and replay. Keep
all seven goals open unless their full criteria are independently met. No
production, paper, commit or push changes, and no new critic tuning in this run.
