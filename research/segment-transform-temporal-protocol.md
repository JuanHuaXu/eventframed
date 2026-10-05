# Temporal compatibility diagnostic

Retain the previous service fixture, four readers/64 recalls,16 future-dated
writes, persistent temporary LibraVDB, capacity16,100ms age, fixed272-frame
context-aware fitter and GOMAXPROCS4. Both arms have shadow enabled. Compare
strict snapshot matching with the existing TemporalReuse policy, three pairs
alternating order. No new guard implementation or production change.

Record started processors and context-interrupted processors in addition to
request timing, overlap and all terminal statuses. This distinguishes jobs
that enter the fitter from ones rejected before it. A stale non-interrupted
job may still reflect post-fit expiry or snapshot motion; do not overattribute.

Frozen diagnostic success: temporal completion >=52/64 in every trial and
no fewer completions than strict; zero errors and full terminal accounting;
temporal p99<=1.10*strict p99. This comparison alone cannot prove non-harm
relative to shadow-off. Preserve all failed cells. No semantic learning claim.
