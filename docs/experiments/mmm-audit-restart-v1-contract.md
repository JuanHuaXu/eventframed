# Restarted audit evidence, consumed-data screen

The single-start audit margin experiment detected only 16/128 changed
schedules by frame 511. Pre-change evidence may drain its product before
the changed regime arrives. Freeze this rescue before scoring.

Use the same complete 256-schedule finalized audit stream, paired score
W=reference correctness minus live correctness, tolerance 0.15, and one-sided
rate 0.5. Open exactly eight e-processes at 1-based audit indices
1, 2, 4, 8, 16, 32, 64, 128. Before opening, a process equals one. After
opening, each score multiplies it by 1+0.5(W-0.15). Define wealth as the
arithmetic mean of all eight processes and nominate on first wealth>=20.
No start selection from outcomes, case identity, model fit, external split, or
known change time. Compare with both previous single-start processes and
the original external gate. Report by511/543 counts, stable and pre-change
alarms, delays, and audit updates across every cell. Preserve all failures.

Under the idealized null E[W|audit history]<=0.15, each component is a
nonnegative supermartingale, and their fixed average also is. The nominal
Ville alpha of 0.05 applies to that scalar null under those assumptions.
It does not certify the target-law diameter, correct causal split, or unknown
real-world selection and missingness. This bank can nominate investigation
only. No changes to served forecasts, Anti-Pigeon gate, or audit cost.

Eight products and one average require constant work per finalized audit;
measure this arithmetic separately and do not call it daemon latency.
This is reused consumed data, not an untouched confirmation test.
