# Frozen optimistic-transfer noise misspecification test

Status: predeclared before results. No predictor or solver changes.

The preceding finite 180-gate pass assumes the measurement likelihood is correct.
This test asks whether its protection survives a wrong noise model, not whether
the conditional finite-family mathematical bound was incorrect.

Use the same ten allocations, sixteen counterfeit masks, uniform sixteen latent
hypotheses and all 1024 binary report vectors. Prediction always uses the original
noise=.20 table, original model priors, optimistic target and epsilon=.01 guard.
Actual generating noise is .10, .15, .20, .25 or .30. The existing table function
maps noise to channel errors; for measured types [0,1,2,7] this changes types
0,1,7 while type2 remains .01. Copy rules, hypothesis prior and schedule do not
change. Actual noise and mask are used only to score the frozen predictions.

For each world, preserve all 180 gates: 160 population nonharm comparisons
against the same misspecified local control (harm <=.01), ten genuine gains
>=.005, ten all-counterfeit confidently-wrong reductions >=.05. Report every
failure, not just the aggregate. These are exact finite expectations, not sampled
confidence intervals. The matched .20 controls must reproduce prior values;
replay the whole experiment and compare the .20 candidate to the archived
optimistic result. Numerical certificate checks continue to concern declared
conditional laws ONLY; do not present them as certificates for the actual world.

This is a stress test, not a production patch or an independently calibrated
likelihood estimate. No fitting on the new outcomes, altered risk budget or
changed thresholds is permitted. A misspecified-world failure falsifies robustness
to that tested misspecification, not the conditional guarantee's algebra.

