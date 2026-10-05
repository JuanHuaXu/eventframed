# Rational joint-policy witness certificate

Use the strongest saved random/normalized-entropy witness, replacing each row
weight by floor(1e9*w) and normalizing the integer weights by their exact sum.
This is a new valid rational dual distribution, not a claim that rounded weights
equal the floating-point witness. Do not optimize or tune after quantization.

Source noise values are exact decimal rationals10/100,15/100,20/100,25/100,30/100;
the context source error is1/100. Sum over eight distinct latent bit assignments
(the fourth bit in the original16-state model is irrelevant and duplicated).
Represent all history joint masses over common denominator8*100^10. Copied
renewals contribute indicators; independent renewals contribute Bernoulli factors.

Decode frozen forecast doubles as exact binary rationals. Recompute control
risks for both raw vectors and their exact simplex normalization. Use the larger
upper bound of the two control conventions. Random path multiplicities have
denominator4^6; deterministic path counts are integers. Recompute true expected
control scores, not their previously rounded JSON headline values.

At each state, integer weighted outcome masses v have denominator D. The joint
oracle's minimum Brier cost is exactly ((sum v)^2-sum(v^2))/(D*sum v), with0
when total mass is0. Use integer division to bound this rational from below
and above in units2^-80. Add/minimize these bounds backward without choosing
an approximate action. Bound control contributions similarly and subtract the
control UPPER bound and final allowance UPPER bound from oracle LOWER bound.
A positive integer result certifies infeasibility for this exact rational source
family, both control forecast conventions, fixed observation budget and arbitrary
state forecasts/observation policies as in the joint oracle derivation.

This is an arithmetic certificate implemented by a small checked program, not
a proof-assistant-verified theorem or a claim about arbitrary real-world sources.
Check integer likelihoods against a separate repeated-product construction,
forecast decoding, exact enclosure arithmetic, and full replay. The rational
source model is the intended decimal formulation; do not silently claim a
certificate for every binary rounding of source probabilities.
