# Envelope batch reuse: unsuccessful rescue

## Decision

Close this envelope design as an unsuccessful general storage rescue. Bounded
reuse helps relative to reading each envelope slice separately, but loses to
the existing indexed batch API in every measured cell. Do not integrate it or
claim the earlier storage-only speedup. Preserve artifacts and test-only code;
no production path was changed. All seven goals remain OPEN.

[Contract](mmm-envelope-batch-read-v1-contract.md),
[raw run](mmm-envelope-batch-read-v1.jsonl),
[summary](mmm-envelope-batch-read-v1-summary.json),
[replay](mmm-envelope-batch-read-v1-summary-replay.json).
Raw SHA-256:
`9ed2345164ba2dc6dbe86a595567e17cff3da4f087ddb7e07d3139216b663ec4`.
All 32 source hashes match captured/local text; summary replay is byte-identical.

## Contract and tests

The candidate reads normalized references in one transaction, loads each needed
envelope at most once into an 8MiB per-call cache, and validates each requested
original's key, source and JSON binding. Output copies have an independent 8MiB
cap. After exhausting cache capacity, bounded slice reads preserve otherwise
valid dispersed requests. There is no persistent cache, corpus-sized state,
dropping of requests or skipped source validation. Errors return no partial list.

Tests cover order, duplicates without aliasing, misses, cancellation, wrong-source
slices despite valid offsets, and output-cap enforcement with caching both on
and off. The full ledger race suite passes (4.629 s), as does vet. The experiment
passes in 4.94 s (5.150 s package). These do not prove full runtime lifecycle or
crash safety. The code remains test-only.

Three rotated trials compare indexed control, cached envelope and cache-disabled
envelope. Each stores 32 batches of 200 bound synthetic originals. Query sizes
50/200 and grouped/scattered layouts each have 32 calls. All 144,000 returned
originals across 36 cells match keys, bytes, sequence and requested source.
Preparation and append lie outside read timings. Layout order is fixed within
each arm; this is a short consumed diagnostic without confidence bounds.

## Results

Mean call time ranges across trials, milliseconds:

| Query | Layout | Indexed batch | Cached envelope | Uncached envelope |
| --- | --- | --- | --- | --- |
| 50 | Grouped | 0.338-0.351 | 0.413-0.422 | 2.410-2.436 |
| 50 | Scattered | 0.352-0.381 | 3.052-3.068 | 3.138-3.196 |
| 200 | Grouped | 1.470-1.521 | 1.555-1.573 | 9.709-9.995 |
| 200 | Scattered | 1.426-1.445 | 4.141-4.198 | 12.669-13.242 |

Cached/indexed ratios are 1.177-1.249 for grouped50, 8.021-8.702 for scattered50,
1.034-1.058 for grouped200, and 2.866-2.934 for scattered200. Reuse reduces
grouped envelope cost substantially compared with its own cache-disabled arm,
but that is not a win against the actual control.

Counters verify one envelope per grouped call versus 32 per scattered call.
Peak cached bytes are 228,200 and 7,300,186 respectively, both within the cap.
The negative control uses one slice read per requested event and caches none.
These observations support repeated-access and read-amplification costs in this
prototype; they do not isolate driver copies, SQL overhead, JSON parsing, page
cache behavior and allocation as separate causal contributions.

## Consequence

The previous reference experiment's roughly13% write improvement at200 events
cannot justify ignoring this read penalty. There is no matched whole-service
benefit and no reason to silently restrict evaluation to grouped access. The
existing indexed batch path remains the control and serving behavior is intact.

Avoid more envelope-size/cache-cap sweeps on these consumed results. A distinct
future storage design would need an explicit random-access cost model and all
identity, retry, feedback and crash-replay obligations before another benchmark.
Next research should revisit a different admission/scheduling mechanism or an
independent open goal, retaining this negative result. Original recovery-protocol
changes still require the pending user decision. No whitepaper or remote edits.
