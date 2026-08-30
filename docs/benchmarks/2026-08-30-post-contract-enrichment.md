# Post-contract enrichment benchmark, 2026-08-30

## Environment

- Base commit: `88461ef` plus the uncommitted contract-12 EventFrame-corpus implementation and benchmark harness
- Go: `go1.27.0 darwin/arm64`
- OS: macOS 26.6.2 (25G83)
- CPU: Apple M4
- Physical memory: 16 GiB
- Inputs: deterministic synthetic conversations only
- Embedding: local 384-dimensional feature hash
- Durable traversal: embedded LibraVDB SQ8

## Existing benchmark suite

The complete sequential warm suite ran for five repetitions per case:

```sh
make benchmark
```

| Case | Mean across five runs | Observed p99 range |
| --- | ---: | ---: |
| Memory recall, 100 events | 0.654 ms | 1.803-1.931 ms |
| Memory recall, 1,000 events | 1.655 ms | 3.061-3.526 ms |
| Memory recall, 10,000 events | 7.314 ms | 8.076-13.964 ms |
| LibraVDB SQ8 recall, 100 events | 7.386 ms | 8.940-9.198 ms |
| LibraVDB SQ8 recall, 1,000 events | 8.362 ms | 9.295-10.038 ms |
| LibraVDB SQ8, 1,000 events, 50% unavailable | 11.981 ms | 13.052-13.175 ms |
| LibraVDB SQ8, 1,000 events, 95% unavailable | 14.650 ms | 15.084-17.907 ms |
| Agency claim, 1 active record | 4.969 ms | 6.117-6.830 ms |
| Agency claim, 100 active records | 5.794 ms | 7.063-9.365 ms |

All recall cases remained below the 100 ms target. Compared with the August 28
baseline, the ordinary 1,000-event SQ8 mean moved from 7.643 ms to approximately
8.4 ms. The repository has accumulated several scoring, audit, and contract
changes since that baseline, so this cross-version difference cannot be
attributed specifically to post-contract enrichment.

### Contract-12 EventFrame-corpus rerun

After raw text was removed from embeddings, nomination/reranking text, and
diversity calculations, the focused suite was rerun with the same five-run
method. These figures supersede the corresponding recall and capture figures
above for the corrected semantic path.

| Case | Mean across five runs | Observed p99 range |
| --- | ---: | ---: |
| Memory recall, 100 events | 0.726 ms | 1.766-1.896 ms |
| Memory recall, 1,000 events | 1.760 ms | 3.161-3.336 ms |
| Memory recall, 10,000 events | 6.893 ms | 7.981-8.513 ms |
| LibraVDB SQ8 recall, 100 events | 7.954 ms | 9.109-11.035 ms |
| LibraVDB SQ8 recall, 1,000 events | 8.702 ms | 9.988-11.014 ms |

The corrected 10,000-event in-memory mean improved by about 5.8%; the corrected
1,000-event SQ8 mean increased by about 4.1%. These are local synthetic results,
not evidence about a remote embedding service or TLS contract endpoint.

## Upgrade-specific microbenchmarks

```sh
go test ./benchmark -run '^$' \
  -bench '^Benchmark(PostContractFrameExtraction|CaptureTurnMemory)$' \
  -benchmem -benchtime=2s -count=5
go test ./internal/api -run '^$' \
  -bench '^BenchmarkProjectOpenClawPacket50$' \
  -benchmem -benchtime=2s -count=5
```

| Boundary | Mean | Allocation |
| --- | ---: | ---: |
| Deterministic 5W1H enrichment | 23.96 us | 4,799 B/op, 70 allocs/op |
| Enrichment + canonical hash embedding + memory persistence | 29.53 us | 13,258 B/op, 96 allocs/op |
| OpenClaw-safe projection, 50 candidates | 4.78 us | 32,768 B/op, 1 alloc/op |

The parser itself is not a material contributor to millisecond-scale production
latency. The projection performs one slice allocation proportional to the packed
candidate count.

## Unix-socket end-to-end probe

This transport probe predates the contract-12 corpus correction. It remains
evidence about the HTTP-over-Unix and durable-write boundary, but the focused
contract-12 rerun above is the applicable semantic-path benchmark.

A built `eventframed` binary and built TypeScript client communicated over a
mode-local Unix socket. This boundary includes JSON serialization, HTTP-over-Unix,
daemon processing, local embedding, and durable LibraVDB commits. It excludes an
external LibraVDB contract server and remote embedding provider.

| Probe | n | Mean | p50 | p95 | p99 | Max |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Fresh raw-turn captures | 2,000 | 47.879 ms | 43.918 ms | 104.252 ms | 110.656 ms | 134.511 ms |
| Fresh structured-event control, separate fresh database | 2,000 | 44.686 ms | 41.128 ms | 91.703 ms | 96.541 ms | 123.343 ms |
| Warm paired raw captures | 500 | 21.986 ms | 21.005 ms | 40.802 ms | 42.031 ms | 44.175 ms |
| Warm paired structured controls | 500 | 22.039 ms | 20.991 ms | 40.047 ms | 42.301 ms | 45.109 ms |
| OpenClaw-projected recall after 2,000 captures | 1,000 | 9.764 ms | 9.952 ms | 10.950 ms | 11.147 ms | 19.135 ms |

The separate fresh runs show substantial write-tail variance and are not a
paired causal comparison. In the alternating paired run against one daemon and
database, raw capture and pre-structured observation were effectively identical;
the raw path was 0.24% faster by mean and had a 0.27 ms lower p99. This supports
the narrower conclusion that moving 5W1H enrichment behind the contract did not
measurably degrade the warm capture path.

Fresh durable capture can exceed 100 ms at the tail. Capture occurs after a
successful agent turn and is not on prompt recall's latency-critical path, but
batching or group commit remains a production optimization candidate if capture
completion latency becomes operationally important.

## Limits

1. The Unix-socket probe was sequential and synthetic, not a concurrent soak.
2. No remote TLS LibraVDB contract or external embedding provider was enabled.
3. The fresh raw and structured runs used separate databases; only the 500-pair
   alternating run controls shared runtime and storage conditions.
4. The probe measures latency and allocation, not extraction accuracy.
