# V4 Patch Reasoning And Scope

Confirmed upstream cause: SearchWithEf always called the physical base. Its
ordinary HNSW returns ErrEmptyIndex at zero size. The wrapper returned this error
before merging the live overlay. Collection treats that error as an empty shard,
so the sink sees zero results although storage and overlay contain records.

Competing explanations were missing durable rows, availability-filter mismatch
and empty-base early return. New 0/1/2/64-entry standalone witnesses all fail with
the same base error, with no storage or availability machinery involved. The two
actual service/store failures provide independent sink evidence. Repair only
skips the base call at physical size zero; error injection at nonzero size proves
other base failures remain visible. The cold witness includes filtered search,
abort visibility, deletion and snapshot/reopen. All V3 ownership/profile/atomic
and key-alignment checks remain required.

Happy path: mutation reserve -> immutable candidate prepare -> durable WAL ->
pointer publication. Readers see only the last committed generation. New base
empty state is not logical collection empty state. Search must merge every live
overlay candidate even when base has no records. No arbitrary error suppression,
evidence dropping, score adjustment or retrospective threshold change is allowed.

Cost: one constant-time base Size check and the same O(cap*D) overlay scan.
Compaction copies O(ND) canonical payload and reconstructs the entire HNSW,
including sort/insertion graph work; it is NOT an O(N) time claim. This does not
establish bounded compaction latency or total retained memory. V3's142.531ms maximum survives
as historical failure, not an excluded outlier.

Related-fix check: upstream HEAD unchanged at18515d4ce0620b24fd2512b0f9023b8b1f4d8355;
all currently listed ten issues/PRs contain no wrapper-specific cold-overlay fix.
The wrapper is our isolated research prototype, not upstream installed code.

The first V4 ordinary selector used guessed transaction/key test names. Its ten
core tests plus WAL test passed, but it did not run the intended sibling checks.
The explicit preflight-full ordinary/race commands correct that coverage gap;
retain both receipts, never count nonexistent selector matches as passing cases.
No durable instructions, module cache, production or private datasets edited.
