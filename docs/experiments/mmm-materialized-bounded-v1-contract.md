# Bounded materialized read comparison

Repeat materialized-source-v1:3 rotated trials,batch50/200,32 appends/cell,
all original records checked after reopen. Candidate reads now use the original
bounded SQL projection and readServiceAdmission validator. Key validation and
canonical lookup-key construction are inside read timing. Original code unchanged.
Record all samples and source snapshots. No timing gates changed or samples removed.

Before collection, test absent/canceled/invalid keys, source/owner mismatch,
wrong payload type, oversized identity and payload; verify oversized columns
are NULL before Go materialization. Both reads must reject malformed originals.
Write preparation remains counted. This tests read-contract parity on the
admissible fixture domain, not legacy writer equivalence or whole-goal success.
No feedback/migration/crash/concurrency extension, production use, or push.
