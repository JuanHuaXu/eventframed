# Pending-feedback loss budget: exploratory contract

Before execution: use the same 84 consumed cadence records and issued expert
forecasts as feedback-v1. No refits, tuning grid, new confirmation or production
changes. Keep the feedback-v1 proposed weights (equal prior, squared loss,
learning rate 1, once-per-origin arrived evidence). Add a separate publication
guard. The expert weight learner still scores its two original experts, not
the guarded mixture.

For forecast i, allow cumulative excess squared loss at most .01*(i+1).
Track actual guarded-minus-incumbent excess for labels already arrived, and
the maximum possible excess over y in {0,1} for every pending or permanently
missing label. Never release a missing-label reserve without evidence.
For b incumbent, c challenger, d=c-b and chosen weight w, new excess is
w*w*d*d + 2*w*d*(b-y). Its worst case is a*w*w+k*w, with a=d*d and
k=max(2*d*b,2*d*(b-1)) >= 0. Choose the largest weight no greater than the
proposed weight that fits the remaining budget. Use a stable quadratic root;
if d=0 retain the proposed weight because the law does not move.

Induction: replacing a pending worst-case reserve with an observed excess
cannot increase the ledger. Admission bounds the new ledger by .01*(i+1).
Actual cumulative excess, including hidden outcomes, is at most that ledger.
This is a pathwise realized-loss statement for bounded binary Brier loss,
not a per-realization expected-Brier guarantee or a population confidence
certificate. Verify all prefixes, not only terminal loss. Floating-point
checks use 1e-12 tolerance; do not claim exact machine arithmetic proof.

Report expected/realized Brier versus Markov, unguarded feedback mixture and
fixed half-mixture; number of records with expected harm >.01; weights and
clipping frequency. Retain all candidate fitting costs. Passing the ledger
invariant alone is insufficient: useful pooled gains and stationary harm
must be examined separately. No seven-goal completion claim from this pilot.

Test every outcome sequence for a small delayed/missing fixture, no-feedback
and identical-expert controls, current/future-label poisoning, arrival reserve
release, and independent retrospective prefix-bound evaluation. Research
classification: incomplete incumbent protection is confirmed; this is a new
candidate policy, not a production bug repair. Its falsifier is lost utility
or persisting unacceptable expected-score harm despite the loss budget.
