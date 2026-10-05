# Source-Bound Lexical-First Serving Research

2026-10-04, design only. Native calibration measured173.816msserial p99, mostly
native Search170.302ms. Non-native stages average2.984ms/p994.543ms. This motivates
testing a source-bound lexical-first foreground plus bounded semantic background,
NOT assuming a cached-index benchmark completes loaded continuous learning.

## Existing Machinery and Missing Contract

Read actual `internal/store/libravdbstore/research_sort_published_view_test.go`,
`research_published_payload_v1_test.go`, `internal/store/research_snapshot.go`
and `internal/service/research_shadow.go`. They supply journal-certified exact
LSN leases, payload projection, bounded publication age, poison/cancel and
future-motion compatibility under explicit one-owner assumptions. They do not
make the new public corpus index a live source-bound lexical service. In
particular, a closed-copy verification-trace hash is not a live store LSN.

Minimum snapshot bundle: source payloads/digests/span IDs, availability cutoff,
owner/tenant/collection, certified LSN/runtime/evidence epoch, corpus-term-stat
scope, feature contract, model fingerprint and publication time. Capture this
bundle atomically; do not pin the epoch AFTER retrieving candidates. Unknown
motion or changed visible payload/term statistics invalidates it. Build and
validate a replacement outside the serving lock, then publish only if its captured
source/model dependencies still match. Stale builds never gain serving authority.

Availability cutoff must govern corpus DF/mean length BEFORE building index
statistics, not merely filter returned candidates. A future-only source can
otherwise change scores even when absent from results. An LSN lease alone does
not prove caller-as-of eligibility; history-bucketed indexes need exact declared
cutoffs or a validated incremental temporal-statistics structure.

## Experiment That Would Prove More

Use actual stored5W1H payloads and the existing real durable journal/lease path
in a NEW owned research store. Controls: current semantic nomination, lexical-only,
and lexical-first with bounded background semantic work, at SAME offered queries,
writes, outcomes, worker count and deadlines. Count changed nominee universe;
do not describe route changes as merely faster ranking of the same candidates.

Measure complete offer-to-Recall and offered-label-to-publication ages, backlog,
drops/errors, durable acknowledgements, rollback/reopen and RSS/allocations.
Include visible/future writes, unrelated and referenced edits/deletes, model
updates, cancellation, expired views and a deliberately late background build.
Preserve all requested arrivals; admission rejection is a failure under the
original completion gate, not a shorter workload. An async semantic job is
charged to total compute and cannot confer authority before its evidence/audit.

Source-IDF duplicates term maps and reaches310.55MiB in the current offline
prototype; concurrent full rebuilds may multiply this cost. Bounded job queues,
versioned immutable sharing and explicit release/retention accounting are needed
before calling memory use controlled. A cheap lexical response can still lose
novel semantic answers; matched outcome quality is mandatory, not just latency.

This step is NOT implemented or tested yet. It remains a viable goal6 lead,
alongside delayed-score alignment, useful Anti-Pigeon split outcomes and
equal-total-cost observation. Keep production/private corpora unchanged.
Methodological latency source: [Dean & Barroso2013](https://research.google/pubs/the-tail-at-scale/).
