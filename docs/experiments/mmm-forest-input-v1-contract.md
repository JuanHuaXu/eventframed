# Frozen four-arm input-model screen

Fresh seed base2026092210; other seed offsets and576cells match the shape of
the prior tree screen, not its draws. Same threeinput laws, two targets,
noise.05/.25/.5, n64/128,16fits/cell, masks0/1/3/31/511.
Arms uniform/histogram/fulltree/heldoutforest share all training observations
and subset outcome learner. Forest reserves its chronological second half for
input validation; no refit or split-ratio tuning. Earlier failed result retained.

Primary forest copied-field/bit/noise.05/mask1 mean gain>=.005 and lower
mean-minus3.5SE>0 at both sizes. All180cell/mask non-harm lower gain bounds
versus uniform>=-.01. Full-input probabilities agree1e-12. Report tree and
histogram contrasts, especially higher-order XOR failures, without replacing
primary criteria. This is exploratory fixed-mask evidence, not whole-goal
validation. Capture sources before collection and replay the summary.
