# Public Pooling And Native Contract Results

2026-10-04. Goal5 infrastructure progress, not quality confirmation. All seven
whole research goals remain OPEN. The previous research goal turn made progress;
the intervening social exchange did not advance research. This turn adds a
verified native launch repair and a prospective fit-only retrieval diagnostic.

## Complete Document Pool

The isolated source-preserving adapter supplies one native index record per
original source document, linking every title/abstract span outside scored text.
ALL5,183documents /10,869spans are retained. Actual pooled text is8,582,765bytes,
metadata2,826,303bytes; largest input10,579bytes below frozen16KiB cap. This
public-document instantiation preserves scientific text rather than claiming
general compression, episodic extraction or independent evidence from spans.

Full test-dependent source closure171files,6terminal code-zero commands.
Ten actual race roots execute3times each with no skips; vet and separate full
Node reconstruction PASS. Fourteen corrupted pools and twelve corrupted native
pilot transcripts are rejected. Source/copy/raw-log/command/cost readback PASS
in `research/public-task-pilot/scifact-pool-v1/artifact-audit.json`.

At768FP32 dimensions, pooled document key/vector payload is24,515,307bytes;
with conservative query payload16,169,598bytes, total40,684,905bytes /6,014entries
fits the existing12k/64MiB research cache. This is NOT RSS or a native internal
cache bound. Do not also populate a span cache and silently count only documents.

Three measurements, three operations each, AppleM4/darwin/arm64/Go1.27.1:

| Operation | Time Per Operation | Allocated Bytes Per Operation |
| --- | --- | --- |
| Build all5,183 pooled records | 86.79-89.04ms | 162.81-162.83MB |
| Bind full nomination/ranking set50 | .222-.228ms | 151,656 |
| Bind full nomination/ranking set200 | .881-.975ms | 630,957 |

Offline construction and isolated binding costs are NOT full serving p99, loaded
latency, quality gains, CPU utilization or retained-memory measurements.

## Preserved Native Failures

Owned research daemon only: fresh HOME/cwd/config/store, explicit private Unix
socket, minimal child environment, read-only hashed installed public assets.
No launchctl, production endpoint/config/model writes or private chats.

| Variant | Actual Outcome | Sampled Peak Daemon RSS |
| --- | --- | --- |
| v1 | Refused mutually exclusive db_path/agent_db_root | Not sampled |
| v2 ONNX | Bare cognitive_scanner.bin missing | 1,561,664KiB |
| v3 ONNX | Owned monitor stopped above fixed2GiB ceiling; client failed | 2,114,032KiB |
| v4 Q8GGUF | Socket ready, but InsertText failed bare libllama lookup | 234,144KiB |
| v5 Q8GGUF | Same failure; nice wrapper removed supplied library path | 233,408KiB |
| v6 Q8GGUF | Direct owned spawn plus process-local priority10; pilot PASS | 763,184KiB |

v4 was a separately declared backend experiment, NOT equivalent ONNX performance.
All earlier failed roots remain untouched. A real direct-vs-nice environment
probe proves the wrapper drops DYLD_LIBRARY_PATH. v6 changes only owned launch
and equivalent priority scheduling, with complete inverse equality to v5.
No macOS protection, installed library path or global configuration changed.
Every managed job for these six pilots is terminal. RSS sampled every500ms may
miss peaks; it is not total machine memory, an enforced hard limit or GPU memory.
The daemon overrides requested Go memory budget using physical RAM, so the
monitor, not GOMEMLIMIT, supplies the observed stop rule.

## Technical Pilot

The exact frozen client imported first8corpus documents plus one future-only
copy in737.87ms. Both actual source-title queries recovered their source and
passed SearchTextCollections -> RankCandidates with complete returned sets.
The future copy was retrievable without exclusion and absent with exclusion:
the temporal control is nonvacuous. Search8.25/12.75ms, rank.72/1.69ms, n=2;
no p99 or general runtime compliance claim. Full request/response metadata,
identities, source digests and span links independently checked.

Important limitation: daemon logs show raw8hits -> post-L7 returned1hit for each
query. K=50 is a maximum request, not proof of a50candidate frontier. Ranking
preserved the ACTUAL one-row frontier, which does not validate reranking a
50-200document set. No guessed bypass, changed ingestion threshold or gold
padding is introduced. A native stale dirty-anchor warning is preserved; this
pilot does not diagnose or repair the native micro-temporal implementation.
The GGUF log reports512-token context. Complete framed input coverage does not
prove every model-side suffix was consumed.

Next: the separately frozen full-corpus/351fit-only native diagnostic in
`scifact-native-fit-v1-protocol.md`, measuring nomination, ranking and total
cost before selecting any rescue. Calibration180/confirmation300 remain unused
by prediction. This pilot is NOT agent answer accuracy, probabilistic calibration,
Anti-Pigeon certification or completed Goal5. Production and whitepaper untouched;
no commit, push or deployment.
