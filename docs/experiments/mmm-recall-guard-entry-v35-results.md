# Guard entry versus source validation v35: results

Date: 2026-10-01. Research-only Goal 6 diagnostic against the
[frozen protocol](mmm-recall-guard-entry-v35-protocol.md).
A private context-tagged timestamp in the test fixture split
the v34 pre-callback interval without changing production guard
or learner behavior. Fresh enabled-only 6 ms and 4 ms fixtures
each completed 192 full 200-event Recalls, 64 durable labels,
256 future-only concurrent writes, zero drops and all exact
nomination/no-future/as-of, mutation rejection, journal reopen,
replay and per-label phase-conservation checks.

| Offer gap | Offer p99 | Frontier age p99 | Guard entry + compatibility p50 / p99 | Candidate validation p50 / p99 |
| --- | ---: | ---: | ---: | ---: |
| 6 ms | 37.701 ms | 46.504 ms | **13.765 / 19.984 ms** | 1.227 / 2.079 ms |
| 4 ms | 45.912 ms | **418.718 ms** | **14.514 / 27.006 ms** | 0.872 / 1.874 ms |

For the five oldest 4 ms labels, guard entry was 11.97–13.64 ms
and candidate validation 0.83–0.99 ms. The two intervals summed
exactly to each label's pre-callback duration. Source reads and
feature validation are therefore **not** the dominant serial cost
in this fixture. Guard entry includes waiting on the shared writer
gate and its snapshot/lineage compatibility check. The timestamp
does not separate those two operations; concurrent future writes
are a plausible cause, not yet established. A writer-on/off
control is the next falsifier before any lock design change.

Command:

```sh
EVENTFRAME_RUN_RECALL_GUARD_ENTRY_V35=1 go test ./internal/service -run '^TestResearchRecallGuardEntryV35$' -count=1 -v -timeout 5m
```

At-run SHA-256:

```text
6edab255832839d04ea9d527a2cf32805adff0b24ac5f5aa143d7448b08af9ea  docs/experiments/mmm-recall-guard-entry-v35-protocol.md
f0ffa0ed43e00ccf7a4343083cd8156a2398676eacfcfb1fee0edc8de12e12fb  internal/service/research_recall_guard_entry_v35_test.go
563ceaeccbc20ecc450858698f725c4e82657500513e62a9579d2bd1e0b7ec2a  internal/service/research_recall_decoupled_drain_v29_test.go
2584f8b508c88ce978f4e978f0093249a7d55f487c64dea792677033c25f9fd9  internal/service/research_recall_live_learning_v26_test.go
```

No production code changed. This is not a higher-rate pass or
whole Goal 6 completion.
