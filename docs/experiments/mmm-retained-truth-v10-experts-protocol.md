# V10 expert bottleneck diagnostic

Frozen 2026-10-01 after observing the v10 outcome-family failure and its
selected-view oracle gap. This is consumed-data diagnosis, not a new
confirmation experiment or a candidate forecast. Use the archived v10 full
journals only. Include both input modes, confirmation `shift128` and
`shift256`, all 24 trajectories per cell, and both all post-change frames
and the first 64 post-change frames.

For each arm's *issued forecast on its actual selected view*, score final
forecast and the archived base, challenger, long-window and neutral experts
against the realized outcome. Also score the archived short-count and tree
inner predictions on the same arm/view. Do not choose an expert using the
same frame's outcome, do not refit, and do not transfer predictions between
arms with different selected views. Compare average Brier and observed-bit
count; report every arm, scenario and input mode. Exact reconstruction of
each arm's final Brier from the full journal must agree with the original
v10 summary. Preserve source/input hashes and measure diagnostic runtime.

If an individual expert improves substantially but the issued mix does not,
that is evidence for a weighting/selection hypothesis only. If no expert
improves, that supports a capacity/support hypothesis. Neither observation
proves causality or validates a new policy. The retrospective scores are
not independent confirmation or production latency evidence.
