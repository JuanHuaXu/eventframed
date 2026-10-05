# Exact mixed-counterfeit control

Declared after the v14 fixed-trace failure, before this control's evaluation.
Use the unchanged priors and likelihoods, noise20. Fixed16-credit schedule:
ordinary first reports from types0,1,2,7, then renewed measurements from types
0,1,0,1,2,7. No adaptive selection or hidden truth enters inference.

Enumerate all1024 binary report vectors and16 equally likely hypotheses under
four true measurement mechanisms: all renewals genuine, all counterfeit,
even-type renewals counterfeit, odd-type renewals counterfeit. Counterfeit
reports equal that type's already acquired ordinary first report; genuine ones
are independent conditional draws. The source mechanism is not given to models.
The ordinary first observations are conditionally independent across types.

Score exact expected final Brier, accuracy and confidently-wrong probability for
certain-fresh, local-uncertain and hierarchical models on the same fixed schedule.
Also compute the true-law Bayes risk on the complete acquired vector as an oracle
floor, not a deployed forecast. Require hierarchical nonharm versus local within
.01 Brier in every mechanism, all-genuine gain>=.005, and at least .05 reduction
in false confidence versus certain-fresh on all-counterfeit. This control is not
a substitute for closed-loop unknown-mode validation even if all gates pass.

Check all true-law vector masses sum to1, posterior score never beats the oracle
floor, and deterministic replay. Do not tune priors or the schedule after results.
Full-support local inference processes all report vectors; impossible vectors
under a given true law contribute zero risk rather than being invented as data.
