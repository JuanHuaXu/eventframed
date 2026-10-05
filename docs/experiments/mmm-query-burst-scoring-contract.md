# Independent query-burst scoring contract

The quality gates remain those in mmm-query-burst-protocol.md. This document
records verification mechanics, not a post-result change in success criteria.

The collector retains seven arms: natural, periodic random/entropy/disagreement,
and burst random/entropy/disagreement. Each uses the same four expert fitters
and32/64 evidence windows. The new periodic and burst arms share a32-origin pool.

The JavaScript audit independently reconstructs:
- natural/paid arrival times and single-use evidence admission;
- surprise against the served forecast actually issued for that origin;
- expiry-aware mixer admission and a direct-probability four-state Markov
  recursion (Go uses a log-space checkpointed journal);
- each periodic/burst query clock and bounded eligible pool;
- exact seeded random choices and maximal entropy/disagreement scores within
  numerical tolerance1e-10, including possible near-ties;
- every fit's full ordered origin list;
- full and terminal64 Brier from each recorded forecast and external Q.

No fit theorem is assumed: expert forecasts are inputs to the independent mixer
audit. Expert fitting retains its original Go tests; this audit is not a second
independent implementation of the fitters.

Both file streams attach their line iterators before any await, preserving
headers while hashing. The original source hash must equal the collector header.
Require2688 unique grouped trajectories and24,084,480 probability checks.
Paired cost mismatch is reported and prevents an equal-cost gate from passing;
it does not exclude the trajectory. Natural is a no-paid-label comparison, not
an equal-acquisition-cost control.

Tests before full scoring passed on two handcrafted256-frame tapes: complete
immediate feedback and entirely missing feedback. Every forecast is0.5, giving
exact0.25 Brier. These check zero versus31 acquisitions, late-label accounting,
and rejection of invalid fits, reveal times and origins. Three original Go
immediate-feedback trajectories also passed the direct-probability recursion.
The full pass additionally mutates four artifacts to verify rejection of bad
reveal timing, trigger bits, fit origins and baseline forecasts.

Diagnostics: triggerClocks counts clocks with at least one surprising arrival;
paidTriggers counts surprising arrivals for which paid delivery is available
by that clock, including ties with natural arrival. It is not proof payment
was necessary for the arrival. EmptyQueryClocks in burst mode counts all empty
clocks, whereas periodic mode records empty scheduled-query clocks; do not
compare those raw counts as equivalent denominators. Query clocks and per-block
counts are retained for narrower timing analyses.

Run summaries are deterministic scoring replays, not independent recollections.
Report elapsed whole-collector runtime separately from serving-path latency.
