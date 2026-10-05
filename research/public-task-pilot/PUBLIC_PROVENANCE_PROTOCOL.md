# Public-fact provenance boundary diagnostic

Use the13 existing structured NASA facts in facts.json. No new facts, private
sessions, model generation, source fetches or hypothetical sensor accuracy.
Construct explicit historical-fact5W1H records: mission=Who, event=What,
target=Where, semantic date=When, archive retrieval=How. This is a direct test
of internal/epistemic, not CaptureTurn extraction, signing or serving end-to-end.

For each fact compare four repeat-retrieval transformations. Content and semantic
time stay identical; only retrieval metadata changes. Local source-event fixture
IDs are deterministic hashes of recorded URLs, representing a stable upstream
origin mapping rather than actual production database identities.

1. Stable source lineage, new event/run/session/tool-call IDs.
2. No source-event lineage, different tool-call IDs, same producer.
3. Same stable source-event lineage, different ingestion producer.
4. Source-event IDs changed with the fetch attempt, same producer.

The factual repetition is known from test construction. Record both exact group
equality and Correlated(...,.8), without claiming those results authenticate
truth or independent sources. Different keys mean the grouping component did
not recognize correlation; they do not mean another layer granted confidence.
Report all outcomes, including uncovered transformations. Do not call a core
boundary diagnostic an end-to-end security exploit or mutate the implementation.

These are52 paired transformations of13 facts across3 mission clusters and
2 source pages, not52 independent empirical samples. This corpus is consumed.
Hash inputs and implementation, replay exact output, and retain absence of
stable origins as an explicit observation condition. No PII or generated text.
