# V50 Brier primitive: standalone cost

2026-10-04. All required V49 study/audit commands were terminal before timing.
`research/brier-v50-cost/` freezes seven source files/copies and records three
code-zero commands: rerun six-root race suite, vet, and repeated benchmarks.
A separate auditor rechecks both primitive source/log sets, actual test roots,
the preceding normal-study completion hash, and independently parses benchmark
columns against the stored nine samples.

Go benchmark environment: darwin/arm64, Apple M4. Three repeats per benchmark,
one-second requested benchtime. The reported ranges are repeat averages in
ns/op, NOT per-operation quantiles or statistical confidence bounds.

| Operation | Observed ns/op range | B/op | allocs/op |
| --- | ---: | ---: | ---: |
| Three-expert strong substitution | 86.11-86.30 | 0 | 0 |
| Eight-expert strong substitution | 195.3-195.9 | 0 | 0 |
| Three-expert immediate issue/resolve cycle | 133.9-136.8 | 0 | 0 |

The immediate cycle includes owned original advice, issued ticket, strong
substitution, resolution receipt, log-weight normalization, and a prior reset
every 4,096 admitted observations. Initial model construction is excluded from
that benchmark. Advice is fixed, labels deterministic; these are computational
measurements, not fresh learning-quality data. Pure substitution excludes child
expert prediction, and none of the operations includes retrieval, durable
persistence, queue waiting, delayed replay, or background publication.

The primitive is an inexpensive prospective component. No advantage over a
matched alternative, loaded daemon latency/freshness guarantee, sample-efficiency
win, posterior truth authority, or agent outcome benefit is established.
Correctness scope and limitations are in [the preflight result](mmm-brier-v50-preflight-results.md).
All seven whole goals remain open; next is a bounded, independently audited
integration with untouched quality cohorts and unchanged original gates.
