# Eager Immutable Core V42: Technical Preflight Contract

2026-10-03 local. A new test-only wrapper; no production mutation/adoption.
V37's normal loaded trials all failed original100/250ms serving/freshness
screens. Its warm nomination path was eligible in only41/1032requests.
Source inspection confirms that V36 builds an immutable core on the FIRST
subsequent Search rather than at durable publication. This is one upstream
read-side cost, not proof it causes the whole loaded tail. Native persistence,
archive-owner contention and backlog remain competing explanations.

Hypothesis: preparing the read core after existing full durability reduces
first-recall owner acquisition. Falsifier: a held authority owner still prevents
a fresh warm request from reaching the existing afterProjection boundary, or
publication reuses a gapped/native-uncommitted head. Full acknowledgement still
must await the original journal durability. No throughput/freshness benefit is
inferred from this barrier alone. New code does not replace any native write,
receipt validation, source authenticity or archive ancestry predicate.

Eager wrapper delegates the existing append/outcome/journal operation FIRST.
Only after success it locks authority owner, checks context, poison, witness
head, native head and published head, deep-copies the original read-only
projection, and atomically publishes it. Nil initial certificate is not cached.
Reuse same-head cores only when the read dependencies agree by construction;
ordinary journal-only changes do not alter Sources/Log/Certificate after the
first binding. No borrowed maps. Errors in optional preparation are recorded
but do not rewrite a successful durable mutation as an uncommitted failure;
the original warm head checks force ordinary cold/fail-closed behavior.

Core construction is NOT eliminated: it moves to the writer/journal return
path and adds an owner acquisition. Constructor copies, owner waits, mutation
return time and full offered-load cost must be measured before any loaded
adoption. Cold/disabled path remains the V37 control. Single request bound
native nomination, as-of, frontier, horizon, original forecast, provenance,
packing, journal bytes, cancellation, archive durability and reopen semantics
remain unchanged. This is a publication-policy experiment, not a claim that
read copies were the sole cause or that freshness was achieved.

Publication attempts, actual publication-owner acquisitions and elapsed
preparation time are separately counted, including failures and same-head
reuse checks. They are NOT hidden in old read-side owner/build-hit counters.
An eagerly built head need not be consumed before a later mutation; any loaded
auditor must count unused preparations too rather than reuse V36's per-request
build+hit=129 invariant. Physical copying must include every eager build. These
are measurement requirements, not permission to weaken source/latency gates.

Preflight: fresh-head reachability while owner is held with no early packet;
future/visible mutation source and law preservation; interrupted/unwitnessed
head rejection; context cancellation/foreign stale authority; deferred full
ack remains load-bearing. Tests use isolated temporary LibraVDB+SQLite fixture,
public dense data, no production or private corpus. Gated by explicit env.
Prospective loaded original128Recalls/128writes/16outcomes factorial, external
audit and100/250ms gates are STILL REQUIRED before Goal6 benefit or adoption.

V41 normal collection runs separately. Do not run another CPU-loaded benchmark
during its timed collector. Technical V42 tests may run while the independent
offline V41 audit is active; that audit's rerun timings are not study outcomes.
Record this execution distinction. V42 technical tests are not part of V41's
frozen sources and cannot be used to alter V41 candidates or gates.
