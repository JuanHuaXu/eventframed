# Fixed-rate fresh learning diagnostic

Keep64 distinct histories, actual persistent service,100ms scheduler age,
capacity16,temporal guard and table fitter. Three alternating off/on pairs at
10 and40 requests/second. Four readers dispatch strided IDs4*i+w at absolute
epoch+ID/rate, not a sleep after completion. A late dispatch keeps its original
scheduled time; no omitted arrivals.16 future-dated writes remain concurrent
near the beginning as in earlier tests, not continuous throughout the run.

Capture request duration and scheduled-offer-to-response separately. Processor
trace also includes scheduled-offer-to-return. Scheduler age still begins at
nomination, so offered-time on-time completion uses a conservative lower bound:
completed minus successful returns later than100ms, clamped at0. This cannot
credit a fit that the scheduler rejected. Final validation time after processor
return is not individually traced; thus this is not a strict offered-to-final
publication deadline certificate and must be described accordingly.

Per rate/trial: zero request errors, overlap, full accounting,>=52/64 scheduler
completions AND lower-bound on-time returns, scheduled-request p99<=1.10*off.
Preserve failures at both rates. These are new operating-point diagnostics,
not replacements for the prior saturated-workload failures. No claim about
population p99 or real task usefulness. No cache or prefix preparation.
