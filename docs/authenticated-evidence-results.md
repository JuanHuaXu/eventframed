# Authenticated Evidence Results

Date: 2026-09-05. Apple M4, darwin/arm64, Go 1.27.0, default GOMAXPROCS 10.
All fixtures were isolated temporary databases. No production or personal data
was read, re-signed, or uploaded. No keys were installed in production.

## Functional Results

- Signed-content tampering, wrong tenant/source, cached expiry, revocation,
  and derived-observation admission: rejected by focused tests.
- Replay: exactly one update for eight concurrent copies, in both memory and
  on-disk backends; a re-signed new-journal replay cannot update again.
- Restart: duplicate status and working belief persist across close/reopen.
- Registry change: old-trust belief is not served; new-trust statistics start
  fresh. The scored bundle uses the declared working predictive map.
- Reversibility: after 10,000 positive updates saturate the odds bound, five
  negative updates reverse the preference without requiring a detector reset.
- Split-reset: both backends count the revealing outcome once in the triggering
  child's new filter; siblings retain local evidence without pooled confidence.
- `go test ./...`, focused race checks, and `go vet ./...`: passed.

This validates those implementation contracts in the fixtures. It does not
validate real-world truth authentication, calibrated probabilities, improved
answer accuracy, or resistance to an authorized observer fabricating new IDs.

## Performance

The [raw output](authenticated-evidence-benchmark.txt) contains three runs of
500 operations per configuration. Control is the same binary with the new
authentication/filter options disabled; upgrade enables both. The fixture has
50 events, 32-dimensional hash embeddings, recall/pack budgets 50/10, and a
temporary on-disk LibraVDB. These are service calls, not HTTP/OpenClaw round trips.

| Complete internal path | Control median run mean | Upgrade median run mean | Upgrade p99 across three runs |
| --- | ---: | ---: | ---: |
| Serial recall | 7.153 ms | 7.473 ms | 9.02-9.28 ms |
| Four-worker recall | 2.481 ms/op | 2.504 ms/op | 15.34-19.17 ms |
| Serial durable outcome | 5.603 ms | 5.649 ms | 6.98-7.07 ms |
| Four-worker mixed, 75% recall / 25% outcome | 3.603 ms/op | 3.629 ms/op | 29.05-32.90 ms |

Concurrent `ms/op` is inverse throughput, not individual request latency.
The upgraded mixed-request p50 was 11.01-11.50 ms and maximum 74.96 ms.
Internal database/lock waits are included. Observer-side signing and fixture
preparation are excluded; signature verification is included in outcome timing.
The mixed path also performs ordinary journal persistence and optimistic retries.

Median run means increased about 4.48% for serial recall, 0.92% for concurrent
recall, 0.83% for durable outcome writes, and 0.70% for the concurrent mixed case.
Do not interpret these as causal or statistically established overhead estimates:
three short, sequential configuration runs have storage/scheduling variance.
In the earlier exploratory campaign, serial recall's median run mean decreased
slightly, and upgraded mixed p99 ranged 31.99-52.54 ms (maximum 74.91 ms).
All upgraded complete-request maxima reported in both campaigns stayed below 100ms, but this
does not establish a hard deadline or a production tail-latency bound.

Primitive medians in the saved run:

- Uncached signature verification: 38.12 microseconds, 2466 bytes, 55 allocations.
- Cached signature check plus admission: 2.13 microseconds, 2382 bytes, 53 allocations.
- Working filter update: 0.96 microseconds, 371 bytes, 6 allocations.
- Memory-only outcome call: 3.18 microseconds control, 32.75 microseconds upgrade.

The last comparison exposes the authentication cost hidden by millisecond-scale
durable storage. The upgrade is not free; it simply remains small compared with
the tested complete internal request budget. Hash embeddings and a small corpus
cannot establish performance with external embeddings, remote retrieval, or a
large production corpus. The timing fixtures also do not claim every candidate
passes activation; only the certified evidence-ready subset is belief-conditioned.

## Reproduce

```sh
go test ./...
go vet ./...
go test -race ./internal/epistemic ./internal/bayes ./internal/config ./internal/service ./internal/store/memorystore ./internal/store/libravdbstore ./internal/api
go test ./internal/epistemic ./internal/bayes ./internal/service -run '^$' -bench 'Benchmark(EvidenceVerify|WorkingBelief|AuthenticatedOutcome|EvidenceInternalRequests)$' -benchmem -benchtime=500x -count=3
```

Remaining rollout work is explicit observer key enrollment and held-out
prospective evaluation of the opt-in filter. Old chat logs cannot establish
signature provenance retrospectively. This change was not deployed or pushed.
