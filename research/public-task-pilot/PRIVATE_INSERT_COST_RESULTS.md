# Complete private insertion preparation cost

Diagnostic under [the frozen protocol](PRIVATE_INSERT_COST_PROTOCOL.md), not
a passed sustained-load gate. Raw Go JSON stream: `private-insert-cost.jsonl`.
Go 1.27.1 darwin/arm64, CPU=4, three one-second repetitions per fixed case.
No concurrent benchmark or production changes.

| Graph / insertion | Time range (ms/op) | Bytes/op | Allocations/op | Query evaluations | Pair calls |
|---|---:|---:|---:|---:|---:|
| 800 / new-0 | 4.796-4.806 | 550744 | 5077 | 899 | 5623 |
| 800 / new-7 | 2.717-2.724 | 381328 | 3570 | 903 | 2585 |
| 6400 / new-0 | 9.275-9.481 | 865088 | 6778 | 6628 | 5280 |
| 6400 / new-7 | 6.862-6.870 | 606888 | 4572 | 4756 | 4135 |

Timing covers search, selection, connection edits, intermediate/final immutable
roots and summary preparation. It excludes parsing/setup, vector generation,
durable writes, publication, old-reader retention, concurrency and queueing.
Iterations reuse the same original snapshot. Total benchmark process: 24.282s.

At the original offered load of 100 writes/sec, a serial writer has an average
10ms arrival interval. The slower captured preparation already consumes about
93-95% of that interval, before persistence. This is an illustrative capacity
comparison, not a queueing model, tail estimate, measured maximum throughput or
claim that every insertion has this cost. Allocated bytes are not retained RAM.

## Next action

Proceed to durable insertion publication and bounded reader integration with
separate preparation, transaction and queue timing. The result does not justify
claiming a serving rescue based on component speed alone. A bounded metric-norm
cache is a possible subsequent optimization: current cosine recomputes both
norms for every pair, even though vectors are immutable. Its benefit is untested;
prior pair-cache failures warn against assuming lookup overhead is free.

## Artifact identity

- private_insert.go: `b69e8bc487151048495e4fb3524c3ebb9cd4c863991971cc2a4820a956a75958`
- private_connect.go: `d32e54643edab6b343be831c7b6d2f512e8f8cefbb1637f7e8dd1b8deb169a32`
- private_link.go: `1ef62530a0573657cf53ce20fba2d9dcf11b6bfdd6a248c73b5259db4d62f818`
- private_selection.go: `d94306b3fcd3c50dd793751d36cc83c4f47fba5a82c904cdccaa04bac9db8545`
- layered_search.go: `daa5c93ec713c9adb68ad509102fc13111f60c9b6d2c302a53d0fb5669872569`
- serial-work-control.json: `0e96cf71654b990455382185b7beb1e835b54239381df0b9aab807a8d061bc87`
- private-insert-cost.jsonl: `c41cf9fda1997e1040eb55be2b9afa08a772717ffa9914823ce64a7586f4f4f4`

Command from daemon root:

```sh
RESEARCH_INSERT_CAPTURE=../../research/public-task-pilot/serial-work-control.json go test -json ./internal/researchindex -run '^$' -bench '^BenchmarkPrivateInsert$' -benchtime=1s -count=3 -cpu=4
```
