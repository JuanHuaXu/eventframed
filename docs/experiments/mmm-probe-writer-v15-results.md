# Probe/writer v15: overload without the bound learner

Date: 2026-10-01. [Frozen contract](mmm-probe-writer-v15-contract.md).
The no-learner sibling control completed twice on local persistent LibraVDB.
It reproduces severe writer-arm queueing with either 50 or 200 visible
records. This is **not** a fix or proof of the v14 failure's exact cause.

| Live records | Arm | First offer p99 | Repeat offer p99 | First call p99 | First queue p99 |
| ---: | --- | ---: | ---: | ---: | ---: |
| 50 | quiet | 14.53 ms | 13.91 ms | 14.52 ms | 0.021 ms |
| 50 | writer | 301.31 ms | 288.35 ms | 43.08 ms | 260.47 ms |
| 200 | quiet | 14.63 ms | 14.53 ms | 14.63 ms | 0.022 ms |
| 200 | writer | 787.55 ms | 815.66 ms | 66.54 ms | 734.50 ms |

Each cell contains three trials and 576 completed Recall offers; writer
cells complete 768 future-only writes. Every probe's nominated identities
matched the exact 50 or 200 live IDs, and its packed results excluded future
events. The bound worker was absent: no admission, feedback, or learner-state
publication occurred. Therefore a bound learner is not **necessary** for
overload under this schedule. However, these are not read-only controls:
`Service.Recall` still persists a Bayesian journal. Removing the learner also
changes CPU load and timing, so the higher v15 p99 than v14's 50-cell cannot
be interpreted as the learner improving throughput. The active marker
identifies a shared workload window, not a per-operation lock-overlap trace.

The next discriminator is the same 50/200 records, paced probes, and future
writer using raw as-of `Store.Search` without a service journal. If raw
search queues similarly, investigate the backend read/write path; if it does
not, trace the service's journal/write phases. Both possibilities remain open.

Reproduce:

```sh
EVENTFRAME_RUN_MOTION_PROBE_V15=1 go test ./internal/service -run '^TestResearchMotionProbeWriterOnlyV15$' -count=1 -v
```

Artifacts: [first log](mmm-probe-writer-v15-raw.log),
[repeat log](mmm-probe-writer-v15-repeat.log). SHA-256 of the frozen contract
`909a0ec49f48469461f4d66655a713a7602fb49333076ebe6c564915dfb00809`;
first log `826ab8f3ad48277f50e00c5ba2e025aa6da968ad4716c43aec6622866e93336f`;
repeat log `8b1e872978f30e5998c6fa10a64f6c2720b5b5c3b2623d264c90d37a4cf2a941`.
This is synthetic, local, and small. Goal 6 remains open.
