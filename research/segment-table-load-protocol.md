# Finite-table actual service comparison

Three alternating batch/table pairs,64 distinct workloads, actual Recall/Observe
and persistent temporary LibraVDB. Four readers,16 future-dated writes,
GOMAXPROCS4,100ms age,capacity16,temporal compatibility. No memoization or65ms
allowance in either arm. Candidate differs only in finite beta-log lookup.

Reference forecasts use the previous pairwise method, outside measurement and
never supplied as predictions. Check every returned forecast to1e-12. Keep
deadline-entry/compute/exit traces and all terminal counts. Success still requires
>=52/64 candidate completions, no fewer than batch, p99<=1.10*batch, no request
errors and overlapping writes in all pairs. No shadow-off population non-harm
claim; this is a matched processor comparison. Retain all failed outcomes.
