# Lazy repair cost screen

Reuse the frozen summary-cost workload and its thresholds, without the summary
optimization: CPU4, N800/6400, serial seed,32 inserts,8 forced entry-point deletes.
Only change eager matrix preparation to preparation at the first neighbor below
the existing reconnection threshold. Graph-state hashes must match. Require at
least10% deletion improvement and no more than5% seed/insertion regression in
each size/repeat. Two repeats are a screen, not a production/tail guarantee.
Fresh order: control0,lazy0,lazy1,control1. Counters disabled. Preserve all arms.
