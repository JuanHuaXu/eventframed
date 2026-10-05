# Native Offline Resume Diagnostic

2026-10-04. Goal5 full-corpus preparation, NOT a native memory-leak fix or Goal6
loaded-serving rescue. Prior full fit diagnostic FAILED; retain it unchanged.
This attempt changes only offline import lifecycle and adds verified recovery.

Copy the closed failed research store and its embedding fingerprint, plus its
research auxiliary daemon store/fingerprint, into a fresh owned directory.
Never reuse the original files for writes. Hash each original before/after and
each new copy before starting; refuse incomplete copies. Do NOT copy sockets,
PID/lock files, credentials, private data or launchctl state. Inspect fingerprints
for remote paths: existing immutable installed public model paths only.

Keep SAME5183documents, source pool, Q8GGUF backend/assets, priority10, native
RPC contracts, fixed sampled2GiB ceiling, one concurrent call/one attempt,
30second RPC timeout,20minute full-client limit, fit351queries and query order.
Do not reduce the corpus or loosen the ceiling. This is a new resumed diagnostic,
not a claim of byte-equivalent temporal graph state or of continuous learning.
Any offline restart/model loading/readback cost is counted; native maintenance
is not disabled. All seven WHOLE goals remain OPEN. Confirmation stays closed.

The original trace supplies claimed acknowledged PREFIX identities only. It is
NOT durability proof. Before ANY new insert, use ListByMeta with the source SHA
for EVERY claimed prefix record. Require exactly one returned row and bind it
to the canonical source registry, including exact text, source/availability/
clock/span metadata. Reject unknown, duplicate, missing, nonfinite or changed
rows; record raw responses. Skip no record merely because an old RPC said OK.
Only when ALL prefix reads pass may the remaining original corpus be imported.
Prefix plan is frozen before native startup; no performance-based record choice.

Then run ALL351fit queries: SearchTextCollections K200; RankCandidates
K1=K2=the whole actual nomination set, no hidden prepacking. Empty/undersized
frontiers remain limitations. No label access in client or recovery. Only a
separate evaluator AFTER complete terminal predictions may read FIT citation
labels. Calibration180/confirmation300 untouched by prediction/evaluation.

The first attempt may still exceed the ceiling, fail to recover or panic. Retain
all such results. Source-level root cause for native memory growth/DirtyJournal
panic is still unknown; closed native code and production remain untouched.
Small native pilot and isolated pool successes do not establish full readiness.

Technical checks: correct typed protobuf request/response field numbers against
the original public plugin contract; local in-process RPC regression; missing,
duplicate, changed text/source metadata, nonfinite, canceled and wrong-prefix
controls; frozen complete test dependency closure; race/vet; immutable artifacts
and source hashes; separate independent actual store-response reconstruction.
Do not substitute RPC receipts, log messages or snapshots for sink-visible reads.
