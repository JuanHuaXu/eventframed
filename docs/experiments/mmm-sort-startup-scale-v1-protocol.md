# Incremental publication startup scale v1: frozen private diagnostic

Use one private single-owner sortable Store and the v1 SQLite journal gate.
Grow it by authorized one-event appends to total row counts 35, 259, and
1,027. At each size, fully close the current owner and run five paired
reopens, rotating which arm is measured first. The control opens the same
LibraVDB Store and SQLite marker without journal verification; the candidate
opens through the incremental gate and must verify READY. Never keep two
Stores open at once. Exclude Close from timing and time each Open with Go's
monotonic duration clock. Check the candidate READY result and row count
outside the timed span. Reopen the candidate after each size to resume
appending.

Report every paired total-Open duration plus p50/max and candidate/control
ratios at each size. This is an explanatory diagnostic with **no pass gate**:
the goal is to determine whether restart verification dominates ordinary
Open and how its cost changes across these finite sizes. Do not extrapolate
from ~1k synthetic rows to millions or billions, claim loaded service
latency, or replace the O(N) complexity observation with a small timing.
Preserve unfavorable results and keep production untouched.
