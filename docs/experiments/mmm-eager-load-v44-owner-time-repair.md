# V44 owner-wait trace repair

Confirmed follow-up, 2026-10-04: incremental audit of the original complete load
trace fails `reused core not yet available`. The checker required core build time
to precede preparation `Begin`, but `Begin` precedes acquiring the authority
owner lock. Another caller can publish the core during this wait. The actual
reuse decision is made under the lock, after acquisition. Original sources,
trace and both failed auditor invocations are preserved; they are not recast as
successful verification.

Distinguish possible premature reuse, corrupt timestamps, and a checker imposing
the wrong lifecycle boundary. The source acquires the lock before inspecting
the core, while the old checker compares against pre-lock entry. This proves the
boundary mismatch. The old trace has no lock-acquisition timestamp; do not invent
one or claim it proves the stronger corrected check. Run a prospective repeat.

The isolated V44 trace now records `OwnerAt` immediately after acquisition.
Require `Begin <= OwnerAt <= End`, zero owner time without acquisition, publisher
build after acquisition, and reused core built no later than acquisition. Four
temporal positives include publication during owner wait; five negatives cover
missing/reversed/hidden owner time and genuinely future cores. Full scientific
self-test retains the previous 49 corruptions and adds three owner-time corruptions.
Existing native race controls validate owner-time and canceled-no-owner traces.

No publication predicate, full-durability boundary, workload, forecast law or
100/250 ms gate changes. A new preflight/source freeze and exclusive repeat root
separate this instrumentation revision from the original incomplete audit.
