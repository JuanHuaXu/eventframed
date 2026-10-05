# Canonical source-key preparation component

Test-only implementation: `internal/researchledger/canonical_source_test.go`.
No ledger schema or production writer uses this component yet.

## Invariant and scope

Derive one unambiguous key from decoded tenant, stream, contract, source journal,
and source event. Learner prediction IDs are deliberately excluded. Keep original
payload bytes unchanged, so a canonical source key never makes changed payloads
an exact retry. The future transaction must enforce both source uniqueness and
original byte equality; this component alone establishes neither durability nor
transactional uniqueness.

The parser requires exact critical field names, rejects duplicate decoded keys
in the top-level object and Binding object, and rejects case-folded aliases of
critical fields. This prevents first-wins/last-wins and Go struct/SQL path
disagreement. A tenant mismatch, non-string or missing source field, trailing
JSON, invalid UTF-8, empty or oversized identities also rejects.

Prototype restriction: identities containing NUL or the replacement character
U+FFFD reject, including a literal valid U+FFFD. This conservatively rejects
unpaired-surrogate decoding and is NOT full compatibility with every legacy
accepted string. No existing record is migrated or repaired. A deployment needs
an explicit compatibility policy or a stricter Unicode parser without that
overrestriction. Unknown noncritical fields are retained in original bytes and
their semantics are outside this identity-only component.

## Tests and audit

Equivalent Unicode, slash and field-name escapes produce identical source keys.
Changing each of the five source coordinates changes the key; changing only the
learner ID does not. Input bytes remain unchanged. Duplicate exact/escaped keys,
case aliases, malformed source fields, tenant mismatch and trailing data reject.
The alias check was added during the preintegration audit because Go struct
decoding is case-insensitive whereas SQL JSON paths are exact.

Final `go test -race ./internal/researchledger -run '^TestCanonicalSourceKey$'
-count=1` passed (1.285s); ledger vet passed. These are component checks, not a
proof over all JSON inputs or an end-to-end persistence test.

## Measured cost

Apple M4, darwin/arm64; same 1KiB-padding fixture used in the append diagnostic.
`go test ./internal/researchledger -run '^$' -bench '^BenchmarkCanonicalSourceKey$'
-benchmem -benchtime=500ms -count=3`:

```text
141446 iterations  4216 ns/op  7799 B/op  56 allocs/op
143246 iterations  4203 ns/op  7799 B/op  56 allocs/op
142953 iterations  4201 ns/op  7800 B/op  56 allocs/op
```

The .21-.23ms/50 and .94-.95ms/200 index row-cost differences from the earlier
diagnostic are about 4.3-4.8 microseconds per record. This parser costs roughly
4.2 microseconds before any replacement index insertion. These are separate
experiments, not a paired net-benefit estimate. Simply adding this preparation
on top of the original writer is unlikely to help and is not proposed.

## Next experiment

Measure the entire candidate transaction against the original, including key
preparation, persistent uniqueness, unchanged retries, and indexed reads. Do not
subtract isolated component means and declare a rescue. Consider computing
immutable source identity once at the ownership boundary only if its binding to
the eventual original can be demonstrated; do not move model-dependent validity
checks outside the guard. The high allocation count is a lead, not a reason to
remove ambiguity validation. All seven research goals remain open.
