# Discriminating evidence availability

Frozen before scoring. Consumed v120 diagnostic, not a new predictor or test of
active acquisition. All2688 runs retained; paired schedules not independent.

At clocks128,160,192 choose the best fixed expert from [0,1,2,3,10,11] over
the next64 issued forecasts by expected Brier using hidden Q. Tie-break in arm
order. Compare against generic64 (arm0), a declared incumbent, not a claim that
it is the switch mixer's MAP expert. Report future improvement and arm counts.

For preceding32,64,128 frames (truncate at zero), report number available under
!Missing and origin+Delay<=clock, excluded late/missing counts, and Bernoulli
log likelihood ratio of hindsight expert versus incumbent. Report actual and
teacher-expected ratios on available evidence, and expected ratio on the whole
window. Floor probabilities at1e-12. No current/future labels in these sums.
Also report expected log-score improvement over the future64: best Brier and
best log-score need not agree. This separates objective mismatch from lack of
evidence. All ratios unweighted; do not call them valid e-values or posterior
odds after hindsight expert selection and multiple-window inspection.

Summarize means and counts across32 trajectories separately by phase/case/
schedule/clock/window. Use future Brier gain>=.005 only to label diagnostic
opportunities, never to select a deployable action. Within opportunities count
available expected LLR<=0, actual LLR<=0, and actual LLR>log(19). No statistical
independence or effective-sample-size claim follows from raw label count.

Tests: independent direct-product likelihood identity on short rows; unavailable
label and hidden-Q poisoning leave actual evidence unchanged for a FIXED expert;
future-label poisoning leaves expert selection unchanged (selection uses Q).
Record source/input/protocol hashes and byte-identical full replay.
