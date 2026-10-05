# Fixed computation allowance experiment

Retain the deadline-trace distinct-work fixture and actual service:64 recalls,
four readers,16 future writes,100ms age,capacity16,temporal reuse,GOMAXPROCS4.
Three alternating batch/allowance pairs. Both arms use batch normalization.
At processor entry the allowance arm returns an explicit error without fitting
when remaining deadline<65ms.65ms is a provisional allowance above the previous
58-61ms successful loaded fits, NOT a certified worst-case bound.

Refused jobs remain failed/stale terminals, never completions or labels. No
deadline or snapshot guard changes; no production change. Trace all entry,
compute,exit timings and error strings. The recognized refusal error is separate
from unexpected failures; account for all64 offers and actual forecast returns.

Original completion screen remains>=52/64 and foreground p99<=1.10*control,
with zero request errors and overlap. Additionally report actual completed
counts, context interruptions and total processor time. Reduced wasted work
without meeting completion is a partial result, not a successful rescue.
