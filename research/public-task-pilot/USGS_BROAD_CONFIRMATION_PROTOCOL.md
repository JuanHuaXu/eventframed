# USGS broad-location current-contract confirmation

Frozen after the consumed Tadine fixed-clock diagnostic and **before** fetching
this cohort or running any of its retrieval questions. Reuse the exact two
current-contract arms, rank-score formula, pack settings, Nomic digest,
independent scoring logic and finite gate from
[`USGS_TADINE_CURRENT_CONTRACT.md`](USGS_TADINE_CURRENT_CONTRACT.md).
There is no fitted parameter or cohort-specific tuning.

Fetch one official [USGS catalog API](https://earthquake.usgs.gov/fdsnws/event/1/)
snapshot with `format=geojson`, `starttime=2026-09-15`,
`endtime=2026-09-23`, `minmagnitude=4`, `orderby=time-asc`, and
`limit=2000`. Preserve raw bytes and capture time. Sort eligible features by
event time and USGS ID, require a finite magnitude, nonempty source place and official event URL,
and take the first 24. Do not select by place, magnitude value, rank result or
answer difficulty. The corpus is each selected event's source-derived UTC
timestamp, place, reported magnitude and magnitude type. The source snapshot's
reported magnitude is the answer label. Source IDs and magnitudes are absent
from question text except where ordinary date/time digits coincide.

Generate three questions per event: exact UTC ISO timestamp, natural-language
UTC timestamp with its source place, and an equivalent timestamp labeled
`UTC+11` with its offset explained and source place. `UTC+11` is a mathematical
offset, not an assertion about the place's civil time zone. First 12 events
form the design stratum and next 12 the confirmation stratum, but no tuning
is allowed between them; the same frozen gate applies over 72 questions.
Capture all 24 records at the single recorded snapshot time for both arms;
Recall as-of is that time plus one second, so paired laws share one world age.

Pass requires at least eight paired top-1 gains overall, one in each wording,
no paired top-1 or pack-survival losses, and per-wording survival no worse than
passthrough. Both arms must receive exactly the same 24 source IDs and
pre-rank scores, return the same 24 IDs, and have identical full-frontier
scored laws and matching journals. The task-lexical rank scores must follow
`1-rank/(n+1)`. Sequential isolated Recall p95 must be below 100 ms per arm.
Report source/code hashes and cluster-level event counts, not 72 independent
samples. A pass would support cross-sequence, cross-location retrieval only;
it would not establish answer-generation quality, production load, semantic
time conversion in all locales or Goal 5 completion.
