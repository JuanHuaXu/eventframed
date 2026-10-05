# Processor deadline trace

Keep actual service/temporal guard,100ms age,capacity16,64 distinct histories,
four readers,16 future writes, and GOMAXPROCS4. Three alternating direct/batch
pairs. No policy change, off control or new performance-success threshold.
This diagnoses failures rather than retesting foreground non-harm.

At processor entry/return record monotonic remaining deadline, compute duration,
request fixture ID, fit/context errors. Preserve every entered processor; count
accepted jobs that never reach it separately through aggregate status. Verify
remaining-entry minus compute equals remaining-exit, IDs unique, started count
matches trace length, and success/error totals reconcile. Trace overhead occurs
after fit and before scheduler revalidation and may itself affect tight jobs.

Classify positive-entry budgets below10/25/40ms descriptively. No empirical
42ms mean is a hard lower bound for individual fits. Success before processor
return is not proof of passing the subsequent scheduler freshness check. No
per-job post-check cause is exposed; report that boundary as unobserved.
