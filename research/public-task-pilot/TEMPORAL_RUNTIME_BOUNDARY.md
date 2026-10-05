# Temporal research integration boundary

Follow-up: [calendar contract pilot](CALENDAR_CONTRACT_RESULTS.md) now implements
a research pre-packing adapter through the EXISTING retrieval CandidateRanker
interface. This avoids edits to the feature-only research hook. The checklist
below records the original requirements; some are now verified in the isolated
empty-cache pilot, but durable explanations, concurrent snapshot/retry behavior,
learned-delta interactions and production qualification remain open.

Inspection after the normalization rescue:

- internal/service/service.go calls applyResearchRanking before truncating to
  recallK, journal publication, and packing.Select. This is the relevant full
  frontier boundary, not the already packed ContextPacket.
- internal/service/research_rank.go currently exposes only baseline probability,
  nine-bit features or sparse features, snapshot and AsOf. It does NOT expose
  original query, candidate text, normalized date or temporal relation.
- The callback validates cancellation, snapshot freshness, cardinality and finite
  scores. Any extension must retain those properties and the frontier cap200.
- Existing experiments hash these source files. Do not silently edit them and
  invalidate their historical artifact verification to claim integration.

Required research contract, still NOT implemented:

1. Retain original query independently of optional retrieval-query focus. Parse
   requested relation/anchor/exclusion from the original, never the rewrite.
2. Extract a typed calendar observation from the actual candidate event with
   source-field/provenance and ambiguity status. Do not recover text from fixture
   IDs or read oracle tables. Do not substitute event ingestion/availability time
   for a historical date mentioned in the event's content.
3. Apply contradiction prioritization on the full admitted frontier before
   packing, with explicit unknown/all-contradicted behavior. Existing scores are
   not recalibrated probabilities of temporal compatibility. Keep proper laws
   untouched unless separately modeled and validated.
4. Journal method/version and reasons sufficient to replay the result, preserve
   existing snapshot/as-of checks, and reject cross-tenant or stale evidence.
5. Test with packK smaller than recallK, including a support record initially
   outside packK. Verify preservation of law, unknown handling, cancellation,
   snapshot invalidation and the original query across the real boundary.

The offline JS partition is evidence for the component, not completion of this
contract. Simply sorting a returned four-record pack would miss the central
before-packing failure mode and is not an acceptable replacement for these tests.
