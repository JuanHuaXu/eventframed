# Summary maintenance cost screen

Compare control and summary backend overlays at N800/6400, dimension768, CPU4.
Both include the same disabled probe hooks; no counters are enabled during this
test. Serial deterministic initialization,32 fresh inserts,8 repeated current
entry-point deletions. Time phases separately; vector generation, capture/hash,
and shutdown are outside operation timers. Initial/final hashes must match.

Run order: control0,candidate0,candidate1,control1. No concurrent benchmark jobs.
Preserve all outputs. Screening target: at least10% lower total timed deletion
cost in both repeats and no more than5% initialization/insertion regression in
either repeat at each size. Failure is not permission to drop unfavorable phases.
Two repeats are a diagnostic screen, not a population or tail-latency guarantee.
No claim of sustained serving performance follows, even on a screen pass.
