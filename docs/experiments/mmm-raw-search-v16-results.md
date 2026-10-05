# Raw as-of search v16: backend search alone stays below threshold

Date: 2026-10-01. [Frozen protocol](mmm-raw-search-v16-contract.md).
The isolated `Store.Search` sibling completed all probes and writes with
exact as-of identities. Unlike full `Service.Recall` in v14/v15, it did not
queue under the same nominal 8 ms offer pacing.

| Live records | Arm | Offer p99 | Call p99 | Queue p99 |
| ---: | --- | ---: | ---: | ---: |
| 50 | quiet | 2.15 ms | 2.14 ms | 0.008 ms |
| 50 | writer | 10.19 ms | 10.19 ms | 0.020 ms |
| 200 | quiet | 5.02 ms | 5.01 ms | 0.008 ms |
| 200 | writer | 13.25 ms | 13.25 ms | 0.031 ms |

Three paired trials per cell completed 576 raw searches; writer cells
completed 768 future-only writes and moved runtime version 201 to 457 at
200 records. Every returned set contained exactly the expected 50 or 200
live IDs, with no future event. Raw search therefore is **not sufficient**
to reproduce the hundreds-of-milliseconds full-service overload on this
fixture. It does not prove that journal persistence is the sole culprit:
the raw path omits other service reads, candidate processing, packing, and
the journal write, and those omitted steps also alter scheduling.

Next trace full `Service.Recall` at the same load with named store-operation
durations and a matched journal-only control before modifying storage or
the background learner. Do not replace the full-service performance gate
with the passing raw-search number.

Reproduce:

```sh
EVENTFRAME_RUN_RAW_SEARCH_V16=1 go test ./internal/service -run '^TestResearchRawSearchFutureWriterV16$' -count=1 -v
```

[Raw log](mmm-raw-search-v16-raw.log) SHA-256:
`29e27a83906f68978934aa179a87c2d9e8e645f761c78dd7ee74024256ecf774`;
frozen protocol `30f1919ddd8d6cd9f01f3ed407d3d1a703e3da7631a508187cdb4e1ab0bd929e`;
test source at run `ed3e8e1c0c85c9f9fb537db318189898bf40c84d48aca3d49aa5b2b0f32ec715`.
Research-only, local hash-embedder fixture, no agent outcomes. Goal 6 open.
