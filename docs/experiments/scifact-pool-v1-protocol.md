# Document-Level Public Retrieval Contract

2026-10-04. Isolated Goal 5 preparation and native-contract pilot, frozen before
native import or measurement. All seven WHOLE goals remain open. Existing source,
labels, split plan, failure gates and confirmation partitions remain unchanged.

## Invariant And Hypothesis

Multiple title/abstract spans must not count as independent relevance outcomes
or monopolize a packed document set. This is an accounting invariant, NOT a
diagnosis of the earlier quality failures. Other credible causes are weak
nomination, weak rank features and weak/absent feedback. No production patch.
The falsifier is a missing source tail, duplicate source, out-of-frontier row,
future-only source, changed/unknown identity or pre-ranking truncation accepted
by the new boundary. No special treatment of convenient benchmark examples.

Create one document-level index row from ALL bounded source frames, in source
title/abstract order, separated by two newlines. Text is their actual FrameText,
not raw Content, source ID, gold labels, split role or fabricated conversation.
This is source-description pooling, NOT posterior pooling or truth authority.
The complete abstract is retained in the framed input; this public document
instantiation does not claim compression of scientific content or general
episodic event extraction. It can be larger than a conversational EventFrame.
Maximum pooled embedding input is 16KiB; refuse the whole import if any document
exceeds it, do not silently cut the tail or increase the cap after measuring.

Each pool identity binds the frame contract, first span identity, and source
digest. Every record links ALL span IDs outside scored text. All metadata is
uniform except source/availability/identity fields; authored=false, authority=0,
salience=0 and access_count=0 are NOT claims of scientific reliability. Import
clocks denote local availability, not publication dates. The registry owns
envelopes and byte slices. Search exclusion lists are availability-derived.
Strict response binding rejects unknown IDs, duplicate source rows, altered
text or source metadata, nonfinite scores and future-only returned sources.
Scores are native retrieval signals, NOT probability laws.

Freeze a COMMON 50-200 document frontier for quality work. RankCandidates uses
K1=K2=the entire nominated frontier; local correction then packing may reduce
it to top10. A shortened second-pass response is a protocol failure, not
permission to compare against an easier reduced frontier. Reordering is allowed,
new rows and duplicates are not. No labels enter the serving registry. Original
fit/calibration/confirmation partitions are unused by prediction at this stage.

## Native Pilot Boundary

Use the installed public daemon executable only in an explicitly owned child
process, with research HOME/config/data/agent directories and owned Unix socket.
Never connect to auto-detected or production endpoints; never use auth-secret
discovery, launchctl, provisioning, updates or installed config writes. Supplied
model/runtime files are immutable read-only inputs, hashed before and after.
Pass a minimal environment: no inherited LIBRAVDB settings, credentials or
remote LLM URLs. Local ONNX CPU embedding uses the installed nomic model;
no production LLM or shared embedding endpoint. Concurrency is one RPC at a time.
If config paths are ignored or the server binds a different endpoint, fail and
stop only the owned child. No retry with guessed defaults or hidden fallback.

The prior `serve -h` probe unexpectedly initialized instead of printing help,
then failed before finding ONNX. Preserve this as a CLI assumption failure;
do not fix it by changing global configuration. Follow the documented YAML
and explicit asset/transport environment settings for the owned pilot.

Pilot imports the first eight original corpus records and one future-only copy.
Use two source-title self-probes and a future-source exclusion control. Relevance
labels, official queries and confirmation outcomes are NOT read. Self-title
retrieval is a technical positive control, never meaningful quality validation.
Measure startup/import/search/ranking wall time, actual returned cardinality,
client request/response snapshots and owned process exit. Sample count is too
small for p99, adoption, scientific accuracy, or latency-compliance claims.

## Checks Before Expanding

Complete public registry build, exact source coverage, ownership, cancellation,
metadata poisoning, invalid/future/duplicate responses, shortened ranking,
selection-before-packing, and equal-score tie behavior. Independent full-corpus
reconstruction without candidate imports; at least12 corruption rejections.
Race/vet, all actual source/test dependencies and immutable model/binary hashes.
Measure full registry build/response binding with caps50 and200 after correctness.
No hidden model acquisition/import cost; no allocation or payload number is RSS.
Then a separate frozen fit-only native full-corpus retrieval study, with untouched
confirmation held closed until design and calibration have frozen a candidate.

The old semantic shared-endpoint harness is not reused. Public source schema
and label limitations remain in `scifact-provenance-v3-results.md`.
