# USGS Tadine cross-sequence retrieval transfer

Frozen before source snapshot construction or any retrieval call. This is a
new public retrieval task population, not a generated-answer experiment.

## Source and fixture

Fetch the official [USGS catalog API](https://earthquake.usgs.gov/fdsnws/event/1/)
once with `format=geojson`, `starttime=2026-09-23`,
`endtime=2026-10-01`, `minmagnitude=4`, `orderby=time-asc`, and
`limit=2000`. Preserve the raw response and its SHA-256. Require at least 24
events whose `place` contains `Tadine, New Caledonia`; sort these by event time
and USGS ID, then take the first 24. Do not filter on magnitude or retrieval
behavior after that point. Magnitude and source URL come only from each
selected official feature. All questions refer to the source snapshot's
reported value, not the event-time preliminary estimate.

For each selected event, construct three questions without the magnitude:

- `literal`: exact UTC ISO timestamp.
- `paraphrase`: natural-language UTC date and time.
- `offset`: equivalent timestamp explicitly labeled `UTC+11`, with the offset
  explained in the question. This tests time normalization without claiming
  the civil time zone of a place.

The answer label is the selected USGS event ID and its reported magnitude.
Source IDs and labels are absent from model-facing query text. The corpus is
the 24 source-derived event sentences; there are no invented earthquake facts.
All 72 query cases are untouched by prior retrieval runs. The first 12 events
form a design stratum and the next 12 a confirmation stratum, but the frozen
pass gate applies to all 72 because the ranking policy is unchanged and no
fit or tuning is allowed between strata. The same sequence limits inference.

## Arms and gate

Use the unchanged `focus` and task-plus-lexical `priority` service arms from
the July 2019 USGS test, the same Nomic embedding digest, a fresh in-memory
service per case, recall 50, pack 10, diversity enabled, and numeric ranker
passthrough. Run both arms on all 72 questions with the same snapshot and
capture packet order, full frontier, numeric scores, scored laws, journal
agreement and Recall duration before opening the oracle.

Finite transfer passes only if all conditions hold:

1. Priority gains at least eight paired top-1 answers over focus, including at
   least one gain in each of the three wordings, with zero paired top-1 losses.
2. Priority has zero paired target-survival losses at pack 10; each wording's
   aggregate survival is no worse than focus.
3. Each arm's frontier has exactly the same 24 source records; the numeric
   ranker is unchanged; full-frontier original scores and forecast laws match
   across arms; packet explanations match the durable journal.
4. The sequential isolated Recall nearest-rank p95 is below 100 ms for both
   arms. This is not a loaded or agent-service tail-latency claim.

Report per-wording top-1 and survival, paired gains/losses, cluster-level
counts (24 events, not 72 independent samples), all invariant failures,
latency and source/code hashes. No gate changes after seeing results. A pass
would show cross-sequence retrieval transfer only. It cannot establish
generated answer correctness, calibration, causal truth, production serving,
or Goal 5 completion.
