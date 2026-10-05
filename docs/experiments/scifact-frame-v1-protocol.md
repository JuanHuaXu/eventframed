# Source-preserving public document frames

2026-10-04. Goal 5 dependency, isolated research, no adoption. Frozen before
conversion or cost measurement. Existing SciFact source and family partitions
remain unchanged. No confirmation predictions, model fitting or model calls.

## Hypothesis And Boundary

The old public semantic harness used paired conversation capture, whose first
statement and 320-byte field bounds can omit scientific source content. That
is an observed harness mismatch, not proof it caused earlier ranking failures.
Other explanations include weak rank features, bad nomination, and absent
outcome feedback. No production parser or old experiment is patched.

The new adapter accepts ONLY public document `id`, `title`, `text` records.
Unknown or duplicate JSON fields, malformed UTF-8 or lone Unicode surrogates,
empty fields, duplicate IDs, trailing
JSON, and explicit caps reject. It cannot accept labels, annotations, authors,
query IDs or split roles through this contract. Scope is source descriptions,
NOT fabricated conversations, scientific truth, causal effects, or certified
5W1H extraction. This is a sparse source-assertion EventFrame instantiation.

One title span and as many abstract spans as needed become independent events.
Only `what` and the source-section descriptor `where` are populated. Unknown
who/when/why/how stay empty. The Event envelope uses an explicit import instant
for occurred/observed/available, never an inferred publication date; semantic
`when` stays empty. Observed/confidence 1 means exact source transcription, NOT
probability that a scientific assertion is true. Content retains original title
and abstract as untrusted metadata. It is never part of FrameText scoring.

Canonical source text is Unicode-whitespace collapse with `strings.Fields` and
one ASCII separator. Partition canonical title and abstract into contiguous,
non-overlapping UTF-8 spans with at most 2048 bytes. Prefer the final ASCII
space within a bounded prefix; exceptionally long tokens can split. Every
canonical byte is covered exactly once. Original bytes remain in metadata.
Section/offset/document/digest identity is outside semantic score text.
IDs are SHA256-bound to contract, tenant, session, import instant, document ID,
raw title/abstract digest, section, and byte interval. Reimport with different
availability must not silently reuse an old idempotency key.

Input: at most 16 MiB JSON, 8192 documents, 128 KiB combined raw title/abstract
per document, 128-byte IDs, 64 frames/document. Invalid/canceled conversion
publishes no partial output. Explicit availability-gated selection returns owned
envelopes and refuses a canceled context. EventFrame field caps are inherited
from the existing model. These are experiment bounds, not billion-record/RSS
guarantees. Full-source metadata is repeated per span for Observe compatibility;
measure the cost rather than hiding it as a constant.

## Required Checks

1. Decode rejection tests; ASCII/multibyte/long-token exact coverage, every field
   bound, deterministic identity, import-time/tenant/source sensitivity.
2. Availability, cancellation, ownership, and metadata-only poisoning controls.
   No invented 5W1H fields. Reversal of corpus order preserves per-document
   identity/frames; no ordinal or label-based identity.
3. All 5183 documents through the actual adapter. Independent Node reconstruction
   checks every canonical span and EventFrame, source/content hashes, ordering,
   IDs, fields, envelopes and exact inventory. At least 12 corrupted outputs
   must reject; the auditor must not import the Go adapter.
4. Real Observe/store/Recall integration with recording deterministic hash
   embedder, no network service, and future-only documents. This is contract
   wiring, NOT a semantic retrieval or answer-benefit experiment.
5. Race/vet and serial benchmarks after correctness. Count ALL document frames,
   canonical bytes, repeated metadata and actual JSON payload. Report bounded
   cache feasibility at dimension 768 including document keys and vectors,
   plus 351 fit/180 calibration/300 confirmation query entry overhead. Exact
   query-role keys are not yet instantiated; conservative upper envelope uses
   16KiB per query key. Report cap failures, do not increase caps after observing.

Freeze source hashes/copies, exact commands, raw logs/artifact hashes and usage.
No quality/adoption gate is evaluated here. Subsequent retrieval must freeze a
common document-level frontier (50-200), deduplicate spans before packing and
feedback, specify source pooling and measure acquisition/import cost. More
spans MUST NOT count as independent confirmations or extra relevant documents.
The unused outcome-labeled fit/calibration/confirmation partitions remain unused
by prediction. All seven whole goals remain OPEN.

Primary source schema: [SciFact](https://github.com/allenai/scifact/blob/master/doc/data.md).
Citation relevance differs from annotated support/contradiction. Dataset terms
remain separate; local conversion is not authorization to redistribute data.
