# Calendar adapter: actual pre-packing integration

Completed60 isolated CaptureTurn/Recall runs, four admitted records and packK=1.
The Go adapter runs through the existing CandidateRanker contract BEFORE packing;
no service.go/research_rank.go edits or production configuration changes.

| Set | Focus control | Calendar adapter |
| --- | --- | --- |
| W3C written | 4/8 | 7/8 |
| ESA written | 6/8 | 8/8 |
| ESA ISO | 4/8 | 7/8 |

PASS the integration screen: all offline normalized top1 outcomes reproduced,
all control-correct cases preserved, and eight missed answers promoted into the
single final slot. Their actual input positions were2-4, so this is not sorting
an already successful final pack. Sets share underlying facts and are consumed;
the eight rescues are not eight independent domains or fresh scientific trials.

All proper forecast laws for ALL four frontier events are exactly equal across
control/adapter for every question (maximum absolute difference0). The verifier
uses full Bayesian report forecasts, not merely the packed candidate. It also
recomputes the partition independently with the strict JS implementation on
the actual captured what fields. Ranker trace preserves candidate text/metadata.

Original query is request-context-bound and checked against tenant and canonical
focused query. Dates come from real Event.FrameText what fields, not corpus-ID
lookup, forecast labels or ingestion timestamp metadata. Repeated dates in
when/why do not create false ambiguity because those fields are not independently
parsed. Missing/ambiguous what evidence stays unknown. The adapter is stateless.

Mixed partitions use ordinal ranking scores to survive subsequent score sorting.
Those scores are not calibrated probabilities or packet confidence. Proper laws
remain separate. Empty caches mean later learned-delta interference was NOT
tested. Unknown and all-contradicted inputs retain existing order; the adapter
does not implement an abstention or a durable contradiction certificate.

## Verification

`go test -race ./internal/researchcalendar` passed: valid-date Gregorian-cycle
checks, invalid-date/token cases, original-query and what-field binding, input
immutability, unknown/all-contradicted preservation, tenant/query mismatch,
cancelled requests, duplicate IDs and invalid score/frontier rejection.

`node research/public-task-pilot/check-calendar-contract.mjs` passed: source
hashes,60 cases, four-record before/after frontier, one-record pack, identity and
metadata preservation, full-frontier law equality, JS/Go partition equivalence,
offline outcome equivalence and verified pre-pack rescue ranks.

Artifacts: contract-results.json in [W3C](w3c-focus-v1/contract-results.json),
[ESA written](esa-date-v1/contract-results.json), [ESA ISO](esa-iso-v1/contract-results.json).

This is an isolated memory-backend integration pilot using local embeddings, not
an OpenClaw/LibraVDB production test. No LLM generation, feedback learning,
latency benchmark, calibration improvement or goal completion is claimed.
The W3C publication-status miss and ISO later-than miss remain. Next required:
snapshot/retry and concurrent request-binding tests, learned-delta interactions,
durable explanations and measured serving performance. Existing service-level
checks were retained, but their existence alone does not prove those scenarios.
