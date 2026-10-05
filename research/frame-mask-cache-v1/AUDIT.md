# Patch and measurement audit

Scope: isolated per-message mask reuse, not production promotion.

The frozen repaired extractor computes `unquoted(raw)` for every participating
field lookup. The candidate stores precisely that deterministic string in a
call-local `sourceText`. Raw text, field ordering and span comparisons are
unchanged. The mask retains byte length, so regex offsets still refer to raw
bytes; the earlier quoted-span rejection is neither removed nor relaxed.
Unprepared literal sources retain the original on-demand behavior. The separate
`ready` flag distinguishes a legitimately prepared empty mask from no preparation.

No mask is global, keyed by session, persisted or shared between calls. Strings
are immutable, and no prepared raw text is subsequently reassigned in this
candidate. If a future caller mutates the source struct's raw field, it must
construct a new source rather than reuse its old mask; this helper is private.
Preparation reads only the incoming envelope already used in raw Content. It
does not read the next packet, outcome labels, sessions, files or a database.

The five regex passes still depend on input length. Reuse saves repeated
mask scans and allocations, not an asymptotic constant-time extraction claim.
Eager assistant preparation can add work on early user matches. Timing criteria
therefore include early matches and retain every regression, not only favorable
quote-heavy cases.

Control is the complete repaired published frame source, compiled separately as
`framecontrol` and using the same model types. Differential tests compare entire
events and query strings; identity-aware comparisons include unresolved-reference
and provenance metadata. 2,048 fixed-seed input pairs exercise four entry points
(8,192 comparisons), with another 384 independent parallel turn comparisons.
These are functional checks on authored synthetic inputs, not scientific
confirmation or exhaustive proof over all strings. The inherited semantic
tests remain unchanged; their original source-fidelity test is exercised on the
control, where its source pins apply. Candidate identity instead comes from the
frozen patch and SHA-256 manifest. No failed semantic test is removed.

Matched functions run in the same binary with both arm orders and identical
inputs. Each 45-cell profile has six samples per arm. Setup is outside `b.Loop`;
Go reports component time, cumulative allocation per operation and allocation
count, not peak RSS or loaded service latency. Independent-call race testing is
not a loaded queue or durable freshness test. Public synthetic timing fixtures
do not establish the workload weights of real agent traffic.

No P1/P2 semantic defect was found in this scoped local review. This is a parent
review, not a newly commissioned independent reviewer. The performance verdict
must come from all frozen cells after the live benchmark handle completes.
