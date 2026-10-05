# V39 Preflight Repair

2026-10-03. Original prospective attempt stops BEFORE benchmarks or any normal
outcomes: full package race exit1,234.935s, `corruption 10 accepted`. Its test
assigned immediate to arm5, which is ALREADY immediate. This is a confirmed
no-op technical fixture, not evidence of an auditor accepting changed data,
learner failure or a quality result. All other package tests passed. Failed
log and run/freeze/failure JSON retained in `research/continuing-v39`.

Original auditor source archived verbatim as
`research/continuing-v39/preserved-source/continuing_audit_test.go`; its hash
must match the ORIGINAL freeze. Change only that test target to fixed150,
and assert every technical mutation actually changes the fixture. No model,
collector, audit validation, protocol, seed, threshold or outcome modification.
Parameterize only the benchmark artifact path for the separate output directory;
record that path in the runner command metadata. Fallback remains the original
path, and the same allocation parser/threshold and benchmark hash are enforced.
Re-freeze corrected source and run full race, vet, benchmarks, BOTH new normal
splits and frozen audit in exclusive `research/continuing-v39-recovery`.
Original source freeze is intentionally NOT a live-source match for this one
auditor file; the archived bytes retain its exact failed implementation.

Local reusable lesson: assert a negative-control mutation is non-identity
BEFORE asking whether the validator rejects it. No durable global policy or
instruction change. No normal seed has been consumed by the failed preflight.
