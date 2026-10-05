# Live durable feedback freshness v1

**Correctness passed; the freshness endpoint was too narrow.** The [frozen
v1 contract](mmm-durable-live-freshness-v1-contract.md) ran three persistent
store trials per arm, 64 explicit labels per trial, with a notification
observer and 768 future-dated writes in the writer arm. 762 writer calls
overlapped an active learning operation. Before Close, the observer saw all
192 labels per arm reach the worker's published-snapshot count with zero
failures. The independent ledger and replay checks also passed. A focused
race run reported no race or state failure.

| Mode | Arm | Recall p99 | Guarded feedback p99 | Reported age p99 |
| --- | --- | ---: | ---: | ---: |
| Ordinary | Quiet | 7.34 ms | 0.97 ms | 4.01 ms |
| Ordinary | Future writer | 52.66 ms | 12.15 ms | 1.35 ms |
| `-race` | Quiet | 9.39 ms | 2.26 ms | 8.90 ms |
| `-race` | Future writer | 398.83 ms | 122.49 ms | 6.44 ms |

The low age is **not** evidence that the entire feedback update took only
microseconds. The v1 observer timestamp was taken **after** the guarded
`Durable.Feedback` call returned. The worker can fit and publish while that
call is still returning, so the recorded age measures only the remaining
post-return wait. The v1 contract called it submission-to-publication age;
that interpretation is invalid. The ordinary p99 <100ms Recall screen passed,
and the reported post-return age p99 <250ms passed, but the latter does not
test the intended end-to-end freshness claim. This measurement mistake is
preserved rather than silently redefining v1.

The service helper, test, and read-only Durable observation API source hashes
were respectively `4af3301bc14ca3ef88b9108523661ec923e929eeaba9d32298fc0c86e24faf1b`,
`436430b6decff313039c03b9218b565a2d1df918ce0f1d722b7a8c2328c1300f`,
and `36d1150bb497b929014dd0171c09fba777cf291a2c9760d3f6efa0ba4f71d035`.
No production serving path changed.
