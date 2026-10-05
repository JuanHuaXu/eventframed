# Offline origin-binding adapter

Research-only adapter; no runtime or authority changes. The trusted input is a
read-only host registry keyed by tenant/event identity. Each binding contains a
stable origin and SHA256 of the exact serialized Event. A registry is not built
from untrusted event attributes, tool-call names or claimed URLs during Apply.
Registry creation is a trust prerequisite, NOT authentication implemented here.

Apply verifies the full serialized payload digest, permits only explicit external
observations, and changes only Candidate.EvidenceGroupKey. Its grouping key binds
tenant, origin, kind and all six exact semantic field values including When.
No punctuation-erasing semantic equivalence or cross-tenant grouping. Missing,
changed or unsupported bindings preserve the existing candidate unchanged.
The registry caps256 entries and rejects duplicate identities. This is offline
research code with full-event hashing, not a bounded serving-latency claim.

Use the same13 NASA facts and four metadata variants as the preceding boundary
test. The fixture controller supplies both binding receipts using the known
original source mapping; this measures transport/packing behavior given a
correct registry, not the real-world reliability of constructing that registry.
Compare the prior no-adapter outputs without modifying them.

Negative checks: missing receipt, changed payload, wrong tenant, unsupported
kind, duplicate/empty registry identities, independent origins, changed relation
punctuation, distinct semantic occurrence time, and certified AP split buckets.
Retain original events and confidence laws. Verify public-output replay and
focused race tests. No full ingest/extraction/authentication/learning claim.
