# V80 equivalent incremental frozen-support ledger

All seven WHOLE goals OPEN. Parent V79 checkpoint SHA256
`c5582c55cf000fbfa2d34d91c1da54f7a6e06fdd0493f2113f403b162b05c839`.

## Patch reasoning gate

Confirmed source and cost findings: every Issue, Reveal and both Pending branches
replay the entire history; each row allocates/copies fixed200-member component
arrays and repeatedly searches the same support. Full queries148/187-188ms and
766MB/1.02GB cumulative allocation. Alternatives are unavoidable law complexity,
publication hashing, noisy timing and genuinely expensive full-history delayed
conditioning. The trace/source show replay/allocation is upstream; hashing is
also real and remains measured. No daemon repair or upstream fix is substituted
for this isolated research lead. A data repair would not fix the same expensive
loops on valid unknown, first/pair and late histories.

Preserve the mathematical law, scored receipt, support order, normalizers,
envelopes, epochs, error handling and accepted history range. A fixed-size four
prefix cache (stride64 original issues) stores immutable inference states. New
issues advance the current state once. Evidence updates start from the latest
cached prefix that EXCLUDES the changed row; invalidate/rebuild later prefixes
only after successful publication. An available prefix is at most63rows before
the changed row; the whole suffix still propagates to the present. Older
evidence falls back to equivalent full replay,
not ignored or dropped. Refresh changes support epoch and rebuilds caches from
scratch. Hypothetical queries read but never mutate published caches. Reuse two
component buffers and support-index maps within a replay; retain arithmetic
summation/order. No indifference projection or selective observations added.

Falsifiers: any independent dense or V79 law/normalizer/receipt/branch mismatch
over existing2e-11 tolerance; stale prefix surviving an earlier reveal/refresh;
query or rejected update mutating cache/publication; future label seen before
reveal; losing old-query support/fail-closed behavior; cache count exceeding4;
dimension or configured history gate changed. Test prefix boundaries63/64/65,
multiple old/recent/reverse reveals, unsupported pairs, current future forks,
uncapped small controls and full150/200-member history. Do not infer full-policy
trace equivalence from point forecasts alone.

Freeze candidate, tests and compiler closure before unit/race/vet/benchmarks.
Matched raw benchmark command will include V79 reference and V80 operations
under equivalent states with latest, old and nearby delayed evidence. Report
both allocations and serial duration; no loaded serving/peak-RSS claim. Measure
retained cache/support type sizes and actual active workload separately from
empty constructors. Original max-history support memory remains a limitation;
four cached prefixes alone do not guarantee a whole8MiB bound.

This is not improved scientific quality: V79 noise-support exclusion and V77
approximation defects remain. After equivalence and cost gates, unchanged broad
controlled recovery/non-harm/equal-TOTAL-cost and untouched-agent/loaded-serving
requirements still apply. No production/private/sealed/reserved-seed access,
dirty tracked edits, whitepaper/publication or commit/push/install/deploy changes.
