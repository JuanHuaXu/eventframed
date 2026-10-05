# Canonical numerical tie policy comparison

Diagnostic on the already-consumed 80-regime grid, not fresh confirmation.
Keep every saved mass and forecast in count-planning-exact.json unchanged.
Add tie_one, tie_two, tie_full and tie_entropy policies. At each decision,
choose the first type in [0,1,2,7] whose objective is within absolute 1e-12
of the computed minimum. For recursive planning propagate the selected cost,
not the unattained minimum. Bound remaining depth by six renewals.

Reconstruct branches from child/parent model evidence masses. Check their sum,
count all action changes relative to the archived strict-argmin policies, and
bound any full-horizon cost increase by six times the tie tolerance plus 1e-12
numerical slack. Evaluate all ten policies in all five noise values and all
16 copy masks with unchanged forecasts and exact path multiplicities.

Compare tie_two/full with random, archived entropy, and tie_entropy. Keep the
existing final nonharm allowance 0.01 and positive learning-area condition
(except mask15). Record paired policy differences separately. No success
criterion is weakened. Even a finite pass would not complete direction7.

Verification: byte-exact replay; archived-policy scores must be bit-identical
to exact-regime-evaluation.json; separate backward scoring in six representative
regimes; decision and tolerance checks in the compiler. Do not rewrite archived
sources or results. No model-prior change or production deployment.
