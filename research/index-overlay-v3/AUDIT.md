# V3 Audit And Boundary

Inherited negative results are authoritative: V1 checkpoint-lock and key-arena
failures, V2 complete geometry study and service Stats panic. Do not overwrite
old sources or report V2's partial service run as completed integration.

## Confirmed Findings And Repairs

- Checkpoint lock cycle: separate writer reservation from reader generation
  locks. Old snapshots/queries remain readable before Commit; prepared state
  publishes only after durable WAL success. Earlier passing core tests did not
  exercise the storage callback; the dedicated failing witness did.
- Key-arena capacity: align each ID's own allocation to8bytes, including rename
  source IDs.12case sibling regression has8unaligned failures and4aligned passes
  before repair; all12pass after. Same common repair in control/candidate.
- Profile adapter: preserve standard typed base RawVectorStoreProfile fields
  rather than replacing them with overlay-only diagnostics. New profile-shape
  regression fails before repair; service panic is the independent sink witness.
- Ownership: provider-free base HNSW returns borrowed vectors. Collection.Search
  assumes ownership, and normalized index payload is not always canonical record
  payload. Two pre-patch regressions fail: mutation changes later search, and
  provider-backed results expose index vectors. Standalone output now clones
  retained canonical index vectors; storage-backed output omits them for ordinary
  authoritative hydration. Graph never reads mutable provider rows.

Initial V3 protocol's one-new-change wording is superseded explicitly by the
pre-timing OWNERSHIP_PROTOCOL.md. Both changes and tests are in frozen source
pins before timing. No dimension, cap, candidate count, gate or criterion altered.
Initial profile extras counted logical live vectors, not retained shadow copies;
canonical_vector_bytes now also includes retained base records plus live overlay
copies. This is a payload diagnostic, not heap/peak-RSS or total memory bound.

## Not Established

The base is immutable between synchronous compactions; future-only compaction
can still change navigation among eligible old events. Exact overlay processing
does not prove counterfactual independence of the ANN base. Full merged public
packet/frontier/law audit must preserve any such failure separately from quality.
Ordinary/race mechanics tests do not demonstrate real agent usefulness, valid
Anti-Pigeon external error control, Bayesian scientific convergence or learning.
No background compactor, sustained queue/RSS, kill recovery, cross-epoch native
contracts or production deployment is claimed. All seven original goals OPEN.

## Setup History

V1 read-only file copy and missing sibling lexer were local setup errors; ML=0
was an invalid test fixture. V2 broad preflight regex selected an unconfigured
cost test. Raw outputs distinguish these from actual capability failures. The
local reusable lesson is to inspect copied module replacements/permissions and
use explicit preflight names. No live/global instruction or installed cache edit.
