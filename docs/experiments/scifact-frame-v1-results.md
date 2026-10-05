# Source-Preserving Public Frames

2026-10-04. Goal 5 preparation advances; all seven WHOLE goals remain OPEN.
No predictions on fit/calibration/confirmation, no fitting, no semantic model
call, no production/private-corpus/whitepaper change, no commit or push.

## Executed Evidence

The isolated [adapter](../../internal/researchpublicframe/document.go) now
converts the complete, unchanged official public corpus into sparse
source-assertion EventFrames, not fabricated user/assistant turns. The full
source is retained as untrusted metadata; only bounded source spans and source
section enter canonical FrameText. Unknown who/when/why/how remain empty.
Import clocks are not publication dates. Confidence 1 denotes transcription
fidelity, not scientific truth. This is NOT a certified semantic event extractor.

| Quantity | Independently reconstructed result |
| --- | ---: |
| Documents | 5,183 |
| Bounded source frames | 10,869 |
| Maximum frames/document | 6 |
| Maximum semantic field bytes | 2,048 |
| Complete canonical title/abstract bytes | 7,772,270 |
| Repeated original-source metadata bytes | 16,997,477 |
| Actual generated JSON bytes | 46,215,235 |
| Document embedding-cache key/vector payload at dimension 768 | 41,982,699 |
| Document frames plus 831 partition-query entries | 11,700 |
| Conservative cache payload including 16KiB/query key envelope | 58,152,297 |

Thus this corpus fits the UNCHANGED 12,000-entry/64MiB key/vector cache caps.
This does not bound map overhead, runtime allocation, RSS, persistence, or future
corpus growth. Each source title happens to fit one span; longer titles follow
the same bounded partition as abstracts. Canonical whitespace collapse is
explicit; original bytes remain available. Very long tokens can split across
spans, which can affect future retrieval even though source coverage is complete.

## Correctness And Repairs

Six actual test roots execute three times under race detection, with no skips:
strict decoding/bounds, strict JSON identity, exact Unicode/long-token coverage,
availability/ownership/cancellation, real Observe/store/Recall wiring, and all
5,183 documents. Vet passes. The wiring uses a recording deterministic HASH
embedder and memory store, never an external endpoint. Four actual source-frame
embeddings and a two-candidate as-of packet confirm that full source tails reach
the embedding sink and future-only documents do not reach that packet. This is
contract integration, NOT native LibraVDB or semantic-answer evidence.

The independent Node auditor imports no Go adapter. It reconstructs every
source span, envelope, field, identity, digest, clock, inventory and byte
coverage from the original public corpus; 20 deliberate corruptions reject.
IDs/digests/labels/split roles do not enter score text. Reversing corpus order
preserves per-source identities. Metadata-only injections cannot alter FrameText;
this does NOT prove prompt-injection resistance of an agent consuming metadata.

The bug hunt reproduced three actual preflight failures: duplicate JSON identity
fields and two lone Unicode surrogate cases were silently accepted by the
standard decoder. The new contract now rejects duplicate/unknown fields and
non-scalar escaped characters while preserving valid pairs and literal escaped
text. JSON syntax remains handled by the standard parser. Original failing
source copies and terminal log remain in
`research/public-task-pilot/scifact-frame-v1-strict-preflight`; no old parser
or historical evidence was rewritten. A total raw-source cap also applies when
the typed conversion is called directly rather than through Decode.

The first five-command run froze 21 runtime sources but omitted imports used
only by the external integration tests. It remains preserved. A supplemental
`go list -test -deps` closure freezes 168 sources including service/store/embed
dependencies and Go's generated testmain. All five commands rerun terminal PASS;
the corpus output is BIT-identical. Separate artifact readback verifies source
and copy hashes, terminal commands, exact test execution counts, both raw corpus
outputs, regression preservation and all benchmark lines. This is reproducible
verification, not a new independent dataset or quality confirmation.

## Measured Cost

Complete dependency repeat, Go1.27.1/darwin/arm64/Apple M4, serial after
correctness; three repetitions, three timed operations per sample:

| Offline operation | Sample mean time | Allocated bytes/op | Allocations/op |
| --- | ---: | ---: | ---: |
| Convert all 5,183 source records | 51.38-52.77ms | 86.09-86.10MB | 442,067-442,087 |
| Serialize all 10,869 frames | 39.79-42.58ms | 131.49-131.50MB | 86,974-87,103 |

These exclude source download/JSON decoding, embeddings, native index import,
model acquisition, persistence and serving. Repeated metadata and JSON encoding
are material costs, not hidden constants. Sample means are not request p99 or
confidence intervals; no loaded sub-100ms guarantee or performance adoption.
Do not derive faster live serving from these offline microbenchmarks.

## Next Required Experiment

Freeze source pooling and a common DOCUMENT-level nomination frontier of 50-200
before ranking comparisons. Deduplicate title/abstract spans before packing and
feedback; multiple spans are not independent confirmations or extra relevant
documents. Use the native LibraVDB contracts and an owned isolated semantic
embedding process/cache, not a shared production endpoint. Full acquisition,
import, correction, packing and timing must be counted. Unknown/unjudged
scientific usefulness must not be relabeled as confirmed false evidence.

The 351 fit/180 calibration/300 confirmation query partitions remain unused by
prediction. Freeze design and success criteria before any confirmation run;
citation recovery is the declared benchmark target, not truth verification or
causal authority. See [source provenance](scifact-provenance-v3-results.md) and
the [frozen framing protocol](scifact-frame-v1-protocol.md).

Authoritative execution artifacts are under
`research/public-task-pilot/scifact-frame-v1-closure`: `freeze.json`,
`commands.json`, `manifest.json`, `audit.json`, `artifact-audit.json`, all five
raw logs and `frames.json`. Original output SHA256 is
`a5837e4738186b271d9e1c03d8887347e06cba57b8b575f96aaae6d5b7b973fd`.
Latest weekly usage at verified readback: 21%.
