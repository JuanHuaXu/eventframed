# Batch publication open-loop v1: frozen private backend screen

Run three matched pairs in rotated control/candidate order, with a fresh
single-owner private LibraVDB Store and SQLite journal for each arm. Seed
200 eligible EventFrames and four future-available sentinel EventFrames.
The two arms receive the same 256 later eligible events. The sentinels
have the closest query vectors, so a failed as-of filter has a visible
chance to enter the returned top ten. All data are synthetic and contain
no user sessions.

Offer 256 event writes and 192 exact-LSN vector searches at nominal 4 ms
intervals on separate producers. The control publishes each event with
one receipt and journal entry. The candidate publishes groups of at most
16 events, with at most 16 ms of intentional batching wait measured from
the oldest offered event. Events already queued behind earlier writes may
exceed that age; count that queue delay in offer-to-ack. Both use the same test-only single-owner
gate and its snapshot/LSN motion guard. Record actual offer gaps, group
sizes, DB+publication call durations, per-event offer-to-ack age, search
call and offer-to-completion time, search errors, and returned IDs. Do not
hide queue or dwell time inside a writer-call metric.

After the load, assert 256 unique acknowledged IDs, no future sentinel in
any completed search, no unknown or duplicate search IDs, exact row and
journal counts, durable presence and sortable key for every offered event,
and successful full-journal verification after the owner closes and
reopens. Report pooled p50/p95/p99 and per-pair p99. The candidate component
screen requires zero integrity violations, no search errors, writer
offer-to-ack p99 below 250 ms, search call p99 below 100 ms, and a search
call p99 ratio no worse than 1.10 versus control. These are backend
diagnostic thresholds, not a Goal 6 completion criterion.

No service-level Recall, labels, background learner, agent outcome or
cross-process writer fence is included. The test cannot establish a 4 ms
learning-freshness guarantee. Keep unfavorable results and production
untouched.
