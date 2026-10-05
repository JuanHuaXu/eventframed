# Rational certificate for the finite observation-model barrier

[Protocol](EXACT_JOINT_PROTOCOL.md), [certificate](exact-joint-certificate.json),
[verification](exact-joint-verification.json), [arithmetic tests](test-exact-joint.mjs).

The random/normalized-entropy screen now has a positive lower bound constructed
with integer arithmetic and outward rational enclosures, not just floating-point
agreement. The weighted minimum violation is at least

605940659664608577417 / 1208925819614629174706176,

approximately .0005012223660321566. Positivity is checked on the integer
numerator. The displayed decimal is informational, not used for certification.

## What is certified

The source probabilities are exact rationals with common decimal interpretation:
noise10/100,15/100,20/100,25/100,30/100; context-source error1/100. All sixteen
copy masks and the original finite observation/query contract are retained.
The candidate may choose both forecasts and observation actions adaptively,
as in the joint oracle. The certificate covers the specified sufficient-state
family and its mixtures, not arbitrary sources or real-world tasks.

The rational witness uses floor(1e9*w_i) from saved round126, normalized by exact
integer sum999999961. All retained weight is on learning-area rows. Final rows
have total weight0 and the final-allowance term is0. Consequently this witness
rules out simultaneous nonharm on all weighted area comparisons even WITHOUT
the final-error constraint. The desired strict area improvements are therefore
unattainable in this exact finite model. This is not merely an overly tight .01
terminal allowance or an unfinished optimizer run.

Both control conventions are covered: original stored binary forecast vectors,
and their exact normalization onto the probability simplex. Control expected
risks were recomputed from exact source masses and integer path counts, not
read from rounded headline scores. We subtract the larger aggregate control
upper bound. The min/max arithmetic and rational weight normalization preserve
a valid dual witness; no assumption of optimizer convergence is needed.

## Arithmetic construction

Each history joint outcome mass has common denominator8*100^10. The irrelevant
fourth latent bit is summed out. An independent16-latent repeated-product test
checks this reduction. Weighted masses have integer numerators v_y; the minimized
Brier state cost is ((sum v)^2-sum(v_y^2))/(D*sum v). Its lower and upper bounds
are exact integer divisions in units2^-80. Backward sums and minima preserve
enclosures without selecting an approximate optimum.

The final oracle enclosure width is1008 units2^-80, about8.34e-22. Control and
allowance upper bounds are included separately before testing positivity.
This is a checked-program arithmetic certificate, not proof-assistant verification.
It certifies the intended exact-decimal source family, not every alternative
binary rounding of those probabilities.

## Verification

- Full byte-exact certificate replay and source/witness hash checks.
- 15360 independent integer likelihood comparisons against a16-latent
  repeated-product implementation.
- 3200 exact enclosure inequalities checked by cross multiplication.
- Eight binary-fraction fixtures, including the smallest subnormal, plus four
  invalid inputs.
- All48048 states processed; bound recombination and positive numerator checked
  independently from the certificate's stored arithmetic components.

The remaining logical assumptions are explicit: the finite likelihood family,
sufficient-state reduction, additive scoring, and the Brier best-response
derivation. A proof assistant has not verified those assumptions or the program.

## Research consequence

The acquisition/forecast tuning branch under this fixed information model is
exhausted for the full area-dominance requirement. Do not relabel its failures
or spend further rounds tuning equivalent policies. Direction7 itself remains
open: richer observable evidence, independently informative source/provenance
checks, or a different genuine task environment can change the information
available. Any new option must have equal availability and charged cost for
candidate and controls, and it must not reveal hidden truth by construction.
No empirical or production result follows. All seven whole research directions
remain open; the paper and runtime are unchanged.
