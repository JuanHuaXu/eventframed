# Admission phase diagnostic v65

Frozen before measurement. Measure source resolution, durable preflight, actual
forecast/record staging plus encoding, and atomic append separately. Optional
private timing pointer, nil in all constructors; no default clock calls. Install
only before exclusive owner calls in tests; no serving or feedback authority.

Reuse v56's resolved-source isolated fixture:50/200 records per batch, cold or
64-label trained histories with per-label processing barriers,3 trials,32
fresh/retry/verified-discard cycles:12 cells. Verify original/retry equality,
terminal records, lifecycle counts and reopen; hash complete forecast histories.
Test timing-enabled versus disabled original parity and duration containment.
Run researchmemory race suite and vet before non-race measurement.

This is phase attribution, not an optimization success screen. Source and
durable preflight are upper bounds on potentially movable work, not assertions
of purity: they read owner-protected state and allocate planned identities.
Stage mutates actual worker state and is not reusable preview data. Append
includes SQLite validation/transaction/commit, not just fsync. Source/durable
mutex waiting and return/clock overhead are outside named phase sums. No claim
that those waits vanish in concurrent service use. Report fresh and retry
separately; measurements perturb timings. No default, durability or lock change.
