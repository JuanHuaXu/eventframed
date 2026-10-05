# Packed source-index candidate: rejected

Before performance testing, we examined replacing five scalar JSON extraction
index expressions with two multiple-path JSON extraction expressions. This
would retain the original row representation and fewer extraction calls, but
its output is JSON text rather than a tuple of decoded SQL strings.

## Falsifier

The [six-case artifact](mmm-packed-index-semantics-v1.json) compares equality
under scalar extraction with equality of packed array text. Unicode escapes,
escaped slashes, and equivalent newline escapes each produce equal decoded
strings but unequal packed keys. Ordinary identical text and changed whitespace
match; a deliberate delimiter-boundary pair remains distinct.

For example, `alpha` and `\u0061lpha` name the same source journal but produce
different packed keys. The separate `TestPackedIndexDuplicateWitness` installs
the proposed index only in a temporary database and proves both duplicate-source
records commit. The original scalar index rejects the same batch atomically,
leaving zero rows. This is an observed constraint bypass in the rejected
candidate, not a defect in the original index.

## Verification and decision

The opt-in semantics collector passed under the race detector (1.294s), and
its output matches the original artifact byte-for-byte. A passing collector
means observations were collected correctly, NOT that the candidate passed.
The duplicate witness also passed under the race detector; ledger vet passed.
Raw SHA-256:
`6b6e81ae91f228c7b5b39cb2bd014c07b97a2ae80935a5daf78039a70d0f6fcf`.

No performance experiment was run because the necessary equivalence gate failed.
No production schema, admission code, or guard changed. Keep this candidate as
a negative control. A later typed/materialized representation must derive keys
from decoded strings, preserve field boundaries and source-to-payload agreement,
and handle alternate valid JSON encodings without creating distinct identities.
Canonical Go output alone is not a sufficient general API invariant: append
accepts valid JSON bytes, and retries intentionally compare original exact bytes.

All seven research goals remain open. Next investigate a genuinely canonical
representation with explicit validation cost rather than this lexical shortcut.
