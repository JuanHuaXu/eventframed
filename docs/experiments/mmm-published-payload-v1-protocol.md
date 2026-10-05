# Published-LSN EventFrame payload projection v1: frozen backend probe

The current certified published reader returns IDs only. Full Recall needs
the EventFrame body and retrieval score from the same read view. In a
private LibraVDB fixture, try an exact-LSN SQL projection of `event_json`,
`corpus_text`, `raw_content`, and vector distance on the existing
availability-only schema; report whether the query binds and returns
decodable metadata, without assuming success.

Repeat with a separate private tenant collection that explicitly declares
those three payload fields plus the sortable availability key as metadata
string columns. Seed three EventFrames: two eligible at as-of `.120Z` and
one future at `.125Z`. Publish a verified marker, then append one eligible
event at `.110Z` through the incremental gate. A query at the old certified
LSN must return exactly the original two bodies; at the new certified LSN
it must return those two plus the appended body, never the future body.
Decode each row through the same EventFrame payload validator as Store
Search and verify raw content, identity, sort key, and finite score.

Run close/reopen and repeat the new-LSN query. A SQL binder failure,
missing projected field, body mismatch, future leak, or old-LSN admission
of the newly appended event fails the declared-schema candidate. This is
only a backend read-view feasibility study: graph/certificate/posterior
state, frontier journal routing, external retrievers, full Service Recall,
loaded latency and production migration are out of scope.
