# Source-owner lifecycle cost v48

Correctness PASS; substantial API-level overhead identified. This is an isolated
cold fixture, not loaded serving or learning validation. Protocol was frozen
before execution and all samples/trials retained.

Command: `EVENTFRAME_SOURCE_OWNER_COST_ARTIFACT=.../mmm-source-owner-cost-v48.jsonl go test ./internal/researchmemory -run '^TestSourceOwnerCostStudy$' -count=1 -v`.
Completed in 7.04s on Go 1.27.1 darwin/arm64, Apple M4, 10 logical CPUs/GOMAXPROCS.
Twelve cells, 384 cycles, 48,000 originals and 48,000 typed terminals passed
original/retry/readback/restart checks. No labels, fitting or production data.
All source hashes and phase sums were independently recomputed from JSONL.

| Size | Trial | Control mean cycle ms | Source mean cycle ms | Source cycle p95 ms | Control/source restart ms |
|---|---|---|---|---|---|
| 50 | 0 | 2.724 | 11.051 | 12.026 | 13.592 / 24.808 |
| 50 | 1 | 2.403 | 11.013 | 11.986 | 13.440 / 24.400 |
| 50 | 2 | 2.455 | 11.078 | 12.355 | 13.466 / 24.459 |
| 200 | 0 | 8.548 | 43.721 | 44.390 | 50.983 / 97.136 |
| 200 | 1 | 8.616 | 44.270 | 48.456 | 50.795 / 97.897 |
| 200 | 2 | 8.599 | 43.440 | 44.460 | 51.191 / 97.496 |

At size 200, source admission means were 7.75-7.79ms versus 2.37-2.39ms;
exact retries 8.78-8.82ms versus 2.43-2.51ms; readback 6.41-6.46ms versus
1.51-1.53ms; discard 20.45-21.27ms versus 2.21-2.23ms. The source API performs
per-source indexed resolution and single-item cleanup, while the control has
batch readback/cleanup. These figures cannot all be attributed to index writes.

After 6,400 originals plus terminals, closed source database size was 7,979,008
bytes versus 7,307,264 bytes for control. Startup includes full replay and source
index validation, not a lifetime-constant or online lookup latency guarantee.

Every measured cycle p95 is below 250ms, but that is only an isolated diagnostic:
it excludes queueing, service guards, retrieval, concurrency and fitting. It does
not re-establish v43 loaded completion with the new owner, or the sub-100ms serving
goal. Default paths remain unchanged and all research directions remain open.

Next: batch source-key discard under the existing atomic terminal-write contract.
It must resolve all sources and preflight all terminals before any write, preserve
original identity, reject late conflicts without partial discard, and stop on
uncertain commits. Compare against the retained per-item source control before
considering further indexed-read optimization or loaded integration.

Artifact: `mmm-source-owner-cost-v48.jsonl`.
SHA-256: `c0eb2d30627e869d844eccd2554e2b78a12e3cda7dc2cfc05ae2afdab76a1d65`.
